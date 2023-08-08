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

package util

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func createContainerWithImage(name string) *types.Container {
	return &types.Container{Image: types.Image{Name: name}}
}

func createContainerWithMounts(source string) *types.Container {
	return &types.Container{Mounts: []types.MountPoint{{Destination: "testDestination1", Source: "testSource1", PropagationMode: "private"},
		{Destination: "testDestination1", Source: source, PropagationMode: "private"}}}
}

func createContainerWithConfig(cmd []string) *types.Container {
	return &types.Container{Config: &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: cmd}}
}

func createContainerWithIOConfig(openstdin bool) *types.Container {
	return &types.Container{IOConfig: &types.IOConfig{AttachStderr: true, AttachStdin: true, AttachStdout: true, OpenStdin: openstdin, StdinOnce: true, Tty: true}}
}

func createContainerWithHostConfig(hostConfig *types.HostConfig) *types.Container {
	return &types.Container{HostConfig: hostConfig}
}

func copyHostConfig(source *types.HostConfig) *types.HostConfig {
	return &types.HostConfig{
		Devices:           source.Devices,
		NetworkMode:       source.NetworkMode,
		Privileged:        source.Privileged,
		RestartPolicy:     source.RestartPolicy,
		Runtime:           source.Runtime,
		ExtraHosts:        source.ExtraHosts,
		ExtraCapabilities: source.ExtraCapabilities,
		PortMappings:      source.PortMappings,
		LogConfig:         source.LogConfig,
		Resources:         source.Resources,
	}
}

func TestDetermineUpdateAction(t *testing.T) {
	testCases := map[string]struct {
		current        *types.Container
		desired        *types.Container
		expectedResult ActionType
	}{
		"test_current_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: ActionCreate,
		},
		"test_image_name_equal": {
			current:        createContainerWithImage("name1"),
			desired:        createContainerWithImage("name1"),
			expectedResult: ActionCheck,
		},
		"test_image_name_not_equal": {
			current:        createContainerWithImage("name1"),
			desired:        createContainerWithImage("name2"),
			expectedResult: ActionRecreate,
		},
		"test_mounts_equal": {
			current:        createContainerWithMounts("testSource1"),
			desired:        createContainerWithMounts("testSource1"),
			expectedResult: ActionCheck,
		},
		"test_mounts_not_equal": {
			current:        createContainerWithMounts("testSource1"),
			desired:        createContainerWithMounts("notequal"),
			expectedResult: ActionRecreate,
		},
		"test_container_config_equal": {
			current:        createContainerWithConfig([]string{"testCmd"}),
			desired:        createContainerWithConfig([]string{"testCmd"}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_empty_and_nil_not_equal": {
			current:        createContainerWithConfig([]string{}),
			desired:        createContainerWithConfig(nil),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_elements_empty_not_equal": {
			current:        createContainerWithConfig([]string{""}),
			desired:        createContainerWithConfig([]string{""}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_elements_nil_not_equal": {
			current:        createContainerWithConfig(nil),
			desired:        createContainerWithConfig(nil),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_one_empty_element_and_one_not_empty_not_equal": {
			current:        createContainerWithConfig([]string{"testCmd"}),
			desired:        createContainerWithConfig(nil),
			expectedResult: ActionRecreate,
		},
		"test_container_config_not_equal": {
			current: &types.Container{
				Config: &types.ContainerConfiguration{Env: []string{"testEnv1", "testEnv2"}, Cmd: []string{"testCmd1", "testCmd2"}},
			},
			desired: &types.Container{
				Config: &types.ContainerConfiguration{Env: []string{"testEnv1", "notequal"}, Cmd: []string{"testCmd1", "testCmd2"}},
			},
			expectedResult: ActionRecreate,
		},
		"test_ioconfig_equal": {
			current:        createContainerWithIOConfig(true),
			desired:        createContainerWithIOConfig(true),
			expectedResult: ActionCheck,
		},
		"test_ioconfig_not_equal": {
			current:        createContainerWithIOConfig(false),
			desired:        createContainerWithIOConfig(true),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig0_equal_privileged": {
			current:        createContainerWithHostConfig(&types.HostConfig{Privileged: true}),
			desired:        createContainerWithHostConfig(&types.HostConfig{Privileged: true}),
			expectedResult: ActionCheck,
		},
		"test_hostconfig0_not_equal_privileged": {
			current:        createContainerWithHostConfig(&types.HostConfig{Privileged: true}),
			desired:        createContainerWithHostConfig(&types.HostConfig{Privileged: false}),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig0_equal_capabilities": {
			current:        createContainerWithHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			desired:        createContainerWithHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			expectedResult: ActionCheck,
		},
		"test_hostconfig0_not_equal_capabilities": {
			current:        createContainerWithHostConfig(&types.HostConfig{ExtraCapabilities: []string{"test"}}),
			desired:        createContainerWithHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig1_equal": {
			current: createContainerWithHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "3m", MemorySwap: "-1",
				},
			}),
			desired: createContainerWithHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "3m", MemorySwap: "-1",
				},
			}),
			expectedResult: ActionCheck,
		},
		"test_hostconfig1_not_equal": {
			current: createContainerWithHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "3m", MemorySwap: "-1",
				},
			}),
			desired: createContainerWithHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "34m", MemorySwap: "-1",
				},
			}),
			expectedResult: ActionUpdate,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, DetermineUpdateAction(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualImage(t *testing.T) {
	t.Run("test_image_equal", func(t *testing.T) {
		current := types.Image{Name: "name"}
		desired := types.Image{Name: "name"}

		res := isEqualImage(current, desired)

		assert.True(t, res)
	})

	t.Run("test_image_not_equal", func(t *testing.T) {
		current := types.Image{Name: "name1"}
		desired := types.Image{Name: "name2"}

		res := isEqualImage(current, desired)

		assert.False(t, res)
	})
}

func TestIsEqualContainerConfig(t *testing.T) {
	defaultContainerConfig := &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}}
	testCases := map[string]struct {
		current        *types.ContainerConfiguration
		desired        *types.ContainerConfiguration
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.ContainerConfiguration{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.ContainerConfiguration{},
		},
		"test_current_desired_equal": {
			current:        defaultContainerConfig,
			desired:        defaultContainerConfig,
			expectedResult: true,
		},
		"test_env_not_equal": {
			current: defaultContainerConfig,
			desired: &types.ContainerConfiguration{Env: []string{"testNotEqual"}, Cmd: []string{"testCmd"}},
		},
		"test_env_equal_ordering_not_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv", "testEnv2"}, Cmd: []string{"testCmd"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv2", "testEnv"}, Cmd: []string{"testCmd"}},
			expectedResult: true,
		},
		"test_cmd_not_equal": {
			current: defaultContainerConfig,
			desired: &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testNotEqual"}},
		},
		"test_cmd_equal_ordering_not_equal": {
			current: &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd", "testCmd2"}},
			desired: &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd2", "testCmd"}},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualContainerConfig(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualHostConfig0(t *testing.T) {
	testCases := map[string]struct {
		current        *types.HostConfig
		desired        *types.HostConfig
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.HostConfig{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.HostConfig{},
		},
		"test_all_equal": {
			current: internalHostConfig,
			desired: func(hostConfig *types.HostConfig) *types.HostConfig {
				return copyHostConfig(hostConfig)
			}(internalHostConfig),
			expectedResult: true,
		},
		"test_privileged_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.Privileged = true
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_networkmode_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.NetworkMode = "true"
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_devices_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.Devices = append(copy.Devices, types.DeviceMapping{PathOnHost: "testPathOnHost", PathInContainer: "testPathInContainer", CgroupPermissions: "rwm"})
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_extracapabilities_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.ExtraCapabilities = []string{"testExtraCapabilities"}
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_extrahosts_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.ExtraCapabilities = []string{"testExtraHosts"}
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_portmappings_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.PortMappings = append(copy.PortMappings, types.PortMapping{Proto: "tcp", ContainerPort: 80, HostPort: 80, HostIP: "0.0.0.0", HostPortEnd: 80})
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_logconfig_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.LogConfig = &types.LogConfiguration{
					DriverConfig: &types.LogDriverConfiguration{
						Type:     types.LogConfigDriverJSONFile,
						MaxFiles: 1,
					},
				}
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualHostConfig0(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualHostConfig1(t *testing.T) {
	testCases := map[string]struct {
		current        *types.HostConfig
		desired        *types.HostConfig
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.HostConfig{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.HostConfig{},
		},
		"test_resources_equal": {
			current: internalHostConfig,
			desired: func(hostConfig *types.HostConfig) *types.HostConfig {
				return copyHostConfig(hostConfig)
			}(internalHostConfig),
			expectedResult: true,
		},
		"test_resources_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.Resources = &types.Resources{
					Memory: "10M",
				}
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
		"test_restartpolicy_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.RestartPolicy = &types.RestartPolicy{
					Type: "unless-stopped",
				}
				return copy
			}(copyHostConfig(internalHostConfig)),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualHostConfig1(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualResources(t *testing.T) {
	defaultResources := &types.Resources{
		Memory:            "4m",
		MemoryReservation: "3m",
		MemorySwap:        "-1",
	}
	testCases := map[string]struct {
		current        *types.Resources
		desired        *types.Resources
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.Resources{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.Resources{},
		},
		"test_resources_equal": {
			current:        defaultResources,
			desired:        defaultResources,
			expectedResult: true,
		},
		"test_resources_not_equal": {
			current: defaultResources,
			desired: &types.Resources{
				Memory:            "4m",
				MemoryReservation: "3m",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualResources(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualRestartPolicy(t *testing.T) {
	defaultRestartPolicy := &types.RestartPolicy{
		MaximumRetryCount: 0,
		RetryTimeout:      0,
		Type:              "always",
	}
	testCases := map[string]struct {
		current        *types.RestartPolicy
		desired        *types.RestartPolicy
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.RestartPolicy{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.RestartPolicy{},
		},
		"test_restartpolicy_equal": {
			current:        defaultRestartPolicy,
			desired:        defaultRestartPolicy,
			expectedResult: true,
		},
		"test_restartpolicy_not_equal": {
			current: defaultRestartPolicy,
			desired: &types.RestartPolicy{
				RetryTimeout: 5,
				Type:         "always",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualRestartPolicy(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualLog(t *testing.T) {
	defaultDriverConfig := &types.LogDriverConfiguration{
		Type:     "json-file",
		MaxFiles: 1,
		MaxSize:  "100M",
		RootDir:  "testRootDir",
	}
	defaultModeConfig := &types.LogModeConfiguration{
		Mode:          "non-blocking",
		MaxBufferSize: "1m",
	}
	testCases := map[string]struct {
		current        *types.LogConfiguration
		desired        *types.LogConfiguration
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.LogConfiguration{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.LogConfiguration{},
		},
		"test_current_driverconfig_nil_desired_driverconfig_not_nil": {
			current: &types.LogConfiguration{
				DriverConfig: nil,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   &types.LogModeConfiguration{},
			},
		},
		"test_current_driverconfig_nil_desired_driverconfig_nil": {
			current: &types.LogConfiguration{
				DriverConfig: nil,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: nil,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			expectedResult: true,
		},
		"test_current_driverconfig_not_nil_desired_driverconfig_nil": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: nil,
				ModeConfig:   &types.LogModeConfiguration{},
			},
		},
		"test_current_driverconfig_not_nil_desired_driverconfig_not_nil_equal": {
			current: &types.LogConfiguration{
				DriverConfig: defaultDriverConfig,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: defaultDriverConfig,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			expectedResult: true,
		},
		"test_current_driverconfig_not_nil_desired_driverconfig_not_nil_not_equal": {
			current: &types.LogConfiguration{
				DriverConfig: defaultDriverConfig,
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type: "json-file",
				},
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
		"test_current_modeconfig_nil_desired_modeconfig_not_nil": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   nil,
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   &types.LogModeConfiguration{},
			},
		},
		"test_current_modeconfig_nil_desired_modeconfig_nil": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   nil,
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   nil,
			},
			expectedResult: true,
		},
		"test_current_modeconfig_not_nil_desired_modeconfig_nil": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   nil,
			},
		},
		"test_current_modeconfig_not_nil_desired_modeconfig_not_nil_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   defaultModeConfig,
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   defaultModeConfig,
			},
			expectedResult: true,
		},
		"test_current_modeconfig_not_nil_desired_modeconfig_not_nil_not_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig:   defaultModeConfig,
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          "non-blocking",
					MaxBufferSize: "10000m",
				},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualLog(testCase.current, testCase.desired))
		})
	}
}

func TestIsEqualIOConfig(t *testing.T) {
	setIOConfig := func(attachStderr, openstdin, tty bool) *types.IOConfig {
		return &types.IOConfig{AttachStderr: attachStderr, OpenStdin: openstdin, Tty: tty}
	}
	testCases := map[string]struct {
		current        *types.IOConfig
		desired        *types.IOConfig
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			desired: &types.IOConfig{},
		},
		"test_current_not_nil_desired_nil": {
			current: &types.IOConfig{},
		},
		"test_IOConfig_equal": {
			current:        setIOConfig(true, true, true),
			desired:        setIOConfig(true, true, true),
			expectedResult: true,
		},
		"test_IOConfig_equal_AttachStderr_not_equal": {
			current:        setIOConfig(false, true, true),
			desired:        setIOConfig(true, true, true),
			expectedResult: true,
		},
		"test_IOConfig_not_equal": {
			current: setIOConfig(false, false, false),
			desired: setIOConfig(true, true, true),
		},
		"test_IOConfig_openstdin_not_equal": {
			current: setIOConfig(false, false, false),
			desired: setIOConfig(true, true, false),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.expectedResult, isEqualIOConfig(testCase.current, testCase.desired))
		})
	}
}

func TestCompareSliceSet(t *testing.T) {
	testCases := map[string]struct {
		slice1 interface{}
		slice2 interface{}
		match  bool
	}{
		"test_equal_same_order": {
			slice1: []string{"a", "b", "c"},
			slice2: []string{"a", "b", "c"},
			match:  true,
		},
		"test_equal_mixed_order": {
			slice1: []string{"a", "b", "c"},
			slice2: []string{"b", "a", "c"},
			match:  true,
		},
		"test_equal_duplicates": {
			slice1: []string{"x", "x", "y"},
			slice2: []string{"y", "x", "x"},
			match:  true,
		},
		"test_equal_duplicates_diff_count": {
			slice1: []string{"x", "x", "y"},
			slice2: []string{"y", "y", "x"},
			match:  true,
		},
		"test_unequal": {
			slice1: []string{"x", "x", "y"},
			slice2: []string{"x", "y", "z"},
			match:  false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, testCase.match, compareSliceSet(testCase.slice1, testCase.slice2))
			testutil.AssertEqual(t, testCase.match, compareSliceSet(testCase.slice2, testCase.slice1))
		})
	}
}
