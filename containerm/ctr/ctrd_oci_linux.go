// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package ctr

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/containerd/containers"
	crtdoci "github.com/containerd/containerd/oci"
	"github.com/docker/docker/pkg/stringid"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/oci"
	"github.com/eclipse-kanto/container-management/containerm/oci/linux"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/opencontainers/runc/libcontainer/devices"
	rsystem "github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// WithCommonOptions sets common options:
// - hostname
func WithCommonOptions(c *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		//setup hostname in spec
		s.Hostname = c.HostName

		//setup env hostname
		if s.Process.Env == nil {
			s.Process.Env = []string{}
		}
		s.Process.Env = append(s.Process.Env, "HOSTNAME="+c.HostName)

		s.Process.Terminal = c.IOConfig.Tty

		if s.Process.Terminal {
			s.Process.Env = append(s.Process.Env, "TERM=xterm")
		}

		return nil
	}
}

// WithProcessOptions sets the container's root process configuration
// - env variables
// - etc.
func WithProcessOptions(c *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		if c.Config != nil && c.Config.Env != nil && len(c.Config.Env) > 0 {
			//setup env hostname
			if s.Process.Env == nil {
				s.Process.Env = []string{}
			}
			s.Process.Env = append(s.Process.Env, c.Config.Env...)
		}
		return nil
	}
}

// WithMounts sets the network resolution files generated
//e.g. c.getRootResourceDir("resolv.conf"), "hostname", "hosts"
func WithMounts(container *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		if s.Mounts == nil {
			s.Mounts = []specs.Mount{}
		}
		opts := []string{"rbind"}

		for _, mnt := range container.Mounts {
			optsMnt := append(opts, mnt.PropagationMode)
			s.Mounts = append(s.Mounts, specs.Mount{Destination: mnt.Destination, Source: mnt.Source, Type: "bind", Options: optsMnt})
		}
		// ensure network binds
		optsMnt := append(opts, types.RPrivatePropagationMode)
		s.Mounts = append(s.Mounts, specs.Mount{Destination: "/etc/resolv.conf", Source: container.ResolvConfPath, Type: "bind", Options: optsMnt})
		s.Mounts = append(s.Mounts, specs.Mount{Destination: "/etc/hostname", Source: container.HostnamePath, Type: "bind", Options: optsMnt})
		s.Mounts = append(s.Mounts, specs.Mount{Destination: "/etc/hosts", Source: container.HostsPath, Type: "bind", Options: optsMnt})

		// remove /run that is propagated automatically to tmpfs by the default spec generated from the image
		mpIdxToRemove := -1
		for idx, specMount := range s.Mounts {
			if specMount.Destination == "/run" {
				mpIdxToRemove = idx
				break
			}
		}
		if mpIdxToRemove != -1 {
			s.Mounts[mpIdxToRemove] = s.Mounts[len(s.Mounts)-1]
			s.Mounts = s.Mounts[:len(s.Mounts)-1]
		}
		return nil
	}
}

// WithNamespaces sets the enabled and desired namespaces to be used for the container's isolation.
func WithNamespaces(container *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		networkNamespace := specs.LinuxNamespace{Type: specs.NetworkNamespace}
		if util.IsContainerNetworkHost(container) {
			networkNamespace.Path = container.NetworkSettings.SandboxKey
		}
		for i, n := range s.Linux.Namespaces {
			if n.Type == networkNamespace.Type {
				s.Linux.Namespaces[i] = networkNamespace
				return nil
			}
		}
		s.Linux.Namespaces = append(s.Linux.Namespaces, networkNamespace)
		return nil
	}
}

// WithHooks sets the desired OCI hooks for the provided container instance.
func WithHooks(container *types.Container, execRoot string) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		if s.Hooks == nil {
			s.Hooks = &specs.Hooks{}
		}
		var specHook specs.Hook
		for _, cHook := range container.Hooks {
			specHook = specs.Hook{Path: cHook.Path, Args: cHook.Args, Env: cHook.Env /*, Timeout: &cHook.Timeout*/}

			switch cHook.Type {
			case types.HookTypePrestart:
				s.Hooks.Prestart = append(s.Hooks.Prestart, specHook)
			case types.HookTypePoststart:
				s.Hooks.Poststart = append(s.Hooks.Poststart, specHook)
			case types.HookTypePoststop:
				s.Hooks.Poststop = append(s.Hooks.Poststop, specHook)
			default:
				// should never get here since there are only 3 hooktypes - prestart, poststart, poststop
				log.Error("Invalid hook type")
			}
		}
		if container.NetworkSettings.NetworkControllerID != "" {
			// ensure networking via libnetwork for bridged containers only
			if util.IsContainerNetworkBridge(container) {
				target, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(os.Getpid()), "exe"))
				if err != nil {
					return err
				}
				libnetHook := specs.Hook{
					Path: target,
					Args: []string{
						"libnetwork-setkey",
						"-exec-root=" + execRoot,
						container.ID,
						stringid.TruncateID(container.NetworkSettings.NetworkControllerID),
					},
				}
				s.Hooks.Prestart = append(s.Hooks.Prestart, libnetHook)
			}
		}
		return nil
	}
}

// Copyright The PouchContainer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// WithDevices sets the available USB devices
// Package name changed also removed not needed logic and added custom code to handle the specific use case, Bosch.IO GmbH, 2020
func WithDevices(c *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		// Build lists of devices allowed and created within the container.
		var devs []specs.LinuxDevice
		devPermissions := s.Linux.Resources.Devices

		if c.HostConfig.Privileged && !rsystem.RunningInUserNS() {
			hostDevices, err := devices.HostDevices()
			if err != nil {
				return err
			}
			for _, d := range hostDevices {
				devs = append(devs, linux.Device(d))
			}
			devPermissions = []specs.LinuxDeviceCgroup{
				{
					Allow:  true,
					Access: "rwm",
				},
			}
		} else {
			for _, deviceMapping := range c.HostConfig.Devices {
				d, dPermissions, err := linux.DevicesFromPath(deviceMapping.PathOnHost, deviceMapping.PathInContainer, deviceMapping.CgroupPermissions)
				if err != nil {
					return err
				}
				devs = append(devs, d...)
				devPermissions = append(devPermissions, dPermissions...)
			}
			var err error
			devPermissions, err = oci.AppendDevicePermissionsFromCgroupRules(devPermissions /*TODO add device Cgroup Rules in container host configuration*/, []string{})
			if err != nil {
				return err
			}
		}

		s.Linux.Devices = append(s.Linux.Devices, devs...)
		s.Linux.Resources.Devices = devPermissions
		//TODO add GPU drivers handling!!
		return nil

	}
}

// WithResources sets container resource limitation:
func WithResources(c *types.Container) crtdoci.SpecOpts {
	return func(ctx context.Context, _ crtdoci.Client, _ *containers.Container, s *crtdoci.Spec) error {
		if c.HostConfig.Resources == nil {
			return nil
		}

		s.Linux.Resources.Memory = toLinuxMemory(c.HostConfig.Resources)
		return nil
	}
}

func parseMemoryValue(value string) *int64 {
	if value != "" {
		bytes, _ := util.SizeToBytes(value) // already validated
		return &bytes
	}
	// It is possible that a kernel does not support all memory constraints. Starting a container will fail,
	// if unsupported constraint is provided. Do not set any default value, unless provided.
	return nil
}

func toLinuxMemory(resources *types.Resources) *specs.LinuxMemory {
	if resources == nil {
		return nil
	}
	return &specs.LinuxMemory{
		Limit:       parseMemoryValue(resources.Memory),
		Reservation: parseMemoryValue(resources.MemoryReservation),
		Swap:        parseMemoryValue(resources.MemorySwap),
		// updates of swappiness and disableOOMKiller will take place only after the container is restarted
		// kernel resources limit is deprecated in cgroup v1 and cgroup v2 does not have support for kernel resources limit
	}
}
