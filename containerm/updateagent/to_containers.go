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
	"strconv"
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
	var (
		env            []string
		cmd            []string
		extraHosts     []string
		mountPoints    []ctrtypes.MountPoint
		portMappings   []ctrtypes.PortMapping
		deviceMappings []ctrtypes.DeviceMapping
	)

	config := make(map[string]string, len(component.Config))
	for _, keyValuePair := range component.Config {
		switch keyValuePair.Key {
		case keyDevice:
			deviceMapping, err := util.ParseDeviceMapping(keyValuePair.Value)
			if err != nil {
				log.WarnErr(err, "Ignoring invalid device mapping")
			} else {
				deviceMappings = append(deviceMappings, *deviceMapping)
			}
		case keyPort:
			portMapping, err := util.ParsePortMapping(keyValuePair.Value)
			if err != nil {
				log.WarnErr(err, "Ignoring invalid port mapping")
			} else {
				portMappings = append(portMappings, *portMapping)
			}
		case keyHost:
			extraHosts = append(extraHosts, keyValuePair.Value)
		case keyMount:
			mountPoint, err := util.ParseMountPoint(keyValuePair.Value)
			if err != nil {
				log.WarnErr(err, "Ignoring invalid mount point")
			} else {
				mountPoints = append(mountPoints, *mountPoint)
			}
		case keyEnv:
			env = append(env, keyValuePair.Value)
		case keyCmd:
			cmd = append(cmd, keyValuePair.Value)
		default:
			config[keyValuePair.Key] = keyValuePair.Value
		}
	}

	imageName, ok := config[keyImage]
	if !ok {
		imageName = component.ID + ":" + component.Version
	}
	container := &ctrtypes.Container{
		Name: component.ID,
		Image: ctrtypes.Image{
			Name: imageName,
		},
		IOConfig: &ctrtypes.IOConfig{
			Tty:       parseBool(keyTerminal, config),
			OpenStdin: parseBool(keyInteractive, config),
		},
		Mounts: mountPoints,
		HostConfig: &ctrtypes.HostConfig{
			Privileged:   parseBool(keyPrivileged, config),
			NetworkMode:  ctrtypes.NetworkMode(config[keyNetwork]),
			Devices:      deviceMappings,
			ExtraHosts:   extraHosts,
			PortMappings: portMappings,
			LogConfig: &ctrtypes.LogConfiguration{
				DriverConfig: &ctrtypes.LogDriverConfiguration{
					Type:     ctrtypes.LogDriver(config[keyLogDriver]),
					MaxFiles: parseInt(keyLogMaxFiles, config),
					MaxSize:  config[keyLogMaxSize],
					RootDir:  config[keyLogPath],
				},
				ModeConfig: &ctrtypes.LogModeConfiguration{
					Mode:          ctrtypes.LogMode(config[keyLogMode]),
					MaxBufferSize: config[keyLogMaxBufferSize],
				},
			},
		},
	}
	if config[keyMemory] != "" || config[keyMemorySwap] != "" || config[keyMemoryReservation] != "" {
		container.HostConfig.Resources = &ctrtypes.Resources{
			Memory:            config[keyMemory],
			MemorySwap:        config[keyMemorySwap],
			MemoryReservation: config[keyMemoryReservation],
		}
	}

	if env != nil || cmd != nil {
		container.Config = &ctrtypes.ContainerConfiguration{
			Env: env,
			Cmd: cmd,
		}
	}

	if rpType, ok := config[keyRestartPolicy]; ok {
		container.HostConfig.RestartPolicy = &ctrtypes.RestartPolicy{
			Type: ctrtypes.PolicyType(rpType),
		}
		if container.HostConfig.RestartPolicy.Type == ctrtypes.OnFailure {
			container.HostConfig.RestartPolicy.MaximumRetryCount = parseInt(keyRestartMaxRetries, config)
			container.HostConfig.RestartPolicy.RetryTimeout = time.Duration(parseInt(keyRestartTimeout, config)) * time.Second
		}
	}

	util.FillDefaults(container)
	if err := util.ValidateContainer(container); err != nil {
		return container, err
	}

	return container, nil
}

func parseBool(key string, config map[string]string) bool {
	value, ok := config[key]
	if !ok {
		return false
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		log.Warn("Unknown boolean value for key %s = %s", key, value)
		return false
	}
	return result
}

func parseInt(key string, config map[string]string) int {
	value, ok := config[key]
	if !ok {
		return 0
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		log.Warn("Unknown integer value for key %s = %s", key, value)
		return 0
	}
	return result
}
