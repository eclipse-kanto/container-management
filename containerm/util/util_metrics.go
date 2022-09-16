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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math"
	"sync"
	"time"
)

// CalculateCPUPercent calculates the CPU percentage in range [0-100].
// Returns error if there is missing data needed for the calculation.
func CalculateCPUPercent(cpu *types.CPUStats, previousCPU *types.CPUStats) (float64, error) {
	if cpu != nil && previousCPU != nil {
		if cpu.Total == 0 || previousCPU.Total == 0 {
			return 0, log.NewErrorf("no total system CPU usage")
		}
		cpuDelta := float64(cpu.Used) - float64(previousCPU.Used)
		systemDelta := float64(cpu.Total) - float64(previousCPU.Total)
		if systemDelta > 0 {
			return math.Min(100, math.Max(0, cpuDelta/systemDelta*100)), nil
		}
		return 0, log.NewErrorf("unexpected system CPU delta: %f", systemDelta)
	}
	return 0, log.NewErrorf("no CPU data")
}

// CalculateMemoryPercent calculates the memory percentage.
// Returns error if there is missing data needed for the calculation.
func CalculateMemoryPercent(memory *types.MemoryStats) (float64, error) {
	if memory != nil && memory.Total != 0 {
		percent := float64(memory.Used) / float64(memory.Total) * 100.0
		return math.Min(100, math.Max(0, percent)), nil
	}
	return 0, log.NewErrorf("no memory data: %+v", memory)
}

var (
	machineMemory       uint64
	detectMachineMemory sync.Once
)

// GetMemoryLimit takes a limit in bytes. If the machine memory is read successfully and is lower than
// the provided limit, returns the machine memory in bytes. Otherwise, returns the provided limit.
func GetMemoryLimit(limit uint64) uint64 {
	detectMachineMemory.Do(func() {
		var (
			vm  *mem.VirtualMemoryStat
			err error
		)
		if vm, err = mem.VirtualMemory(); err != nil {
			log.ErrorErr(err, "could not get machine memory")
		} else if vm.Total <= 0 {
			log.Error("unexpected value for machine memory: %d", vm.Total)
		} else {
			machineMemory = vm.Total
		}
	})
	if limit > machineMemory && machineMemory > 0 {
		return machineMemory
	}
	return limit
}

// GetSystemCPUUsage returns the system CPU usage as nanoseconds the CPU has spent performing different work.
// Returns error if cpu times could not be read.
func GetSystemCPUUsage() (uint64, error) {
	if times, err := cpu.Times(false); err == nil {
		aggregated := times[0]
		total := time.Duration(aggregated.User+aggregated.System+aggregated.Idle+aggregated.Nice+aggregated.Iowait+aggregated.Irq+
			aggregated.Softirq+aggregated.Steal) * time.Second
		return uint64(total.Nanoseconds()), nil
	}
	return 0, log.NewErrorf("could not get system CPU usage")
}
