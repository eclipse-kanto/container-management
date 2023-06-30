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

package updateagent

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/pkg/errors"
)

func toContainers(components []*types.ComponentWithConfig) ([]*ctrtypes.Container, error) {
	containers := []*ctrtypes.Container{}
	for _, component := range components {
		container, err := toContainer(component)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid configuration for container %s", component.ID)
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func toContainer(component *types.ComponentWithConfig) (*ctrtypes.Container, error) {
	var logDriverConfig *ctrtypes.LogDriverConfiguration
	var logModeConfig *ctrtypes.LogModeConfiguration
	var restartPolicy *ctrtypes.RestartPolicy
	env := []string{}
	cmd := []string{}
	extraHosts := []string{}
	mountPoints := []ctrtypes.MountPoint{}
	portMappings := []ctrtypes.PortMapping{}
	deviceMappings := []ctrtypes.DeviceMapping{}

	container := &ctrtypes.Container{
		Name: component.ID,
		Image: ctrtypes.Image{
			Name: component.ID + ":" + component.Version,
		},
		IOConfig: &ctrtypes.IOConfig{},
		HostConfig: &ctrtypes.HostConfig{
			NetworkMode: ctrtypes.NetworkModeBridge,
		},
	}

	for _, keyValuePair := range component.Config {
		switch keyValuePair.Key {
		case keyImage:
			container.Image.Name = keyValuePair.Value
		case keyTerminal:
			container.IOConfig.Tty = parseBool(keyValuePair.Key, keyValuePair.Value)
		case keyInteractive:
			container.IOConfig.OpenStdin = parseBool(keyValuePair.Key, keyValuePair.Value)
		case keyPrivileged:
			container.HostConfig.Privileged = parseBool(keyValuePair.Key, keyValuePair.Value)
		case keyRestartPolicy:
			restartPolicy = newOrGetRestartPolicy(restartPolicy)
			restartPolicy.Type = ctrtypes.PolicyType(keyValuePair.Value)
		case keyRestartMaxRetries:
			if count := parseInt(keyValuePair.Key, keyValuePair.Value); count != -1 {
				restartPolicy = newOrGetRestartPolicy(restartPolicy)
				restartPolicy.MaximumRetryCount = count
			}
		case keyRestartTimeout:
			if timeout := parseInt(keyValuePair.Key, keyValuePair.Value); timeout != -1 {
				restartPolicy = newOrGetRestartPolicy(restartPolicy)
				restartPolicy.RetryTimeout = time.Duration(timeout) * time.Second
			}
		case keyDevice:
			if deviceMapping := parseDeviceMapping(keyValuePair.Value); deviceMapping != nil {
				deviceMappings = append(deviceMappings, *deviceMapping)
			}
		case keyPort:
			if portMapping := parsePortMapping(keyValuePair.Value); portMapping != nil {
				portMappings = append(portMappings, *portMapping)
			}
		case keyNetwork:
			container.HostConfig.NetworkMode = ctrtypes.NetworkMode(keyValuePair.Value)
		case keyHost:
			extraHosts = append(extraHosts, keyValuePair.Value)
		case keyMount:
			if mountPoint := parseMountPoint(keyValuePair.Value); mountPoint != nil {
				mountPoints = append(mountPoints, *mountPoint)
			}
		case keyEnv:
			env = append(env, keyValuePair.Value)
		case keyCmd:
			cmd = append(cmd, keyValuePair.Value)
		case keyLogDriver:
			logDriverConfig = newOrGetLogDriverConfig(logDriverConfig)
			logDriverConfig.Type = ctrtypes.LogDriver(keyValuePair.Value)
		case keyLogMaxFiles:
			if count := parseInt(keyValuePair.Key, keyValuePair.Value); count != -1 {
				logDriverConfig = newOrGetLogDriverConfig(logDriverConfig)
				logDriverConfig.MaxFiles = count
			}
		case keyLogMaxSize:
			logDriverConfig = newOrGetLogDriverConfig(logDriverConfig)
			logDriverConfig.MaxSize = keyValuePair.Value
		case keyLogPath:
			logDriverConfig = newOrGetLogDriverConfig(logDriverConfig)
			logDriverConfig.RootDir = keyValuePair.Value
		case keyLogMode:
			logModeConfig = newOrGetLogModeConfig(logModeConfig)
			logModeConfig.Mode = ctrtypes.LogMode(keyValuePair.Value)
		case keyLogMaxBufferSize:
			logModeConfig = newOrGetLogModeConfig(logModeConfig)
			logModeConfig.MaxBufferSize = keyValuePair.Value
		case keyMemory:
			container.HostConfig.Resources = newOrGetResources(container.HostConfig.Resources)
			container.HostConfig.Resources.Memory = keyValuePair.Value
		case keyMemoryReservation:
			container.HostConfig.Resources = newOrGetResources(container.HostConfig.Resources)
			container.HostConfig.Resources.MemoryReservation = keyValuePair.Value
		case keyMemorySwap:
			container.HostConfig.Resources = newOrGetResources(container.HostConfig.Resources)
			container.HostConfig.Resources.MemorySwap = keyValuePair.Value
		}
	}

	if container.HostConfig.Privileged && len(deviceMappings) > 0 {
		return nil, fmt.Errorf("cannot have a  privileged container with specified devices")
	}

	if restartPolicy != nil {
		if restartPolicy.Type == ctrtypes.Always || restartPolicy.Type == ctrtypes.No || restartPolicy.Type == ctrtypes.UnlessStopped {
			restartPolicy.MaximumRetryCount = 0
			restartPolicy.RetryTimeout = 0
		} else if restartPolicy.Type != ctrtypes.OnFailure {
			log.Warn("Unknown restart policy configuration %s", restartPolicy.Type)
			restartPolicy = nil
		}
		container.HostConfig.RestartPolicy = restartPolicy
	}
	if logDriverConfig != nil {
		if logDriverConfig.Type == ctrtypes.LogConfigDriverNone {
			// MaxFiles, MaxSize and RootDir are only valid for JSON file log
			logDriverConfig.MaxFiles = 0
			logDriverConfig.MaxSize = ""
			logDriverConfig.RootDir = ""
		} else if logDriverConfig.Type != ctrtypes.LogConfigDriverJSONFile {
			log.Warn("Unknown log driver configuration %s", logDriverConfig.Type)
			logDriverConfig = nil
		}
	}
	if logModeConfig != nil {
		if logModeConfig.Mode == ctrtypes.LogModeBlocking {
			logModeConfig.MaxBufferSize = ""
		} else if logModeConfig.Mode != ctrtypes.LogModeNonBlocking {
			log.Warn("Unknown log mode configuration %s", logModeConfig.Mode)
			logModeConfig = nil
		}
	}
	container.HostConfig.LogConfig = &ctrtypes.LogConfiguration{
		DriverConfig: logDriverConfig,
		ModeConfig:   logModeConfig,
	}
	if len(env) > 0 || len(cmd) > 0 {
		container.Config = &ctrtypes.ContainerConfiguration{
			Env: env,
			Cmd: cmd,
		}
	}
	if len(mountPoints) > 0 {
		container.Mounts = mountPoints
	}
	if len(deviceMappings) > 0 {
		container.HostConfig.Devices = deviceMappings
	}
	if len(portMappings) > 0 {
		container.HostConfig.PortMappings = portMappings
	}
	if len(extraHosts) > 0 {
		container.HostConfig.ExtraHosts = extraHosts
	}

	util.FillDefaults(container)

	if err := util.ValidateContainer(container); err != nil {
		return nil, err
	}
	return container, nil
}

func newOrGetRestartPolicy(restartPolicy *ctrtypes.RestartPolicy) *ctrtypes.RestartPolicy {
	if restartPolicy != nil {
		return restartPolicy
	}
	return &ctrtypes.RestartPolicy{
		MaximumRetryCount: 1,
		RetryTimeout:      time.Duration(30) * time.Second,
	}
}

func newOrGetLogDriverConfig(logDriverConfig *ctrtypes.LogDriverConfiguration) *ctrtypes.LogDriverConfiguration {
	if logDriverConfig != nil {
		return logDriverConfig
	}
	return &ctrtypes.LogDriverConfiguration{
		Type:     ctrtypes.LogConfigDriverJSONFile,
		MaxFiles: 2,
		MaxSize:  "100M",
	}
}

func newOrGetLogModeConfig(logModeConfig *ctrtypes.LogModeConfiguration) *ctrtypes.LogModeConfiguration {
	if logModeConfig != nil {
		return logModeConfig
	}
	return &ctrtypes.LogModeConfiguration{
		Mode:          ctrtypes.LogModeBlocking,
		MaxBufferSize: "1M",
	}
}

func newOrGetResources(resources *ctrtypes.Resources) *ctrtypes.Resources {
	if resources != nil {
		return resources
	}
	return &ctrtypes.Resources{}
}

func parseDeviceMapping(device string) *ctrtypes.DeviceMapping {
	pair := strings.Split(strings.TrimSpace(device), ":")
	if len(pair) == 2 {
		return &ctrtypes.DeviceMapping{
			PathOnHost:        pair[0],
			PathInContainer:   pair[1],
			CgroupPermissions: "rwm",
		}
	}
	if len(pair) == 3 {
		if len(pair[2]) == 0 || len(pair[2]) > 3 {
			log.Warn("incorrect cgroup permissions format for device mapping [%s]", device)
			return nil
		}
		for i := 0; i < len(pair[2]); i++ {
			if (pair[2])[i] != "w"[0] && (pair[2])[i] != "r"[0] && (pair[2])[i] != "m"[0] {
				log.Warn("incorrect cgroup permissions for device mapping [%s]", device)
				return nil
			}
		}
		return &ctrtypes.DeviceMapping{
			PathOnHost:        pair[0],
			PathInContainer:   pair[1],
			CgroupPermissions: pair[2],
		}
	}
	log.Warn("incorrect configuration value for device mapping [%s]", device)
	return nil
}

func parsePortMapping(mapping string) *ctrtypes.PortMapping {
	var err error
	var protocol string
	var containerPort int64
	var hostIP string
	var hostPort int64
	var hostPortEnd int64

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
			log.Warn("Incorrect host ip port mapping configuration %s", mapping)
			return nil
		}
	} else if len(addressAndPorts) != 2 { // len==2: host address not specified, e.g. 80:80
		log.Warn("Incorrect port mapping configuration %s", mapping)
		return nil
	}
	hostPortWithRange := strings.Split(strings.TrimSpace(addressAndPorts[hostPortIdx]), "-")
	if len(hostPortWithRange) == 2 {
		hostPortEnd, err = strconv.ParseInt(hostPortWithRange[1], 10, 32)
		if err != nil {
			log.WarnErr(err, "Incorrect host range port mapping configuration %s", mapping)
			return nil
		}
		hostPort, err = strconv.ParseInt(hostPortWithRange[0], 10, 32)
	} else {
		hostPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx], 10, 32)
	}
	if err != nil {
		log.WarnErr(err, "Incorrect host port mapping configuration %s", mapping)
		return nil
	}
	containerPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx+1], 10, 32)
	if err != nil {
		log.WarnErr(err, "Incorrect container port mapping configuration %s", mapping)
		return nil
	}
	return &ctrtypes.PortMapping{
		Proto:         protocol,
		ContainerPort: uint16(containerPort),
		HostIP:        hostIP,
		HostPort:      uint16(hostPort),
		HostPortEnd:   uint16(hostPortEnd),
	}
}

func parseMountPoint(mp string) *ctrtypes.MountPoint {
	mount := strings.Split(strings.TrimSpace(mp), ":")
	if len(mount) < 2 || len(mount) > 3 {
		log.Warn("Incorrect number of parameters of the mount point %s", mp)
		return nil
	}
	mountPoint := &ctrtypes.MountPoint{
		Destination: mount[1],
		Source:      mount[0],
	}
	if len(mount) == 2 {
		// if propagation mode is omitted, "rprivate" is set as default
		mountPoint.PropagationMode = ctrtypes.RPrivatePropagationMode
	} else {
		mountPoint.PropagationMode = mount[2]
	}
	return mountPoint
}

func parseBool(key string, value string) bool {
	result, err := strconv.ParseBool(value)
	if err != nil {
		log.Warn("Unknown boolean value for key %s = %s", key, value)
		return false
	}
	return result
}

func parseInt(key string, value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		log.Warn("Unknown integer value for key %s = %s", key, value)
		return -1
	}
	return result
}
