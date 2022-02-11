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

import (
	"sync"
)

// Container represents the container instance
type Container struct {
	sync.Mutex

	// ID the ID is system-internally generated
	ID string `json:"container_id"`
	// Name is a user-defined name of the container - the ID is set if none is provided
	Name string `json:"container_name"`
	// Image is the image information for the container
	Image Image `json:"image"`
	// DomainName is the domain name set inside the container
	DomainName string `json:"domain_name"`
	// HostName is the hostname for the container
	HostName string `json:"host_name"`
	// ResolvConfPath is the path to the container's resolv.conf file
	ResolvConfPath string `json:"resolv_conf_path,omitempty"`
	// HostsPath is the path to the container's hosts file
	HostsPath string `json:"hosts_path,omitempty"`
	// HostnamePath is the path to the container's hostname file
	HostnamePath string `json:"hostname_path,omitempty"`
	// Mounts is the mounts for the container
	Mounts []MountPoint `json:"mount_points"`
	// Hooks is to perform on container start/stop, etc.
	Hooks []Hook `json:"hooks"`
	// Config is the configuration of the container's root process
	Config *ContainerConfiguration `json:"config"`
	// HostConfig is the host configuration for the container
	HostConfig *HostConfig `json:"host_config"`
	// IOConfig is the IO configuration for the container
	IOConfig *IOConfig `json:"io_config"`
	// NetworkSettings is the network settings for the container
	NetworkSettings *NetworkSettings `json:"network_settings"`
	// State is the container's state
	State *State `json:"state"`
	// Created is the time of the container's creation
	Created string `json:"created,omitempty"`
	// RestartCount is the metric for the container showing how many restart retries have been performed on it
	RestartCount int `json:"restart_count"`
	// ManuallyStopped is the flag indicating whether the container has been manually stopped or internally by the system
	ManuallyStopped bool `json:"manually_stopped"`
	// StartedSuccessfullyBefore is the flag indicating if the container has ever been started successfully before
	StartedSuccessfullyBefore bool `json:"started_successfully_before"`
}
