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

package things

import (
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	containerID          = "test-id"
	imageName            = "image.url"
	name                 = "name"
	domain               = "domain"
	hostName             = "hostname"
	env                  = "test-env"
	cmd                  = "test-cmd"
	mountSrc             = "/proc"
	mountDest            = "/proc"
	mountPropagationMode = string(types.RPrivatePropagationMode)

	hostConfigPrivileged                 = true
	hostConfigNetType                    = types.NetworkModeHost
	hostConfigContainerPort              = 80
	hostConfigHostPort                   = 81
	hostConfigHostPortEnd                = 82
	hostConfigHostIP                     = "192.168.1.101"
	hostConfigDeviceHost                 = "/dev/ttyACM0"
	hostConfigDeviceContainer            = "/dev/ttyACM1"
	hostConfigDevicePerm                 = "rwm"
	hostConfigRestartPolicyMaxRetry      = 5
	hostConfigRestartPolicyTimeout       = time.Duration(30) * time.Second
	hostConfigRestartPolicyType          = types.OnFailure
	hostConfigLogConfigDriverType        = types.LogConfigDriverJSONFile
	hostConfigLogConfigMaxFiles          = 2
	hostConfigLogConfigMaxSize           = "100M"
	hostConfigLogConfigBufferSize        = "5M"
	hostConfigLogConfigMode              = types.LogModeNonBlocking
	hostConfigRuntime                    = "some-runtime-config"
	hostConfigResourcesMemory            = "500M"
	hostConfigResourcesMemoryReservation = "300M"
	hostConfigResourcesMemorySwap        = "1G"
)

var (
	internalImage  = types.Image{Name: imageName, DecryptConfig: &types.DecryptConfig{}}
	internalMounts = []types.MountPoint{{
		Destination:     mountDest,
		Source:          mountSrc,
		PropagationMode: mountPropagationMode,
	}}
	envVar                      = []string{env}
	cmdVar                      = []string{cmd}
	hostConfigExtraHosts        = []string{"ctrhost:host_ip"}
	hostConfigExtraCapabilities = []string{"CAP_NET_ADMIN"}
	internalHostConfig          = &types.HostConfig{
		Privileged:        hostConfigPrivileged,
		ExtraHosts:        hostConfigExtraHosts,
		ExtraCapabilities: hostConfigExtraCapabilities,
		NetworkMode:       hostConfigNetType,
		PortMappings: []types.PortMapping{{
			ContainerPort: hostConfigContainerPort,
			HostPort:      hostConfigHostPort,
			HostIP:        hostConfigHostIP,
			HostPortEnd:   hostConfigHostPortEnd,
		}},
		Devices: []types.DeviceMapping{{
			PathOnHost:        hostConfigDeviceHost,
			PathInContainer:   hostConfigDeviceContainer,
			CgroupPermissions: hostConfigDevicePerm,
		}},
		RestartPolicy: &types.RestartPolicy{
			MaximumRetryCount: hostConfigRestartPolicyMaxRetry,
			RetryTimeout:      hostConfigRestartPolicyTimeout,
			Type:              hostConfigRestartPolicyType,
		},
		LogConfig: &types.LogConfiguration{
			DriverConfig: &types.LogDriverConfiguration{
				Type:     hostConfigLogConfigDriverType,
				MaxFiles: hostConfigLogConfigMaxFiles,
				MaxSize:  hostConfigLogConfigMaxSize,
			},
			ModeConfig: &types.LogModeConfiguration{
				Mode:          hostConfigLogConfigMode,
				MaxBufferSize: hostConfigLogConfigBufferSize,
			},
		},
		Runtime: hostConfigRuntime,
		Resources: &types.Resources{
			Memory:            hostConfigResourcesMemory,
			MemoryReservation: hostConfigResourcesMemoryReservation,
			MemorySwap:        hostConfigResourcesMemorySwap,
		},
	}

	internalIOConfig = &types.IOConfig{
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		OpenStdin:    true,
		StdinOnce:    true,
	}

	internalContainerConfig = &types.ContainerConfiguration{
		Env: envVar,
		Cmd: cmdVar,
	}
)

func TestFromAPIContainerConfig(t *testing.T) {
	ctr := &types.Container{
		ID:         containerID,
		Name:       name,
		Image:      internalImage,
		DomainName: domain,
		HostName:   hostName,
		Mounts:     internalMounts,
		HostConfig: internalHostConfig,
		IOConfig:   internalIOConfig,
		Config:     internalContainerConfig,
	}

	ctrParsed := fromAPIContainerConfig(ctr)
	t.Run("test_from_api_container_config_domain_name", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.DomainName, ctrParsed.DomainName)
	})
	t.Run("test_from_api_container_config_mounts_len", func(t *testing.T) {
		testutil.AssertEqual(t, len(ctr.Mounts), len(ctrParsed.MountPoints))
	})
	t.Run("test_from_api_container_config_mounts", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.Mounts[0], toAPIMountPoint(ctrParsed.MountPoints[0]))
	})
	t.Run("test_from_api_container_config_devices", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.Devices[0], toAPIDevice(ctrParsed.Devices[0]))
	})
	t.Run("test_from_api_container_config_privileged", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.Privileged, ctrParsed.Privileged)
	})
	t.Run("test_from_api_container_config_restart_policy", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.RestartPolicy, toAPIRestartPolicy(ctrParsed.RestartPolicy))
	})
	t.Run("test_from_api_container_config_extra_caps", func(t *testing.T) {
		ctr.HostConfig.Privileged = false
		ctrParsed = fromAPIContainerConfig(ctr)
		testutil.AssertEqual(t, ctr.HostConfig.ExtraCapabilities, ctrParsed.ExtraCapabilities)
	})
	t.Run("test_from_api_container_config_extra_hosts", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.ExtraHosts, ctrParsed.ExtraHosts)
	})
	t.Run("test_from_api_container_config_extra_port_mappings_len", func(t *testing.T) {
		testutil.AssertEqual(t, len(ctr.HostConfig.PortMappings), len(ctrParsed.PortMappings))
	})
	t.Run("test_from_api_container_config_extra_port_mappings", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.PortMappings[0], toAPIPortMapping(ctrParsed.PortMappings[0]))
	})
	t.Run("test_from_api_container_config_open_stdin", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.IOConfig.OpenStdin, ctrParsed.OpenStdin)
	})
	t.Run("test_from_api_container_config_tty", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.IOConfig.Tty, ctrParsed.Tty)
	})
	t.Run("test_from_api_container_config_log", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.LogConfig, toAPILogConfiguration(ctrParsed.Log))
	})
	t.Run("test_from_api_container_config_env", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.Config.Env, ctrParsed.Env)
	})
	t.Run("test_from_api_container_config_cmd", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.Config.Cmd, ctrParsed.Cmd)
	})
	t.Run("test_from_api_container_config_resources", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.Resources, toAPIResources(ctrParsed.Resources))
	})
	t.Run("test_from_api_container_config_host_name", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostName, ctrParsed.HostName)
	})
	t.Run("test_from_api_container_decrypt_config", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.Image.DecryptConfig, toAPIDecryptionConfiguration(ctrParsed.Decryption))
	})
	t.Run("test_from_api_container_config_networkMode", func(t *testing.T) {
		testutil.AssertEqual(t, ctr.HostConfig.NetworkMode, ctrParsed.NetworkMode.toAPINetworkMode())
	})
}

var (
	testContainerConfig = &configuration{
		DomainName: domain,
		MountPoints: []*mountPoint{{
			Destination:     mountPointDestination,
			Source:          mountPointSource,
			PropagationMode: rprivate,
		}},
		HostName:   hostName,
		Env:        envVar,
		Cmd:        cmdVar,
		Decryption: &decryptionConfiguration{},
		Devices:    []*device{{}},
		Privileged: hostConfigPrivileged,
		RestartPolicy: &restartPolicy{
			MaxRetryCount: hostConfigRestartPolicyMaxRetry,
			RetryTimeout:  hostConfigRestartPolicyTimeout.Seconds(),
			RpType:        onFailure,
		},
		NetworkMode:       host,
		ExtraCapabilities: hostConfigExtraCapabilities,
		ExtraHosts:        hostConfigExtraHosts,
		PortMappings:      []*portMapping{{}},
		OpenStdin:         internalIOConfig.OpenStdin,
		Tty:               internalIOConfig.Tty,
		Log: &logConfiguration{
			Type:          testLogDriverType,
			MaxFiles:      testLogMaxFiles,
			MaxSize:       testLogMaxSize,
			Mode:          testLogMode,
			MaxBufferSize: testLogBufferSize,
		},
		Resources: &resources{
			Memory:            testMemory,
			MemoryReservation: testMemoryReservation,
			MemorySwap:        testMemorySwap,
		},
	}
)

func TestToAPIContainerConfig(t *testing.T) {
	ctrParsed := toAPIContainerConfig(testContainerConfig)
	t.Run("test_to_api_container_config_domain_name", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.DomainName, ctrParsed.DomainName)
	})
	t.Run("test_to_api_container_config_mounts_len", func(t *testing.T) {
		testutil.AssertEqual(t, len(testContainerConfig.MountPoints), len(ctrParsed.Mounts))
	})
	t.Run("test_to_api_container_config_mounts", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.MountPoints[0], fromAPIMountPoint(ctrParsed.Mounts[0]))
	})
	t.Run("test_to_api_container_config_devices", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Devices[0], fromAPIDevice(ctrParsed.HostConfig.Devices[0]))
	})
	t.Run("test_to_api_container_config_privileged", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Privileged, ctrParsed.HostConfig.Privileged)
	})
	t.Run("test_to_api_container_config_restart_policy", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.RestartPolicy, fromAPIRestartPolicy(ctrParsed.HostConfig.RestartPolicy))
	})
	t.Run("test_to_api_container_config_extra_caps", func(t *testing.T) {
		copyTestContainerConfig := *testContainerConfig
		copyTestContainerConfig.Privileged = false
		ctrParsedExtraCapabilities := toAPIContainerConfig(&copyTestContainerConfig)
		testutil.AssertEqual(t, copyTestContainerConfig.ExtraCapabilities, ctrParsedExtraCapabilities.HostConfig.ExtraCapabilities)
	})
	t.Run("test_to_api_container_config_extra_hosts", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.ExtraHosts, ctrParsed.HostConfig.ExtraHosts)
	})
	t.Run("test_to_api_container_config_extra_port_mappings_len", func(t *testing.T) {
		testutil.AssertEqual(t, len(testContainerConfig.PortMappings), len(ctrParsed.HostConfig.PortMappings))
	})
	t.Run("test_to_api_container_config_extra_port_mappings", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.PortMappings[0], fromAPIPortMapping(ctrParsed.HostConfig.PortMappings[0]))
	})
	t.Run("test_to_api_container_config_open_stdin", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.OpenStdin, ctrParsed.IOConfig.OpenStdin)
	})
	t.Run("test_to_api_container_config_tty", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Tty, ctrParsed.IOConfig.Tty)
	})
	t.Run("test_to_api_container_config_log", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Log, fromAPILogConfiguration(ctrParsed.HostConfig.LogConfig))
	})
	t.Run("test_to_api_container_config_env", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Env, ctrParsed.Config.Env)
	})
	t.Run("test_to_api_container_config_cmd", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Cmd, ctrParsed.Config.Cmd)
	})
	t.Run("test_to_api_container_config_resources", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Resources, fromAPIResources(ctrParsed.HostConfig.Resources))
	})
	t.Run("test_to_api_container_config_host_name", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.HostName, ctrParsed.HostName)
	})
	t.Run("test_to_api_container_decrypt_config", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.Decryption, fromAPIDecryptionConfiguration(ctrParsed.Image.DecryptConfig))
	})
	t.Run("test_to_api_container_config_networkMode", func(t *testing.T) {
		testutil.AssertEqual(t, testContainerConfig.NetworkMode, fromAPINetworkMode(ctrParsed.HostConfig.NetworkMode))
	})
}
