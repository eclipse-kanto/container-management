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

// MemoryUnlimited - no memory constraint
const MemoryUnlimited = "-1"

// Resources of the container
type Resources struct {

	// Hard memory usage limit
	Memory string `json:"memory,omitempty"`

	// Soft memory usage limit
	MemoryReservation string `json:"memory_reservation,omitempty"`

	// Swap + memory usage limit
	MemorySwap string `json:"memory_swap,omitempty"`
}
