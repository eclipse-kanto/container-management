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

package ctr

import (
	"math"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/util"

	statsV1 "github.com/containerd/cgroups/stats/v1"
	statsV2 "github.com/containerd/cgroups/v2/stats"
	ctrdTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	protoTypes "github.com/gogo/protobuf/types"
)

func TestToMetrics(t *testing.T) {
	const (
		currentPIDs     = uint64(65)
		cpuUsage        = uint64(16643328222)
		memUsage        = uint64(462376960)
		memLimit        = uint64(1073741824)
		memInactiveFile = uint64(512)
		readBytes       = uint64(1024)
		writeBytes      = uint64(2048)
	)

	var (
		now = time.Now()

		v1Metrics = &statsV1.Metrics{
			Pids: &statsV1.PidsStat{Current: currentPIDs},
			CPU: &statsV1.CPUStat{
				Usage: &statsV1.CPUUsage{
					Total: cpuUsage * 1000,
				},
			},
			Memory: &statsV1.MemoryStat{
				TotalInactiveFile: memInactiveFile,
				Usage: &statsV1.MemoryEntry{
					Limit: memLimit,
					Usage: memUsage,
				},
			},
			Blkio: &statsV1.BlkIOStat{
				IoServiceBytesRecursive: []*statsV1.BlkIOEntry{
					{Op: "Read", Value: readBytes},
					{Op: "read", Value: readBytes},
					{Op: "Write", Value: writeBytes},
					{Op: "write", Value: writeBytes},
					{Op: "", Value: 4096},
				},
			},
		}
		v2Metrics = &statsV2.Metrics{
			Pids: &statsV2.PidsStat{Current: currentPIDs},
			CPU: &statsV2.CPUStat{
				UsageUsec: cpuUsage,
			},
			Memory: &statsV2.MemoryStat{
				InactiveFile: memInactiveFile,
				UsageLimit:   memLimit,
				Usage:        memUsage,
			},
			Io: &statsV2.IOStat{
				Usage: []*statsV2.IOEntry{
					{Rbytes: readBytes, Wbytes: writeBytes},
					{Rbytes: readBytes, Wbytes: writeBytes},
				},
			},
		}
		testCPU = &types.CPUStats{
			Used: cpuUsage * 1000,
		}
		testMemory = &types.MemoryStats{
			Total: memLimit,
			Used:  memUsage - memInactiveFile,
		}
		testIO = &types.IOStats{
			Read:  2 * readBytes,
			Write: 2 * writeBytes,
		}
		testPIDs = currentPIDs
		testTime = now

		m = &ctrdTypes.Metric{
			Timestamp: now,
			ID:        testContainerID,
			Data:      nil,
		}
	)

	expectedMachineMemLimit := func() *types.MemoryStats {
		return &types.MemoryStats{
			Total: util.GetMemoryLimit(math.MaxUint64),
			Used:  testMemory.Used,
		}
	}()

	tests := map[string]struct {
		data           interface{}
		expectedCPU    *types.CPUStats
		expectedMemory *types.MemoryStats
		expectedIO     *types.IOStats
		expectedPIDs   uint64
		expectedTime   time.Time

		expectedErr error
	}{
		"test_valid_v1_data": {
			data:           v1Metrics,
			expectedCPU:    testCPU,
			expectedMemory: testMemory,
			expectedIO:     testIO,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
		"test_valid_v2_data": {
			data:           v2Metrics,
			expectedCPU:    testCPU,
			expectedMemory: testMemory,
			expectedIO:     testIO,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
		"test_error_invalid_data": {
			data:        &protoTypes.Any{},
			expectedErr: log.NewErrorf("type with url : not found"),
		},
		"test_error_invalid_data_type": {
			data:        v1Metrics.Memory,
			expectedErr: log.NewErrorf("unexpected metrics type = %T for container with ID = %s", v1Metrics.Memory, testContainerID),
		},
		"test_valid_v1_data_without_blkio": {
			data: func() interface{} {
				withoutBlkio := *v1Metrics
				withoutBlkio.Blkio = nil
				return &withoutBlkio
			}(),
			expectedCPU:    testCPU,
			expectedMemory: testMemory,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
		"test_valid_v2_data_without_io": {
			data: func() interface{} {
				withoutIO := *v2Metrics
				withoutIO.Io = nil
				return &withoutIO
			}(),
			expectedCPU:    testCPU,
			expectedMemory: testMemory,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
		"test_empty_v1_data": {
			data:         &statsV1.Metrics{},
			expectedTime: testTime,
		},
		"test_empty_v2_data": {
			data:         &statsV2.Metrics{},
			expectedTime: testTime,
		},
		"test_valid_v1_data_without_memory_limit": {
			data: func() interface{} {
				withoutMemLimit := *v1Metrics
				withoutMemLimit.Memory = &statsV1.MemoryStat{
					TotalInactiveFile: v1Metrics.Memory.TotalInactiveFile,
					Usage: &statsV1.MemoryEntry{
						Usage: v1Metrics.Memory.Usage.Usage,
						Limit: 9223372036854771712,
					},
				}
				return &withoutMemLimit
			}(),
			expectedCPU:    testCPU,
			expectedMemory: expectedMachineMemLimit,
			expectedIO:     testIO,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
		"test_valid_v2_data_without_memory_limit": {
			data: func() interface{} {
				withoutMemLimit := *v2Metrics
				withoutMemLimit.Memory = &statsV2.MemoryStat{
					InactiveFile: v2Metrics.Memory.InactiveFile,
					Usage:        v2Metrics.Memory.Usage,
					UsageLimit:   9223372036854771712,
				}
				return &withoutMemLimit
			}(),
			expectedCPU:    testCPU,
			expectedMemory: expectedMachineMemLimit,
			expectedIO:     testIO,
			expectedPIDs:   testPIDs,
			expectedTime:   testTime,
			expectedErr:    nil,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			d, e := typeurl.MarshalAny(testCase.data)
			if e != nil {
				t.Fatal(e)
			}
			m.Data = d
			cpu, mem, io, pids, time, err := toMetrics(m, testContainerID)
			testutil.AssertError(t, testCase.expectedErr, err)
			if testCase.expectedCPU != nil {
				testutil.AssertNotNil(t, cpu)
				testutil.AssertEqual(t, testCase.expectedCPU.Used, cpu.Used)
				testutil.AssertTrue(t, cpu.Total > 0)
			} else {
				testutil.AssertNil(t, cpu)
			}
			testutil.AssertEqual(t, testCase.expectedMemory, mem)
			testutil.AssertEqual(t, testCase.expectedIO, io)
			testutil.AssertEqual(t, testCase.expectedPIDs, pids)
			testutil.AssertEqual(t, testCase.expectedTime, time)
		})
	}
}
