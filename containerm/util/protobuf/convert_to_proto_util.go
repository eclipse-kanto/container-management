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
	apitypescontainers "github.com/eclipse-kanto/container-management/containerm/api/types/containers"
	apitypessysinfo "github.com/eclipse-kanto/container-management/containerm/api/types/sysinfo"
	internaltypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	sysinfointernaltypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
)

// ToProtoContainer converts an internal Container instance to a types.Container one
func ToProtoContainer(intenralContainer *internaltypes.Container) *apitypescontainers.Container {
	if intenralContainer == nil {
		return nil
	}
	var (
		mounts []*apitypescontainers.MountPoint
		hooks  []*apitypescontainers.Hook
	)

	if intenralContainer.Mounts != nil {
		mounts = []*apitypescontainers.MountPoint{}
		for _, mnt := range intenralContainer.Mounts {
			mounts = append(mounts, ToProtoMountPoint(&mnt))
		}
	}

	if intenralContainer.Hooks != nil {
		hooks = []*apitypescontainers.Hook{}
		for _, hook := range intenralContainer.Hooks {
			hooks = append(hooks, ToProtoHook(&hook))
		}
	}

	return &apitypescontainers.Container{
		Id:              intenralContainer.ID,
		Name:            intenralContainer.Name,
		Image:           ToProtoImage(&intenralContainer.Image),
		HostName:        intenralContainer.HostName,
		DomainName:      intenralContainer.DomainName,
		ResolvConfPath:  intenralContainer.ResolvConfPath,
		HostsPath:       intenralContainer.HostsPath,
		HostnamePath:    intenralContainer.HostnamePath,
		Mounts:          mounts,
		Hooks:           hooks,
		Config:          ToProtoConfig(intenralContainer.Config),
		HostConfig:      ToProtoHostConfig(intenralContainer.HostConfig),
		IoConfig:        ToProtoIOConfig(intenralContainer.IOConfig),
		NetworkSettings: ToProtoNetworkSettings(intenralContainer.NetworkSettings),
		State:           ToProtoState(intenralContainer.State),
		Created:         intenralContainer.Created,
		ManuallyStopped: intenralContainer.ManuallyStopped,
		RestartCount:    int64(intenralContainer.RestartCount),
	}
}

// ToProtoState converts an internal State instance to a types.State one
func ToProtoState(internState *internaltypes.State) *apitypescontainers.State {
	if internState == nil {
		return nil
	}
	return &apitypescontainers.State{
		Pid:        internState.Pid,
		StartedAt:  internState.StartedAt,
		Error:      internState.Error,
		ExitCode:   internState.ExitCode,
		FinishedAt: internState.FinishedAt,
		Exited:     internState.Exited,
		Dead:       internState.Dead,
		Restarting: internState.Restarting,
		Paused:     internState.Paused,
		Running:    internState.Running,
		Status:     internState.Status.String(),
		OomKilled:  internState.OOMKilled,
	}
}

// ToProtoDeviceMapping converts an internal DeviceMapping instance to a types.DeviceMapping one
func ToProtoDeviceMapping(internDevMapping *internaltypes.DeviceMapping) *apitypescontainers.DeviceMapping {
	if internDevMapping == nil {
		return nil
	}
	return &apitypescontainers.DeviceMapping{
		PathOnHost:        internDevMapping.PathOnHost,
		PathInContainer:   internDevMapping.PathInContainer,
		CgroupPermissions: internDevMapping.CgroupPermissions,
	}
}

// ToProtoEndpointSettings converts an internal EndpointSettings instance to a types.EndpointSettings one
func ToProtoEndpointSettings(internEpSettings *internaltypes.EndpointSettings) *apitypescontainers.EndpointSettings {
	if internEpSettings == nil {
		return nil
	}
	return &apitypescontainers.EndpointSettings{
		Id:         internEpSettings.ID,
		Gateway:    internEpSettings.Gateway,
		IpAddress:  internEpSettings.IPAddress,
		MacAddress: internEpSettings.MacAddress,
		NetworkId:  internEpSettings.NetworkID,
	}
}

// ToProtoHook converts an internal Hook instance to a types.Hook one
func ToProtoHook(inernalHook *internaltypes.Hook) *apitypescontainers.Hook {
	if inernalHook == nil {
		return nil
	}
	return &apitypescontainers.Hook{
		Path:    inernalHook.Path,
		Args:    append([]string{}, inernalHook.Args...),
		Env:     append([]string{}, inernalHook.Env...),
		Timeout: int32(inernalHook.Timeout),
		Type:    inernalHook.Type.String(),
	}
}

// ToProtoImage converts an internal Image instance to a types.Image one
func ToProtoImage(internalImage *internaltypes.Image) *apitypescontainers.Image {
	if internalImage == nil {
		return nil
	}
	return &apitypescontainers.Image{
		Name:          internalImage.Name,
		DecryptConfig: ToProtoDecryptConfig(internalImage.DecryptConfig),
	}
}

// ToProtoDecryptConfig converts an internal DecryptDate instance to a types.DecryptDate one
func ToProtoDecryptConfig(DecryptConfig *internaltypes.DecryptConfig) *apitypescontainers.DecryptConfig {
	if DecryptConfig == nil {
		return nil
	}
	return &apitypescontainers.DecryptConfig{
		Keys:       DecryptConfig.Keys,
		Recipients: DecryptConfig.Recipients,
	}
}

// ToProtoMountPoint converts an internal MountPoint instance to a types.MountPoint one
func ToProtoMountPoint(internalMountPoint *internaltypes.MountPoint) *apitypescontainers.MountPoint {
	if internalMountPoint == nil {
		return nil
	}
	return &apitypescontainers.MountPoint{
		Destination:     internalMountPoint.Destination,
		Source:          internalMountPoint.Source,
		PropagationMode: internalMountPoint.PropagationMode,
	}
}

// ToProtoNetworkSettings converts an internal NetworkSettings instance to a types.NetworkSettings one
func ToProtoNetworkSettings(internalNetworkSettings *internaltypes.NetworkSettings) *apitypescontainers.NetworkSettings {
	if internalNetworkSettings == nil {
		return nil
	}
	var networks map[string]*apitypescontainers.EndpointSettings

	if internalNetworkSettings.Networks != nil {
		networks = make(map[string]*apitypescontainers.EndpointSettings)
		for netName, epSettings := range internalNetworkSettings.Networks {
			networks[netName] = ToProtoEndpointSettings(epSettings)
		}
	}

	return &apitypescontainers.NetworkSettings{
		Networks:            networks,
		SandboxId:           internalNetworkSettings.SandboxID,
		SandboxKey:          internalNetworkSettings.SandboxKey,
		NetworkControllerId: internalNetworkSettings.NetworkControllerID,
	}
}

// ToProtoPortMappings converts an internal PortMapping instance to a types.PortMapping one
func ToProtoPortMappings(internalPortMappings []internaltypes.PortMapping) []*apitypescontainers.PortMapping {
	if internalPortMappings == nil {
		return nil
	}
	var protoPortMappings []*apitypescontainers.PortMapping
	protoPortMappings = []*apitypescontainers.PortMapping{}
	for _, mapping := range internalPortMappings {
		protoPortMappings = append(protoPortMappings, &apitypescontainers.PortMapping{
			Protocol:      mapping.Proto,
			HostIp:        mapping.HostIP,
			HostPort:      int64(mapping.HostPort),
			HostPortEnd:   int64(mapping.HostPortEnd),
			ContainerPort: int64(mapping.ContainerPort),
		})
	}

	return protoPortMappings
}

// ToProtoRestartPolicy converts an internal RestartPolicy instance to a types.RestartPolicy one
func ToProtoRestartPolicy(internalRestartPolicy *internaltypes.RestartPolicy) *apitypescontainers.RestartPolicy {
	if internalRestartPolicy == nil {
		return nil
	}
	return &apitypescontainers.RestartPolicy{
		MaximumRetryCount: int64(internalRestartPolicy.MaximumRetryCount),
		RetryTimeout:      (int64)(internalRestartPolicy.RetryTimeout.Seconds()),
		Type:              string(internalRestartPolicy.Type),
	}
}

// ToProtoConfig converts an internal instance ContainerConfiguration to a types.ContainerConfiguration one
func ToProtoConfig(internalConfig *internaltypes.ContainerConfiguration) *apitypescontainers.ContainerConfiguration {
	if internalConfig == nil {
		return nil
	}
	return &apitypescontainers.ContainerConfiguration{
		Env: internalConfig.Env,
		Cmd: internalConfig.Cmd,
	}
}

// ToProtoHostConfig converts an internal instance HostConfig to a types.HostConfig one
func ToProtoHostConfig(internalHostConfig *internaltypes.HostConfig) *apitypescontainers.HostConfig {
	if internalHostConfig == nil {
		return nil
	}
	var devices []*apitypescontainers.DeviceMapping

	if internalHostConfig.Devices != nil {
		devices = []*apitypescontainers.DeviceMapping{}
		for _, grpcDevMapping := range internalHostConfig.Devices {
			devices = append(devices, ToProtoDeviceMapping(&grpcDevMapping))
		}
	}

	return &apitypescontainers.HostConfig{
		Devices:           devices,
		NetworkMode:       string(internalHostConfig.NetworkMode),
		Privileged:        internalHostConfig.Privileged,
		RestartPolicy:     ToProtoRestartPolicy(internalHostConfig.RestartPolicy),
		Runtime:           string(internalHostConfig.Runtime),
		ExtraHosts:        internalHostConfig.ExtraHosts,
		ExtraCapabilities: internalHostConfig.ExtraCapabilities,
		PortMappings:      ToProtoPortMappings(internalHostConfig.PortMappings),
		LogConfig:         ToProtoLogConfig(internalHostConfig.LogConfig),
		Resources:         ToProtoResource(internalHostConfig.Resources),
	}
}

// ToProtoIOConfig converts an internal IOConfig instance to a types.IOConfig one
func ToProtoIOConfig(internalIOConfig *internaltypes.IOConfig) *apitypescontainers.IOConfig {
	if internalIOConfig == nil {
		return nil
	}
	return &apitypescontainers.IOConfig{
		AttachStderr: internalIOConfig.AttachStderr,
		AttachStdin:  internalIOConfig.AttachStdin,
		AttachStdout: internalIOConfig.AttachStdout,
		OpenStdin:    internalIOConfig.OpenStdin,
		StdinOnce:    internalIOConfig.StdinOnce,
		Tty:          internalIOConfig.Tty,
	}
}

// ToProtoProjectInfo converts an internal ProjectInfo instance to a types.ProjectInfo one
func ToProtoProjectInfo(projectInfo sysinfointernaltypes.ProjectInfo) *apitypessysinfo.ProjectInfo {
	return &apitypessysinfo.ProjectInfo{
		ProjectVersion: projectInfo.ProjectVersion,
		BuildTime:      projectInfo.BuildTime,
		ApiVersion:     projectInfo.APIVersion,
		GitCommit:      projectInfo.GitCommit,
	}
}

// ToProtoLogConfig converts an internal LogConfiguration instance to a types.LogConfiguration one
func ToProtoLogConfig(internalLogConfig *internaltypes.LogConfiguration) *apitypescontainers.LogConfiguration {
	if internalLogConfig == nil {
		return nil
	}
	return &apitypescontainers.LogConfiguration{
		DriverConfig: ToProtoLogDriverConfig(internalLogConfig.DriverConfig),
		ModeConfig:   ToProtoLogModeConfig(internalLogConfig.ModeConfig),
	}
}

// ToProtoResource converts an internal Resource instance to a types.Resource one
func ToProtoResource(internalResource *internaltypes.Resources) *apitypescontainers.Resources {
	if internalResource == nil {
		return nil
	}
	return &apitypescontainers.Resources{
		Memory:            internalResource.Memory,
		MemoryReservation: internalResource.MemoryReservation,
		MemorySwap:        internalResource.MemorySwap,
	}
}

// ToProtoLogDriverConfig converts an internal LogDriverConfiguration instance to a types.LogDriverConfiguration one
func ToProtoLogDriverConfig(internalLogDriverConfig *internaltypes.LogDriverConfiguration) *apitypescontainers.LogDriverConfiguration {
	if internalLogDriverConfig == nil {
		return nil
	}
	return &apitypescontainers.LogDriverConfiguration{
		Type:     string(internalLogDriverConfig.Type),
		MaxFiles: int64(internalLogDriverConfig.MaxFiles),
		MaxSize:  internalLogDriverConfig.MaxSize,
		RootDir:  internalLogDriverConfig.RootDir,
	}
}

// ToProtoLogModeConfig converts an internal LogModeConfiguration instance to a types.LogModeConfiguration one
func ToProtoLogModeConfig(internalLogModeConfig *internaltypes.LogModeConfiguration) *apitypescontainers.LogModeConfiguration {
	if internalLogModeConfig == nil {
		return nil
	}
	return &apitypescontainers.LogModeConfiguration{
		Mode:          string(internalLogModeConfig.Mode),
		MaxBufferSize: internalLogModeConfig.MaxBufferSize,
	}
}

// ToProtoStopOptions converts an internal StopOpts instance to a types.StopOpts one
func ToProtoStopOptions(intenralStopOpts *internaltypes.StopOpts) *apitypescontainers.StopOptions {
	if intenralStopOpts == nil {
		return nil
	}
	return &apitypescontainers.StopOptions{
		Timeout: intenralStopOpts.Timeout,
		Force:   intenralStopOpts.Force,
		Signal:  intenralStopOpts.Signal,
	}
}

// ToProtoUpdateOptions converts an internal UpdateOpts instance to a types.UpdateOpts one
func ToProtoUpdateOptions(intenralUpdateOpts *internaltypes.UpdateOpts) *apitypescontainers.UpdateOptions {
	if intenralUpdateOpts == nil {
		return nil
	}
	return &apitypescontainers.UpdateOptions{
		RestartPolicy: ToProtoRestartPolicy(intenralUpdateOpts.RestartPolicy),
		Resources:     ToProtoResource(intenralUpdateOpts.Resources),
	}
}
