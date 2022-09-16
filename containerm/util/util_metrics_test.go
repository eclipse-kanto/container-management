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
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"math"
	"testing"
)

func TestCalculateCPUPercent(t *testing.T) {
	testCPUStats := &types.CPUStats{
		Total: 100,
		Used:  15,
	}
	tests := map[string]struct {
		cpu           *types.CPUStats
		prevCPU       *types.CPUStats
		expectedValue float64
		expectedErr   error
	}{
		"test_err_cpu_is_nil": {
			prevCPU:     testCPUStats,
			expectedErr: log.NewErrorf("no CPU data"),
		},
		"test_err_previous_cpu_is_nil": {
			cpu:         testCPUStats,
			expectedErr: log.NewErrorf("no CPU data"),
		},
		"test_err_no_cpu_delta": {
			cpu:         testCPUStats,
			prevCPU:     testCPUStats,
			expectedErr: log.NewErrorf("unexpected system CPU delta: %f", float64(0)),
		},
		"test_err_no_cpu_total_data": {
			cpu:         &types.CPUStats{Used: 5, Total: 0},
			prevCPU:     &types.CPUStats{Used: 10, Total: 0},
			expectedErr: log.NewErrorf("no total system CPU usage"),
		},
		"test_cpu_20_percents": {
			cpu:           testCPUStats,
			prevCPU:       &types.CPUStats{Used: 5, Total: 50},
			expectedValue: 20,
		},
		"test_cpu_percent_is_in range": {
			cpu:           testCPUStats,
			prevCPU:       &types.CPUStats{Total: 90},
			expectedValue: 100,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual, actualErr := CalculateCPUPercent(testCase.cpu, testCase.prevCPU)
			testutil.AssertError(t, testCase.expectedErr, actualErr)
			testutil.AssertEqual(t, testCase.expectedValue, actual)
		})
	}
}

func TestCalculateMemoryPercent(t *testing.T) {
	tests := map[string]struct {
		memory        *types.MemoryStats
		expectedValue float64
		expectedErr   error
	}{
		"test_err_memory_is_nil": {
			expectedErr: log.NewErrorf("no memory data: %+v", nil),
		},
		"test_err_memory_total_is_zero": {
			memory: &types.MemoryStats{
				Total: 0,
				Used:  30,
			},
			expectedErr: log.NewErrorf("no memory data: %+v", &types.MemoryStats{Total: 0, Used: 30}),
		},
		"test_memory_30_percents": {
			memory: &types.MemoryStats{
				Total: 100,
				Used:  30,
			},
			expectedValue: 30,
		},
		"test_memory_percent_is_in range": {
			memory: &types.MemoryStats{
				Total: 100,
				Used:  120,
			},
			expectedValue: 100,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual, actualErr := CalculateMemoryPercent(testCase.memory)
			testutil.AssertError(t, testCase.expectedErr, actualErr)
			testutil.AssertEqual(t, testCase.expectedValue, actual)
		})
	}
}

func TestGetMemoryLimit(t *testing.T) {
	tests := map[string]struct {
		limit uint64
		equal bool
	}{
		"test_memory_limit_below_machine_memory": {
			limit: 100 * 1024 * 1024,
			equal: true,
		},
		"test_memory_limit_above_machine_memory": {
			limit: math.MaxUint64,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			if testCase.equal {
				testutil.AssertEqual(t, testCase.limit, GetMemoryLimit(testCase.limit))
			} else {
				testutil.AssertNotEqual(t, testCase.limit, GetMemoryLimit(testCase.limit))
			}
		})
	}
}

func TestGetSystemCPUUsage(t *testing.T) {
	var (
		cpu1, cpu2 uint64
		err        error
	)

	cpu1, err = GetSystemCPUUsage()
	testutil.AssertNil(t, err)

	cpu2, err = GetSystemCPUUsage()
	testutil.AssertNil(t, err)

	testutil.AssertTrue(t, cpu2 >= cpu1)
}
