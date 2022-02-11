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

package types

// NetworkMode represents the network mode for the container
type NetworkMode string

const (
	// NetworkModeBridge means that the container is connected to the default bridge network interface of the engine and is assigned an IP
	NetworkModeBridge NetworkMode = "bridge"
	// NetworkModeHost means that the container shares the network stack of the host
	NetworkModeHost NetworkMode = "host"
)

// HostConfig defines the resources, behavior, etc. that the host must manage on the container
type HostConfig struct {
	Devices       []DeviceMapping   `json:"devices"`
	NetworkMode   NetworkMode       `json:"network_mode"`
	Privileged    bool              `json:"privileged"`
	RestartPolicy *RestartPolicy    `json:"restart_policy"`
	Runtime       string            `json:"runtime"`
	ExtraHosts    []string          `json:"extra_hosts"`
	PortMappings  []PortMapping     `json:"port_mappings"`
	LogConfig     *LogConfiguration `json:"log_config"`
	Resources     *Resources        `json:"resources"`
}
