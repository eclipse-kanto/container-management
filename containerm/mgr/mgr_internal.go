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
	"path/filepath"
	"runtime"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"

	"golang.org/x/sync/semaphore"
)

const (
	sigterm = "SIGTERM"
)

func (mgr *containerMgr) getContainerMetaPath(containerID string) string {
	return filepath.Join(mgr.metaPath, containersRootDir, containerID)
}

func (mgr *containerMgr) fillCurrentDefaults(ctrs []*types.Container) {
	if ctrs != nil && len(ctrs) > 0 {
		for _, ctr := range ctrs {
			if util.FillDefaults(ctr) {
				if _, err := mgr.containerRepository.Save(ctr); err != nil {
					log.ErrorErr(err, "error while updating container's configuration for container ID=%s", ctr.ID)
				} else {
					log.Debug("successfully added and stored missing default configurations for container ID=%s", ctr.ID)
				}
			} else {
				log.Debug("no configuration changes required for container ID=%s after checking for missing default values", ctr.ID)
			}
		}
	}
}

func (mgr *containerMgr) exitedAndRelease(container *types.Container, exitCode int64, exitErr error, oomKilled bool, cleanup func() error) error {
	container.Lock()
	defer container.Unlock()

	ctx := context.Background()

	if cleanup != nil {
		if err := cleanup(); err != nil {
			return err
		}
	}

	if err := mgr.updateConfigToExited(ctx, container, exitCode, exitErr, oomKilled); err != nil {
		log.WarnErr(err, "could not clear resources for container id = %s", container.ID)
	}

	if mgr.getContainerFromCache(container.ID) == nil {
		log.Debug("container with id = %s is removed - no further processing of releasing the resources", container.ID)
		return nil
	}

	// check status and restart policy
	if !container.State.Exited {
		log.Debug("the container's state is not Exited - will not process event")
		return nil
	}
	mgr.applyRestartPolicy(ctx, container)
	return nil
}

func (mgr *containerMgr) applyRestartPolicy(ctx context.Context, container *types.Container) {
	restart, wait, err := mgr.getContainerRestartManager(container).shouldRestart(uint32(container.State.ExitCode), container.ManuallyStopped, util.CalculateUptime(container))
	if err == nil && restart {
		container.RestartCount++
	} else {
		log.Debug("container ID = %s exited and its restart policy does not require auto-start - leaving as exited", container.ID)
	}
	if err == nil && restart {
		go func() {
			err := <-wait
			if err == nil {
				if err = mgr.processStartContainer(ctx, container.ID, false); err != nil {
					log.DebugErr(err, "failed to restart container id = %s", container.ID)
				}
			}
			if err != nil {
				mgr.updateConfigToStopped(ctx, container, -1, err, false)
			}
		}()
	}
}

func (mgr *containerMgr) releaseContainerResources(container *types.Container) error {
	log.Debug("will clean all managed resources for container %s ", container.ID)

	ctx := context.Background()
	// release container client resources
	if err := mgr.ctrClient.ReleaseContainerResources(ctx, container); err != nil {
		return err
	}

	// release container network resources
	if err := mgr.netMgr.ReleaseNetworkResources(ctx, container); err != nil {
		return err
	}
	return nil
}

// the mgr.containersLock must be used when calling this method
func (mgr *containerMgr) containersToArray() []*types.Container {
	if mgr.containers == nil || len(mgr.containers) == 0 {
		log.Debug("no containers available")
		return nil
	}

	ctrs := make([]*types.Container, 0, len(mgr.containers))

	for _, ctr := range mgr.containers {
		ctrs = append(ctrs, ctr)
	}
	return ctrs
}
func (mgr *containerMgr) updateConfigToStopped(ctx context.Context, c *types.Container, exitCode int64, err error, releaseContainerResources bool) error {
	var (
		code   int64
		errMsg string
	)
	if exitCode != -1 {
		code = exitCode
	}
	if err != nil {
		errMsg = err.Error()
	}
	util.SetContainerStatusStopped(c, code, errMsg)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersStopped, c); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", c)
	}

	defer func() {
		if _, err := mgr.containerRepository.Save(c); err != nil {
			log.ErrorErr(err, "error while updating container's configuration to stopped for container id =%s", c.ID)
		}
	}()

	if releaseContainerResources {
		return mgr.releaseContainerResources(c)
	}
	return nil
}

func (mgr *containerMgr) updateConfigToExited(ctx context.Context, c *types.Container, exitCode int64, err error, oomKilled bool) error {
	var (
		code   int64
		errMsg string
	)
	if exitCode != -1 {
		code = exitCode
	}
	if err != nil {
		errMsg = err.Error()
	}
	util.SetContainerStatusExited(c, code, errMsg, oomKilled)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersExited, c); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", c)
	}

	defer func() {
		if _, err := mgr.containerRepository.Save(c); err != nil {
			log.ErrorErr(err, "error while updating container's configuration to exited for container id =%s", c.ID)
		}
	}()

	return mgr.releaseContainerResources(c)
}

func (mgr *containerMgr) addContainerToCache(container *types.Container) {
	mgr.containersLock.Lock()
	defer mgr.containersLock.Unlock()
	mgr.containers[container.ID] = container
}

func (mgr *containerMgr) removeContainerFromCache(id string) {
	mgr.containersLock.Lock()
	defer mgr.containersLock.Unlock()
	delete(mgr.containers, id)
}

func (mgr *containerMgr) getContainerFromCache(id string) *types.Container {
	mgr.containersLock.RLock()
	defer mgr.containersLock.RUnlock()
	return mgr.containers[id]
}

func (mgr *containerMgr) publishContainerStateChangedEvent(ctx context.Context, action types.EventAction, container *types.Container) error {
	return mgr.eventsMgr.Publish(ctx, types.EventTypeContainers, action, container)
}

func (mgr *containerMgr) removeContainerRestartManager(container *types.Container) {
	log.Debug("removing restartManager for container id = %s", container.ID)
	if resMan := mgr.restartCtrsMgrCache.get(container.ID); resMan != nil {
		resMan.cancel()
		mgr.restartCtrsMgrCache.remove(container.ID)
	}
}

func (mgr *containerMgr) getContainerRestartManager(container *types.Container) *restartManager {
	log.Debug("getting restartManager for container id = %s", container.ID)
	resMan := mgr.restartCtrsMgrCache.get(container.ID)
	if resMan == nil {
		log.Debug("initializing restartManager for container id = %s", container.ID)
		newResMan := newRestartManager(container.HostConfig.RestartPolicy, container.RestartCount)
		mgr.restartCtrsMgrCache.put(container.ID, newResMan)
		return newResMan
	}
	log.Debug("restartManager for container id = %s exists - returning", container.ID)
	return resMan
}

func (mgr *containerMgr) resetContainerRestartManager(container *types.Container, resetCounter bool) {
	log.Debug("resetting restartManager for container id = %s", container.ID)
	if resMan := mgr.restartCtrsMgrCache.get(container.ID); resMan != nil {
		mgr.removeContainerRestartManager(container)
	}
	log.Debug("cancelling RestartCount for container id = %s", container.ID)
	if resetCounter {
		container.RestartCount = 0
	}
}

func (mgr *containerMgr) cancelContainerRestartManager(container *types.Container) {
	log.Debug("cancelling restartManager for container id = %s", container.ID)
	mgr.getContainerRestartManager(container).cancel()
}

func (mgr *containerMgr) processStartContainer(ctx context.Context, id string, resetResMan bool) error {
	container := mgr.getContainerFromCache(id)
	if container == nil {
		return log.NewErrorf(noSuchContainerErrorMsg, id)
	}

	container.Lock()
	defer container.Unlock()

	// check if container's status is paused
	if container.State.Paused {
		return log.NewErrorf("the container with id = %s is paused - cannot start it - try unpausing it instead", container.ID)
	}

	// check if container's status is running
	if container.State.Running {
		return log.NewErrorf("the container with id = %s is already running", container.ID)
	}

	if container.State.Dead {
		return log.NewErrorf("the container with id = %s is dead - cannot start it", container.ID)
	}
	var (
		err error
		pid int64
	)
	defer func() {
		if err != nil {
			mgr.releaseContainerResources(container)
		}
	}()

	//add container to network manager
	err = mgr.netMgr.Manage(ctx, container)
	if err != nil {
		return err
	}

	//connect to default network
	err = mgr.netMgr.Connect(ctx, container)
	if err != nil {
		return err
	}

	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}

	if resetResMan {
		mgr.resetContainerRestartManager(container, true)
		container.ManuallyStopped = false
	}

	pid, err = mgr.ctrClient.StartContainer(ctx, container, "")
	if err != nil {
		_ = mgr.updateConfigToStopped(ctx, container, -1, err, true)
		return err
	}

	container.StartedSuccessfullyBefore = true

	util.SetContainerStatusRunning(container, pid)
	// publish event
	if pubErr := mgr.publishContainerStateChangedEvent(ctx, types.EventActionContainersRunning, container); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event for container %+v", container)
	}
	if _, errMeta := mgr.containerRepository.Save(container); errMeta != nil {
		log.ErrorErr(errMeta, failedConfigStoringErrorMsg)
	}
	return nil
}

func (mgr *containerMgr) startRestoredContainers(ctx context.Context, containers []*types.Container) {
	ctrsToRestart := make(map[*types.Container]chan struct{})
	for _, ctr := range containers {
		if !util.IsContainerDead(ctr) && !util.IsContainerRunningOrPaused(ctr) {
			mgr.resetContainerRestartManager(ctr, false)
			if res, _, _ := mgr.getContainerRestartManager(ctr).shouldRestart(uint32(ctr.State.ExitCode), ctr.ManuallyStopped, util.CalculateUptime(ctr)); res && ctr.StartedSuccessfullyBefore {
				ctrsToRestart[ctr] = make(chan struct{})
			}
		}
	}
	parallelLimit := util.CalculateParallelLimit(len(containers), 128*runtime.NumCPU())

	// Re-used for all parallel startup jobs.
	var group sync.WaitGroup
	sem := semaphore.NewWeighted(int64(parallelLimit))

	for c, notifier := range ctrsToRestart {
		group.Add(1)
		go func(c *types.Container, chNotify chan struct{}) {
			_ = sem.Acquire(context.Background(), 1)
			log.Debug("Starting container %s", c.ID)
			if err := mgr.processStartContainer(ctx, c.ID, true); err != nil {
				log.ErrorErr(err, "failed to start container %s", c.ID)
			}
			close(chNotify)

			sem.Release(1)
			group.Done()
		}(c, notifier)
	}
	group.Wait()
}

func (mgr *containerMgr) stopContainer(ctx context.Context, container *types.Container, stopOpts *types.StopOpts, manuallyStopped bool) error {
	container.Lock()
	defer container.Unlock()

	if !util.IsContainerRunningOrPaused(container) {
		return log.NewErrorf("cannot perform stop operation for container: %s, with state: %s", container.ID, container.State.Status.String())
	}
	container.ManuallyStopped = manuallyStopped
	exitCode, _, exitErr := mgr.ctrClient.DestroyContainer(ctx, container, stopOpts, false)
	if exitErr != nil {
		container.ManuallyStopped = false
		return exitErr
	}
	return mgr.updateConfigToStopped(ctx, container, exitCode, exitErr, true)
}

func (mgr *containerMgr) stopManagerService(ctx context.Context) error {
	log.Debug("stopping container manager service")
	mgr.containersLock.Lock()
	defer mgr.containersLock.Unlock()

	ctrs := mgr.containersToArray()
	compoundErr := &errorUtil.CompoundError{}
	for _, ctr := range ctrs {
		log.Debug("cancelling restart manager for container ID = %s", ctr.ID)
		mgr.cancelContainerRestartManager(ctr)
		log.Debug("stopping container ID = %s", ctr.ID)
		opts := mgr.getContainerStopOptions(true)
		if err := mgr.stopContainer(ctx, ctr, opts, false); err != nil {
			log.WarnErr(err, "error while stopping container ID = %s on service exit", ctr.ID)
			compoundErr.Append(err)
		}
	}
	log.Debug("finished containers stop on container manager service exit")
	if compoundErr.Size() > 0 {
		return compoundErr
	}
	return nil
}

func (mgr *containerMgr) fillContainerStopDefaults(stopOpts *types.StopOpts) {
	if stopOpts != nil {
		if stopOpts.Timeout == 0 {
			stopOpts.Timeout = int64(mgr.defaultCtrsStopTimeout.Seconds())
		}
		if stopOpts.Signal == "" {
			stopOpts.Signal = sigterm
		}
	}
}
func (mgr *containerMgr) getContainerStopOptions(force bool) *types.StopOpts {
	return &types.StopOpts{
		Timeout: int64(mgr.defaultCtrsStopTimeout.Seconds()),
		Force:   force,
		Signal:  sigterm,
	}
}
