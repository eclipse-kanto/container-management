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

package things

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type resources struct {
	Memory            string `json:"memory,omitempty"`
	MemoryReservation string `json:"memoryReservation,omitempty"`
	MemorySwap        string `json:"memorySwap,omitempty"`
}

func toAPIResources(r *resources) *types.Resources {
	return &types.Resources{
		Memory:            r.Memory,
		MemoryReservation: r.MemoryReservation,
		MemorySwap:        r.MemorySwap,
	}
}

func fromAPIResources(r *types.Resources) *resources {
	return &resources{
		Memory:            r.Memory,
		MemoryReservation: r.MemoryReservation,
		MemorySwap:        r.MemorySwap,
	}
}
