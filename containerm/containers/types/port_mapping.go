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

// PortMapping mappings from the host to a container
type PortMapping struct {
	// Protocol
	Proto string `json:"proto"`
	// Container port
	ContainerPort uint16 `json:"container_port"`
	// Host IP
	HostIP string `json:"host_ip"`
	// Host port
	HostPort uint16 `json:"host_port"`
	// Host port
	HostPortEnd uint16 `json:"host_port_end"`
}
