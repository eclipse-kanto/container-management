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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

type networkMode string

const (
	bridge networkMode = "BRIDGE"
	host   networkMode = "HOST"
)

func (n networkMode) toAPINetworkMode() types.NetworkMode {
	switch n {
	case bridge:
		return types.NetworkModeBridge
	case host:
		return types.NetworkModeHost
	default:
		return types.NetworkMode(n)
	}
}

func fromAPINetworkMode(n types.NetworkMode) networkMode {
	switch n {
	case types.NetworkModeBridge:
		return bridge
	case types.NetworkModeHost:
		return host
	default:
		return networkMode(n)
	}
}
