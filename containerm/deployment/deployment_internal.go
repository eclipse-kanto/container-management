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

		ctr, createErr := d.ctrMgr.Create(ctx, container)
		if createErr != nil {
			log.WarnErr(createErr, "could not create container with name = %s and image name = %s", container.Name, container.Image.Name)
		} else {
			log.Debug("successfully created container with ID = %s", ctr.ID)
			if startErr := d.ctrMgr.Start(ctx, ctr.ID); startErr != nil {
				log.WarnErr(startErr, "could not start container with ID = %s", ctr.ID)
			} else {
				log.Debug("successfully started container with ID = %s", ctr.ID)
			}
		}
	}
	log.Debug("finished initial containers deploy")
}
