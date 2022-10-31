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

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type configuration struct {
	DomainName  string                   `json:"domainName,omitempty"`
	MountPoints []*mountPoint            `json:"mountPoints,omitempty"`
	HostName    string                   `json:"hostName,omitempty"`
	Env         []string                 `json:"env,omitempty"`
	Cmd         []string                 `json:"cmd,omitempty"`
	Decryption  *decryptionConfiguration `json:"decryption,omitempty"`
	// host resources
	Devices       []*device      `json:"devices,omitempty"`
	Privileged    bool           `json:"privileged,omitempty"`
	RestartPolicy *restartPolicy `json:"restartPolicy,omitempty"`
	ExtraHosts    []string       `json:"extraHosts,omitempty"`
	PortMappings  []*portMapping `json:"portMappings,omitempty"`
	NetworkMode   networkMode    `json:"networkMode,omitempty"`
	// IO Config
	OpenStdin bool              `json:"openStdin,omitempty"`
	Tty       bool              `json:"tty,omitempty"`
	Log       *logConfiguration `json:"log,omitempty"`
	Resources *resources        `json:"resources,omitempty"`
}

func fromAPIContainerConfig(ctr *types.Container) *configuration {
	cfg := &configuration{
		DomainName: ctr.DomainName,
	}

	if ctr.HostConfig != nil {
		cfg.Privileged = ctr.HostConfig.Privileged
		if ctr.HostConfig.RestartPolicy != nil {
			cfg.RestartPolicy = fromAPIRestartPolicy(ctr.HostConfig.RestartPolicy)
		}
		if ctr.HostConfig.ExtraHosts != nil && len(ctr.HostConfig.ExtraHosts) > 0 {
			cfg.ExtraHosts = ctr.HostConfig.ExtraHosts
		}
		if ctr.HostConfig.Devices != nil {
			cfg.Devices = []*device{}
			for _, dev := range ctr.HostConfig.Devices {
				cfg.Devices = append(cfg.Devices, fromAPIDevice(dev))
			}
		}
		if ctr.HostConfig.PortMappings != nil && len(ctr.HostConfig.PortMappings) > 0 {
			cfg.PortMappings = []*portMapping{}
			for _, pm := range ctr.HostConfig.PortMappings {
				cfg.PortMappings = append(cfg.PortMappings, fromAPIPortMapping(pm))
			}
		}
		if ctr.HostConfig.LogConfig != nil {
			cfg.Log = fromAPILogConfiguration(ctr.HostConfig.LogConfig)
		}
		if ctr.HostConfig.Resources != nil {
			cfg.Resources = fromAPIResources(ctr.HostConfig.Resources)
		}
		cfg.NetworkMode = fromAPINetworkMode(ctr.HostConfig.NetworkMode)
	}
	if ctr.Mounts != nil && len(ctr.Mounts) > 0 {
		for _, mp := range ctr.Mounts {
			cfg.MountPoints = append(cfg.MountPoints, fromAPIMountPoint(mp))
		}
	}
	if ctr.IOConfig != nil {
		cfg.OpenStdin = ctr.IOConfig.OpenStdin
		cfg.Tty = ctr.IOConfig.Tty
	}

	if ctr.Config != nil {
		if ctr.Config.Env != nil && len(ctr.Config.Env) > 0 {
			cfg.Env = ctr.Config.Env
		}
		if ctr.Config.Cmd != nil && len(ctr.Config.Cmd) > 0 {
			cfg.Cmd = ctr.Config.Cmd
		}
	}
	if ctr.Image.DecryptConfig != nil {
		cfg.Decryption = fromAPIDecryptionConfiguration(ctr.Image.DecryptConfig)
	}
	if len(ctr.HostName) > 0 {
		cfg.HostName = ctr.HostName
	}
	return cfg
}

func toAPIContainerConfig(cfg *configuration) *types.Container {
	if cfg == nil {
		return &types.Container{}
	}
	ctr := &types.Container{
		DomainName: cfg.DomainName,
		IOConfig: &types.IOConfig{
			OpenStdin: cfg.OpenStdin,
			Tty:       cfg.Tty,
		},
	}
	ctr.HostConfig = &types.HostConfig{
		Privileged: cfg.Privileged,
	}

	if cfg.RestartPolicy != nil {
		ctr.HostConfig.RestartPolicy = toAPIRestartPolicy(cfg.RestartPolicy)
	}
	if cfg.ExtraHosts != nil && len(cfg.ExtraHosts) > 0 {
		ctr.HostConfig.ExtraHosts = cfg.ExtraHosts
	}
	if cfg.Devices != nil && len(cfg.Devices) > 0 {
		ctr.HostConfig.Devices = []types.DeviceMapping{}
		for _, dev := range cfg.Devices {
			ctr.HostConfig.Devices = append(ctr.HostConfig.Devices, toAPIDevice(dev))
		}
	}

	if cfg.PortMappings != nil && len(cfg.PortMappings) > 0 {
		ctr.HostConfig.PortMappings = []types.PortMapping{}
		for _, pm := range cfg.PortMappings {
			ctr.HostConfig.PortMappings = append(ctr.HostConfig.PortMappings, toAPIPortMapping(pm))
		}
	}
	if cfg.MountPoints != nil && len(cfg.MountPoints) > 0 {
		ctr.Mounts = []types.MountPoint{}
		for _, mp := range cfg.MountPoints {
			ctr.Mounts = append(ctr.Mounts, toAPIMountPoint(mp))
		}
	}
	if cfg.Log != nil {
		ctr.HostConfig.LogConfig = toAPILogConfiguration(cfg.Log)
	}
	if cfg.Resources != nil {
		ctr.HostConfig.Resources = toAPIResources(cfg.Resources)
	}
	if (cfg.Env != nil && len(cfg.Env) > 0) || (cfg.Cmd != nil && len(cfg.Cmd) > 0) {
		ctr.Config = &types.ContainerConfiguration{
			Env: cfg.Env,
			Cmd: cfg.Cmd,
		}
	}
	if cfg.Decryption != nil {
		ctr.Image.DecryptConfig = toAPIDecryptionConfiguration(cfg.Decryption)
	}
	if len(cfg.HostName) > 0 {
		ctr.HostName = cfg.HostName
	}
	ctr.HostConfig.NetworkMode = cfg.NetworkMode.toAPINetworkMode()
	return ctr
}
