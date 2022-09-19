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

// NetworkSettings container network settings
type NetworkSettings struct {
	// networks
	Networks map[string]*EndpointSettings `json:"networks"`
	// underlying sandbox id
	SandboxID string `json:"sandbox_id,omitempty"`
	// underlying sandbox key
	SandboxKey string `json:"sandbox_key,omitempty"`
	// the underlying network controller id
	NetworkControllerID string `json:"network_controller_id,omitempty"`
}
