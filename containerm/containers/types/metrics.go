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

// Metrics container metrics.
type Metrics struct {
	CPU       *CPUMetrics    `json:"cpu,omitempty"`
	Memory    *MemoryMetrics `json:"memory,omitempty"`
	IO        *IOMetrics     `json:"io,omitempty"`
	Network   *IOMetrics     `json:"network,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	PIDs      uint64         `json:"pids"`
}

// CPUMetrics container stats regarding CPU.
type CPUMetrics struct {
	// Total is the total system CPU time in nanoseconds.
	Total uint64 `json:"total"`
	// Used is container's processes CPU time in nanoseconds.
	Used uint64 `json:"used"`
}

// MemoryMetrics container stats regarding Memory.
type MemoryMetrics struct {
	// Total is the container memory limit in bytes.
	// If container does not have memory limit set, machine memory is used.
	Total uint64 `json:"total"`
	// Used memory used by a container in bytes.
	Used uint64 `json:"used"`
}

// IOMetrics container stats regarding IO.
type IOMetrics struct {
	// Read is the number of bytes that has been read.
	Read uint64 `json:"read"`
	// Write is the number of bytes that has been written.
	Write uint64 `json:"write"`
}