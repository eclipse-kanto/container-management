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
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/libnetwork"
	libnetTypes "github.com/docker/docker/libnetwork/types"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/network"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/golang/mock/gomock"
)

type prepare func(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager
type prepareInit func(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager
type assertContainer func(t *testing.T, mgrConfig *config, container *types.Container)

const (
	testDirsRoot      = "../pkg/testutil/network-test-root"
	testCtrID         = "test-ctr-id"
	testCtrSandboxID  = "sb-id"
	testCtrSandboxKey = "sb-key"
	testCtrEndpointID = "ep-id"

	testNetworkControllerID = "libnetctrl-id"

	netSettingsIPGW = "216.51.200.201"
	netSettingsIP   = "216.58.208.238"
	netSettingsMac  = "00:a0:c9:14:c8:29"
)

var (
	defaultCfg = &config{
		netType:  "bridge",
		metaPath: testDirsRoot + "/meta",
		execRoot: testDirsRoot + "/exec",
		bridgeConfig: bridgeConfig{
			name: "test0",
		},
	}
)

func newDefaultMgrConfig() *config {
	return &config{
		netType:  "bridge",
		metaPath: testDirsRoot + "/meta",
		execRoot: testDirsRoot + "/exec",
		bridgeConfig: bridgeConfig{
			name: "test0",
		},
	}
}

func newDefaultContainer() *types.Container {
	return &types.Container{
		ID: testCtrID,
		HostConfig: &types.HostConfig{
			NetworkMode: types.NetworkModeBridge,
		},
	}
}
func newDefaultConnectedContainer() *types.Container {
	return &types.Container{
		ID: testCtrID,
		HostConfig: &types.HostConfig{
			NetworkMode: types.NetworkModeBridge,
		},
		NetworkSettings: &types.NetworkSettings{
			Networks: map[string]*types.EndpointSettings{
				defaultCfg.netType: {
					ID: testCtrEndpointID,
				},
			},
			SandboxID: testCtrSandboxID,
		},
	}
}

func TestManage(t *testing.T) {
	tests := map[string]struct {
		mgrConfig         *config
		container         *types.Container
		prepareMgrForTest prepare
		assertCtr         assertContainer
		expectedErr       error
	}{
		"netmgr_test_manage_no_ctrl": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareNilCtrl,
			expectedErr:       log.NewErrorf("no network controller to connect to default network"),
		},
		"netmgr_test_manage_default": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_host_mode": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeHost,
				},
			},
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_extra_hosts": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					ExtraHosts:  []string{"test:1.2.3.4"},
					PortMappings: []types.PortMapping{{
						ContainerPort: 80,
						HostPort:      80,
					}},
				},
			},
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_extra_hosts_bad_key": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					ExtraHosts:  []string{"test:host_ip_"},
					PortMappings: []types.PortMapping{{
						ContainerPort: 80,
						HostPort:      80,
					}},
				},
			},
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_extra_hosts_bad_key_host_mode": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeHost,
					ExtraHosts:  []string{"test:host_ip"},
					PortMappings: []types.PortMapping{{
						ContainerPort: 80,
						HostPort:      80,
					}},
				},
			},
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_extra_hosts_reserved_key_bridge": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					ExtraHosts:  []string{"test:host_ip"},
					PortMappings: []types.PortMapping{{
						ContainerPort: 80,
						HostPort:      80,
					}},
				},
			},
			prepareMgrForTest: prepareDefault,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_existing_sb": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareDefaultExistingSb,
			assertCtr:         assertManagedContainer,
			expectedErr:       nil,
		},
		"netmgr_test_manage_default_existing_sb_destroy_error": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareDestroySbFailed,
			assertCtr:         assertManagedContainer,
			expectedErr:       log.NewErrorf("failed to destroy container sandbox"),
		},
		"netmgr_test_manage_default_existing_sb_create_error": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareNewSbFailed,
			assertCtr:         assertManagedContainer,
			expectedErr:       log.NewErrorf("failed to create container sandbox"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer func() {
				defer os.RemoveAll(testDirsRoot)
				controller.Finish()
			}()
			if dirsErr := util.MkDirs(testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot); dirsErr != nil {
				t.Fatalf("could not create the test directories meta and exec : %s, %s", testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot)
			}

			testMgr := testCase.prepareMgrForTest(controller, testCase.mgrConfig, testCase.container)
			err := testMgr.Manage(context.Background(), testCase.container)
			testutil.AssertError(t, testCase.expectedErr, err)
			// assert container
			if testCase.assertCtr != nil {
				testCase.assertCtr(t, testCase.mgrConfig, testCase.container)
			}
		})
	}

}

func TestConnect(t *testing.T) {
	tests := map[string]struct {
		mgrConfig         *config
		container         *types.Container
		prepareMgrForTest prepare
		assertCtr         assertContainer
		expectedErr       error
	}{
		"netmgr_test_connect_no_network": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareConnectErrorGettingNetwork,
			expectedErr:       log.NewErrorf("no network [%s] found while connecting container %s ", string(newDefaultContainer().HostConfig.NetworkMode), newDefaultContainer().ID),
		},
		"netmgr_test_connect_no_sandbox": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareConnectErrorGettingSb,
			expectedErr:       log.NewErrorf("no network sandbox for container %s ", newDefaultContainer().ID),
		},
		"netmgr_test_connect_no_ep": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*types.EndpointSettings{"bridge": {
						ID: "test-ep-id",
					},
					},
				},
			},
			prepareMgrForTest: prepareConnectErrorGettingEp,
			expectedErr:       log.NewErrorf("no endpoint"),
			assertCtr:         assertConnectedContainerFailed,
		},
		"netmgr_test_connect_no_ep_delete": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*types.EndpointSettings{"bridge": {
						ID: "test-ep-id",
					},
					},
				},
			},
			prepareMgrForTest: prepareConnectErrorGettingEpDelete,
			expectedErr:       log.NewErrorf("error deleting endpoint"),
			assertCtr:         assertConnectedContainerFailed,
		},
		"netmgr_test_connect_err_joining": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*types.EndpointSettings{"bridge": {
						ID: "test-ep-id",
					},
					},
				},
			},
			prepareMgrForTest: prepareConnectErrorJoiningEp,
			expectedErr:       log.NewErrorf("error joining endpoint"),
			assertCtr:         assertConnectedContainerFailed,
		},
		"netmgr_test_connect_err_creating": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*types.EndpointSettings{"bridge": {
						ID: "test-ep-id",
					},
					},
				},
			},
			prepareMgrForTest: prepareConnectErrorCreatingEp,
			expectedErr:       log.NewErrorf("error creating endpoint"),
			assertCtr:         assertConnectedContainerFailed,
		},
		"netmgr_test_connect_from_scratch": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
			},
			prepareMgrForTest: prepareConnectFullNoCtrNetworks,
			assertCtr:         assertConnectedContainer,
		},
		"netmgr_test_connect_with_other_nets": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					SandboxKey:          testCtrSandboxKey,
					SandboxID:           testCtrSandboxID,
					NetworkControllerID: testNetworkControllerID,
					Networks: map[string]*types.EndpointSettings{"bridge": {
						ID: testCtrEndpointID,
					},
					},
				},
			},
			prepareMgrForTest: prepareConnectFullWithOtherCtrNetworks,
			assertCtr:         assertConnectedContainer,
		},
		"netmgr_test_connect_with_nets": {
			mgrConfig: defaultCfg,
			container: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					SandboxKey:          testCtrSandboxKey,
					SandboxID:           testCtrSandboxID,
					NetworkControllerID: testNetworkControllerID,
				},
			},
			prepareMgrForTest: prepareConnectFullWithNetSettings,
			assertCtr:         assertConnectedContainer,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer func() {
				os.RemoveAll(testDirsRoot)
				controller.Finish()
			}()
			if dirsErr := util.MkDirs(testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot); dirsErr != nil {
				t.Fatalf("could not create the test directories meta and exec : %s, %s", testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot)
			}

			testMgr := testCase.prepareMgrForTest(controller, testCase.mgrConfig, testCase.container)
			err := testMgr.Connect(context.Background(), testCase.container)
			testutil.AssertError(t, testCase.expectedErr, err)
			// assert container
			if testCase.assertCtr != nil {
				testCase.assertCtr(t, testCase.mgrConfig, testCase.container)
			}
		})
	}

}

func TestInitialize(t *testing.T) {
	tests := map[string]struct {
		mgrConfig         *config
		prepareMgrForTest prepareInit
		expectedErr       error
	}{
		"netmgr_test_init_no_sbs": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitNoSbs,
		},
		"netmgr_test_init_no_sbs_error_deleting_old_bridge": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitNoSbsErrorDeletingOldBridge,
			expectedErr:       log.NewError("error deleting old bridge"),
		},
		"netmgr_test_init_no_sbs_error_creating_old_bridge": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitNoSbsErrorCreatingNewBridge,
			expectedErr:       log.NewError("error creating default bridge"),
		},
		"netmgr_test_init_no_sbs_error_getting_default_bridge": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitNoSbsErrorDeletingNewBridge,
			expectedErr:       log.NewError("error deleting default bridge"),
		},
		"netmgr_test_init_no_sbs_no_host_net_existing": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitNoSbsNoExistingHostNet,
		},
		"netmgr_test_init_with_sbs": {
			mgrConfig: &config{
				netType:  "bridge",
				metaPath: testDirsRoot + "/meta",
				execRoot: testDirsRoot + "/exec",
				bridgeConfig: bridgeConfig{
					name: "test0",
				},
				activeSandboxes: map[string]interface{}{
					"test": "test",
				},
			},
			prepareMgrForTest: prepareInitWithSbs,
		},
		"netmgr_test_init_with_sbs_bridge_failed": {
			mgrConfig: &config{
				netType:  "bridge",
				metaPath: testDirsRoot + "/meta",
				execRoot: testDirsRoot + "/exec",
				bridgeConfig: bridgeConfig{
					name: "test0",
				},
				activeSandboxes: map[string]interface{}{
					"test": "test",
				},
			},
			prepareMgrForTest: prepareInitWithSbsDefaultBridgeError,
			expectedErr:       log.NewError("default bridge failed"),
		},
		"netmgr_test_init_host_net_failed": {
			mgrConfig:         defaultCfg,
			prepareMgrForTest: prepareInitHostNetFail,
			expectedErr:       log.NewErrorf("could not create host network: %v", log.NewErrorf("no host net")),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer func() {
				os.RemoveAll(testDirsRoot)
				controller.Finish()
			}()
			if dirsErr := util.MkDirs(testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot); dirsErr != nil {
				t.Fatalf("could not create the test directories meta and exec : %s, %s", testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot)
			}

			testMgr := testCase.prepareMgrForTest(controller, testCase.mgrConfig)
			err := testMgr.Initialize(context.Background())
			testutil.AssertError(t, testCase.expectedErr, err)

		})
	}

}

func TestReleaseNetworkResources(t *testing.T) {
	tests := map[string]struct {
		mgrConfig         *config
		container         *types.Container
		prepareMgrForTest prepare
		expectedErr       error
	}{
		"netmgr_test_rnr_nil_settings": {
			mgrConfig:         defaultCfg,
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareReleaseResourcesNilSettings,
		},
		"netmgr_test_rnr_default": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesFull,
		},
		"netmgr_test_rnr_err_getting_sandbox": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesGetSbErr,
			expectedErr:       log.NewError("error getting sandbox"),
		},
		"netmgr_test_rnr_err_sb_nil": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesSbNil,
			expectedErr:       log.NewErrorf("failed to get sandbox with id = %s", testCtrSandboxID),
		},
		"netmgr_test_rnr_err_eps_nil": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesEpsNil,
			expectedErr:       log.NewErrorf("no endpoints in sandbox with id = %s", testCtrSandboxID),
		},
		"netmgr_test_rnr_err_ep_missing": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesMissingEp,
			expectedErr:       log.NewErrorf("the endpoint %s is not connected", testCtrEndpointID),
		},
		"netmgr_test_rnr_err_ep_leave_sb_err": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesEpLeaveSbError,
			expectedErr:       log.NewErrorf("error while leaving network %s", testCtrEndpointID),
		},
		"netmgr_test_rnr_err_ep_delete_err": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesEpDeleteError,
			expectedErr:       log.NewErrorf("error while deleting endpoint with id = %s", testCtrEndpointID),
		},
		"netmgr_test_rnr_err_sb_delete_err": {
			mgrConfig:         defaultCfg,
			container:         newDefaultConnectedContainer(),
			prepareMgrForTest: prepareReleaseResourcesSbDeleteError,
			expectedErr:       log.NewError("error deleting sandbox"),
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer func() {
				os.RemoveAll(testDirsRoot)
				controller.Finish()
			}()
			if dirsErr := util.MkDirs(testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot); dirsErr != nil {
				t.Fatalf("could not create the test directories meta and exec : %s, %s", testCase.mgrConfig.metaPath, testCase.mgrConfig.execRoot)
			}

			testMgr := testCase.prepareMgrForTest(controller, testCase.mgrConfig, testCase.container)
			err := testMgr.ReleaseNetworkResources(context.Background(), testCase.container)
			testutil.AssertError(t, testCase.expectedErr, err)
			if err == nil {
				testutil.AssertNil(t, testCase.container.NetworkSettings)
			}
		})
	}
}

func TestDispose(t *testing.T) {
	t.Run("netmgr_test_disconnect", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockLibnetMgr := mocks.NewMockNetworkController(controller)
		testMgr := &libnetworkMgr{defaultCfg, mockLibnetMgr}

		mockLibnetMgr.EXPECT().Stop().Times(1)
		testMgr.Dispose(context.Background())
	})
}
func TestDisconnect(t *testing.T) {
	t.Run("netmgr_test_disconnect", func(t *testing.T) {
		testMgr := &libnetworkMgr{defaultCfg, nil}

		testutil.AssertNil(t, testMgr.Disconnect(context.Background(), newDefaultContainer(), true))
	})
}

func TestRestore(t *testing.T) {
	tests := map[string]struct {
		mgrConfig         *config
		containers        []*types.Container
		expectedMgrConfig *config
	}{
		"netmgr_test_restore_no_sbs_no_ctrs": {
			mgrConfig:         newDefaultMgrConfig(),
			expectedMgrConfig: newDefaultMgrConfig(),
		},
		"netmgr_test_restore_ctrs_no_net_settings": {
			mgrConfig:         newDefaultMgrConfig(),
			containers:        []*types.Container{newDefaultContainer()},
			expectedMgrConfig: newDefaultMgrConfig(),
		},
		"netmgr_test_restore_ctrs_with_net_settings": {
			mgrConfig: newDefaultMgrConfig(),
			containers: []*types.Container{{
				ID:             testCtrID,
				HostName:       "test-host",
				DomainName:     "test-domain",
				HostsPath:      "hosts",
				ResolvConfPath: "resolv.conf",
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				NetworkSettings: &types.NetworkSettings{
					Networks:            nil,
					SandboxID:           testCtrSandboxID,
					SandboxKey:          testCtrSandboxKey,
					NetworkControllerID: testNetworkControllerID,
				},
			}},
			expectedMgrConfig: &config{
				netType:  "bridge",
				metaPath: testDirsRoot + "/meta",
				execRoot: testDirsRoot + "/exec",
				bridgeConfig: bridgeConfig{
					name: "test0",
				},
				activeSandboxes: map[string]interface{}{
					testCtrSandboxID: []libnetwork.SandboxOption{}, /*here are exemplary as the real ones are asserted in a different test case*/
				},
			},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			defer os.RemoveAll(testDirsRoot)
			if mkdirErr := os.MkdirAll(filepath.Join(testCase.mgrConfig.execRoot, "libnetwork"), 0777); mkdirErr != nil {
				t.Fatal("failed to create libnetwork socket dir with the proper test permissions")
			}

			testNetMgr := &libnetworkMgr{
				config: testCase.mgrConfig,
			}
			err := testNetMgr.Restore(context.Background(), testCase.containers)
			testutil.AssertError(t, nil, err)
			_, dirErr := os.Stat(testCase.mgrConfig.execRoot)
			testutil.AssertError(t, nil, dirErr)
			_, dirErr = os.Stat(testCase.mgrConfig.metaPath)
			testutil.AssertError(t, nil, dirErr)

			testutil.AssertNotNil(t, testNetMgr.netController)
			if testCase.expectedMgrConfig.activeSandboxes != nil {
				testutil.AssertEqual(t, len(testCase.expectedMgrConfig.activeSandboxes), len(testCase.mgrConfig.activeSandboxes))
				for sbKey := range testCase.mgrConfig.activeSandboxes {
					testutil.AssertNotNil(t, testCase.mgrConfig.activeSandboxes[sbKey])
				}
			}
		})
	}
}

func TestStats(t *testing.T) {
	tests := map[string]struct {
		container         *types.Container
		prepareMgrForTest prepare
		expectedIOStats   *types.IOStats
		expectedErr       error
	}{
		"netmgr_test_stats_default": {
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareStatsDefault,
			expectedIOStats:   &types.IOStats{Read: 2048, Write: 4096},
		},
		"netmgr_test_stats_missing_sandbox": {
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareStatsMissingSb,
			expectedErr:       log.NewErrorf("no network sandbox for container %s ", testCtrID),
		},
		"netmgr_test_stats_statistics_error": {
			container:         newDefaultContainer(),
			prepareMgrForTest: prepareStatsErrorGettingStatistics,
			expectedErr:       log.NewError("error getting statistics"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)

			testMgr := testCase.prepareMgrForTest(controller, nil, testCase.container)
			ioStats, err := testMgr.Stats(context.Background(), testCase.container)
			testutil.AssertError(t, testCase.expectedErr, err)
			testutil.AssertEqual(t, testCase.expectedIOStats, ioStats)
		})
	}

}

func assertManagedContainer(t *testing.T, mgrConfig *config, container *types.Container) {
	// assert container
	ctrMetaPath := getContainerNetMetaPath(mgrConfig, container.ID)
	expectedHostName := container.HostName
	if util.IsContainerNetworkHost(container) {
		expectedHostName, _ = os.Hostname()
	}
	testutil.AssertEqual(t, filepath.Join(ctrMetaPath, "resolv.conf"), container.ResolvConfPath)
	testutil.AssertEqual(t, filepath.Join(ctrMetaPath, "hostname"), container.HostnamePath)
	testutil.AssertEqual(t, filepath.Join(ctrMetaPath, "hosts"), container.HostsPath)
	testutil.AssertEqual(t, expectedHostName, container.HostName)
}

func assertConnectedContainerFailed(t *testing.T, mgrConfig *config, container *types.Container) {
	// assert container
	testutil.AssertEqual(t, 0, len(container.NetworkSettings.Networks))
}

func assertConnectedContainer(t *testing.T, mgrConfig *config, container *types.Container) {
	// assert container
	ctrNetSettings := container.NetworkSettings
	testutil.AssertNotNil(t, ctrNetSettings)
	testutil.AssertEqual(t, testCtrSandboxID, ctrNetSettings.SandboxID)
	testutil.AssertEqual(t, testCtrSandboxKey, ctrNetSettings.SandboxKey)
	testutil.AssertEqual(t, testNetworkControllerID, ctrNetSettings.NetworkControllerID)
	testutil.AssertTrue(t, len(ctrNetSettings.Networks) >= 1)
	epSettings := ctrNetSettings.Networks[string(container.HostConfig.NetworkMode)]
	testutil.AssertNotNil(t, epSettings)
	testutil.AssertEqual(t, testCtrEndpointID, epSettings.ID)
	testutil.AssertEqual(t, netSettingsMac, epSettings.MacAddress)
	testutil.AssertEqual(t, netSettingsIPGW, epSettings.Gateway)
	testutil.AssertEqual(t, netSettingsIP, epSettings.IPAddress)
	testutil.AssertEqual(t, mgrConfig.bridgeConfig.name, epSettings.NetworkID)
}

func prepareNilCtrl(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	return &libnetworkMgr{config, nil}
}
func prepareDefault(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Times(1)
	mockLibnetMgr.EXPECT().NewSandbox(container.ID, gomock.Any()).Times(1).Return(mockSb, nil)
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareDefaultExistingSb(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockLibnetMgr.EXPECT().SandboxDestroy(container.ID).Times(1).Return(nil)
	mockLibnetMgr.EXPECT().NewSandbox(container.ID, gomock.Any()).Times(1).Return(mockSb, nil)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareStatsDefault(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockSb.EXPECT().Statistics().Times(1).Return(
		map[string]*libnetTypes.InterfaceStatistics{
			"test0": {RxBytes: 1024, TxBytes: 2048},
			"test1": {RxBytes: 1024, TxBytes: 2048},
		}, nil)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareStatsMissingSb(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Times(1)
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareStatsErrorGettingStatistics(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockSb.EXPECT().Statistics().Times(1).Return(nil, log.NewError("error getting statistics"))

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareDestroySbFailed(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockLibnetMgr.EXPECT().SandboxDestroy(container.ID).Times(1).Return(log.NewErrorf("failed to destroy container sandbox"))
	mockLibnetMgr.EXPECT().NewSandbox(container.ID, gomock.Any()).Times(0)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareNewSbFailed(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Times(1)
	mockLibnetMgr.EXPECT().NewSandbox(container.ID, gomock.Any()).Return(mockSb, log.NewErrorf("failed to create container sandbox")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareConnectErrorGettingNetwork(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(nil, log.NewErrorf("no network"))
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectErrorGettingSb(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Times(1)
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectErrorGettingEp(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockNetwork.EXPECT().Name().Times(4).Return("bridge")
	mockNetwork.EXPECT().EndpointByID(container.NetworkSettings.Networks["bridge"].ID).Times(1).Return(nil, log.NewErrorf("no endpoint"))
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectErrorGettingEpDelete(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockNetwork.EXPECT().Name().Times(4).Return("bridge")
	mockNetwork.EXPECT().EndpointByID(container.NetworkSettings.Networks["bridge"].ID).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Delete(true).Return(log.NewErrorf("error deleting endpoint")).Times(1)
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectErrorJoiningEp(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockNetwork.EXPECT().Name().Times(4).Return("bridge")
	mockNetwork.EXPECT().EndpointByID(container.NetworkSettings.Networks["bridge"].ID).Times(1).Return(nil, libnetwork.ErrNoSuchEndpoint(container.NetworkSettings.Networks["bridge"].ID))
	mockNetwork.EXPECT().CreateEndpoint(container.ID+"-ep", gomock.Any()).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Join(mockSb).Times(1).Return(log.NewErrorf("error joining endpoint"))
	mockEp.EXPECT().Delete(true).Times(1).Return(nil)
	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareConnectFullNoCtrNetworks(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)
	mockEpInfo := mocks.NewMockEndpointInfo(gomockCtrl)
	mockIFaceInfo := mocks.NewMockInterfaceInfo(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockNetwork.EXPECT().CreateEndpoint(container.ID+"-ep", gomock.Any()).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Join(mockSb).Return(nil).Times(1)

	// mock the conversion
	mockNetwork.EXPECT().ID().Return(config.bridgeConfig.name)
	mockEp.EXPECT().Info().Return(mockEpInfo).Times(1)
	mockEp.EXPECT().ID().Return(testCtrEndpointID).Times(1)
	ipGw := net.ParseIP(netSettingsIPGW)
	mockEpInfo.EXPECT().Gateway().Return(ipGw).Times(1)
	mockEpInfo.EXPECT().Iface().Return(mockIFaceInfo).Times(1)
	ip := net.ParseIP(netSettingsIP)
	mockIFaceInfo.EXPECT().Address().Return(&net.IPNet{IP: ip}).Times(1)
	mac, _ := net.ParseMAC(netSettingsMac)
	mockIFaceInfo.EXPECT().MacAddress().Return(mac).Times(1)

	mockSb.EXPECT().ID().Return(testCtrSandboxID).Times(1)
	mockSb.EXPECT().Key().Return(testCtrSandboxKey).Times(1)
	mockLibnetMgr.EXPECT().ID().Return(testNetworkControllerID).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareConnectFullWithOtherCtrNetworks(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)
	mockEpInfo := mocks.NewMockEndpointInfo(gomockCtrl)
	mockIFaceInfo := mocks.NewMockInterfaceInfo(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockLibnetMgr.EXPECT().ID().Return(testNetworkControllerID).Times(1)
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockSb.EXPECT().ID().Return(testCtrSandboxID).Times(1)
	mockSb.EXPECT().Key().Return(testCtrSandboxKey).Times(1)
	mockNetwork.EXPECT().Name().Return(bridgeNetworkName).Times(4)
	mockNetwork.EXPECT().EndpointByID(container.NetworkSettings.Networks["bridge"].ID).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Delete(true).Times(1)
	mockNetwork.EXPECT().CreateEndpoint(container.ID+"-ep", gomock.Any()).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Join(mockSb).Return(nil).Times(1)

	// mock the conversion
	mockNetwork.EXPECT().ID().Return(config.bridgeConfig.name)
	mockEp.EXPECT().Info().Return(mockEpInfo).Times(1)
	mockEp.EXPECT().ID().Return(testCtrEndpointID).Times(1)
	ipGw := net.ParseIP(netSettingsIPGW)
	mockEpInfo.EXPECT().Gateway().Return(ipGw).Times(1)
	mockEpInfo.EXPECT().Iface().Return(mockIFaceInfo).Times(1)
	ip := net.ParseIP(netSettingsIP)
	mockIFaceInfo.EXPECT().Address().Return(&net.IPNet{IP: ip}).Times(1)
	mac, _ := net.ParseMAC(netSettingsMac)
	mockIFaceInfo.EXPECT().MacAddress().Return(mac).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectFullWithNetSettings(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)
	mockEpInfo := mocks.NewMockEndpointInfo(gomockCtrl)
	mockIFaceInfo := mocks.NewMockInterfaceInfo(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockLibnetMgr.EXPECT().ID().Return(testNetworkControllerID).Times(1)
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockSb.EXPECT().ID().Return(testCtrSandboxID).Times(1)
	mockSb.EXPECT().Key().Return(testCtrSandboxKey).Times(1)
	mockNetwork.EXPECT().CreateEndpoint(container.ID+"-ep", gomock.Any()).Times(1).Return(mockEp, nil)
	mockEp.EXPECT().Join(mockSb).Return(nil).Times(1)

	// mock the conversion
	mockNetwork.EXPECT().ID().Return(config.bridgeConfig.name)
	mockEp.EXPECT().Info().Return(mockEpInfo).Times(1)
	mockEp.EXPECT().ID().Return(testCtrEndpointID).Times(1)
	ipGw := net.ParseIP(netSettingsIPGW)
	mockEpInfo.EXPECT().Gateway().Return(ipGw).Times(1)
	mockEpInfo.EXPECT().Iface().Return(mockIFaceInfo).Times(1)
	ip := net.ParseIP(netSettingsIP)
	mockIFaceInfo.EXPECT().Address().Return(&net.IPNet{IP: ip}).Times(1)
	mac, _ := net.ParseMAC(netSettingsMac)
	mockIFaceInfo.EXPECT().MacAddress().Return(mac).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareConnectErrorCreatingEp(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(string(container.HostConfig.NetworkMode)).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().WalkSandboxes(gomock.Any()).Do(func(walker libnetwork.SandboxWalker) {
		for _, sb := range mockLibnetMgr.Sandboxes() {
			if walker(sb) {
				return
			}
		}
	}).Times(1)
	mockLibnetMgr.EXPECT().Sandboxes().Times(1).Return([]libnetwork.Sandbox{mockSb})
	mockSb.EXPECT().ContainerID().Times(1).Return(container.ID)
	mockNetwork.EXPECT().Name().Times(4).Return("bridge")
	mockNetwork.EXPECT().EndpointByID(container.NetworkSettings.Networks["bridge"].ID).Times(1).Return(nil, libnetwork.ErrNoSuchEndpoint(container.NetworkSettings.Networks["bridge"].ID))
	mockNetwork.EXPECT().CreateEndpoint(container.ID+"-ep", gomock.Any()).Times(1).Return(nil, log.NewErrorf("error creating endpoint"))
	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareInitNoSbs(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockLibnetMgr.EXPECT().NetworkByName(config.bridgeConfig.name).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockNetwork.EXPECT().Name().Return(bridgeNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NewNetwork(config.netType, bridgeNetworkName, "", gomock.Any()).Times(1).Return(mockNetwork, nil)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareInitNoSbsErrorDeletingOldBridge(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(config.bridgeConfig.name).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(log.NewError("error deleting old bridge")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareInitNoSbsErrorDeletingNewBridge(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(config.bridgeConfig.name).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(log.NewError("error deleting default bridge")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareInitNoSbsErrorCreatingNewBridge(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(config.bridgeConfig.name).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockLibnetMgr.EXPECT().NewNetwork(config.netType, bridgeNetworkName, "", gomock.Any()).Return(nil, log.NewError("error creating default bridge")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareInitNoSbsNoExistingHostNet(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(nil, nil)
	mockLibnetMgr.EXPECT().NewNetwork(libnetworkDriverHost, hostNetworkName, "", gomock.Any()).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(config.bridgeConfig.name).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Delete().Return(nil).Times(1)
	mockNetwork.EXPECT().Name().Return(bridgeNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NewNetwork(config.netType, bridgeNetworkName, "", gomock.Any()).Times(1).Return(mockNetwork, nil)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareInitWithSbs(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(bridgeNetworkName).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareInitWithSbsDefaultBridgeError(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockNetwork := mocks.NewMockNetwork(gomockCtrl)

	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(mockNetwork, nil)
	mockNetwork.EXPECT().Name().Return(hostNetworkName).Times(1)
	mockLibnetMgr.EXPECT().NetworkByName(bridgeNetworkName).Times(1).Return(nil, log.NewError("default bridge failed"))

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareInitHostNetFail(gomockCtrl *gomock.Controller, config *config) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockLibnetMgr.EXPECT().NetworkByName(hostNetworkName).Times(1).Return(nil, nil)
	mockLibnetMgr.EXPECT().NewNetwork(libnetworkDriverHost, hostNetworkName, "", gomock.Any()).Return(nil, log.NewErrorf("no host net"))

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesNilSettings(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	return &libnetworkMgr{config, nil}
}
func prepareReleaseResourcesFull(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockEp.EXPECT().ID().Return(container.NetworkSettings.Networks[config.netType].ID).Times(1)
	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{mockEp}).Times(1)
	mockEp.EXPECT().Leave(mockSb).Return(nil).Times(1)
	mockEp.EXPECT().Delete(false).Return(nil).Times(1)

	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{}).Times(1)
	mockSb.EXPECT().Delete().Return(nil).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesGetSbErr(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(nil, log.NewError("error getting sandbox")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}

func prepareReleaseResourcesSbNil(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(nil, nil).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesEpsNil(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockSb.EXPECT().Endpoints().Return(nil).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesMissingEp(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockEp.EXPECT().ID().Return("random").Times(1)
	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{mockEp}).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesEpLeaveSbError(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockEp.EXPECT().ID().Return(container.NetworkSettings.Networks[config.netType].ID).Times(1)
	mockEp.EXPECT().Name().Return(testCtrEndpointID).Times(1)
	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{mockEp}).Times(1)
	mockEp.EXPECT().Leave(mockSb).Return(log.NewError("error leaving sandbox")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesEpDeleteError(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockEp.EXPECT().ID().Return(container.NetworkSettings.Networks[config.netType].ID).Times(2)
	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{mockEp}).Times(1)
	mockEp.EXPECT().Leave(mockSb).Return(nil).Times(1)
	mockEp.EXPECT().Delete(false).Return(log.NewError("error deleting endpoint")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
func prepareReleaseResourcesSbDeleteError(gomockCtrl *gomock.Controller, config *config, container *types.Container) ContainerNetworkManager {
	mockLibnetMgr := mocks.NewMockNetworkController(gomockCtrl)
	mockSb := mocks.NewMockSandbox(gomockCtrl)
	mockEp := mocks.NewMockEndpoint(gomockCtrl)

	mockLibnetMgr.EXPECT().SandboxByID(container.NetworkSettings.SandboxID).Return(mockSb, nil).Times(1)

	mockEp.EXPECT().ID().Return(container.NetworkSettings.Networks[config.netType].ID).Times(1)
	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{mockEp}).Times(1)
	mockEp.EXPECT().Leave(mockSb).Return(nil).Times(1)
	mockEp.EXPECT().Delete(false).Return(nil).Times(1)

	mockSb.EXPECT().Endpoints().Return([]libnetwork.Endpoint{}).Times(1)
	mockSb.EXPECT().Delete().Return(log.NewError("error deleting sandbox")).Times(1)

	return &libnetworkMgr{config, mockLibnetMgr}
}
