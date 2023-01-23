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

package network

import (
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func newLibnetworkMgr(netConfig config) (ContainerNetworkManager, error) {
	if err := util.MkDirs(netConfig.execRoot); err != nil {
		return nil, err
	}

	if err := util.MkDir(netConfig.metaPath); err != nil {
		return nil, err
	}

	return &libnetworkMgr{config: &netConfig, bridgeConnectedContainers: make(map[string]*types.Container)}, nil
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	netMgrOpts := registryCtx.Config.([]NetOpt)
	netMgrCreateOpts := &netOpts{}
	applyOptsNet(netMgrCreateOpts, netMgrOpts...)

	var (
		err       error
		netConfig config
	)
	//convert opts to libnet config
	netConfig, err = netMrgOptsToLibnetConfig(netMgrCreateOpts)
	if err != nil {
		return nil, err
	}

	//create libnetwork manager
	return newLibnetworkMgr(netConfig)
}
