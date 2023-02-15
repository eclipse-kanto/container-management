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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

// IsContainerRunningOrPaused returns true if the container is running or paused
func IsContainerRunningOrPaused(c *types.Container) bool {
	return c.State.Running || c.State.Paused
}

// IsContainerCreated returns if container is in the created state or not
func IsContainerCreated(c *types.Container) bool {
	return c.State.Status == types.Created
}

// IsContainerDead returns whether the container is in Dead state
func IsContainerDead(c *types.Container) bool {
	return c.State.Status == types.Dead
}

// SetContainerStatusCreated sets the container state to created updating all required fields and flags
func SetContainerStatusCreated(c *types.Container) {
	c.State.Status = types.Created
	c.Created = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Pid = -1
	c.State.ExitCode = 0
	c.State.Dead = false
	c.State.Paused = false
	c.State.Running = false
	c.State.Restarting = false
	c.State.Exited = false
}

// SetContainerStatusRunning sets the container state to running updating all required fields and flags
func SetContainerStatusRunning(c *types.Container, pid int64) {
	c.State.Status = types.Running
	c.State.StartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Pid = pid
	c.State.ExitCode = 0
	c.State.Dead = false
	c.State.Paused = false
	c.State.Running = true
	c.State.Restarting = false
	c.State.Exited = false
	c.State.Error = ""
	c.State.OOMKilled = false
}

// SetContainerStatusStopped sets the container state to stopped updating all required fields and flags
func SetContainerStatusStopped(c *types.Container, exitCode int64, errMsg string) {
	c.State.Status = types.Stopped
	c.State.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Pid = -1
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.State.Dead = false
	c.State.Paused = false
	c.State.Running = false
	c.State.Restarting = false
	c.State.Exited = false
}

// SetContainerStatusExited sets the container state to exited updating all required fields and flags
func SetContainerStatusExited(c *types.Container, exitCode int64, errMsg string, oomKilled bool) {
	c.State.Status = types.Exited
	c.State.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Pid = -1
	c.State.ExitCode = exitCode
	c.State.Error = errMsg
	c.State.Dead = false
	c.State.Paused = false
	c.State.Running = false
	c.State.Restarting = false
	c.State.Exited = true
	if oomKilled {
		c.State.OOMKilled = true
		if c.State.Error == "" {
			c.State.Error = "OOM Killed"
		}
	}
}

// SetContainerStatusPaused sets the container state to paused updating all required fields and flags
func SetContainerStatusPaused(c *types.Container) {
	c.State.Status = types.Paused
	c.State.Dead = false
	c.State.Paused = true
	c.State.Running = false
	c.State.Restarting = false
	c.State.Exited = false
}

// SetContainerStatusUnpaused is added for completion as the container's state is actually running
func SetContainerStatusUnpaused(c *types.Container) {
	c.State.Status = types.Running
	c.State.Dead = false
	c.State.Paused = false
	c.State.Running = true
	c.State.Restarting = false
	c.State.Exited = false
}

// SetContainerStatusDead sets the container state to dead updating all required fields and flags
func SetContainerStatusDead(c *types.Container) {
	c.State.Status = types.Dead
	c.State.Dead = true
	c.State.Paused = false
	c.State.Running = false
	c.State.Restarting = false
	c.State.Exited = false
}

// IsRestartPolicyAlways checks if the restart policy type is set to always
func IsRestartPolicyAlways(policy *types.RestartPolicy) bool {
	return policy != nil && policy.Type == types.Always
}

// IsRestartPolicyNone checks if the restart policy type is set to no
func IsRestartPolicyNone(policy *types.RestartPolicy) bool {
	return policy == nil || policy.Type == types.No
}

// IsRestartPolicyUnlessStopped checks if the restart policy is set to unless-stopped
func IsRestartPolicyUnlessStopped(policy *types.RestartPolicy) bool {
	return policy != nil && policy.Type == types.UnlessStopped
}

// IsRestartPolicyOnFailure checks if the restart policy is set to on-failure
func IsRestartPolicyOnFailure(policy *types.RestartPolicy) bool {
	return policy != nil && policy.Type == types.OnFailure
}

const (
	jsonFileLogConfigDefaultMaxSize = "100M"
	jsonFileLogConfigDefaultMaxFile = 2

	logNonBlockingDefaultMaxBufferSize = "1M"
)

// FillDefaults sets all default configurations which are not required as an input but are required for processing the container's configuration
func FillDefaults(container *types.Container) bool {
	changesMade := false
	if container.ID == "" {
		log.Debug("container ID is not set - will generate a new one")
		container.ID = uuid.New().String()
		changesMade = true
	}
	if container.Name == "" {
		log.Debug("container name is not set - will set it equal to the ID")
		container.Name = container.ID
		changesMade = true
	}
	if container.DomainName == "" {
		log.Debug("container's domain name is not set - setting a default one")
		container.DomainName = fmt.Sprintf("%s-domain", container.Name)
		changesMade = true
	}
	if container.HostName == "" {
		log.Debug("container's host name is not set - setting a default one")
		container.HostName = fmt.Sprintf("%s-host", container.Name)
		changesMade = true
	}
	if container.HostConfig == nil {
		log.Debug("container's host config is not provided - setting a default one")
		container.HostConfig = &types.HostConfig{
			NetworkMode: types.NetworkModeBridge,
			Privileged:  false,
		}
		changesMade = true
	}
	if container.HostConfig.LogConfig == nil {
		container.HostConfig.LogConfig = &types.LogConfiguration{}
		changesMade = true
	}
	changesMade = fillLogDriverConfig(container) || changesMade
	changesMade = fillLogModeConfig(container) || changesMade

	if container.HostConfig.RestartPolicy == nil {
		log.Debug("restart policy in host config is not set - setting to default - %s", types.UnlessStopped)
		container.HostConfig.RestartPolicy = &types.RestartPolicy{
			Type: types.UnlessStopped,
		}
		changesMade = true
	}

	if container.HostConfig.NetworkMode == "" {
		log.Debug("network mode in host config is not set - setting to default - %s", types.NetworkModeBridge)
		container.HostConfig.NetworkMode = types.NetworkModeBridge
		changesMade = true
	}

	if container.HostConfig.Devices != nil {
		changesMade = fillDevices(container) || changesMade
	}

	if container.HostConfig.PortMappings != nil {
		changesMade = fillPortMappings(container) || changesMade
	}

	if container.HostConfig.Runtime == "" {
		log.Debug("container's runtime is not set - setting a default one")
		container.HostConfig.Runtime = types.RuntimeTypeV2runcV2
		changesMade = true
	}

	if container.IOConfig == nil {
		log.Debug("IO config is not set - setting it to default")
		container.IOConfig = &types.IOConfig{
			AttachStderr: false, /* explicit disabling for all */
			AttachStdin:  false,
			AttachStdout: false,
			OpenStdin:    false,
			StdinOnce:    false,
			Tty:          false,
		}
		changesMade = true
	}

	if container.Mounts != nil {
		for idx, mount := range container.Mounts {
			propMode := mount.PropagationMode
			if propMode == "" {
				log.Debug("missing propagation mode for mountpoint[%s, %s] - setting it to default - rprivate", mount.Destination, mount.Source)
				propMode = types.RPrivatePropagationMode
				changesMade = true
			}
			container.Mounts[idx] = types.MountPoint{
				Destination:     mount.Destination,
				Source:          mount.Source,
				PropagationMode: propMode,
			}
		}
	}

	if changesMade {
		log.Debug("added default values that updated the container's configuration")
	}
	return changesMade
}

// FillMemorySwap sets the swap memory of a container. Memory swap should be filled only once during creation.
func FillMemorySwap(container *types.Container) {
	resources := container.HostConfig.Resources
	if resources == nil {
		return
	}
	if resources.Memory != "" && resources.MemorySwap == "" {
		// set swap to 2 * resources limit. Skip error, container is later validated.
		resources.MemorySwap, _ = SizeRecalculate(resources.Memory, func(size float64) float64 {
			return size * 2
		})
		return
	}
	if resources.MemorySwap == types.MemoryUnlimited {
		resources.MemorySwap = ""
		return
	}
}

func fillDevices(container *types.Container) bool {
	changesMade := false
	for idx, dev := range container.HostConfig.Devices {
		if dev.PathInContainer == "" || dev.CgroupPermissions == "" {
			pathInContainer := dev.PathInContainer
			if pathInContainer == "" {
				log.Debug("path in container for mapped device %s is not provided - setting it equal to the path on host", dev.PathOnHost)
				pathInContainer = dev.PathOnHost
				changesMade = true
			}
			cgroupPerms := dev.CgroupPermissions
			if cgroupPerms == "" {
				log.Debug("cgroup permissions for mapped device %s are not set - setting them to default - rwm", dev.PathOnHost)
				cgroupPerms = "rwm"
				changesMade = true
			}
			container.HostConfig.Devices[idx] = types.DeviceMapping{
				PathOnHost:        dev.PathOnHost,
				PathInContainer:   pathInContainer,
				CgroupPermissions: cgroupPerms,
			}
		}
	}
	return changesMade
}
func fillLogDriverConfig(container *types.Container) bool {
	changesMade := false
	logCfg := container.HostConfig.LogConfig
	if logCfg.DriverConfig == nil {
		log.Debug("log driver configuration is not set - setting it to default - %s", types.LogConfigDriverJSONFile)
		logCfg.DriverConfig = &types.LogDriverConfiguration{
			Type:     types.LogConfigDriverJSONFile,
			MaxFiles: jsonFileLogConfigDefaultMaxFile,
			MaxSize:  jsonFileLogConfigDefaultMaxSize,
		}
		changesMade = true
	} else if logCfg.DriverConfig.Type == "" {
		log.Debug("log driver configuration is not set - setting it to default - %s", types.LogConfigDriverJSONFile)
		logCfg.DriverConfig.Type = types.LogConfigDriverJSONFile
		changesMade = true
	}
	if logCfg.DriverConfig.Type == types.LogConfigDriverJSONFile {
		if logCfg.DriverConfig.MaxFiles == 0 {
			log.Debug("log driver max files configuration is not set - setting it to default - %v", jsonFileLogConfigDefaultMaxFile)
			logCfg.DriverConfig.MaxFiles = jsonFileLogConfigDefaultMaxFile
			changesMade = true
		}
		if logCfg.DriverConfig.MaxSize == "" {
			log.Debug("log driver max size configuration is not set - setting it to default - %v", jsonFileLogConfigDefaultMaxSize)
			logCfg.DriverConfig.MaxSize = jsonFileLogConfigDefaultMaxSize
			changesMade = true
		}
	} else if logCfg.DriverConfig.Type == types.LogConfigDriverNone {
		if logCfg.DriverConfig.MaxSize != "" || logCfg.DriverConfig.MaxFiles != 0 {
			log.Debug("log driver configuration none is not set - discarding file options")
			logCfg.DriverConfig.MaxSize = ""
			logCfg.DriverConfig.MaxFiles = 0
			changesMade = true
		}
	}
	return changesMade
}

func fillLogModeConfig(container *types.Container) bool {
	changesMade := false
	logCfg := container.HostConfig.LogConfig
	if logCfg.ModeConfig == nil {
		log.Debug("log driver mode is not set - setting it to default - %s", types.LogModeBlocking)
		logCfg.ModeConfig = &types.LogModeConfiguration{
			Mode: types.LogModeBlocking,
		}
		changesMade = true
	} else if logCfg.ModeConfig.Mode == "" {
		log.Debug("log driver mode is not set - setting it to default - %s", types.LogModeBlocking)
		logCfg.ModeConfig.Mode = types.LogModeBlocking
		changesMade = true
	}
	if logCfg.ModeConfig.Mode == types.LogModeBlocking {
		if logCfg.ModeConfig.MaxBufferSize != "" {
			log.Debug("log mode is set to %s - discarding max buffer size", types.LogModeBlocking)
			logCfg.ModeConfig.MaxBufferSize = ""
			changesMade = true
		}
	} else if logCfg.ModeConfig.Mode == types.LogModeNonBlocking {
		if logCfg.ModeConfig.MaxBufferSize == "" {
			log.Debug("log mode is set to %s but max buffer size is not set - setting max buffer size %s", logNonBlockingDefaultMaxBufferSize)
			logCfg.ModeConfig.MaxBufferSize = logNonBlockingDefaultMaxBufferSize
			changesMade = true
		}
	}
	return changesMade
}

func fillPortMappings(container *types.Container) bool {
	changesMade := false
	for idx, mapping := range container.HostConfig.PortMappings {
		hostIP := mapping.HostIP
		if hostIP == "" {
			log.Debug("host IP in port mapping is not set - setting them to default - 0.0.0.0")
			hostIP = "0.0.0.0"
			changesMade = true
		}
		hostPortEnd := mapping.HostPortEnd
		if hostPortEnd == 0 {
			log.Debug("HostPortEnd in port mapping is not set - setting them to default - HostPort")
			hostPortEnd = mapping.HostPort
			changesMade = true
		}
		proto := mapping.Proto
		if proto == "" {
			log.Debug("protocol in port mapping is not set - setting them to default - tcp")
			proto = "tcp"
			changesMade = true
		}
		container.HostConfig.PortMappings[idx] = types.PortMapping{
			Proto:         proto,
			ContainerPort: mapping.ContainerPort,
			HostIP:        hostIP,
			HostPort:      mapping.HostPort,
			HostPortEnd:   hostPortEnd,
		}
	}
	return changesMade
}

// CalculateUptime calculates the uptime of a container instance
func CalculateUptime(container *types.Container) time.Duration {
	zeroDuration := 0 * time.Second
	if container.State == nil || container.State.StartedAt == "" || container.State.FinishedAt == "" {
		return zeroDuration
	}
	finishedAtTime, err := time.Parse(time.RFC3339, container.State.FinishedAt)
	if err != nil {
		return zeroDuration
	}
	startedAtTime, err := time.Parse(time.RFC3339, container.State.StartedAt)
	if err != nil {
		return zeroDuration
	}
	return finishedAtTime.Sub(startedAtTime)
}

// CalculateParallelLimit calculates the limit for starting parallel jobs
func CalculateParallelLimit(n int, limit int) int {
	const overhead = 2
	var rlim unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rlim); err != nil {
		log.WarnErr(err, "could not find container-management's RLIMIT_NOFILE to double-check startup parallelism factor")
		return limit
	}
	softRlim := int(rlim.Cur)
	if softRlim > overhead*n {
		return limit
	}
	if softRlim > overhead*limit {
		return limit
	}
	log.Warn("container-management's open file ulimit (%v) is too small - consider increasing it (>= %v)", softRlim, overhead*limit)
	return softRlim / overhead
}

// IsContainerNetworkBridge returns true if the network mode is bridge
func IsContainerNetworkBridge(container *types.Container) bool {
	return container.HostConfig != nil && container.HostConfig.NetworkMode == types.NetworkModeBridge
}

// IsContainerNetworkHost returns true if the network mode is host
func IsContainerNetworkHost(container *types.Container) bool {
	return container.HostConfig != nil && container.HostConfig.NetworkMode == types.NetworkModeHost
}

// CopyContainer creates a new container instance from the provided parameter
func CopyContainer(source *types.Container) types.Container {
	return types.Container{
		ID:                        source.ID,
		Name:                      source.Name,
		Image:                     source.Image,
		DomainName:                source.DomainName,
		HostName:                  source.HostName,
		ResolvConfPath:            source.ResolvConfPath,
		HostsPath:                 source.HostsPath,
		HostnamePath:              source.HostnamePath,
		Mounts:                    source.Mounts,
		Hooks:                     source.Hooks,
		Config:                    source.Config,
		HostConfig:                source.HostConfig,
		IOConfig:                  source.IOConfig,
		NetworkSettings:           source.NetworkSettings,
		State:                     source.State,
		Created:                   source.Created,
		RestartCount:              source.RestartCount,
		ManuallyStopped:           source.ManuallyStopped,
		StartedSuccessfullyBefore: source.StartedSuccessfullyBefore,
	}
}

// GetImageHost retrieves the host name of the container imager registry for the provided image
func GetImageHost(imageRef string) string {
	imageHost := strings.Split(imageRef, "/")[0]
	log.Debug("image registry host for image ref = %s is %s", imageRef, imageHost)
	return imageHost
}

// ReadContainer reads container from file
func ReadContainer(path string) (*types.Container, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var ctr *types.Container
	if err = json.NewDecoder(reader).Decode(&ctr); err != nil {
		return nil, err
	}
	return ctr, nil
}
