// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package util

import (
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

// CalculateCPUPercent calculates the CPU percentage in range [0-100]
func CalculateCPUPercent(cpu *types.CPUStats, previousCPU *types.CPUStats) float64 {
	if cpu != nil && previousCPU != nil {
		cpuDelta := float64(cpu.Total) - float64(previousCPU.Total)
		systemDelta := float64(cpu.SystemTotal) - float64(previousCPU.SystemTotal)

		if systemDelta > 0.0 && cpuDelta > 0.0 {
			return cpuDelta / systemDelta * 100.0
		}
	}
	return 0
}

// CalculateMemoryPercent calculates the memory percentage
func CalculateMemoryPercent(memory *types.MemoryStats) float64 {
	if memory != nil && memory.Total != 0 {
		return float64(memory.Used) / float64(memory.Total) * 100.0
	}
	return 0
}
