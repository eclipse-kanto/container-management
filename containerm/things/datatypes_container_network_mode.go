// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

import (
	"encoding/json"
	"fmt"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

type networkMode uint8

const (
	bridge networkMode = iota
	host
)

// String representation of the networkMode
func (n networkMode) String() string {
	return string(n.toAPINetworkMode())
}

func (n networkMode) toAPINetworkMode() types.NetworkMode {
	return []types.NetworkMode{types.NetworkModeBridge, types.NetworkModeHost}[n]
}

func fromAPINetworkMode(n types.NetworkMode) networkMode {
	return map[types.NetworkMode]networkMode{types.NetworkModeBridge: bridge, types.NetworkModeHost: host}[n]
}

func (n *networkMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *networkMode) UnmarshalJSON(data []byte) error {
	nmString := ""
	if err := json.Unmarshal(data, &nmString); err != nil {
		return err
	}
	switch nmString {
	case bridge.String():
		*n = bridge
	case host.String():
		*n = host
	default:
		return fmt.Errorf("invalid network mode %s", nmString)
	}
	return nil
}
