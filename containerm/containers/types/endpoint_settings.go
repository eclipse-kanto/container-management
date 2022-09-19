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

// EndpointSettings represents an endpoint settings for connecting to a specific network
type EndpointSettings struct {
	ID         string `json:"id,omitempty"`
	Gateway    string `json:"gateway"`
	IPAddress  string `json:"ip_address"`
	MacAddress string `json:"mac_address"`
	NetworkID  string `json:"network_id,omitempty"`
}
