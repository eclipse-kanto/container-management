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
	"time"

	apitypescontainers "github.com/eclipse-kanto/container-management/containerm/api/types/containers"
	apitypessysinfo "github.com/eclipse-kanto/container-management/containerm/api/types/sysinfo"
	internaltypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	sysinfointernaltypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
)

// ToInternalContainer converts a types.Container instance to an internal Container one
func ToInternalContainer(grpcContainer *apitypescontainers.Container) *internaltypes.Container {
	if grpcContainer == nil {
		return nil
	}
	var (
		mounts []internaltypes.MountPoint
		hooks  []internaltypes.Hook
	)

	if grpcContainer.Mounts != nil {
		mounts = []internaltypes.MountPoint{}
		for _, mnt := range grpcContainer.Mounts {
			mounts = append(mounts, *ToInternalMountPoint(mnt))
		}
	}

	if grpcContainer.Hooks != nil {
		hooks = []internaltypes.Hook{}
		for _, hook := range grpcContainer.Hooks {
			hooks = append(hooks, *ToInternalHook(hook))
		}
	}

	return &internaltypes.Container{
		ID:              grpcContainer.Id,
		Name:            grpcContainer.Name,
		Image:           *ToInternalImage(grpcContainer.Image),
		DomainName:      grpcContainer.DomainName,
		HostName:        grpcContainer.HostName,
		ResolvConfPath:  grpcContainer.ResolvConfPath,
		HostsPath:       grpcContainer.HostsPath,
		HostnamePath:    grpcContainer.HostnamePath,
		Mounts:          mounts,
		Hooks:           hooks,
		Config:          ToInternalConfig(grpcContainer.Config),
		HostConfig:      ToInternalHostConfig(grpcContainer.HostConfig),
		IOConfig:        ToInternalIOConfig(grpcContainer.IoConfig),
		NetworkSettings: ToInternalNetworkSettings(grpcContainer.NetworkSettings),
		State:           ToInternalState(grpcContainer.State),
		Created:         grpcContainer.Created,
		ManuallyStopped: grpcContainer.ManuallyStopped,
		RestartCount:    int(grpcContainer.RestartCount),
	}
}

// ToInternalDeviceMapping converts a types.DeviceMapping instance to an internal DeviceMapping one
func ToInternalDeviceMapping(grpcDevMapping *apitypescontainers.DeviceMapping) *internaltypes.DeviceMapping {
	return &internaltypes.DeviceMapping{
		PathOnHost:        grpcDevMapping.PathOnHost,
		PathInContainer:   grpcDevMapping.PathInContainer,
		CgroupPermissions: grpcDevMapping.CgroupPermissions,
	}
}

// ToInternalEndpointSettings converts a types.EndpointSettings instance to an internal EndpointSettings one
func ToInternalEndpointSettings(grpcEpSettings *apitypescontainers.EndpointSettings) *internaltypes.EndpointSettings {
	return &internaltypes.EndpointSettings{
		ID:         grpcEpSettings.Id,
		Gateway:    grpcEpSettings.Gateway,
		IPAddress:  grpcEpSettings.IpAddress,
		MacAddress: grpcEpSettings.MacAddress,
		NetworkID:  grpcEpSettings.NetworkId,
	}
}

// ToInternalHook converts a types.Hook instance to an internal Hook one
func ToInternalHook(grpcHook *apitypescontainers.Hook) *internaltypes.Hook {
	return &internaltypes.Hook{
		Path:    grpcHook.Path,
		Args:    append([]string{}, grpcHook.Args...),
		Env:     append([]string{}, grpcHook.Env...),
		Timeout: int(grpcHook.Timeout),
		Type:    ToInternalHookType(grpcHook.Type),
	}
}

// ToInternalImage converts a types.Image instance to an internal Image one
func ToInternalImage(grpcImage *apitypescontainers.Image) *internaltypes.Image {
	if grpcImage == nil {
		return nil
	}
	return &internaltypes.Image{
		Name:          grpcImage.Name,
		DecryptConfig: ToInternalDecryptConfig(grpcImage.DecryptConfig),
	}
}

// ToInternalDecryptConfig converts a types.DecryptConfig instance to an internal DecryptConfig one
func ToInternalDecryptConfig(grpcDecryptConfig *apitypescontainers.DecryptConfig) *internaltypes.DecryptConfig {
	if grpcDecryptConfig == nil {
		return nil
	}
	return &internaltypes.DecryptConfig{
		Keys:       grpcDecryptConfig.Keys,
		Recipients: grpcDecryptConfig.Recipients,
	}
}

// ToInternalMountPoint converts a types.MountPoint instance to an internal MountPoint one
func ToInternalMountPoint(grpcMountPoint *apitypescontainers.MountPoint) *internaltypes.MountPoint {
	if grpcMountPoint == nil {
		return nil
	}
	return &internaltypes.MountPoint{
		Destination:     grpcMountPoint.Destination,
		Source:          grpcMountPoint.Source,
		PropagationMode: grpcMountPoint.PropagationMode,
	}
}

// ToInternalNetworkSettings converts a types.NetworkSettings instance to an internal NetworkSettings one
func ToInternalNetworkSettings(grpcNetworkSettings *apitypescontainers.NetworkSettings) *internaltypes.NetworkSettings {
	if grpcNetworkSettings == nil {
		return nil
	}
	var networks map[string]*internaltypes.EndpointSettings

	if grpcNetworkSettings.Networks != nil {
		networks = make(map[string]*internaltypes.EndpointSettings)
		for netName, epSettings := range grpcNetworkSettings.Networks {
			networks[netName] = ToInternalEndpointSettings(epSettings)
		}
	}

	return &internaltypes.NetworkSettings{
		Networks:            networks,
		SandboxID:           grpcNetworkSettings.SandboxId,
		SandboxKey:          grpcNetworkSettings.SandboxKey,
		NetworkControllerID: grpcNetworkSettings.NetworkControllerId,
	}
}

// ToInternalPortMappings converts a types.PortMapping instance to an internal PortMapping one
func ToInternalPortMappings(grpcPortMappings []*apitypescontainers.PortMapping) []internaltypes.PortMapping {
	if grpcPortMappings == nil {
		return nil
	}
	var mappings []internaltypes.PortMapping
	mappings = []internaltypes.PortMapping{}
	for _, mapping := range grpcPortMappings {
		mappings = append(mappings, internaltypes.PortMapping{
			Proto:         mapping.Protocol,
			ContainerPort: uint16(mapping.ContainerPort),
			HostIP:        mapping.HostIp,
			HostPort:      uint16(mapping.HostPort),
			HostPortEnd:   uint16(mapping.HostPortEnd),
		})
	}
	return mappings
}

// ToInternalRestartPolicy converts a types.RestartPolicy instance to an internal RestartPolicy one
func ToInternalRestartPolicy(grpcRestartPolicy *apitypescontainers.RestartPolicy) *internaltypes.RestartPolicy {
	if grpcRestartPolicy == nil {
		return nil
	}
	return &internaltypes.RestartPolicy{
		MaximumRetryCount: int(grpcRestartPolicy.MaximumRetryCount),
		RetryTimeout:      time.Duration(grpcRestartPolicy.RetryTimeout) * time.Second,
		Type:              internaltypes.PolicyType(grpcRestartPolicy.Type),
	}
}

// ToInternalConfig converts a types.ContainerConfiguration instance to an internal ContainerConfiguration one
func ToInternalConfig(grpcConfig *apitypescontainers.ContainerConfiguration) *internaltypes.ContainerConfiguration {
	if grpcConfig == nil {
		return nil
	}
	return &internaltypes.ContainerConfiguration{
		Env: grpcConfig.Env,
		Cmd: grpcConfig.Cmd,
	}
}

// ToInternalHostConfig converts a types.HostConfig instance to internal HostConfig one
func ToInternalHostConfig(grpcHostConfig *apitypescontainers.HostConfig) *internaltypes.HostConfig {
	if grpcHostConfig == nil {
		return nil
	}
	var devices []internaltypes.DeviceMapping

	if grpcHostConfig.Devices != nil {
		devices = []internaltypes.DeviceMapping{}
		for _, grpcDevMapping := range grpcHostConfig.Devices {
			devices = append(devices, *ToInternalDeviceMapping(grpcDevMapping))
		}
	}

	return &internaltypes.HostConfig{
		Devices:           devices,
		NetworkMode:       internaltypes.NetworkMode(grpcHostConfig.NetworkMode),
		Privileged:        grpcHostConfig.Privileged,
		RestartPolicy:     ToInternalRestartPolicy(grpcHostConfig.RestartPolicy),
		Runtime:           internaltypes.Runtime(grpcHostConfig.Runtime),
		ExtraHosts:        grpcHostConfig.ExtraHosts,
		ExtraCapabilities: grpcHostConfig.ExtraCapabilities,
		PortMappings:      ToInternalPortMappings(grpcHostConfig.PortMappings),
		LogConfig:         ToInternalLogConfig(grpcHostConfig.LogConfig),
		Resources:         ToInternalResources(grpcHostConfig.Resources),
	}
}

// ToInternalIOConfig converts a types.IOConfig instance to an internal IOConfig one
func ToInternalIOConfig(grpcIOConfig *apitypescontainers.IOConfig) *internaltypes.IOConfig {
	if grpcIOConfig == nil {
		return nil
	}
	return &internaltypes.IOConfig{
		AttachStderr: grpcIOConfig.AttachStderr,
		AttachStdin:  grpcIOConfig.AttachStdin,
		AttachStdout: grpcIOConfig.AttachStdout,
		OpenStdin:    grpcIOConfig.OpenStdin,
		StdinOnce:    grpcIOConfig.StdinOnce,
		Tty:          grpcIOConfig.Tty,
	}
}

// ToInternalProjectInfo converts a types.ProjectInfo instance to an internal ProjectInfo one
func ToInternalProjectInfo(grpcProjectInfo *apitypessysinfo.ProjectInfo) sysinfointernaltypes.ProjectInfo {
	if grpcProjectInfo == nil {
		return sysinfointernaltypes.ProjectInfo{}
	}

	return sysinfointernaltypes.ProjectInfo{
		ProjectVersion: grpcProjectInfo.ProjectVersion,
		BuildTime:      grpcProjectInfo.BuildTime,
		APIVersion:     grpcProjectInfo.ApiVersion,
		GitCommit:      grpcProjectInfo.GitCommit,
	}
}

// ToInternalState converts a types.State instance to an internal State one
func ToInternalState(grpcState *apitypescontainers.State) *internaltypes.State {
	if grpcState == nil {
		return nil
	}
	return &internaltypes.State{
		Pid:        grpcState.Pid,
		StartedAt:  grpcState.StartedAt,
		Error:      grpcState.Error,
		ExitCode:   grpcState.ExitCode,
		FinishedAt: grpcState.FinishedAt,
		Exited:     grpcState.Exited,
		Dead:       grpcState.Dead,
		Restarting: grpcState.Restarting,
		Paused:     grpcState.Paused,
		Running:    grpcState.Running,
		Status:     ToInternalStatus(grpcState.Status),
		OOMKilled:  grpcState.OomKilled,
	}
}

// ToInternalHookType converts a gRPC hook type to an internal HookType one
func ToInternalHookType(grpcHookType string) internaltypes.HookType {
	switch grpcHookType {
	case internaltypes.HookTypePrestart.String():
		return internaltypes.HookTypePrestart
	case internaltypes.HookTypePoststart.String():
		return internaltypes.HookTypePoststart
	case internaltypes.HookTypePoststop.String():
		return internaltypes.HookTypePoststop
	default:
		return internaltypes.HookTypeUnknown
	}
}

// ToInternalStatus converts a gRPC status to an internal status one
func ToInternalStatus(grpcStatus string) internaltypes.Status {
	switch grpcStatus {
	case internaltypes.Creating.String():
		return internaltypes.Creating
	case internaltypes.Created.String():
		return internaltypes.Created
	case internaltypes.Running.String():
		return internaltypes.Running
	case internaltypes.Stopped.String():
		return internaltypes.Stopped
	case internaltypes.Paused.String():
		return internaltypes.Paused
	case internaltypes.Exited.String():
		return internaltypes.Exited
	case internaltypes.Dead.String():
		return internaltypes.Dead
	default:
		return internaltypes.Unknown
	}
}

// ToInternalLogConfig converts a types.LogConfiguration to an internal LogConfiguration one
func ToInternalLogConfig(grpcLogConfig *apitypescontainers.LogConfiguration) *internaltypes.LogConfiguration {
	if grpcLogConfig == nil {
		return nil
	}
	return &internaltypes.LogConfiguration{
		DriverConfig: ToInternalLogDriverConfig(grpcLogConfig.DriverConfig),
		ModeConfig:   ToInternalLogModeConfig(grpcLogConfig.ModeConfig),
	}
}

// ToInternalResources converts a types.Resources to an internal Resources one
func ToInternalResources(resources *apitypescontainers.Resources) *internaltypes.Resources {
	if resources == nil {
		return nil
	}
	return &internaltypes.Resources{
		Memory:            resources.Memory,
		MemoryReservation: resources.MemoryReservation,
		MemorySwap:        resources.MemorySwap,
	}
}

// ToInternalLogDriverConfig converts a types.LogDriverConfiguration to an internal LogDriverConfiguration one
func ToInternalLogDriverConfig(grpcLogDriverConfig *apitypescontainers.LogDriverConfiguration) *internaltypes.LogDriverConfiguration {
	if grpcLogDriverConfig == nil {
		return nil
	}
	return &internaltypes.LogDriverConfiguration{
		Type:     internaltypes.LogDriver(grpcLogDriverConfig.Type),
		MaxFiles: int(grpcLogDriverConfig.MaxFiles),
		MaxSize:  grpcLogDriverConfig.MaxSize,
		RootDir:  grpcLogDriverConfig.RootDir,
	}
}

// ToInternalLogModeConfig converts a types.LogModeConfiguration to an internal LogModeConfiguration
func ToInternalLogModeConfig(grpcLogModeConfig *apitypescontainers.LogModeConfiguration) *internaltypes.LogModeConfiguration {
	if grpcLogModeConfig == nil {
		return nil
	}
	return &internaltypes.LogModeConfiguration{
		Mode:          internaltypes.LogMode(grpcLogModeConfig.Mode),
		MaxBufferSize: grpcLogModeConfig.MaxBufferSize,
	}
}

// ToInternalStopOptions converts a types.StopOptions to an internal StopOptions
func ToInternalStopOptions(grpcStopOptions *apitypescontainers.StopOptions) *internaltypes.StopOpts {
	if grpcStopOptions == nil {
		return nil
	}
	return &internaltypes.StopOpts{
		Timeout: grpcStopOptions.Timeout,
		Force:   grpcStopOptions.Force,
		Signal:  grpcStopOptions.Signal,
	}
}

// ToInternalUpdateOptions converts a types.UpdateOptions to an internal UpdateOptions
func ToInternalUpdateOptions(grpcUpdateOptions *apitypescontainers.UpdateOptions) *internaltypes.UpdateOpts {
	if grpcUpdateOptions == nil {
		return &internaltypes.UpdateOpts{}
	}
	return &internaltypes.UpdateOpts{
		RestartPolicy: ToInternalRestartPolicy(grpcUpdateOptions.RestartPolicy),
		Resources:     ToInternalResources(grpcUpdateOptions.Resources),
	}
}
