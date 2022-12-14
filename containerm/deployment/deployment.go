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
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const (
	// DeploymentManagerServiceLocalID local ID for the deploymentMgr service.
	DeploymentManagerServiceLocalID = "container-management.service.local.v1.service-deploymentMgr-manager"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       DeploymentManagerServiceLocalID,
		Type:     registry.DeploymentManagerService,
		InitFunc: registryInit,
	})
}

type deploymentMgr struct {
	initialDeployPath string
	ctrMgr            mgr.ContainerManager
	deploymentLock    sync.RWMutex
	disposeLock       sync.RWMutex
	disposed          bool
}

func (d *deploymentMgr) InitialDeploy(ctx context.Context) error {
	listCtrs, err := d.ctrMgr.List(ctx)
	if err != nil {
		return err
	}
	if len(listCtrs) > 0 {
		log.Debug("there are loaded container resources, will skip initial containers deploy")
		return nil
	}

	if _, err = os.Stat(d.initialDeployPath); os.IsNotExist(err) {
		log.Debug("the initial deploy containers directory does not exist - will exit deploying ")
		return nil
	}

	var ctrs []*types.Container
	err = filepath.WalkDir(d.initialDeployPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			ctr, readErr := util.ReadContainer(path)
			if readErr != nil {
				log.ErrorErr(readErr, "error reading container configuration from file = %s", path)
			} else {
				ctrs = append(ctrs, ctr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	go d.processInitialDeploy(ctx, ctrs)

	return nil
}

func (d *deploymentMgr) Dispose(ctx context.Context) error {
	d.disposeLock.Lock()
	d.disposed = true
	d.disposeLock.Unlock()

	d.deploymentLock.RLock()
	d.deploymentLock.RUnlock()
	return nil
}
