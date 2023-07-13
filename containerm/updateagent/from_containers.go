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

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
)

const (
	defaultRestartCount                     = 0
	defaultRestartPolicyType                = ctrtypes.UnlessStopped
	defaultRestartPolicyMaximumRetryCount   = 0
	defaultRestartPolicyMaximumRetryTimeout = 0
	defaultHostConfigNetrowkMode            = ctrtypes.NetworkModeBridge
	defaultLogConfigDriverConfigType        = ctrtypes.LogConfigDriverJSONFile
	defaultLogConfigMaxSize                 = "100M"
	defaultLogConfigModeConfigMode          = ctrtypes.LogModeBlocking
	defaultLogConfigMaxFiles                = 2
	defaultLogConfigModeConfigMaxBufferSize = "1M"
	defaultExitCode                         = 0
)

func fromContainers(containers []*ctrtypes.Container, verbose bool) []*types.SoftwareNode {
	softwareNodes := make([]*types.SoftwareNode, len(containers))
	for i, container := range containers {
		softwareNodes[i] = fromContainer(container, verbose)
	}
	return softwareNodes
}

// The default values for some container config options and runtime state values are skipped and not included in the result unless verbose parameter is set to true.
func fromContainer(container *ctrtypes.Container, verbose bool) *types.SoftwareNode {
	params := []*types.KeyValuePair{}

	if verbose || len(container.Image.Name) > 0 {
		appendParameter(&params, keyImage, container.Image.Name)
	}
	if verbose || container.DomainName != container.Name+"-domain" {
		appendParameter(&params, keyDomainName, container.DomainName)
	}
	if verbose || container.HostName != container.Name+"-host" {
		appendParameter(&params, keyHostName, container.HostName)
	}
	if container.IOConfig != nil {
		params = append(params, ioConfigParameters(container.IOConfig, verbose)...)
	}
	if container.HostConfig != nil {
		params = append(params, hostConfigParameters(container.HostConfig, verbose)...)
	}
	if len(container.Mounts) > 0 {
		params = append(params, mountPointParameters(container.Mounts)...)
	}
	if container.Config != nil {
		params = append(params, containerConfigParameters(container.Config)...)
	}
	if container.State != nil {
		params = append(params, stateParameters(container.State, verbose)...)
	}
	appendParameter(&params, keyCreated, container.Created)
	if verbose || container.RestartCount != defaultRestartCount {
		appendParameter(&params, keyRestartCount, strconv.FormatInt(int64(container.RestartCount), 10))
	}
	if verbose || (container.ManuallyStopped && container.State.Status != ctrtypes.Running) {
		appendParameter(&params, keyManuallyStopped, strconv.FormatBool(container.ManuallyStopped))
	}
	if verbose || (container.StartedSuccessfullyBefore && container.State.Status != ctrtypes.Running) {
		appendParameter(&params, keyStartedSuccessfullyBefore, strconv.FormatBool(container.StartedSuccessfullyBefore))
	}

	return &types.SoftwareNode{
		InventoryNode: types.InventoryNode{
			ID:         container.Name,
			Version:    findContainerVersion(container.Image.Name),
			Parameters: params,
		},
		Type: types.SoftwareTypeContainer,
	}
}

func hostConfigParameters(hostConfig *ctrtypes.HostConfig, verbose bool) []*types.KeyValuePair {
	kvPair := []*types.KeyValuePair{}
	if verbose || hostConfig.Privileged {
		appendParameter(&kvPair, keyPrivileged, strconv.FormatBool(hostConfig.Privileged))
	}

	if hostConfig.RestartPolicy != nil {
		if verbose || hostConfig.RestartPolicy.Type != defaultRestartPolicyType {
			appendParameter(&kvPair, keyRestartPolicy, string(hostConfig.RestartPolicy.Type))
		}
		if verbose || hostConfig.RestartPolicy.MaximumRetryCount != defaultRestartPolicyMaximumRetryCount {
			appendParameter(&kvPair, keyRestartMaxRetries, strconv.FormatInt(int64(hostConfig.RestartPolicy.MaximumRetryCount), 10))
		}
		if verbose || hostConfig.RestartPolicy.RetryTimeout != defaultRestartPolicyMaximumRetryTimeout {
			appendParameter(&kvPair, keyRestartTimeout, hostConfig.RestartPolicy.RetryTimeout.String())
		}
	}
	for _, device := range hostConfig.Devices {
		appendParameter(&kvPair, keyDevice, util.DeviceMappingToString(&device))
	}
	for _, portMapping := range hostConfig.PortMappings {
		appendParameter(&kvPair, keyPort, util.PortMappingToString(&portMapping))
	}
	if verbose || (len(hostConfig.NetworkMode) > 0 && hostConfig.NetworkMode != defaultHostConfigNetrowkMode) {
		appendParameter(&kvPair, keyNetwork, string(hostConfig.NetworkMode))
	}
	for _, host := range hostConfig.ExtraHosts {
		appendParameter(&kvPair, keyHost, host)
	}
	if hostConfig.LogConfig != nil {
		if hostConfig.LogConfig.DriverConfig != nil {
			if verbose || (len(hostConfig.LogConfig.DriverConfig.Type) != 0 && hostConfig.LogConfig.DriverConfig.Type != defaultLogConfigDriverConfigType) {
				appendParameter(&kvPair, keyLogDriver, string(hostConfig.LogConfig.DriverConfig.Type))
			}
			if verbose || hostConfig.LogConfig.DriverConfig.MaxFiles != defaultLogConfigMaxFiles {
				appendParameter(&kvPair, keyLogMaxFiles, strconv.FormatInt(int64(hostConfig.LogConfig.DriverConfig.MaxFiles), 10))
			}

			if verbose || (len(hostConfig.LogConfig.DriverConfig.MaxSize) > 0 && hostConfig.LogConfig.DriverConfig.MaxSize != defaultLogConfigMaxSize) {
				appendParameter(&kvPair, keyLogMaxSize, hostConfig.LogConfig.DriverConfig.MaxSize)
			}
			if verbose || len(hostConfig.LogConfig.DriverConfig.RootDir) > 0 {
				appendParameter(&kvPair, keyLogPath, hostConfig.LogConfig.DriverConfig.RootDir)
			}
		}
		if hostConfig.LogConfig.ModeConfig != nil {
			if verbose || (len(hostConfig.LogConfig.ModeConfig.Mode) > 0 && hostConfig.LogConfig.ModeConfig.Mode != defaultLogConfigModeConfigMode) {
				appendParameter(&kvPair, keyLogMode, string(hostConfig.LogConfig.ModeConfig.Mode))
			}
			if verbose || (len(hostConfig.LogConfig.ModeConfig.MaxBufferSize) > 0 && hostConfig.LogConfig.ModeConfig.MaxBufferSize != defaultLogConfigModeConfigMaxBufferSize) {
				appendParameter(&kvPair, keyLogMaxBufferSize, hostConfig.LogConfig.ModeConfig.MaxBufferSize)
			}
		}
	}
	if hostConfig.Resources != nil {
		if verbose || len(hostConfig.Resources.Memory) > 0 {
			appendParameter(&kvPair, keyMemory, hostConfig.Resources.Memory)
		}
		if verbose || len(hostConfig.Resources.MemoryReservation) > 0 {
			appendParameter(&kvPair, keyMemoryReservation, hostConfig.Resources.MemoryReservation)
		}
		if verbose || len(hostConfig.Resources.MemorySwap) > 0 {
			appendParameter(&kvPair, keyMemorySwap, hostConfig.Resources.MemorySwap)
		}
	}
	return kvPair
}

func mountPointParameters(mounts []ctrtypes.MountPoint) []*types.KeyValuePair {
	kvPair := make([]*types.KeyValuePair, len(mounts))
	for i, mount := range mounts {
		kvPair[i] = &types.KeyValuePair{Key: keyMount, Value: util.MountPointToString(&mount)}
	}
	return kvPair
}

func containerConfigParameters(config *ctrtypes.ContainerConfiguration) []*types.KeyValuePair {
	kvPair := make([]*types.KeyValuePair, len(config.Env)+len(config.Cmd))
	for i, env := range config.Env {
		kvPair[i] = &types.KeyValuePair{Key: keyEnv, Value: env}
	}
	for i, cmd := range config.Cmd {
		kvPair[i+len(config.Env)] = &types.KeyValuePair{Key: keyCmd, Value: cmd}
	}
	return kvPair
}

func stateParameters(containerState *ctrtypes.State, verbose bool) []*types.KeyValuePair {
	kvPair := []*types.KeyValuePair{}
	appendParameter(&kvPair, keyStatus, containerState.Status.String())
	if verbose || (len(containerState.FinishedAt) > 0 && containerState.Status != ctrtypes.Running) {
		appendParameter(&kvPair, keyFinishedAt, containerState.FinishedAt)
	}
	if verbose || (containerState.ExitCode != defaultExitCode && containerState.Status != ctrtypes.Running) {
		appendParameter(&kvPair, keyExitCode, strconv.FormatInt(containerState.ExitCode, 10))
	}
	return kvPair
}

func ioConfigParameters(ioconfig *ctrtypes.IOConfig, verbose bool) []*types.KeyValuePair {
	kvPair := []*types.KeyValuePair{}
	if verbose || ioconfig.Tty {
		appendParameter(&kvPair, keyTerminal, strconv.FormatBool(ioconfig.Tty))
	}
	if verbose || ioconfig.OpenStdin {
		appendParameter(&kvPair, keyInteractive, strconv.FormatBool(ioconfig.OpenStdin))
	}
	return kvPair
}

func appendParameter(kv *[]*types.KeyValuePair, key string, value string) {
	*kv = append(*kv, &types.KeyValuePair{Key: key, Value: value})
}
