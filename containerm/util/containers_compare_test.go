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

func setupContainerImage(Image types.Image) *types.Container {
	return &types.Container{Image: Image}
}

func setupContainerMounts(Mounts []types.MountPoint) *types.Container {
	return &types.Container{Mounts: Mounts}
}

func setupContainerConfig(config *types.ContainerConfiguration) *types.Container {
	return &types.Container{Config: config}
}

func setupContainerIOConfig(ioConfig *types.IOConfig) *types.Container {
	return &types.Container{IOConfig: ioConfig}
}

func setupContainerHostConfig(hostConfig *types.HostConfig) *types.Container {
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
			current:        setupContainerImage(types.Image{Name: "name1"}),
			desired:        setupContainerImage(types.Image{Name: "name1"}),
			expectedResult: ActionCheck,
		},
		"test_image_name_not_equal": {
			current:        setupContainerImage(types.Image{Name: "name1"}),
			desired:        setupContainerImage(types.Image{Name: "name2"}),
			expectedResult: ActionRecreate,
		},
		"test_mounts_equal": {
			current: setupContainerMounts([]types.MountPoint{{Destination: "testDestination1", Source: "testSource1", PropagationMode: "private"},
				{Destination: "testDestination1", Source: "testSource1", PropagationMode: "private"}}),
			desired: setupContainerMounts([]types.MountPoint{{Destination: "testDestination1", Source: "testSource1", PropagationMode: "private"},
				{Destination: "testDestination1", Source: "testSource1", PropagationMode: "private"}}),
			expectedResult: ActionCheck,
		},
		"test_mounts_not_equal": {
			current: setupContainerMounts([]types.MountPoint{{Destination: "testDestination", Source: "testSource", PropagationMode: "private"},
				{Destination: "testDestination", Source: "testSource", PropagationMode: "private"}}),
			desired: setupContainerMounts([]types.MountPoint{{Destination: "testDestination", Source: "testSource", PropagationMode: "private"},
				{Destination: "testDestination", Source: "notequal", PropagationMode: "private"}}),
			expectedResult: ActionRecreate,
		},
		"test_container_config_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_empty_and_nil_not_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_elements_empty_not_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{}}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_elements_nil_not_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}}),
			expectedResult: ActionCheck,
		},
		"test_container_config_cmd_one_empty_element_and_one_not_empty_not_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{}}),
			expectedResult: ActionRecreate,
		},
		"test_container_config_not_equal": {
			current:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv1", "testEnv2"}, Cmd: []string{"testCmd1", "testCmd2"}}),
			desired:        setupContainerConfig(&types.ContainerConfiguration{Env: []string{"testEnv1", "notequal"}, Cmd: []string{"testCmd1", "testCmd2"}}),
			expectedResult: ActionRecreate,
		},
		"test_ioconfig_equal": {
			current:        setupContainerIOConfig(&types.IOConfig{AttachStderr: true, AttachStdin: true, AttachStdout: true, OpenStdin: true, StdinOnce: true, Tty: true}),
			desired:        setupContainerIOConfig(&types.IOConfig{AttachStderr: true, AttachStdin: true, AttachStdout: true, OpenStdin: true, StdinOnce: true, Tty: true}),
			expectedResult: ActionCheck,
		},
		"test_ioconfig_not_equal": {
			current: setupContainerIOConfig(&types.IOConfig{
				AttachStderr: true, AttachStdin: true, AttachStdout: true, OpenStdin: false, StdinOnce: true, Tty: true,
			}),
			desired: setupContainerIOConfig(&types.IOConfig{
				AttachStderr: true, AttachStdin: true, AttachStdout: true, OpenStdin: true, StdinOnce: true, Tty: true,
			}),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig0_equal_privileged": {
			current:        setupContainerHostConfig(&types.HostConfig{Privileged: true}),
			desired:        setupContainerHostConfig(&types.HostConfig{Privileged: true}),
			expectedResult: ActionCheck,
		},
		"test_hostconfig0_not_equal_privileged": {
			current:        setupContainerHostConfig(&types.HostConfig{Privileged: true}),
			desired:        setupContainerHostConfig(&types.HostConfig{Privileged: false}),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig0_equal_capabilities": {
			current:        setupContainerHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			desired:        setupContainerHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			expectedResult: ActionCheck,
		},
		"test_hostconfig0_not_equal_capabilities": {
			current:        setupContainerHostConfig(&types.HostConfig{ExtraCapabilities: []string{"test"}}),
			desired:        setupContainerHostConfig(&types.HostConfig{ExtraCapabilities: []string{"CAP_NET_ADMIN"}}),
			expectedResult: ActionRecreate,
		},
		"test_hostconfig1_equal": {
			current: setupContainerHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "3m", MemorySwap: "-1",
				},
			}),
			desired: setupContainerHostConfig(&types.HostConfig{
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
			current: setupContainerHostConfig(&types.HostConfig{
				RestartPolicy: &types.RestartPolicy{
					MaximumRetryCount: 0, RetryTimeout: 0, Type: "always",
				},
				Resources: &types.Resources{
					Memory: "4m", MemoryReservation: "3m", MemorySwap: "-1",
				},
			}),
			desired: setupContainerHostConfig(&types.HostConfig{
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
			res := DetermineUpdateAction(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
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
	testCases := map[string]struct {
		current        *types.ContainerConfiguration
		desired        *types.ContainerConfiguration
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.ContainerConfiguration{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.ContainerConfiguration{},
			desired:        nil,
			expectedResult: false,
		},
		"test_current_desired_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}},
			expectedResult: true,
		},
		"test_env_not_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv2"}, Cmd: []string{"testCmd"}},
			expectedResult: false,
		},
		"test_env_equal_ordering_not_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv", "testEnv2"}, Cmd: []string{"testCmd"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv2", "testEnv"}, Cmd: []string{"testCmd"}},
			expectedResult: true,
		},
		"test_cmd_not_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testNotEqual"}},
			expectedResult: false,
		},
		"test_cmd_equal_ordering_not_equal": {
			current:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd", "testCmd2"}},
			desired:        &types.ContainerConfiguration{Env: []string{"testEnv"}, Cmd: []string{"testCmd2", "testCmd"}},
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualContainerConfig(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
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
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.HostConfig{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.HostConfig{},
			desired:        nil,
			expectedResult: false,
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
			expectedResult: false,
		},
		"test_networkmode_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.NetworkMode = "true"
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
		},
		"test_devices_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.Devices = append(copy.Devices, types.DeviceMapping{PathOnHost: "testPathOnHost", PathInContainer: "testPathInContainer", CgroupPermissions: "rwm"})
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
		},
		"test_extracapabilities_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.ExtraCapabilities = []string{"testExtraCapabilities"}
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
		},
		"test_extrahosts_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.ExtraCapabilities = []string{"testExtraHosts"}
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
		},
		"test_portmappings_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.PortMappings = append(copy.PortMappings, types.PortMapping{Proto: "tcp", ContainerPort: 80, HostPort: 80, HostIP: "0.0.0.0", HostPortEnd: 80})
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
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
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualHostConfig0(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
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
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.HostConfig{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.HostConfig{},
			desired:        nil,
			expectedResult: false,
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
			expectedResult: false,
		},
		"test_restartpolicy_not_equal": {
			current: internalHostConfig,
			desired: func(copy *types.HostConfig) *types.HostConfig {
				copy.RestartPolicy = &types.RestartPolicy{
					Type: "unless-stopped",
				}
				return copy
			}(copyHostConfig(internalHostConfig)),
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualHostConfig1(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
		})
	}
}

func TestIsEqualResources(t *testing.T) {
	testCases := map[string]struct {
		current        *types.Resources
		desired        *types.Resources
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.Resources{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.Resources{},
			desired:        nil,
			expectedResult: false,
		},
		"test_resources_equal": {
			current: &types.Resources{
				Memory:            "4m",
				MemoryReservation: "3m",
				MemorySwap:        "-1",
			},
			desired: &types.Resources{
				Memory:            "4m",
				MemoryReservation: "3m",
				MemorySwap:        "-1",
			},
			expectedResult: true,
		},
		"test_resources_not_equal": {
			current: &types.Resources{
				Memory:            "4m",
				MemoryReservation: "3m",
				MemorySwap:        "-1",
			},
			desired: &types.Resources{
				Memory:            "4m",
				MemoryReservation: "3m",
			},
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualResources(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
		})
	}
}

func TestIsEqualRestartPolicy(t *testing.T) {
	testCases := map[string]struct {
		current        *types.RestartPolicy
		desired        *types.RestartPolicy
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.RestartPolicy{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.RestartPolicy{},
			desired:        nil,
			expectedResult: false,
		},
		"test_restartpolicy_equal": {
			current: &types.RestartPolicy{
				MaximumRetryCount: 0,
				RetryTimeout:      0,
				Type:              "always",
			},

			desired: &types.RestartPolicy{
				MaximumRetryCount: 0,
				RetryTimeout:      0,
				Type:              "always",
			},
			expectedResult: true,
		},
		"test_restartpolicy_not_equal": {
			current: &types.RestartPolicy{
				RetryTimeout: 0,
				Type:         "always",
			},
			desired: &types.RestartPolicy{
				RetryTimeout: 5,
				Type:         "always",
			},
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualRestartPolicy(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
		})
	}
}

func TestIsEqualLog(t *testing.T) {
	testCases := map[string]struct {
		current        *types.LogConfiguration
		desired        *types.LogConfiguration
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.LogConfiguration{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.LogConfiguration{},
			desired:        nil,
			expectedResult: false,
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
			expectedResult: false,
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
			expectedResult: false,
		},
		"test_current_driverconfig_not_nil_desired_driverconfig_not_nil_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type:     "json-file",
					MaxFiles: 1,
					MaxSize:  "100M",
					RootDir:  "testRootDir",
				},
				ModeConfig: &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type:     "json-file",
					MaxFiles: 1,
					MaxSize:  "100M",
					RootDir:  "testRootDir",
				},
				ModeConfig: &types.LogModeConfiguration{},
			},
			expectedResult: true,
		},
		"test_current_driverconfig_not_nil_desired_driverconfig_not_nil_not_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type:     "json-file",
					MaxFiles: 1,
					MaxSize:  "100M",
					RootDir:  "testRootDir",
				},
				ModeConfig: &types.LogModeConfiguration{},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type: "json-file",
				},
				ModeConfig: &types.LogModeConfiguration{},
			},
			expectedResult: false,
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
			expectedResult: false,
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
			expectedResult: false,
		},
		"test_current_modeconfig_not_nil_desired_modeconfig_not_nil_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          "non-blocking",
					MaxBufferSize: "1m",
				},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          "non-blocking",
					MaxBufferSize: "1m",
				},
			},
			expectedResult: true,
		},
		"test_current_modeconfig_not_nil_desired_modeconfig_not_nil_not_equal": {
			current: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          "non-blocking",
					MaxBufferSize: "1m",
				},
			},
			desired: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          "non-blocking",
					MaxBufferSize: "10000m",
				},
			},
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualLog(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
		})
	}
}

func TestIsEqualIOConfig(t *testing.T) {
	testCases := map[string]struct {
		current        *types.IOConfig
		desired        *types.IOConfig
		expectedResult bool
	}{
		"test_current_nil_desired_nil": {
			current:        nil,
			desired:        nil,
			expectedResult: true,
		},
		"test_current_nil_desired_not_nil": {
			current:        nil,
			desired:        &types.IOConfig{},
			expectedResult: false,
		},
		"test_current_not_nil_desired_nil": {
			current:        &types.IOConfig{},
			desired:        nil,
			expectedResult: false,
		},
		"test_IOConfig_equal_AttachStderr_not_equal": {
			current: &types.IOConfig{
				AttachStderr: false,

				OpenStdin: true,
				Tty:       true,
			},
			desired: &types.IOConfig{
				AttachStderr: true,

				OpenStdin: true,
				Tty:       true,
			},
			expectedResult: true,
		},
		"test_IOConfig_not_equal": {
			current: &types.IOConfig{
				AttachStderr: false,

				OpenStdin: false,
				Tty:       false,
			},
			desired: &types.IOConfig{
				AttachStderr: true,

				OpenStdin: true,
				Tty:       true,
			},
			expectedResult: false,
		},
		"test_IOConfig_openstdin_not_equal": {
			current: &types.IOConfig{
				AttachStderr: false,

				OpenStdin: false,
				Tty:       false,
			},
			desired: &types.IOConfig{
				AttachStderr: true,

				OpenStdin: true,
				Tty:       false,
			},
			expectedResult: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := isEqualIOConfig(testCase.current, testCase.desired)
			testutil.AssertEqual(t, testCase.expectedResult, res)
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
