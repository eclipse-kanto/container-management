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

package types

// NetworkMode represents the network mode for the container
type NetworkMode string

// Runtime represents the runtime for the container
type Runtime string

const (
	// NetworkModeBridge means that the container is connected to the default bridge network interface of the engine and is assigned an IP
	NetworkModeBridge NetworkMode = "bridge"
	// NetworkModeHost means that the container shares the network stack of the host
	NetworkModeHost NetworkMode = "host"

	// RuntimeTypeV1 is the runtime type name for containerd shim interface v1 version.
	RuntimeTypeV1 Runtime = "io.containerd.runtime.v1.linux"
	// RuntimeTypeV2runscV1 is the runtime type name for gVisor containerd shim implement the shim v2 api.
	RuntimeTypeV2runscV1 Runtime = "io.containerd.runsc.v1"
	// RuntimeTypeV2kataV2 is the runtime type name for kata-runtime containerd shim implement the shim v2 api.
	RuntimeTypeV2kataV2 Runtime = "io.containerd.kata.v2"
	// RuntimeTypeV2runcV1 is the runtime type name for runc containerd shim implement the shim v2 api.
	RuntimeTypeV2runcV1 Runtime = "io.containerd.runc.v1"
	// RuntimeTypeV2runcV2 is the version 2 runtime type name for runc containerd shim implement the shim v2 api.
	RuntimeTypeV2runcV2 Runtime = "io.containerd.runc.v2"
)

// HostConfig defines the resources, behavior, etc. that the host must manage on the container
type HostConfig struct {
	Devices           []DeviceMapping   `json:"devices"`
	NetworkMode       NetworkMode       `json:"network_mode"`
	Privileged        bool              `json:"privileged"`
	RestartPolicy     *RestartPolicy    `json:"restart_policy"`
	Runtime           Runtime           `json:"runtime"`
	ExtraHosts        []string          `json:"extra_hosts"`
	ExtraCapabilities []string          `json:"extra_capabilities"`
	PortMappings      []PortMapping     `json:"port_mappings"`
	LogConfig         *LogConfiguration `json:"log_config"`
	Resources         *Resources        `json:"resources"`
}
