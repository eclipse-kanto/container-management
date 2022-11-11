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
	"path/filepath"
	"testing"

	libnetconfig "github.com/docker/docker/libnetwork/config"
	"github.com/docker/docker/libnetwork/netlabel"
	"github.com/docker/docker/libnetwork/options"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

var (
	testNetMetaPath = filepath.Join(testDirsRoot, "meta")
	testNetExecRoot = filepath.Join(testDirsRoot, "exec")
)

func TestDriverNetworkOptions(t *testing.T) {
	tests := map[string]struct {
		netConfig             config
		expectedDriverOptions []libnetconfig.Option
	}{
		"driver_opts_default": {
			netConfig: config{
				netType:  bridgeNetworkName,
				metaPath: testNetMetaPath,
				execRoot: testNetExecRoot,
				bridgeConfig: bridgeConfig{
					ipTables:      true,
					ipForward:     true,
					userlandProxy: false,
				},
			},
			expectedDriverOptions: []libnetconfig.Option{libnetconfig.OptionDriverConfig(bridgeNetworkName,
				options.Generic{netlabel.GenericData: options.Generic{
					"EnableIPForwarding":  true,
					"EnableIPTables":      true,
					"EnableUserlandProxy": false,
				}}),
			},
		},
		"driver_opts_no_ip_tables": {
			netConfig: config{
				netType:  bridgeNetworkName,
				metaPath: testNetMetaPath,
				execRoot: testNetExecRoot,
				bridgeConfig: bridgeConfig{
					ipTables:      false,
					ipForward:     false,
					userlandProxy: true,
				},
			},
			expectedDriverOptions: []libnetconfig.Option{libnetconfig.OptionDriverConfig(bridgeNetworkName,
				options.Generic{netlabel.GenericData: options.Generic{
					"EnableIPForwarding":  false,
					"EnableIPTables":      false,
					"EnableUserlandProxy": true,
				}}),
			},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			cfgExpected := &libnetconfig.Config{Daemon: libnetconfig.DaemonCfg{
				DriverCfg: make(map[string]interface{}),
			},
			}
			resultDriverOpts := driverNetworkOptions(testCase.netConfig)
			cfgActual := &libnetconfig.Config{Daemon: libnetconfig.DaemonCfg{
				DriverCfg: make(map[string]interface{}),
			},
			}
			testutil.AssertEqual(t, len(testCase.expectedDriverOptions), len(resultDriverOpts))
			testutil.AssertEqual(t, 1, len(resultDriverOpts))
			testCase.expectedDriverOptions[0](cfgExpected)
			resultDriverOpts[0](cfgActual)

			testutil.AssertEqual(t, cfgExpected.Daemon.DriverCfg[testCase.netConfig.netType], cfgActual.Daemon.DriverCfg[testCase.netConfig.netType])
		})
	}

}
func TestBuildNetworkControllerOptions(t *testing.T) {
	tests := map[string]struct {
		netConfig                 *config
		expectedControllerOptions []libnetconfig.Option
		expectedError             error
	}{
		"ctrl_opts_nil": {},
		"ctrl_opts_no_sbs": {
			netConfig: &config{
				netType:  bridgeNetworkName,
				metaPath: testNetMetaPath,
				execRoot: testNetExecRoot,
				bridgeConfig: bridgeConfig{
					ipTables:      false,
					ipForward:     false,
					userlandProxy: true,
					mtu:           1500,
				},
			},
			expectedControllerOptions: []libnetconfig.Option{
				libnetconfig.OptionExperimental(false),
				//init directories
				libnetconfig.OptionDataDir(testNetMetaPath),
				libnetconfig.OptionExecRoot(testNetExecRoot),
				libnetconfig.OptionDefaultDriver(bridgeNetworkName),
				libnetconfig.OptionDefaultNetwork(bridgeNetworkName),

				libnetconfig.OptionDriverConfig(bridgeNetworkName,
					options.Generic{netlabel.GenericData: options.Generic{
						"EnableIPForwarding":  false,
						"EnableIPTables":      false,
						"EnableUserlandProxy": true,
					}}),
				libnetconfig.OptionNetworkControlPlaneMTU(1500),
			},
		},
		"ctrl_opts_with_sbs": {
			netConfig: &config{
				netType:  bridgeNetworkName,
				metaPath: testNetMetaPath,
				execRoot: testNetExecRoot,
				bridgeConfig: bridgeConfig{
					ipTables:      false,
					ipForward:     false,
					userlandProxy: true,
					mtu:           1500,
				},
				activeSandboxes: map[string]interface{}{
					"sb1": "sb1",
				},
			},
			expectedControllerOptions: []libnetconfig.Option{
				libnetconfig.OptionExperimental(false),
				//init directories
				libnetconfig.OptionDataDir(testNetMetaPath),
				libnetconfig.OptionExecRoot(testNetExecRoot),
				libnetconfig.OptionActiveSandboxes(map[string]interface{}{
					"sb1": "sb1",
				}),
				libnetconfig.OptionDefaultDriver(bridgeNetworkName),
				libnetconfig.OptionDefaultNetwork(bridgeNetworkName),

				libnetconfig.OptionDriverConfig(bridgeNetworkName,
					options.Generic{netlabel.GenericData: options.Generic{
						"EnableIPForwarding":  false,
						"EnableIPTables":      false,
						"EnableUserlandProxy": true,
					}}),
				libnetconfig.OptionNetworkControlPlaneMTU(1500),
			},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			if testCase.netConfig == nil {
				testutil.AssertNil(t, testCase.expectedControllerOptions)
				return
			}
			cfgExpected := &libnetconfig.Config{Daemon: libnetconfig.DaemonCfg{
				DriverCfg: make(map[string]interface{}),
			},
			}
			resultControllerOpts, err := buildNetworkControllerOptions(testCase.netConfig)
			cfgActual := &libnetconfig.Config{Daemon: libnetconfig.DaemonCfg{
				DriverCfg: make(map[string]interface{}),
			},
			}
			testutil.AssertError(t, testCase.expectedError, err)
			testutil.AssertEqual(t, len(testCase.expectedControllerOptions), len(resultControllerOpts))

			for _, opt := range testCase.expectedControllerOptions {
				opt(cfgExpected)
			}

			for _, opt := range resultControllerOpts {
				opt(cfgActual)
			}

			testutil.AssertEqual(t, cfgExpected.Daemon.Experimental, cfgActual.Daemon.Experimental)
			testutil.AssertEqual(t, cfgExpected.Daemon.DataDir, cfgActual.Daemon.DataDir)
			testutil.AssertEqual(t, cfgExpected.Daemon.ExecRoot, cfgActual.Daemon.ExecRoot)
			testutil.AssertEqual(t, cfgExpected.ActiveSandboxes, cfgActual.ActiveSandboxes)
			testutil.AssertEqual(t, cfgExpected.Daemon.DefaultDriver, cfgActual.Daemon.DefaultDriver)
			testutil.AssertEqual(t, cfgExpected.Daemon.DefaultNetwork, cfgActual.Daemon.DefaultNetwork)
			testutil.AssertEqual(t, cfgExpected.Daemon.DriverCfg[testCase.netConfig.netType], cfgActual.Daemon.DriverCfg[testCase.netConfig.netType])
			testutil.AssertEqual(t, cfgExpected.Daemon.NetworkControlPlaneMTU, cfgActual.Daemon.NetworkControlPlaneMTU)
		})
	}
}

func TestNetMrgOptsToLibnetConfig(t *testing.T) {
	netOptsToTest := &netOpts{
		netType:       bridgeNetworkName,
		metaPath:      testNetMetaPath,
		execRoot:      testNetExecRoot,
		disableBridge: false,
		name:          bridgeNetworkName,
		ipV4:          "ipV4",
		fixedCIDRv4:   "fixedCIDRv4",
		gatewayIPv4:   "gatewayIPv4",
		enableIPv6:    false,
		mtu:           1500,
		icc:           false,
		ipTables:      true,
		ipForward:     true,
		ipMasq:        false,
		userlandProxy: false,
	}
	expectedCfg := config{
		netType:  netOptsToTest.netType,
		metaPath: netOptsToTest.metaPath,
		execRoot: netOptsToTest.execRoot,
		bridgeConfig: bridgeConfig{
			disableBridge: netOptsToTest.disableBridge,
			name:          netOptsToTest.name,
			ipV4:          netOptsToTest.ipV4,
			fixedCIDRv4:   netOptsToTest.fixedCIDRv4,
			gatewayIPv4:   netOptsToTest.gatewayIPv4,
			enableIPv6:    netOptsToTest.enableIPv6,
			mtu:           netOptsToTest.mtu,
			icc:           netOptsToTest.icc,
			ipTables:      netOptsToTest.ipTables,
			ipForward:     netOptsToTest.ipForward,
			ipMasq:        netOptsToTest.ipMasq,
			userlandProxy: netOptsToTest.userlandProxy,
		},
		activeSandboxes: make(map[string]interface{}),
	}
	resultCfg, err := netMrgOptsToLibnetConfig(netOptsToTest)
	testutil.AssertError(t, nil, err)

	testutil.AssertEqual(t, expectedCfg.netType, resultCfg.netType)
	testutil.AssertEqual(t, expectedCfg.metaPath, resultCfg.metaPath)
	testutil.AssertEqual(t, expectedCfg.execRoot, resultCfg.execRoot)
	testutil.AssertEqual(t, expectedCfg.activeSandboxes, resultCfg.activeSandboxes)
	testutil.AssertEqual(t, expectedCfg.bridgeConfig, resultCfg.bridgeConfig)

}
