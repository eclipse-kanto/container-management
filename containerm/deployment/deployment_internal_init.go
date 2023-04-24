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
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	initOpts := registryCtx.Config.([]Opt)

	options := &opts{}
	applyOpts(options, initOpts...)

	mgrService, err := registryCtx.Get(registry.ContainerManagerService)
	if err != nil {
		return nil, err
	}

	//initialize the deployment manager local service
	return newDeploymentMgr(options.mode, options.metaPath, options.ctrPath, mgrService.(mgr.ContainerManager))
}

func newDeploymentMgr(mode Mode, metaPath, ctrPath string, ctrMgr mgr.ContainerManager) (Manager, error) {
	if err := util.MkDir(metaPath); err != nil {
		return nil, err
	}

	return &deploymentMgr{
		mode:     mode,
		metaPath: metaPath,
		ctrPath:  ctrPath,
		ctrMgr:   ctrMgr,
	}, nil
}
