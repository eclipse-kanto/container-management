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

package util

import (
	"time"

	internaltypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
)

const (
	containerID = "test-id"
	imageName   = "image.url"
	name        = "name"
	domain      = "domain"
	host        = "host"

	mountSrc             = "/proc"
	mountDest            = "/proc"
	mountPropagationMode = string(internaltypes.RPrivatePropagationMode)

	hookPath    = "hookPath"
	hookArg1    = "arg1"
	hookEnv1    = "env1"
	hookTimeout = 10000
	hookType    = internaltypes.HookTypePoststart

	configEnv1 = "VAR1"
	configEnv2 = "VAR2="
	configEnv3 = "VAR3=test"
	configEnv4 = "VAR4=\"test string\""
	configEnv5 = "VAR5=test,comma"
	configEnv6 = "_VAR6=test_underscore"

	hostConfigPrivileged                 = false
	hostConfigNetType                    = "bridge"
	hostConfigContainerPort              = 80
	hostConfigHostPort                   = 81
	hostConfigHostPortEnd                = 82
	hostConfigHostIP                     = "192.168.1.101"
	hostConfigDeviceHost                 = "/dev/ttyACM0"
	hostConfigDeviceContainer            = "/dev/ttyACM1"
	hostConfigDevicePerm                 = "rwm"
	hostConfigRestartPolicyMaxRetry      = 5
	hostConfigRestartPolicyTimeout       = time.Duration(30) * time.Second
	hostConfigRestartPolicyType          = internaltypes.OnFailure
	hostConfigLogConfigDriverType        = internaltypes.LogConfigDriverJSONFile
	hostConfigLogConfigMaxFiles          = 2
	hostConfigLogConfigMaxSize           = "100M"
	hostConfigLogConfigBufferSize        = "5M"
	hostConfigLogConfigMode              = internaltypes.LogModeNonBlocking
	hostConfigRuntime                    = "some-runtime-config"
	hostConfigResourcesMemory            = "200M"
	hostConfigResourcesMemoryReservation = "150M"
	hostConfigResourcesMemorySwap        = "500M"

	networkName              = "name"
	networkSettingID         = "containerId"
	networkSettingGateway    = "192.168.150.150"
	networkSettingIPAddress  = "192.168.1.101"
	networkSettingMacAddress = "aa:bb:00:11:22:33"
	networkSettingNetworkID  = "kanto-cm0"
)

var (
	internalImage  = internaltypes.Image{Name: imageName}
	internalMounts = []internaltypes.MountPoint{{
		Destination:     mountDest,
		Source:          mountSrc,
		PropagationMode: mountPropagationMode,
	}}

	hookArgs = append([]string{}, hookArg1)
	hookEnv  = append([]string{}, hookEnv1)

	internalHooks = []internaltypes.Hook{{
		Path:    hookPath,
		Args:    hookArgs,
		Env:     hookEnv,
		Timeout: hookTimeout,
		Type:    hookType,
	}}

	configEnv               = []string{configEnv1, configEnv2, configEnv3, configEnv4, configEnv5, configEnv6}
	internalContainerConfig = internaltypes.ContainerConfiguration{Env: configEnv}

	hostConfigExtraHosts        = []string{"ctrhost:host_ip"}
	hostConfigExtraCapabilities = []string{"CAP_NET_ADMIN"}
	internalHostConfig          = &internaltypes.HostConfig{
		Privileged:        hostConfigPrivileged,
		ExtraHosts:        hostConfigExtraHosts,
		ExtraCapabilities: hostConfigExtraCapabilities,
		NetworkMode:       hostConfigNetType,
		PortMappings: []internaltypes.PortMapping{{
			ContainerPort: hostConfigContainerPort,
			HostPort:      hostConfigHostPort,
			HostIP:        hostConfigHostIP,
			HostPortEnd:   hostConfigHostPortEnd,
		}},
		Devices: []internaltypes.DeviceMapping{{
			PathOnHost:        hostConfigDeviceHost,
			PathInContainer:   hostConfigDeviceContainer,
			CgroupPermissions: hostConfigDevicePerm,
		}},
		RestartPolicy: &internaltypes.RestartPolicy{
			MaximumRetryCount: hostConfigRestartPolicyMaxRetry,
			RetryTimeout:      hostConfigRestartPolicyTimeout,
			Type:              hostConfigRestartPolicyType,
		},
		LogConfig: &internaltypes.LogConfiguration{
			DriverConfig: &internaltypes.LogDriverConfiguration{
				Type:     hostConfigLogConfigDriverType,
				MaxFiles: hostConfigLogConfigMaxFiles,
				MaxSize:  hostConfigLogConfigMaxSize,
			},
			ModeConfig: &internaltypes.LogModeConfiguration{
				Mode:          hostConfigLogConfigMode,
				MaxBufferSize: hostConfigLogConfigBufferSize,
			},
		},
		Resources: &internaltypes.Resources{
			Memory:            hostConfigResourcesMemory,
			MemoryReservation: hostConfigResourcesMemoryReservation,
			MemorySwap:        hostConfigResourcesMemorySwap,
		},
		Runtime: hostConfigRuntime,
	}

	internalIOConfig = &internaltypes.IOConfig{
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		OpenStdin:    true,
		StdinOnce:    true,
	}

	internalNetworkSettings = &internaltypes.NetworkSettings{
		SandboxID:  "sandboxId",
		SandboxKey: "sandboxKey",
		Networks: map[string]*internaltypes.EndpointSettings{
			networkName: {
				ID:         networkSettingID,
				Gateway:    networkSettingGateway,
				IPAddress:  networkSettingIPAddress,
				MacAddress: networkSettingMacAddress,
				NetworkID:  networkSettingNetworkID,
			},
		}}
)
