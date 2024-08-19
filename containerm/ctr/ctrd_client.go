// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

package ctr

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	// ContainerdClientServiceLocalID sets service local ID for containerd client.
	ContainerdClientServiceLocalID = "container-management.service.local.v1.service-containerd-client"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       ContainerdClientServiceLocalID,
		Type:     registry.ContainerClientService,
		InitFunc: registryInit,
	})
}

type containerdClient struct {
	sync.Mutex
	rootExec           string
	metaPath           string
	registriesResolver containerImageRegistriesResolver
	ctrdCache          *containerInfoCache
	ioMgr              containerIOManager
	logsMgr            containerLogsManager
	decMgr             containerDecryptMgr
	verifier           containerVerifier
	spi                containerdSpi
	eventsCancel       context.CancelFunc
	runcRuntime        types.Runtime
	imageExpiry        time.Duration
	imageExpiryDisable bool
	imagesExpiryLock   sync.Mutex
	imagesWatcher      resourcesWatcher
}

func (ctrdClient *containerdClient) RemoveContainerImage(ctx context.Context, image types.Image) error {
	log.Info("successfully removed container image")
	Image, err := ctrdClient.spi.GetImage(ctx, image.Name)
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Warn("image = %s has already been removed - no usage check or clean up will be performed", image.Name)
		}
		return err
	}
	rmErr := ctrdClient.removeUnusedImage(ctx, Image)
	if rmErr == errImageIsInUse {
		log.Debug("image = %s is in use - will not remove it", image.Name)
		return nil
	}
	if rmErr == nil {
		log.Info("successfully removed container image %s", image.Name)
	}
	return rmErr
}

// -------------------------------------- ContainerdAPIClient implementation with Containerd -------------------------------------
// CreateContainer creates all resources needed in the underlying container management so that a container can be successfully started
func (ctrdClient *containerdClient) CreateContainer(ctx context.Context, container *types.Container, checkpointDir string) error {
	log.Debug("creating container resources in container client")
	var (
		err   error
		image containerd.Image
	)
	defer func() {
		if err != nil {

			if ioErr := ctrdClient.ioMgr.ClearIO(container.ID); ioErr != nil {
				log.ErrorErr(ioErr, "error clearing IO for container id = %s", container.ID)
			}
			ctrdClient.clearSnapshot(ctx, container.ID)
		}
	}()
	log.Debug("creating container IOs ")
	if _, err = ctrdClient.ioMgr.InitIO(container.ID, container.IOConfig.OpenStdin); err != nil {
		log.ErrorErr(err, "failed to initialise IO for container ID = %s", container.ID)
		return err
	}
	log.Debug("successfully created container IOs")

	if err = ctrdClient.initLogDriver(container); err != nil {
		log.ErrorErr(err, "failed to initialize logger for container ID = %s", container.ID)
		return err
	}
	log.Debug("successfully initialized container logger")

	image, err = ctrdClient.pullImage(ctx, container.Image)
	if err != nil {
		log.ErrorErr(err, "error while trying to get container image with ID = %s for container ID = %s ", container.Image.Name, container.ID)
		return err
	}

	return ctrdClient.createSnapshot(ctx, container.ID, image, container.Image)
}

// DestroyContainer kill container and delete it.
func (ctrdClient *containerdClient) DestroyContainer(ctx context.Context, container *types.Container, stopOpts *types.StopOpts, clearResources bool) (int64, time.Time, error) {
	var (
		ctrInfo  *containerInfo
		code     int64
		killErr  error
		exitTime time.Time
	)

	defer func() {
		if clearResources && killErr == nil {
			log.Debug("closing IOs for container id = %s while destroying ", container.ID)
			if err := ctrdClient.ioMgr.ClearIO(container.ID); err != nil {
				log.WarnErr(err, "failed to clear container IO resources during container id = %s destroy", container.ID)
			}
			log.Debug("cleared IOs while destroying container id = %s", container.ID)

			ctrdClient.clearSnapshot(ctx, container.ID)

			if cleanupErr := ctrdClient.handleImageExpiryOnRemove(ctx, container.Image.Name); cleanupErr != nil {
				log.WarnErr(cleanupErr, "could not clean up resources for image %s", container.Image.Name)
			}
		}
	}()

	ctrInfo = ctrdClient.ctrdCache.get(container.ID)
	if ctrInfo == nil {
		return -1, time.Now(), log.NewErrorf("container with ID = %s does not exist", container.ID)
	}
	// if you call DestroyContainer to stop a container, will skip the hooks.
	// the caller need to execute the all hooks.
	ctrInfo.ctrInfoLock.Lock()
	ctrInfo.skipExitHooks = true
	ctrInfo.ctrInfoLock.Unlock()
	defer func() {
		ctrInfo.ctrInfoLock.Lock()
		ctrInfo.skipExitHooks = false
		ctrInfo.ctrInfoLock.Unlock()

		fmt.Printf("does it come here")
	}()

	if task := ctrInfo.getTask(); task != nil {
		code, exitTime, killErr = ctrdClient.killTask(ctx, ctrInfo, stopOpts)
		if killErr != nil {
			return code, exitTime, killErr
		}
		log.Debug("closing task's IOs now to release the streams for container id = %s", ctrInfo.c.ID)
		err := ctrInfo.task.CloseIO(ctx, containerd.WithStdinCloser)
		if err != nil {
			log.ErrorErr(err, "could not close task IOs for container id = %s", ctrInfo.c.ID)
		}
		log.Debug("deleting task for container id = %s", ctrInfo.c.ID)
		_, err = ctrInfo.task.Delete(ctx)
		if err != nil {
			log.ErrorErr(err, "error while deleting task for container id = %s, container delete was not successful ", ctrInfo.c.ID)
		} else {
			ctrInfo.task = nil
			log.Debug("deleted the task for container id = %s successfully", ctrInfo.c.ID)
		}
	}

	if ctrInfo.container != nil {
		log.Debug("deleting container for container id = %s", ctrInfo.c.ID)
		err := ctrInfo.container.Delete(ctx)
		if err != nil {
			log.ErrorErr(err, "error while deleting containerd container for container id = %s, container delete was not successful ", ctrInfo.c.ID)
		} else {
			ctrInfo.container = nil
			log.Debug("deleted the underlying container for container id = %s successfully", ctrInfo.c.ID)
		}
	}
	ctrInfo = ctrdClient.ctrdCache.remove(container.ID)

	return code, exitTime, killErr

}

// StartContainer starts the underlying container
func (ctrdClient *containerdClient) StartContainer(ctx context.Context, container *types.Container, checkpointDir string) (int64, error) {
	var (
		ctrdContainer containerd.Container
		createOpts    []containerd.NewContainerOpts
		image         containerd.Image
		ctrInfo       *containerInfo
		err           error
	)

	if ctrdContainer, err = ctrdClient.spi.LoadContainer(ctx, container.ID); err != nil && !errdefs.IsNotFound(err) {
		log.ErrorErr(err, "error trying to check for existing container ID = %s", container.ID)
		return -1, err
	}

	if ctrdContainer != nil {
		log.Debug("container with ID = %s already exists ", container.ID)
		return -1, log.NewErrorf("container with ID = %s already exists", container.ID)
	}

	if _, err = ctrdClient.spi.GetSnapshot(ctx, container.ID); err != nil {
		return -1, log.NewErrorf("snapshot for container with ID = %s does not exist", container.ID)
	}

	defer func() {
		if err != nil {
			if ctrdContainer != nil {
				if cleanupErr := ctrdContainer.Delete(ctx); cleanupErr != nil {
					log.ErrorErr(cleanupErr, "could not delete container for container id =%s", container.ID)
				}
			}
		}
	}()
	image, err = ctrdClient.getImage(ctx, container.Image)
	if err != nil {
		log.ErrorErr(err, "missing image ID = %s for container with ID = %s", container.Image.Name, container.ID)
		return -1, err
	}

	ctrdClient.configureRuncRuntime(container)
	createOpts, err = ctrdClient.generateNewContainerOpts(container, image)
	if err != nil {
		log.ErrorErr(err, "failed to generate create opts for image ID = %s for container with ID = %s", container.Image.Name, container.ID)
		return -1, err
	}
	ctrdContainer, err = ctrdClient.spi.CreateContainer(ctx, container.ID, createOpts...)
	if err != nil {
		log.ErrorErr(err, "error creating new container with ID = %s", container.ID)
		return -1, err
	}

	if !ctrdClient.ioMgr.ExistsIO(container.ID) {
		if _, err = ctrdClient.ioMgr.InitIO(container.ID, container.IOConfig.OpenStdin); err != nil {
			log.ErrorErr(err, "failed to initialise IO for container ID = %s", container.ID)
			return -1, err
		}
	} else {
		log.Debug("container IOs already created - will use them")
	}
	log.Debug("successfully created container IOs")

	if err = ctrdClient.initLogDriver(container); err != nil {
		log.ErrorErr(err, "failed to initialize logger for container ID = %s", container.ID)
		return -1, err
	}

	defer func() {
		if err != nil {
			if ctrInfo != nil && ctrInfo.task != nil {
				ctrInfo.task.Delete(ctx)
			}
			ctrdClient.ctrdCache.remove(container.ID)
			ctrdClient.ReleaseContainerResources(ctx, container)
		}
	}()

	ctrInfo, err = ctrdClient.createTask(ctx, container.IOConfig, container.ID, checkpointDir, ctrdContainer)
	if err != nil {
		log.ErrorErr(err, "error creating task for container ID = %s", container.ID)
		return -1, err
	}
	ctrInfo.c = container
	ctrdClient.ctrdCache.add(ctrInfo)

	if err = ctrInfo.getTask().Start(ctx); err != nil {
		return -1, err
	}
	return int64(ctrInfo.getTask().Pid()), nil
}

// AttachContainer attaches the container's IO
func (ctrdClient *containerdClient) AttachContainer(ctx context.Context, container *types.Container, attachConfig *streams.AttachConfig) error {
	ctrIO := ctrdClient.ioMgr.GetIO(container.ID)
	if ctrIO == nil {
		log.Debug("creating IO for container id = %s", container.ID)
		var err error
		defer func() {
			if err != nil {
				if ctrIO != nil {
					if closeErr := ctrdClient.ioMgr.ClearIO(container.ID); closeErr != nil {
						log.ErrorErr(closeErr, "error while clearing container streams for container id = %s", container.ID)
					}
				}
			}
		}()
		log.Debug("creating container IOs ")
		if ctrIO, err = ctrdClient.ioMgr.InitIO(container.ID, container.IOConfig.OpenStdin); err != nil {
			log.ErrorErr(err, "failed to initialise IO for container ID = %s", container.ID)
			return err
		}
		log.Debug("successfully created container IOs")
	}

	attachConfig.Terminal = container.IOConfig.Tty

	// NOTE: the AttachContainerIO might use the hijack's connection as
	// stdin in the AttachConfig. If we close it directly, the stdout/stderr
	// will return the `using closed connection` error. As a result, the
	// Attach will return the error. We need to use pipe here instead of
	// origin one and let the caller closes the stdin by themself.
	if container.IOConfig.OpenStdin && attachConfig.UseStdin {
		oldStdin := attachConfig.Stdin
		pstdinr, pstdinw := io.Pipe()
		go func() {
			defer pstdinw.Close()
			io.Copy(pstdinw, oldStdin)
		}()
		attachConfig.Stdin = pstdinr
		attachConfig.CloseStdin = true
	} else {
		attachConfig.UseStdin = false
	}
	return <-ctrIO.Stream().Attach(ctx, attachConfig)
}

// PauseContainer pause container.
func (ctrdClient *containerdClient) PauseContainer(ctx context.Context, container *types.Container) error {
	var (
		ctrInfo *containerInfo
		err     error
	)
	ctrInfo = ctrdClient.ctrdCache.get(container.ID)
	if ctrInfo == nil {
		return log.NewErrorf("missing container to pause")
	}
	if err = ctrInfo.getTask().Pause(ctx); err != nil {
		return err
	}

	return nil
}

// UnpauseContainer unpauses a container.
func (ctrdClient *containerdClient) UnpauseContainer(ctx context.Context, container *types.Container) error {
	var (
		ctrInfo *containerInfo
		err     error
	)
	ctrInfo = ctrdClient.ctrdCache.get(container.ID)
	if ctrInfo == nil {
		return log.NewErrorf("missing container to unpause")
	}
	if err = ctrInfo.getTask().Resume(ctx); err != nil {
		return err
	}

	return nil
}

// Lists all created containers.
func (ctrdClient *containerdClient) ListContainers(ctx context.Context) ([]*types.Container, error) {
	all := ctrdClient.ctrdCache.getAll()
	cachedResult := make([]*types.Container, len(all))
	for i, ctr := range all {
		cachedResult[i] = ctr.c
	}
	return cachedResult, nil
}

func (ctrdClient *containerdClient) GetContainerInfo(ctx context.Context, id string) (*types.Container, error) {
	ctrInfo := ctrdClient.ctrdCache.get(id)
	if ctrInfo != nil {
		return ctrInfo.c, nil
	}
	return nil, log.NewErrorf("missing container with ID = %s", id)

}

// Restore containerd container
func (ctrdClient *containerdClient) RestoreContainer(ctx context.Context, container *types.Container) error {
	if _, err := ctrdClient.spi.GetSnapshot(ctx, container.ID); err != nil {
		return log.NewErrorf("snapshot for container with ID = %s does not exist", container.ID)
	}

	ctrdContainer, ctrdErr := ctrdClient.spi.LoadContainer(ctx, container.ID)
	if ctrdErr != nil {
		log.ErrorErr(ctrdErr, "failed to retrieve container ID = %s from containerd while restoring", container.ID)
		return ctrdErr
	}

	if !ctrdClient.ioMgr.ExistsIO(container.ID) {
		if _, err := ctrdClient.ioMgr.InitIO(container.ID, container.IOConfig.OpenStdin); err != nil {
			log.ErrorErr(err, "error while initialising IO for container ID = %s", container.ID)
			return err
		}
	} else {
		log.Debug("container IOs already initialized")
	}
	log.Debug("successfully created container IOs")

	if err := ctrdClient.initLogDriver(container); err != nil {
		log.ErrorErr(err, "failed to initialize logger for container ID = %s", container.ID)
		return err
	}

	ctrInfo, taskErr := ctrdClient.loadTask(ctx, container.ID, "", ctrdContainer)
	if taskErr != nil {
		log.ErrorErr(taskErr, "error loading task for container ID = %s while restoring", container.ID)
		if delErr := ctrdContainer.Delete(ctx); delErr != nil {
			log.ErrorErr(delErr, "could not delete loaded container with id = %s", container.ID)
		}
		return taskErr
	}
	ctrInfo.c = container
	ctrdClient.ctrdCache.add(ctrInfo)
	log.Debug("successfully restored container ID = %s", container.ID)
	return nil
}

func (ctrdClient *containerdClient) ReleaseContainerResources(ctx context.Context, container *types.Container) error {
	log.Debug("called container exit hook - will release container resources for container %s ", container.ID)
	// Note - cleanup of tasks , etc. must be called prior to reseting the container IOs

	log.Debug("resetting IOs for container id = %s", container.ID)
	ctrdClient.ioMgr.ResetIO(container.ID)
	return nil
}

func (ctrdClient *containerdClient) SetContainerExitHooks(hooks ...ContainerExitHook) {
	ctrdClient.ctrdCache.setExitHooks(hooks...)
}

func (ctrdClient *containerdClient) UpdateContainer(ctx context.Context, container *types.Container, resources *types.Resources) error {
	if resources == nil {
		return nil
	}
	ctrInfo := ctrdClient.ctrdCache.get(container.ID)
	if ctrInfo == nil {
		return log.NewErrorf("missing container to update with ID = %s", container.ID)
	}
	spec, err := ctrInfo.container.Spec(ctx)
	if err != nil {
		return err
	}

	var lm *specs.LinuxMemory
	if container.HostConfig.Resources == nil {
		lm = toLinuxMemory(resources)
	} else {
		unlimited := func() *int64 { i := int64(-1); return &i }()
		get := func(old, new string) *int64 {
			if new == "" && new != old {
				// if there is an old value and updated to "" then set to unlimited
				return unlimited
			}
			return parseMemoryValue(new)
		}
		lm = &specs.LinuxMemory{
			Limit:       get(container.HostConfig.Resources.Memory, resources.Memory),
			Reservation: get(container.HostConfig.Resources.MemoryReservation, resources.MemoryReservation),
			Swap:        get(container.HostConfig.Resources.MemorySwap, resources.MemorySwap),
		}
	}

	r := &specs.LinuxResources{
		// Currently, runc update could not change device config and skips it. Add it just in case this changes.
		Devices: spec.Linux.Resources.Devices,
		Memory:  lm,
	}
	return ctrInfo.getTask().Update(ctx, containerd.WithResources(r))
}

func (ctrdClient *containerdClient) GetContainerStats(ctx context.Context, container *types.Container) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
	ctrInfo := ctrdClient.ctrdCache.get(container.ID)
	if ctrInfo != nil && ctrInfo.task != nil {
		ctrdMetrics, err := ctrInfo.task.Metrics(ctx)
		if err != nil {
			log.ErrorErr(err, "could not get stats for container ID = %s", container.ID)
			return nil, nil, nil, 0, time.Time{}, err
		}
		return toMetrics(ctrdMetrics, container.ID)
	}

	return nil, nil, nil, 0, time.Time{}, log.NewErrorf("missing container with ID = %s", container.ID)
}

//--------------------------------------EOF ContainerdAPIClient implementation with Containerd -------------------------------------

//----------------------------Disposable-------------------------------------------

func (ctrdClient *containerdClient) Dispose(ctx context.Context) error {
	if ctrdClient.eventsCancel != nil {
		ctrdClient.eventsCancel()
	}
	if ctrdClient.imagesWatcher != nil { // disabled expiry management
		ctrdClient.imagesWatcher.Dispose()

		ctrdClient.imagesExpiryLock.Lock() // wait for all expiry handling to finish
		defer ctrdClient.imagesExpiryLock.Unlock()
	}

	ctrdClient.ctrdCache.setContainerdDead(true)
	return ctrdClient.spi.Dispose(ctx)
}

//----------------------------EOF Disposable-------------------------------------------
