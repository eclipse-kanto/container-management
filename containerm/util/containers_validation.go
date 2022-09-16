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
	"regexp"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	containerNameRegexp      = "^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$"
	extraHostsReservedRegexp = "^(.+):host_ip(_(.+))?$"
	envVarRegexp             = "^[a-zA-Z_]([a-zA-Z0-9_]*)(|=(.*))$"
)

var (
	containerNameRegex      = regexp.MustCompile(containerNameRegexp)
	extraHostsReservedRegex = regexp.MustCompile(extraHostsReservedRegexp)
	envVarRegex             = regexp.MustCompile(envVarRegexp)
)

// ValidateContainer validats all container properties
func ValidateContainer(container *types.Container) error {
	if container.ID == "" {
		log.NewError("container ID must be provided")
	}
	if err := ValidateImage(container.Image); err != nil {
		return log.NewError("the containers image configuration is invalid")
	}
	if err := ValidateName(container.Name); err != nil {
		return err
	}
	if err := ValidateMounts(container.Mounts); err != nil {
		return err
	}
	if container.HostConfig == nil {
		return log.NewError("the containers host config is mandatory and is missing")
	}
	if err := ValidateHostConfig(container.HostConfig); err != nil {
		return err
	}
	if err := ValidateConfig(container.Config); err != nil {
		return err
	}
	if container.IOConfig == nil {
		return log.NewError("the container's IO config is missing")
	}
	return nil
}

// ValidateImage validates the container image
func ValidateImage(img types.Image) error {
	if img.Name == "" {
		return log.NewError("image is not provided")
	}
	return nil
}

// ValidateName validates the container name
func ValidateName(name string) error {
	if name != "" {
		if !containerNameRegex.MatchString(name) {
			return log.NewErrorf("invalid container name format : %s", name)
		}
	}
	return nil
}

// ValidateHostConfig validates the container host configuration
func ValidateHostConfig(hostConfig *types.HostConfig) error {
	if err := ValidateNetworking(hostConfig); err != nil {
		return err
	}
	if err := ValidateDeviceMappings(hostConfig.Devices); err != nil {
		return err
	}
	if err := ValidateLogConfig(hostConfig.LogConfig); err != nil {
		return err
	}
	if err := ValidateRestartPolicy(hostConfig.RestartPolicy); err != nil {
		return err
	}
	if err := ValidateResources(hostConfig.Resources); err != nil {
		return err
	}
	return nil
}

// ValidateResources validates the container resources limitations
func ValidateResources(resources *types.Resources) error {
	if resources == nil {
		return nil
	}

	var (
		swap        int64
		limit       int64
		reservation int64
		err         error
	)

	parse := func(limit string) (int64, error) {
		if limit == "" {
			return 0, nil
		}
		return SizeToBytes(limit)
	}

	if limit, err = parse(resources.Memory); err != nil {
		return log.NewErrorf("invalid format of memory - %s", resources.Memory)
	}
	if reservation, err = parse(resources.MemoryReservation); err != nil {
		return log.NewErrorf("invalid format of memory reservation - %s", resources.MemoryReservation)
	}
	if resources.MemorySwap == types.MemoryUnlimited {
		swap = -1
	} else if swap, err = parse(resources.MemorySwap); err != nil {
		return log.NewErrorf("invalid format of swap memory - %s", resources.MemorySwap)
	}

	if limit > 0 && limit < 3*mb {
		// even for busybox container at least 3M are needed in order to start
		return log.NewErrorf("minimum memory allowed is 3M")
	}

	if swap > 0 {
		if resources.Memory == "" {
			return log.NewErrorf("swap memory - %s, memory must be set as well", resources.MemorySwap)
		}
		if limit > swap {
			return log.NewErrorf("swap memory - %s is less than memory - %s", resources.MemorySwap, resources.Memory)
		}
	}

	if reservation > 0 && resources.Memory != "" && reservation > limit {
		return log.NewErrorf("reservation memory - %s must be lower than memory - %s", resources.MemoryReservation, resources.Memory)
	}

	return nil
}

// ValidateNetworking validates the container networking
func ValidateNetworking(hostConfig *types.HostConfig) error {
	if hostConfig.NetworkMode == "" {
		return log.NewError("network mode is not set")
	}
	if hostConfig.NetworkMode != types.NetworkModeHost && hostConfig.NetworkMode != types.NetworkModeBridge {
		return log.NewErrorf("unsupported network mode %s", hostConfig.NetworkMode)
	}
	if hostConfig.NetworkMode == types.NetworkModeHost {
		if len(hostConfig.PortMappings) != 0 {
			return log.NewError("cannot use port mappings when in host network mode")
		}
		for _, extraHost := range hostConfig.ExtraHosts {
			if extraHostsReservedRegex.MatchString(extraHost) {
				return log.NewError("cannot use the host_ip reserved key or any of its modifications when in host network mode")
			}
		}
	}
	return nil
}

// ValidateDeviceMappings validates all device mappings
func ValidateDeviceMappings(devMappings []types.DeviceMapping) error {
	if devMappings == nil {
		return nil
	}
	for _, devMapping := range devMappings {
		if err := ValidateDeviceMapping(devMapping); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDeviceMapping validates the mapping between the path on the host and the path in the container and checks if the cgroup permissions are set
func ValidateDeviceMapping(devMapping types.DeviceMapping) error {
	if devMapping.PathOnHost == "" || devMapping.PathInContainer == "" {
		return log.NewError("both path on the host and in the container must be specified for a device mapping")
	}
	if devMapping.CgroupPermissions == "" {
		return log.NewErrorf("the cgroup permissions for device mapping %s:%s are not provided", devMapping.PathOnHost, devMapping.PathInContainer)
	}
	return nil
}

// ValidateMounts validates all the cointainer mount points
func ValidateMounts(mounts []types.MountPoint) error {
	if mounts == nil {
		return nil
	}
	for _, mp := range mounts {
		if err := ValidateMountPoint(mp); err != nil {
			return err
		}
	}
	return nil
}

// ValidateMountPoint validates the cointainer mount configuration
func ValidateMountPoint(mp types.MountPoint) error {
	if mp.Source == "" || mp.Destination == "" {
		return log.NewError("source and destination must be set for a mount point")
	}
	propMode := mp.PropagationMode
	isPrivate := propMode == types.RPrivatePropagationMode || propMode == types.PrivatePropagationMode
	isShared := propMode == types.RSharedPropagationMode || propMode == types.SharedPropagationMode
	isSlave := propMode == types.RSlavePropagationMode || propMode == types.SlavePropagationMode
	if !(isPrivate || isShared || isSlave) {
		return log.NewError("propagation mode must be set to one of the supported modes")
	}
	return nil
}

// ValidateLogConfig validates the log configuration
func ValidateLogConfig(logCfg *types.LogConfiguration) error {
	if logCfg == nil {
		return nil
	}
	if logCfg.DriverConfig != nil {
		if !(logCfg.DriverConfig.Type == types.LogConfigDriverJSONFile || logCfg.DriverConfig.Type == types.LogConfigDriverNone) {
			return log.NewError("unsupported log driver configuration")
		}
		if logCfg.DriverConfig.Type == types.LogConfigDriverJSONFile {
			if logCfg.DriverConfig.MaxFiles < 1 {
				return log.NewError("max log files cannot be < 1")
			}
			if logCfg.DriverConfig.MaxSize == "" {
				return log.NewError("max logs size must be set")
			}
			if _, err := SizeToBytes(logCfg.DriverConfig.MaxSize); err != nil {
				return log.NewErrorf("invalid format of max logs size - %s", logCfg.DriverConfig.MaxSize)
			}
		}
	}
	if logCfg.ModeConfig != nil {
		if !(logCfg.ModeConfig.Mode == types.LogModeBlocking || logCfg.ModeConfig.Mode == types.LogModeNonBlocking) {
			return log.NewError("unsupported log mode configuration")
		}
		if logCfg.ModeConfig.Mode == types.LogModeNonBlocking {
			if logCfg.ModeConfig.MaxBufferSize == "" {
				return log.NewError("max buffer size must be set")
			}
			if _, err := SizeToBytes(logCfg.ModeConfig.MaxBufferSize); err != nil {
				return log.NewErrorf("invalid format of max buffer size - %s", logCfg.ModeConfig.MaxBufferSize)
			}
		}
	}
	return nil
}

// ValidateRestartPolicy validates the container restart policy
func ValidateRestartPolicy(rsPolicy *types.RestartPolicy) error {
	if rsPolicy == nil {
		return nil
	}

	switch rsPolicy.Type {
	case types.No:
	case types.UnlessStopped:
	case types.OnFailure:
	case types.Always:
		break
	default:
		return log.NewErrorf("unsupported restart policy type %s", rsPolicy.Type)
	}

	if rsPolicy.Type == types.OnFailure {
		if rsPolicy.MaximumRetryCount < 0 {
			return log.NewErrorf("restart policy max retry count cannot be negative")
		}
		if rsPolicy.RetryTimeout < 0 {
			return log.NewErrorf("restart policy retry timeout cannot be negative")
		}
	} else {
		if rsPolicy.MaximumRetryCount != 0 {
			return log.NewErrorf("cannot use max retry count when the restart policy is %s", rsPolicy.Type)
		}
	}
	return nil
}

// ValidateConfig validates the container config
func ValidateConfig(config *types.ContainerConfiguration) error {
	if config != nil {
		for _, envVar := range config.Env {
			if !envVarRegex.MatchString(envVar) {
				return log.NewErrorf("invalid environmental variable declaration provided : %s", envVar)
			}
		}
	}
	return nil
}
