// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package deployment

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func (d *deploymentMgr) processInitialDeploy(ctx context.Context, containers []*types.Container) {
	d.deploymentLock.Lock()
	defer d.deploymentLock.Unlock()

	log.Debug("starting initial containers deploy")
	for _, container := range containers {
		d.disposeLock.RLock()
		if d.disposed {
			d.disposeLock.RUnlock()
			log.Warn("interrupted initial containers deploy")
			return
		}
		d.disposeLock.RUnlock()

		createAndStartContainer(ctx, d.ctrMgr, container)
	}
	log.Debug("finished initial containers deploy")
}

func (d *deploymentMgr) processUpdate(ctx context.Context, existing []*types.Container, target []*types.Container) {
	d.deploymentLock.Lock()
	defer d.deploymentLock.Unlock()

	log.Debug("starting containers update")

	mapCurrent := util.AsNamedMap(existing)

	for _, desired := range target {
		d.disposeLock.RLock()
		if d.disposed {
			d.disposeLock.RUnlock()
			log.Warn("interrupted containers update")
			return
		}
		d.disposeLock.RUnlock()

		id := desired.ID
		util.FillDefaults(desired)
		desired.ID = id
		current := mapCurrent[desired.Name]

		action := util.DetermineUpdateAction(current, desired)
		switch action {
		case util.ActionCheck:
			ensureContainerRunning(ctx, d.ctrMgr, current)
		case util.ActionCreate:
			createAndStartContainer(ctx, d.ctrMgr, desired)
		case util.ActionRecreate:
			recreateAndStartContainer(ctx, d.ctrMgr, current, desired)
		case util.ActionUpdate:
			updateContainer(ctx, d.ctrMgr, current, desired)
			ensureContainerRunning(ctx, d.ctrMgr, current)
		}
	}
	log.Debug("finished containers update")
}

func ensureContainerRunning(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) {
	if container.State.Running {
		log.Debug("container with ID = %s, name = %s and image name = %s is already running, nothing to do", container.ID, container.Name, container.Image.Name)
	} else if container.State.Paused {
		unpauseContainer(ctx, ctrMgr, container)
	} else {
		startContainer(ctx, ctrMgr, container)
	}
}

func createAndStartContainer(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) {
	ctr, createErr := ctrMgr.Create(ctx, container)
	if createErr != nil {
		log.WarnErr(createErr, "could not create container with name = %s and image name = %s", container.Name, container.Image.Name)
	} else {
		log.Debug("successfully created container with ID = %s, name = %s and image name = %s", ctr.ID, ctr.Name, ctr.Image.Name)
		startContainer(ctx, ctrMgr, ctr)
	}
}

func recreateAndStartContainer(ctx context.Context, ctrMgr mgr.ContainerManager, current *types.Container, desired *types.Container) {
	ctr, createErr := ctrMgr.Create(ctx, desired)
	if createErr != nil {
		log.WarnErr(createErr, "could not create new container with name = %s and image name = %s", desired.Name, desired.Image.Name)
		return
	}
	log.Debug("successfully created new container with ID = %s, name = %s and image name = %s", ctr.ID, ctr.Name, ctr.Image.Name)
	stopped := stopContainer(ctx, ctrMgr, current)
	if startErr := startContainer(ctx, ctrMgr, ctr); startErr != nil {
		log.Warn("restoring old container with ID = %s, name = %s and image name = %s", current.ID, current.Name, current.Image.Name)
		if stopped {
			startContainer(ctx, ctrMgr, current)
		}
		log.Warn("removing not-started new container with ID = %s, name = %s and image name = %s", ctr.ID, ctr.Name, ctr.Image.Name)
		removeContainer(ctx, ctrMgr, ctr)
	} else {
		log.Debug("removing old container with ID = %s, name = %s and image name = %s", current.ID, current.Name, current.Image.Name)
		removeContainer(ctx, ctrMgr, current)
	}
}

func updateContainer(ctx context.Context, ctrMgr mgr.ContainerManager, current *types.Container, desired *types.Container) {
	updateOpts := &types.UpdateOpts{
		RestartPolicy: desired.HostConfig.RestartPolicy,
		Resources:     desired.HostConfig.Resources,
	}
	if updateErr := ctrMgr.Update(ctx, current.ID, updateOpts); updateErr != nil {
		log.WarnErr(updateErr, "could not update container with ID = %s, name = %s and image name = %s", current.ID, current.Name, current.Image.Name)
	} else {
		log.Debug("successfully updated container with ID = %s, name = %s and image name = %s", current.ID, current.Name, current.Image.Name)
	}
}

func startContainer(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) error {
	if startErr := ctrMgr.Start(ctx, container.ID); startErr != nil {
		log.WarnErr(startErr, "could not start container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
		return startErr
	}
	log.Debug("successfully started container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	return nil
}

func unpauseContainer(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) {
	if unpauseErr := ctrMgr.Unpause(ctx, container.ID); unpauseErr != nil {
		log.WarnErr(unpauseErr, "could not unpause container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	} else {
		log.Debug("successfully unpaused container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	}
}

func stopContainer(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) bool {
	if !util.IsContainerRunningOrPaused(container) {
		log.Debug("container with ID = %s, name = %s and image name = %s is not running, nor paused", container.ID, container.Name, container.Image.Name)
		return false
	}
	stopOpts := &types.StopOpts{
		Force:  true,
		Signal: "SIGTERM",
	}
	if stopErr := ctrMgr.Stop(ctx, container.ID, stopOpts); stopErr != nil {
		log.WarnErr(stopErr, "could not stop container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
		return false
	}
	log.Debug("successfully stopped container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	return true
}

func removeContainer(ctx context.Context, ctrMgr mgr.ContainerManager, container *types.Container) {
	if removeErr := ctrMgr.Remove(ctx, container.ID, true, nil); removeErr != nil {
		log.WarnErr(removeErr, "could not remove container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	} else {
		log.Debug("successfully removed container with ID = %s, name = %s and image name = %s", container.ID, container.Name, container.Image.Name)
	}
}
