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

package protobuf

import (
	"testing"
	"time"

	internaltypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	sysinfointernaltypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

/*
	The const and vars declared for this test package are mirrored from util/util_base_test.go

in order to avoid redundant publicly available fields and to retain the package structure as is.
*/
const (
	id        = "test-id"
	imageName = "image.url"
	name      = "name"
	domain    = "domain"
	host      = "host"

	mountSrc             = "/proc"
	mountDest            = "/proc"
	mountPropagationMode = string(internaltypes.RPrivatePropagationMode)

	hookPath    = "hookPath"
	hookArg1    = "arg1"
	hookEnv1    = "env1"
	hookTimeout = 10000
	hookType    = internaltypes.HookTypePoststart

	configEnv1 = "env1"

	hostConfigPrivileged                 = true
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
	hostConfigLogConfigMode              = internaltypes.LogModeBlocking
	hostConfigRuntime                    = "some-runtime-config"
	hostConfigResourcesMemory            = "200M"
	hostConfigResourcesMemoryReservation = "150M"
	hostConfigResourcesMemorySwap        = "500M"

	networkName              = "name"
	networkSettingID         = "testContainerId"
	networkSettingGateway    = "192.168.150.150"
	networkSettingIPAddress  = "192.168.1.101"
	networkSettingMacAddress = "aa:bb:00:11:22:33"
	networkSettingNetworkID  = "kanto-cm0"

	manuallyStopped = true

	restartCount = 10
	key          = "testKey"
	decRecipient = "testRecipient"
)

var (
	internalImage = internaltypes.Image{
		Name: imageName,
	}
	internalImageWithDecryptConfig = internaltypes.Image{
		Name: imageName,
		DecryptConfig: &internaltypes.DecryptConfig{
			Keys:       []string{key},
			Recipients: []string{decRecipient},
		},
	}
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

	configEnv               = []string{configEnv1}
	configArg               = []string{"echo", "test", "command"}
	internalContainerConfig = internaltypes.ContainerConfiguration{
		Env: configEnv,
		Cmd: configArg,
	}

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
				Mode: hostConfigLogConfigMode,
			},
		},
		Runtime: hostConfigRuntime,
		Resources: &internaltypes.Resources{
			Memory:            hostConfigResourcesMemory,
			MemoryReservation: hostConfigResourcesMemoryReservation,
			MemorySwap:        hostConfigResourcesMemorySwap,
		},
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

// Tests conversions for container fields, that are expected to be mapped 1:1 (all info exposed)
func TestConvertContainer(t *testing.T) {
	ctr := &internaltypes.Container{
		ID:         id,
		Name:       name,
		Image:      internalImageWithDecryptConfig,
		DomainName: domain,
		HostName:   host,
		Mounts:     internalMounts,
		Hooks:      internalHooks,
		Config:     &internalContainerConfig,
		HostConfig: internalHostConfig,
		IOConfig:   internalIOConfig,
		// not exposed:
		//NetworkSettings: internalNetworkSettings,
		ManuallyStopped: manuallyStopped,
		RestartCount:    restartCount,
		// not exposed:
		//StartedSuccessfullyBefore: true,
		State: &internaltypes.State{},
	}
	util.SetContainerStatusCreated(ctr)

	t.Run("test_convert_container", func(t *testing.T) {
		testutil.AssertEqual(t, ctr, ToInternalContainer(ToProtoContainer(ctr)))
	})
}

func TestConvertContainerEmpty(t *testing.T) {
	ctr := &internaltypes.Container{
		Image: internalImage,
	}
	t.Run("test_convert_container_empty", func(t *testing.T) {
		testutil.AssertEqual(t, ctr, ToInternalContainer(ToProtoContainer(ctr)))
	})
}

// Tests conversions for container fields, that are not expected to be mapped 1:1 (not all info exposed)
func TestConvertContainerHidden(t *testing.T) {
	ctr := &internaltypes.Container{
		Image:                     internalImage,
		NetworkSettings:           internalNetworkSettings,
		StartedSuccessfullyBefore: true,
		State:                     &internaltypes.State{},
	}
	util.SetContainerStatusCreated(ctr)

	t.Run("test_convert_container_hidden", func(t *testing.T) {
		protoCtr := ToProtoContainer(ctr)
		internalCtr := ToInternalContainer(protoCtr)
		if ctr.NetworkSettings.Networks == nil || len(ctr.NetworkSettings.Networks) != 1 {
			t.Errorf("container networks not converted correctly: %+v, but was: %+v", ctr.NetworkSettings.Networks, internalCtr.NetworkSettings.Networks)
		}
		testutil.AssertEqual(t, ctr.NetworkSettings.Networks, internalCtr.NetworkSettings.Networks)
		if ctr.StartedSuccessfullyBefore == internalCtr.StartedSuccessfullyBefore {
			t.Errorf("container StartedSuccessfullyBefore not expected to be set")
		}
	})
}

func TestToInternalStatus(t *testing.T) {
	tests := map[string]struct {
		grpcStatus string
		expected   internaltypes.Status
	}{
		"test_convert_status_creating": {
			grpcStatus: internaltypes.Creating.String(),
			expected:   internaltypes.Creating,
		},
		"test_convert_status_created": {
			grpcStatus: internaltypes.Created.String(),
			expected:   internaltypes.Created,
		},
		"test_convert_status_running": {
			grpcStatus: internaltypes.Running.String(),
			expected:   internaltypes.Running,
		},
		"test_convert_status_stopped": {
			grpcStatus: internaltypes.Stopped.String(),
			expected:   internaltypes.Stopped,
		},
		"test_convert_status_paused": {
			grpcStatus: internaltypes.Paused.String(),
			expected:   internaltypes.Paused,
		},
		"test_convert_status_exited": {
			grpcStatus: internaltypes.Exited.String(),
			expected:   internaltypes.Exited,
		},
		"test_convert_status_dead": {
			grpcStatus: internaltypes.Dead.String(),
			expected:   internaltypes.Dead,
		},
		"test_convert_status_unknown": {
			grpcStatus: internaltypes.Unknown.String(),
			expected:   internaltypes.Unknown,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := ToInternalStatus(testCase.grpcStatus)
			if actual.String() != testCase.expected.String() {
				t.Errorf("status not converted correctly: %s, but was %s:", testCase.expected.String(), actual.String())
			}
		})
	}
}

func TestToInternalHookType(t *testing.T) {
	tests := map[string]struct {
		grpcHookType string
		expected     internaltypes.HookType
	}{
		"test_convert_hook_type_prestart": {
			grpcHookType: internaltypes.HookTypePrestart.String(),
			expected:     internaltypes.HookTypePrestart,
		},
		"test_convert_hook_type_poststart": {
			grpcHookType: internaltypes.HookTypePoststart.String(),
			expected:     internaltypes.HookTypePoststart,
		},
		"test_convert_hook_type_poststop": {
			grpcHookType: internaltypes.HookTypePoststop.String(),
			expected:     internaltypes.HookTypePoststop,
		},
		"test_convert_hook_type_unknown": {
			grpcHookType: "some-unknown",
			expected:     internaltypes.HookTypeUnknown,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := ToInternalHookType(testCase.grpcHookType)
			if actual.String() != testCase.expected.String() {
				t.Errorf("status not converted correctly: %s, but was %s:", testCase.expected.String(), actual.String())
			}
		})
	}
}

func TestToInternalStopOpts(t *testing.T) {
	stopOpts := &internaltypes.StopOpts{
		Timeout: 20,
		Force:   true,
		Signal:  "SIGTERM",
	}

	t.Run("test_convert_stop_options", func(t *testing.T) {
		testutil.AssertEqual(t, stopOpts, ToInternalStopOptions(ToProtoStopOptions(stopOpts)))
	})

	t.Run("test_convert_stop_options_nil", func(t *testing.T) {
		testutil.AssertNil(t, ToInternalStopOptions(ToProtoStopOptions(nil)))
	})
}

func TestToInternalProjectInfo(t *testing.T) {
	projectInfo := sysinfointernaltypes.ProjectInfo{
		ProjectVersion: "prj-version",
		BuildTime:      "build-time",
		APIVersion:     "api-version",
		GitCommit:      "git-commit",
	}

	t.Run("test_convert_projet_info", func(t *testing.T) {
		testutil.AssertEqual(t, projectInfo, ToInternalProjectInfo(ToProtoProjectInfo(projectInfo)))
	})
}

func TestToInternalUpdateOpts(t *testing.T) {
	updateOpts := &internaltypes.UpdateOpts{
		RestartPolicy: &internaltypes.RestartPolicy{
			MaximumRetryCount: hostConfigRestartPolicyMaxRetry,
			RetryTimeout:      hostConfigRestartPolicyTimeout,
			Type:              hostConfigRestartPolicyType,
		},
		Resources: &internaltypes.Resources{
			Memory:            hostConfigResourcesMemory,
			MemoryReservation: hostConfigResourcesMemoryReservation,
			MemorySwap:        hostConfigResourcesMemorySwap,
		},
	}

	t.Run("test_convert_update_options", func(t *testing.T) {
		testutil.AssertEqual(t, updateOpts, ToInternalUpdateOptions(ToProtoUpdateOptions(updateOpts)))
	})

	t.Run("test_convert_update_options_nil", func(t *testing.T) {
		testutil.AssertEqual(t, &internaltypes.UpdateOpts{RestartPolicy: nil, Resources: nil}, ToInternalUpdateOptions(ToProtoUpdateOptions(nil)))
	})
}
