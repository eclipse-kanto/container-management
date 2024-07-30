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
	"encoding/base64"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

// ParseDeviceMappings converts string representations of container's device mappings to structured DeviceMapping instances.
// The string representation format for a device mapping is defined with ParseDeviceMapping function.
func ParseDeviceMappings(devices []string) ([]types.DeviceMapping, error) {
	var devs []types.DeviceMapping
	for _, devPair := range devices {
		dev, err := ParseDeviceMapping(devPair)
		if err != nil {
			return nil, err
		}
		devs = append(devs, *dev)
	}
	return devs, nil
}

// ParseDeviceMapping converts a single string representation of a container's device mapping to a structured DeviceMapping instance.
// Format: <host_device>:<container_device>[:propagation_mode].
// Both path on host and in container must be set.
// The string representation may contain optional cgroups permissions configuration.
// Possible cgroup permissions options are “r” (read), “w” (write), “m” (mknod) and all combinations of the three are possible. If not set, “rwm” is default device configuration.
// Example: /dev/ttyACM0:/dev/ttyUSB0[:rwm].
func ParseDeviceMapping(device string) (*types.DeviceMapping, error) {
	pair := strings.Split(strings.TrimSpace(device), ":")
	if len(pair) == 2 {
		return &types.DeviceMapping{
			PathOnHost:        pair[0],
			PathInContainer:   pair[1],
			CgroupPermissions: "rwm",
		}, nil
	}
	if len(pair) == 3 {
		if len(pair[2]) == 0 || len(pair[2]) > 3 {
			return nil, log.NewErrorf("incorrect cgroup permissions format for device mapping %s", device)
		}
		for i := 0; i < len(pair[2]); i++ {
			if (pair[2])[i] != "w"[0] && (pair[2])[i] != "r"[0] && (pair[2])[i] != "m"[0] {
				return nil, log.NewErrorf("incorrect cgroup permissions format for device mapping %s", device)
			}
		}
		return &types.DeviceMapping{
			PathOnHost:        pair[0],
			PathInContainer:   pair[1],
			CgroupPermissions: pair[2],
		}, nil
	}
	return nil, log.NewErrorf("incorrect configuration value for device mapping %s", device)
}

// ParseMountPoints converts string representations of container's mounts to structured MountPoint instances.
// The string representation format for a mount point is defined with ParseMountPoint function.
func ParseMountPoints(mps []string) ([]types.MountPoint, error) {
	var mountPoints []types.MountPoint
	for _, mp := range mps {
		mount, err := ParseMountPoint(mp)
		if err != nil {
			return nil, err
		}
		mountPoints = append(mountPoints, *mount)
	}
	return mountPoints, nil
}

// ParseMountPoint converts a single string representation of a container's mount to a structured MountPoint instance.
// Format:
// "Source":"mount_source",
// "Destination":"mount_destination",
// "Propagation_Mode": "propagation_mode",
// "Data":"configuration_data"
// If the propagation mode parameter is omitted, rprivate will be set by default.
// Available propagation modes are: rprivate, private, rshared, shared, rslave, slave.
func ParseMountPoint(mp string) (*types.MountPoint, error) {
	var config types.MountPoint
	if err := json.Unmarshal([]byte(mp), &config); err != nil {
		return nil, log.NewErrorf("error unmarshalling json:", err)
	}
	if config.Source == "" || config.Destination == "" {
		return nil, log.NewErrorf("either mount source, %s or mount destination, %s is invalid", config.Source, config.Destination)
	}
	if config.PropagationMode == "" {
		config.PropagationMode = types.RPrivatePropagationMode
	}
	configInfo, err := base64.StdEncoding.DecodeString(config.Data)
	if err != nil {
		log.WarnErr(err, "error decoding bin64 string, ignoring configuration data", config.Data)
		config.Data = ""
		return &config, nil
	}
	config.Data = string(configInfo)
	return &config, nil
}

// ParsePortMappings converts string representations of container's port mappings to structured PortMapping instances.
// The string representation format for a port mapping is defined with ParsePortMapping function.
func ParsePortMappings(mappings []string) ([]types.PortMapping, error) {
	var portMappings []types.PortMapping
	for _, mapping := range mappings {
		pm, err := ParsePortMapping(mapping)
		if err != nil {
			return nil, err
		}
		portMappings = append(portMappings, *pm)
	}
	return portMappings, nil
}

// ParsePortMapping converts a single string representation of container's port mapping to a structured PortMapping instance.
// Format: [<host-ip>:]<host-port>[-<range>]:<container-port>[/<proto>].
// Most common use-case: 80:80
// Mapping the container’s 80 port to a host port in the 5000-6000 range: 5000-6000:80/udp
// Specifying port protocol (default is tcp): 80:80/udp
// By default the port mapping will set on all network interfaces, but this is also manageable: 0.0.0.0:80-100:80/udp
func ParsePortMapping(mapping string) (*types.PortMapping, error) {
	var (
		err           error
		protocol      string
		containerPort int64
		hostIP        string
		hostPort      int64
		hostPortEnd   int64
	)

	mapping0 := mapping
	mappingWithProto := strings.Split(strings.TrimSpace(mapping), "/")
	mapping = mappingWithProto[0]
	if len(mappingWithProto) == 2 {
		// port is specified, e.g.80:80/tcp
		protocol = mappingWithProto[1]
	}
	addressAndPorts := strings.Split(strings.TrimSpace(mapping), ":")
	hostPortIdx := 0 // if host ip not set
	if len(addressAndPorts) == 3 {
		hostPortIdx = 1
		hostIP = addressAndPorts[0]
		validIP := net.ParseIP(hostIP)
		if validIP == nil {
			return nil, log.NewErrorf("Incorrect host ip port mapping configuration %s", mapping0)
		}
	} else if len(addressAndPorts) != 2 { // len==2: host address not specified, e.g. 80:80
		return nil, log.NewErrorf("Incorrect port mapping configuration %s", mapping0)
	}
	hostPortWithRange := strings.Split(strings.TrimSpace(addressAndPorts[hostPortIdx]), "-")
	if len(hostPortWithRange) == 2 {
		hostPortEnd, err = strconv.ParseInt(hostPortWithRange[1], 10, 32)
		if err != nil {
			return nil, log.NewErrorf("Incorrect host range port mapping configuration %s", mapping0)
		}
		hostPort, err = strconv.ParseInt(hostPortWithRange[0], 10, 32)
	} else {
		hostPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx], 10, 32)
	}
	if err != nil {
		return nil, log.NewErrorf("Incorrect host port mapping configuration %s", mapping0)
	}
	containerPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx+1], 10, 32)
	if err != nil {
		return nil, log.NewErrorf("Incorrect container port mapping configuration %s", mapping0)
	}
	return &types.PortMapping{
		Proto:         protocol,
		ContainerPort: uint16(containerPort),
		HostIP:        hostIP,
		HostPort:      uint16(hostPort),
		HostPortEnd:   uint16(hostPortEnd),
	}, nil
}

// DeviceMappingToString returns the string representation of the given device mapping.
// The string representation format for a device mapping is defined with ParseDeviceMapping function.
func DeviceMappingToString(deviceMapping *types.DeviceMapping) string {
	var device strings.Builder
	if len(deviceMapping.PathOnHost) > 0 {
		device.WriteString(deviceMapping.PathOnHost)
	}
	if len(deviceMapping.PathInContainer) > 0 {
		if device.Len() > 0 {
			device.WriteRune(':')
		}
		device.WriteString(deviceMapping.PathInContainer)
	}
	if len(deviceMapping.CgroupPermissions) > 0 {
		if device.Len() > 0 {
			device.WriteRune(':')
		}
		device.WriteString(deviceMapping.CgroupPermissions)
	}
	return device.String()
}

// MountPointToString returns the string representation of the given mount point.
// The string representation format for a mount point is defined with ParseMountPoint function.
func MountPointToString(mountPoint *types.MountPoint) string {
	return mountPoint.Source + ":" + mountPoint.Destination + ":" + mountPoint.PropagationMode
}

// PortMappingToString returns the string representation of the given port mapping.
// The string representation format for a port mapping is defined with ParsePortMapping function.
func PortMappingToString(portMapping *types.PortMapping) string {
	var ports strings.Builder
	if len(portMapping.HostIP) > 0 && portMapping.HostIP != "0.0.0.0" { //ex. 1.2.3.4:80:80
		ports.WriteString(portMapping.HostIP)
		ports.WriteRune(':')
	}
	if portMapping.HostPort != 0 {
		ports.WriteString(strconv.FormatUint(uint64(portMapping.HostPort), 10))
		if portMapping.HostPortEnd != 0 && portMapping.HostPort != portMapping.HostPortEnd { //ex. 5000-6000:80
			ports.WriteRune('-')
			ports.WriteString(strconv.FormatUint(uint64(portMapping.HostPortEnd), 10))
		}
	}
	if portMapping.ContainerPort != 0 {
		ports.WriteRune(':')
		ports.WriteString(strconv.FormatUint(uint64(portMapping.ContainerPort), 10))
	}
	if len(portMapping.Proto) > 0 { //ex. 80:80/tcp
		ports.WriteRune('/')
		ports.WriteString(portMapping.Proto)
	}
	return ports.String()
}
