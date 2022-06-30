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

package util

import (
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestContainerValidation(t *testing.T) {
	ctr := &types.Container{
		ID:         containerID,
		Name:       name,
		Image:      internalImage,
		DomainName: domain,
		HostName:   host,
		Mounts:     internalMounts,
		Hooks:      internalHooks,
		Config:     &internalContainerConfig,
		HostConfig: internalHostConfig,
		IOConfig:   internalIOConfig,
	}

	t.Run("test_validate_container", func(t *testing.T) {
		err := ValidateContainer(ctr)
		if err != nil {
			t.Log(err)
			t.Error("unexpected validation error")
		}
	})
}

func TestNegativeContainerValidations(t *testing.T) {
	tests := map[string]struct {
		ctr         *types.Container
		expectedErr error
	}{
		"test_validate_invalid_image": {
			ctr:         &types.Container{},
			expectedErr: log.NewError("the containers image configuration is invalid"),
		},
		"test_validate_defaults": {
			ctr: func() *types.Container {
				ctrWithDefaults := &types.Container{
					Image: types.Image{Name: "image"},
				}
				FillDefaults(ctrWithDefaults)
				return ctrWithDefaults
			}(),
			expectedErr: nil,
		},
		"test_validate_invalid_name": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				Name:  ".invalid_name",
			},
			expectedErr: log.NewErrorf("invalid container name format : %s", ".invalid_name"),
		},
		"test_validate_mounts_invalid_mount_src": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				Mounts: []types.MountPoint{{
					Destination:     mountDest,
					Source:          "",
					PropagationMode: mountPropagationMode,
				}},
			},
			expectedErr: log.NewErrorf("source and destination must be set for a mount point"),
		},
		"test_validate_mounts_invalid_mount_dst": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				Mounts: []types.MountPoint{{
					Destination:     "",
					Source:          mountSrc,
					PropagationMode: mountPropagationMode,
				}},
			},
			expectedErr: log.NewErrorf("source and destination must be set for a mount point"),
		},
		"test_validate_mounts_invalid_mount_dst_and_src": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				Mounts: []types.MountPoint{{
					Destination:     "",
					Source:          "",
					PropagationMode: mountPropagationMode,
				}},
			},
			expectedErr: log.NewErrorf("source and destination must be set for a mount point"),
		},
		"test_validate_mounts_invalid_mount_propagation_mode": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				Mounts: []types.MountPoint{{
					Destination:     mountDest,
					Source:          mountSrc,
					PropagationMode: "invalid_mode",
				}},
			},
			expectedErr: log.NewErrorf("propagation mode must be set to one of the supported modes"),
		},
		"test_validate_host_config_nil": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
			},
			expectedErr: log.NewErrorf("the containers host config is mandatory and is missing"),
		},
		"test_validate_host_config_device_mappings_invalid_host_path": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Devices: []types.DeviceMapping{{
						PathOnHost:        "",
						PathInContainer:   hostConfigDeviceContainer,
						CgroupPermissions: hostConfigDevicePerm,
					}},
				},
			},
			expectedErr: log.NewErrorf("both path on the host and in the container must be specified for a device mapping"),
		},
		"test_validate_host_config_device_mappings_invalid_container_path": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Devices: []types.DeviceMapping{{
						PathOnHost:        hostConfigDeviceHost,
						PathInContainer:   "",
						CgroupPermissions: hostConfigDevicePerm,
					}},
				},
			},
			expectedErr: log.NewErrorf("both path on the host and in the container must be specified for a device mapping"),
		},
		"test_validate_host_config_device_mappings_invalid_cgroup_permissions": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Devices: []types.DeviceMapping{{
						PathOnHost:        hostConfigDeviceHost,
						PathInContainer:   hostConfigDeviceContainer,
						CgroupPermissions: "",
					}},
				},
			},
			expectedErr: log.NewErrorf("the cgroup permissions for device mapping %s:%s are not provided", hostConfigDeviceHost, hostConfigDeviceContainer),
		},
		"test_validate_host_config_log_config_invalid_max_size_invalid": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type:     types.LogConfigDriverJSONFile,
							MaxFiles: hostConfigLogConfigMaxFiles,
							MaxSize:  "asdf",
						},
						ModeConfig: &types.LogModeConfiguration{
							Mode: hostConfigLogConfigMode,
						},
					},
				},
			},
			expectedErr: log.NewErrorf("invalid format of max logs size - %s", "asdf"),
		},
		"test_validate_host_config_log_config_invalid_max_size_empty": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type:     types.LogConfigDriverJSONFile,
							MaxFiles: hostConfigLogConfigMaxFiles,
							MaxSize:  "",
						},
						ModeConfig: &types.LogModeConfiguration{
							Mode: hostConfigLogConfigMode,
						},
					},
				},
			},
			expectedErr: log.NewError("max logs size must be set"),
		},
		"test_validate_host_config_log_config_invalid_max_files": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type:     types.LogConfigDriverJSONFile,
							MaxFiles: -1,
							MaxSize:  hostConfigLogConfigMaxSize,
						},
						ModeConfig: &types.LogModeConfiguration{
							Mode: hostConfigLogConfigMode,
						},
					},
				},
			},
			expectedErr: log.NewErrorf("max log files cannot be < 1"),
		},
		"test_validate_host_config_log_config_empty_buffer_size": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type:     types.LogConfigDriverNone,
							MaxFiles: hostConfigLogConfigMaxFiles,
							MaxSize:  hostConfigLogConfigMaxSize,
						},
						ModeConfig: &types.LogModeConfiguration{
							Mode:          hostConfigLogConfigMode,
							MaxBufferSize: "",
						},
					},
				},
			},
			expectedErr: log.NewErrorf("max buffer size must be set"),
		},
		"test_validate_host_config_log_config_invalid_buffer_size": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type:     types.LogConfigDriverNone,
							MaxFiles: hostConfigLogConfigMaxFiles,
							MaxSize:  hostConfigLogConfigMaxSize,
						},
						ModeConfig: &types.LogModeConfiguration{
							Mode:          hostConfigLogConfigMode,
							MaxBufferSize: "adf",
						},
					},
				},
			},
			expectedErr: log.NewErrorf("invalid format of max buffer size - %s", "adf"),
		},
		"test_validate_host_config_io_config_nil": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
				},
				IOConfig: nil,
			},
			expectedErr: log.NewError("the container's IO config is missing"),
		},
		"test_validate_host_config_host_mode_with_ports": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeHost,
					PortMappings: []types.PortMapping{{
						Proto:         "tcp",
						ContainerPort: 80,
						HostIP:        "0.0.0.0",
						HostPort:      80,
						HostPortEnd:   80,
					}},
				},
			},
			expectedErr: log.NewError("cannot use port mappings when in host network mode"),
		},
		"test_validate_host_config_host_mode_with_key_used": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeHost,
					ExtraHosts:  []string{"ctrhost:host_ip"},
				},
			},
			expectedErr: log.NewError("cannot use the host_ip reserved key or any of its modifications when in host network mode"),
		},
		"test_validate_host_config_host_mode_with_key_and_net_if_used": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeHost,
					ExtraHosts:  []string{"ctrhost:host_ip_eth0"},
				},
			},
			expectedErr: log.NewError("cannot use the host_ip reserved key or any of its modifications when in host network mode"),
		},
		"test_validate_host_config_host_mode_unsupported": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: "custom",
				},
			},
			expectedErr: log.NewError("unsupported network mode custom"),
		},
		"test_validate_host_config_invalid_restart_policy_type": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					RestartPolicy: &types.RestartPolicy{
						Type: types.PolicyType("unsupported"),
					},
				},
			},
			expectedErr: log.NewErrorf("unsupported restart policy type %s", types.PolicyType("unsupported")),
		},
		"test_validate_host_config_restart_policy_negative_retry_timeout": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					RestartPolicy: &types.RestartPolicy{
						Type:              types.OnFailure,
						MaximumRetryCount: hostConfigRestartPolicyMaxRetry,
						RetryTimeout:      -1 * time.Second,
					},
				},
			},
			expectedErr: log.NewError("restart policy retry timeout cannot be negative"),
		},
		"test_validate_host_config_restart_policy_negative_max_retry_count": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					RestartPolicy: &types.RestartPolicy{
						Type:              types.OnFailure,
						MaximumRetryCount: -3,
						RetryTimeout:      hostConfigRestartPolicyTimeout,
					},
				},
			},
			expectedErr: log.NewError("restart policy max retry count cannot be negative"),
		},
		"test_validate_host_config_restart_policy_always_with_max_retry_count": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					RestartPolicy: &types.RestartPolicy{
						Type:              types.Always,
						MaximumRetryCount: hostConfigRestartPolicyMaxRetry,
					},
				},
			},
			expectedErr: log.NewErrorf("cannot use max retry count when the restart policy is %s", types.Always),
		},
		"test_validate_host_config_resources_memory_swap_is_less_than_limit": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						Memory:     "200M",
						MemorySwap: "199M",
					},
				},
			},
			expectedErr: log.NewError("swap memory - 199M is less than memory - 200M"),
		},
		"test_validate_host_config_resources_memory_is_less_than_reservation": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						Memory:            "200M",
						MemoryReservation: "300M",
					},
				},
			},
			expectedErr: log.NewError("reservation memory - 300M must be lower than memory - 200M"),
		},
		"test_validate_host_config_resources_memory_swap_is_set_without_limit": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						MemorySwap: hostConfigResourcesMemorySwap,
					},
				},
			},
			expectedErr: log.NewErrorf("swap memory - %s, memory must be set as well", hostConfigResourcesMemorySwap),
		},
		"test_validate_host_config_resources_memory_invalid": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						Memory: "invalid_value",
					},
				},
			},
			expectedErr: log.NewError("invalid format of memory - invalid_value"),
		},
		"test_validate_host_config_resources_memory_reservation_invalid": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						MemoryReservation: "invalid_value",
					},
				},
			},
			expectedErr: log.NewError("invalid format of memory reservation - invalid_value"),
		},
		"test_validate_host_config_resources_memory_swap_invalid": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						MemorySwap: "invalid_value",
					},
				},
			},
			expectedErr: log.NewError("invalid format of swap memory - invalid_value"),
		},
		"test_validate_host_config_resources_memory_too_low": {
			ctr: &types.Container{
				Image: types.Image{Name: "image"},
				HostConfig: &types.HostConfig{
					NetworkMode: types.NetworkModeBridge,
					Resources: &types.Resources{
						Memory:     "2.8M",
						MemorySwap: "-1",
					},
				},
			},
			expectedErr: log.NewErrorf("minimum memory allowed is 3M"),
		},
		"test_validate_config_env_incorrect_format_start_digit": {
			ctr: &types.Container{
				Image:      types.Image{Name: "image"},
				HostConfig: &types.HostConfig{NetworkMode: types.NetworkModeBridge},
				Config: &types.ContainerConfiguration{
					Env: []string{"1VAR=1"},
				},
			},
			expectedErr: log.NewErrorf("invalid environmental variable declaration provided : 1VAR=1"),
		},
		"test_validate_config_env_incorrect_format_start_other": {
			ctr: &types.Container{
				Image:      types.Image{Name: "image"},
				HostConfig: &types.HostConfig{NetworkMode: types.NetworkModeBridge},
				Config: &types.ContainerConfiguration{
					Env: []string{"$VAR=1"},
				},
			},
			expectedErr: log.NewErrorf("invalid environmental variable declaration provided : $VAR=1"),
		},
		"test_validate_config_env_incorrect_format_contains_at_sign": {
			ctr: &types.Container{
				Image:      types.Image{Name: "image"},
				HostConfig: &types.HostConfig{NetworkMode: types.NetworkModeBridge},
				Config: &types.ContainerConfiguration{
					Env: []string{"V@R=1"},
				},
			},
			expectedErr: log.NewErrorf("invalid environmental variable declaration provided : V@R=1"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			err := ValidateContainer(testCase.ctr)
			testutil.AssertError(t, testCase.expectedErr, err)
		})
	}
}
