// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl_2.0
//
// SPDX_License_Identifier: EPL_2.0

package types

import "time"

// Metrics represents all measurements of a container.
type Metrics struct {
	CPU       *CPUStats    `json:"cpu,omitempty"`
	Memory    *MemoryStats `json:"memory,omitempty"`
	IO        *IOStats     `json:"io,omitempty"`
	Network   *IOStats     `json:"network,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	PIDs      uint64       `json:"pids,omitempty"`
}

// CPUStats represents the CPU measurements of a container.
type CPUStats struct {
	// Total represents the total system CPU time in nanoseconds.
	Total uint64 `json:"total,omitempty"`
	// Used represents the container's processes CPU time in nanoseconds.
	Used uint64 `json:"used"`
}

// MemoryStats represents the memory measurements of a container.
type MemoryStats struct {
	// Total represents the container memory limit in bytes.
	// If container does not have memory limit set, machine memory is used.
	Total uint64 `json:"total"`
	// Used represents the memory used by a container in bytes.
	Used uint64 `json:"used"`
}

// IOStats represents the IO measurements of a container.
type IOStats struct {
	// Read represents the number of bytes that has been read.
	Read uint64 `json:"read"`
	// Write represents the number of bytes that has been written.
	Write uint64 `json:"write"`
}
