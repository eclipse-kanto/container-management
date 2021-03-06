// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package network

import (
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

func newLibnetworkMgr(netConfig config) (ConteinerNetworkManager, error) {
	return &libnetworkMgr{&netConfig, nil}, nil
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
