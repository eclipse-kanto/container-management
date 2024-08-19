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

package mgr

import (
	"context"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/network"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	"github.com/eclipse-kanto/container-management/containerm/util"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
)

const (
	// ManagerServiceLocalID local ID for the manager service.
	ManagerServiceLocalID       = "container-management.service.local.v1.service-manager"
	noSuchContainerErrorMsg     = "no such container with id = %s exists"
	failedConfigStoringErrorMsg = "failed to store container's configuration to local storage"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       ManagerServiceLocalID,
		Type:     registry.ContainerManagerService,
		InitFunc: registryInit,
	})
}

type containerMgr struct {
	metaPath               string
	execPath               string
	defaultCtrsStopTimeout time.Duration
	ctrClient              ctr.ContainerAPIClient
	netMgr                 network.ContainerNetworkManager
	eventsMgr              events.ContainerEventsManager

	containers     map[string]*types.Container
	containersLock sync.RWMutex

	restartCtrsMgrCache *restartMgrCache
	containerRepository containerRepository
}

// Load all container data prior to loading the actual containers in the client
func (mgr *containerMgr) Load(ctx context.Context) error {
	if pruneErr := mgr.containerRepository.Prune(); pruneErr != nil {
		log.DebugErr(pruneErr, "could not prune containers")
	}

	readCtrs, err := mgr.containerRepository.ReadAll()

	// update defaults to current to ensure backwards compatibility
	mgr.fillCurrentDefaults(readCtrs)

	if err != nil {
		return err
	}

	if readCtrs == nil || len(readCtrs) == 0 {
		log.Debug("no container configurations loaded")
		return nil
	}

	mgr.containersLock.Lock()
	defer mgr.containersLock.Unlock()
	mgr.containers = make(map[string]*types.Container)
	for _, v := range readCtrs {
		mgr.containers[v.ID] = v
	}
	return nil
}

// Restore recover alive containers only
func (mgr *containerMgr) Restore(ctx context.Context) error {
	mgr.containersLock.RLock()
	defer mgr.containersLock.RUnlock()
	var (
		err error
	)

	ctrs := mgr.containersToArray()
	deadCtrIds := make([]string, 0)

	if err = mgr.netMgr.Restore(ctx, ctrs); err != nil {
		log.ErrorErr(err, "could not restore network resources for running containers")
	}

	if ctrs != nil && len(ctrs) > 0 {
		for _, ctr := range ctrs {
			log.Debug("start loading for container id = %s", ctr.ID)
			if util.IsContainerDead(ctr) {
				log.Warn("will not load dead container with id = %s", ctr.ID)

				pth := mgr.getContainerMetaPath(ctr.ID)
				if err := os.RemoveAll(pth); err != nil {
					log.ErrorErr(err, "failed to Delete local storage for container id = %s", ctr.ID)
				}

				deadCtrIds = append(deadCtrIds, ctr.ID)

				continue
			}

			// Note: while the daemon is starting we must initalize the container IOs as
			// the user may try to start a stopped container in the meanwhile
			// TODO handle after GA version

			// recover only running or paused containers
			if !util.IsContainerRunningOrPaused(ctr) {
				log.Debug("will not load container with id = %s as the container is not active - it is with status %s", ctr.ID, ctr.State.Status.String())
				continue
			}
			mgr.cancelContainerRestartManager(ctr)
			if err = mgr.ctrClient.RestoreContainer(ctx, ctr); err == nil {
				mgr.containerRepository.Save(ctr)
				continue
			}

			log.ErrorErr(err, "could not restore container ID = %s in the underlying container management runtime", ctr.ID)
			if err := mgr.exitedAndRelease(ctr, -1, err, false, nil); err != nil {
				log.ErrorErr(err, "could not execute internal exit handler for container id = %s", ctr.ID)
			}
		}

		for _, ctrID := range deadCtrIds {
			delete(mgr.containers, ctrID)
		}
	} else {
		log.Debug("no containers data loaded from the persistent storage - nothing to restore")
	}

	if err = mgr.netMgr.Initialize(ctx); err != nil {
		log.ErrorErr(err, "could not restore network resources for running containers")
	}

	log.Debug("restarting restored containers compliant with their restart policies")
	mgr.startRestoredContainers(ctx, ctrs)
	log.Debug("finished restarting restored containers")
	return nil
}

// Create a new container.
func (mgr *containerMgr) Create(ctx context.Context, container *types.Container) (*types.Container, error) {

	ctr := mgr.getContainerFromCache(container.ID)
	if ctr != nil {
		return nil, log.NewErrorf("container with id = %s already exists", container.ID)
	}

	container.Lock()
	defer container.Unlock()

	util.FillDefaults(container)
	util.FillMemorySwap(container)

	if err := util.ValidateContainer(container); err != nil {
		log.ErrorErr(err, "configuration for container id = %s is invalid", container.ID)
		return nil, err
	}

	container.State = &types.State{
		Status: types.Creating,
	}

	var err error
	pth := mgr.getContainerMetaPath(container.ID)

	defer func() {
		if err != nil {
			if _, err := os.Stat(pth); !os.IsNotExist(err) {
				if err := os.RemoveAll(pth); err != nil {
					log.ErrorErr(err, "failed to Delete meta path for failed container id = %s ", container.ID)
				}
			}
		}
	}()

	if _, err := os.Stat(pth); os.IsNotExist(err) {
		util.MkDir(pth)
	}

	if err = mgr.ctrClient.CreateContainer(ctx, container, ""); err != nil {
		return nil, err
	}

	// update state
	util.SetContainerStatusCreated(container)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersCreated, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", container)
	}

	// Save state
	_, err = mgr.containerRepository.Save(container)
	if err != nil {
		log.ErrorErr(err, "could not write config for container %+v", container)
	}

	mgr.addContainerToCache(container)

	return container, nil
}

// Get the detailed information of container.
func (mgr *containerMgr) Get(ctx context.Context, id string) (*types.Container, error) {
	return mgr.getContainerFromCache(id), nil
}

// List returns the list of containers.
func (mgr *containerMgr) List(ctx context.Context) ([]*types.Container, error) {
	mgr.containersLock.RLock()
	defer mgr.containersLock.RUnlock()
	ctrs := mgr.containersToArray()
	return ctrs, nil
}

// Start a container.
func (mgr *containerMgr) Start(ctx context.Context, id string) error {
	return mgr.processStartContainer(ctx, id, true)
}

// Attach attaches the container's IO
func (mgr *containerMgr) Attach(ctx context.Context, id string, attachConfig *streams.AttachConfig) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	return mgr.ctrClient.AttachContainer(ctx, container, attachConfig)
}

// Stop a container.
func (mgr *containerMgr) Stop(ctx context.Context, id string, stopOpts *types.StopOpts) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	// just in case there aren't any provided as the internal logic depends on the opts
	if stopOpts == nil {
		stopOpts = &types.StopOpts{}
	}
	mgr.fillContainerStopDefaults(stopOpts)
	if err := util.ValidateStopOpts(stopOpts); err != nil {
		log.ErrorErr(err, "invalid stop options for container id = %s", container.ID)
		return err
	}
	if err := mgr.stopContainer(ctx, container, stopOpts, true); err != nil {
		log.ErrorErr(err, "error stopping container ID = %s", id)
		return err
	}
	container.Lock()
	defer container.Unlock()
	mgr.applyRestartPolicy(context.Background(), container)
	return nil
}

// Update a container.
func (mgr *containerMgr) Update(ctx context.Context, id string, updateOpts *types.UpdateOpts) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	// just in case there aren't any provided as the internal logic depends on the opts
	if updateOpts == nil {
		updateOpts = &types.UpdateOpts{}
	}

	if err := util.ValidateRestartPolicy(updateOpts.RestartPolicy); err != nil {
		log.ErrorErr(err, "will not update container id = %s invalid restart policy", container.ID)
		return err
	}

	if err := util.ValidateResources(updateOpts.Resources); err != nil {
		log.ErrorErr(err, "will not update container id = %s invalid resources", container.ID)
		return err
	}

	container.Lock()
	defer container.Unlock()

	var changesMade bool
	if updateOpts.Resources != nil && !reflect.DeepEqual(updateOpts.Resources, container.HostConfig.Resources) {
		if util.IsContainerRunningOrPaused(container) {
			if err := mgr.ctrClient.UpdateContainer(ctx, container, updateOpts.Resources); err != nil {
				return err
			}
		}
		if (*updateOpts.Resources == types.Resources{}) { // empty, no limits
			container.HostConfig.Resources = nil
		} else {
			container.HostConfig.Resources = updateOpts.Resources
		}
		changesMade = true
	}

	var rpChanged bool
	if updateOpts.RestartPolicy != nil && !reflect.DeepEqual(updateOpts.RestartPolicy, container.HostConfig.RestartPolicy) {
		mgr.resetContainerRestartManager(container, false)
		container.HostConfig.RestartPolicy = updateOpts.RestartPolicy
		changesMade, rpChanged = true, true
	}

	if changesMade {
		if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersUpdated, container); pubErr != nil {
			log.ErrorErr(pubErr, "failed to publish update event for container %+v", container)
		}

		if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
			log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
		}
	}

	if rpChanged && (container.State.Exited || container.State.Status == types.Stopped) {
		mgr.applyRestartPolicy(context.Background(), container)
	}
	return nil
}

// Restart a running container.
func (mgr *containerMgr) Restart(ctx context.Context, id string, timeout int64) error {
	return log.NewErrorf("restart not supported")
}

// Pause a container.
func (mgr *containerMgr) Pause(ctx context.Context, id string) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	container.Lock()
	defer container.Unlock()
	if !container.State.Running {
		return log.NewErrorf("container with id = %s is not running - current status is %s", container.ID, container.State.Status.String())
	}

	if err := mgr.ctrClient.PauseContainer(ctx, container); err != nil {
		return err
	}
	util.SetContainerStatusPaused(container)
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersPaused, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", container)
	}
	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}
	return nil
}

// Unpause a container.
func (mgr *containerMgr) Unpause(ctx context.Context, id string) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	container.Lock()
	defer container.Unlock()

	if !container.State.Paused {
		return log.NewErrorf("container id = %s is not paused - cannot unpause it ", container.ID)
	}

	if err := mgr.ctrClient.UnpauseContainer(ctx, container); err != nil {
		return err
	}
	util.SetContainerStatusUnpaused(container)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersResumed, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", container)
	}
	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}
	return nil
}

// Rename renames a container.
func (mgr *containerMgr) Rename(ctx context.Context, id string, name string) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}

	if err := util.ValidateName(name); err != nil {
		log.ErrorErr(err, "will not rename container id = %s invalid name", container.ID)
		return err
	}

	container.Lock()
	defer container.Unlock()

	if name == "" || name == container.Name {
		return nil
	}

	container.Name = name
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersRenamed, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish rename event for container %+v", container)
	}

	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}
	return nil
}

// Remove removes a container, it may be running or stopped and so on.
func (mgr *containerMgr) Remove(ctx context.Context, id string, force bool, removecache bool, stopOpts *types.StopOpts) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}
	container.Lock()
	defer container.Unlock()

	if util.IsContainerRunningOrPaused(container) && !force {
		return log.NewErrorf("container with id = %s is not stopped - must set the force flag to true to remove it", container.ID)
	}

	if container.State.Dead {
		log.Warn("container with id = %t is already removed", removecache)
		return nil
	}
	// if the container is running and force is set to true - try to stop and remove it
	if (!util.IsContainerRunningOrPaused(container)) || force {
		mgr.cancelContainerRestartManager(container)

		if stopOpts == nil {
			stopOpts = mgr.getContainerStopOptions(force)
		} else {
			mgr.fillContainerStopDefaults(stopOpts)
		}
		if err := util.ValidateStopOpts(stopOpts); err != nil {
			log.ErrorErr(err, "invalid stop options for container id = %s", container.ID)
			return err
		}
		_, _, err := mgr.ctrClient.DestroyContainer(ctx, container, stopOpts, true)
		if err != nil && !(strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found")) {
			mgr.resetContainerRestartManager(container, false)
			log.ErrorErr(err, "removing error while trying to destroy container with id = %s", container.ID)
			return err
		}
		if err := mgr.releaseContainerResources(container); err != nil {
			log.ErrorErr(err, "removing error while trying to release resources for container id = %s", container.ID)
		}
		if removecache {
			if err := mgr.ctrClient.RemoveContainerImage(ctx, container.Image); err != nil {
				log.ErrorErr(err, "unable to remove container Image %s", container.Image)
				return err
			}
		}
	}

	util.SetContainerStatusDead(container)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersRemoved, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", container)
	}
	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}

	err := mgr.containerRepository.Delete(id)

	mgr.removeContainerRestartManager(container)
	mgr.removeContainerFromCache(id)

	if err != nil {
		log.WarnErr(err, "failed to Delete container file with id: %s", id)
	}

	return nil
}

func (mgr *containerMgr) Metrics(ctx context.Context, id string) (*types.Metrics, error) {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return nil, log.NewErrorf(noSuchContainerErrorMsg, id)
	}

	container.Lock()
	defer container.Unlock()

	if !util.IsContainerRunningOrPaused(container) {
		// no metrics for not running container
		return nil, nil
	}

	var (
		m      = &types.Metrics{}
		errCtr error
		errNet error
	)

	m.CPU, m.Memory, m.IO, m.PIDs, m.Timestamp, errCtr = mgr.ctrClient.GetContainerStats(ctx, container)
	m.Network, errNet = mgr.netMgr.Stats(ctx, container)

	// both failed - return compound error, only one failed - log warning and return the metrics
	if errNet != nil {
		if errCtr != nil {
			errs := &errorUtil.CompoundError{}
			errs.Append(errCtr, errNet)
			return nil, errs
		}
		log.WarnErr(errCtr, "could not get network stats for container with ID = %s", container.ID)
	} else if errCtr != nil {
		m.Timestamp = time.Now()
		log.WarnErr(errCtr, "could not get CPU, memory, IO and PIDs stats for container with ID = %s", container.ID)
	}
	return m, nil
}

//--------------------------------- Disposable impl -----------------------------------

func (mgr *containerMgr) Dispose(ctx context.Context) error {
	log.Debug("waiting for any container resources and operations to finish")
	if err := mgr.stopManagerService(ctx); err != nil {
		log.WarnErr(err, "error while stopping containers on stopping the container management service")
	}
	log.Debug("finished clearing container resources and operations")
	var err error
	err = mgr.ctrClient.Dispose(ctx)
	err = mgr.netMgr.Dispose(ctx)
	if err != nil {
		return err
	}
	return nil
}
