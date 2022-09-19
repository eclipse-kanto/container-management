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
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

func TestInit(t *testing.T) {
	nOpts := []NetOpt{
		WithLibNetType(bridgeNetworkName),
		WithLibNetIPTables(true),
		WithLibNetMtu(1500),
		WithLibNetIPForward(true),
		WithLibNetName("test0")}
	registryCtx := &registry.ServiceRegistryContext{
		Config: nOpts,
	}
	netMgr, err := registryInit(registryCtx)
	testutil.AssertError(t, nil, err)
	testNetMgr := netMgr.(*libnetworkMgr)
	testutil.AssertNotNil(t, testNetMgr)

	expectedNetOpts := &netOpts{}
	applyOptsNet(expectedNetOpts, nOpts...)
	expectedCfg, cfgErr := netMrgOptsToLibnetConfig(expectedNetOpts)
	if cfgErr != nil {
		t.Fatal("could not get expected config from net opts")
	}
	testutil.AssertEqual(t, expectedCfg.netType, testNetMgr.config.netType)
	testutil.AssertEqual(t, expectedCfg.bridgeConfig.ipTables, testNetMgr.config.bridgeConfig.ipTables)
	testutil.AssertEqual(t, expectedCfg.bridgeConfig.mtu, testNetMgr.config.bridgeConfig.mtu)
	testutil.AssertEqual(t, expectedCfg.bridgeConfig.ipForward, testNetMgr.config.bridgeConfig.ipForward)
	testutil.AssertEqual(t, expectedCfg.bridgeConfig.name, testNetMgr.config.bridgeConfig.name)
}
