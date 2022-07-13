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

package ctr

import (
	"bufio"
	statsV1 "github.com/containerd/cgroups/stats/v1"
	statsV2 "github.com/containerd/cgroups/v2/stats"
	ctrdTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/docker/docker/pkg/system"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"os"
	"strconv"
	"strings"
	"sync"
)

func toMetrics(ctrdMetrics *ctrdTypes.Metric) (*types.Metrics, error) {
	var (
		metrics     *types.Metrics
		metricsData interface{}
		err         error
	)

	if metricsData, err = typeurl.UnmarshalAny(ctrdMetrics.Data); err != nil {
		return nil, err
	}

	switch metricsData.(type) {
	case *statsV1.Metrics:
		metrics = toMetricsV1(metricsData.(*statsV1.Metrics))
	case *statsV2.Metrics:
		metrics = toMetricsV2(metricsData.(*statsV2.Metrics))
	default:
		return nil, log.NewErrorf("unexpected metrics type = %T ", metricsData)
	}

	metrics.Timestamp = ctrdMetrics.Timestamp
	return metrics, nil
}

func toMetricsV1(ctrdMetrics *statsV1.Metrics) *types.Metrics {
	metrics := &types.Metrics{
		IO: calculateBlkIO(ctrdMetrics.Blkio),
	}
	if ctrdMetrics.Pids != nil {
		metrics.PIDs = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil && ctrdMetrics.CPU.Usage != nil {
		if systemTotal, err := getSystemCPUUsage(); err == nil {
			metrics.CPU = &types.CPUStats{
				Total:       ctrdMetrics.CPU.Usage.Total,
				SystemTotal: systemTotal,
			}
		} else {
			log.WarnErr(err, "could not get system CPU usage")
		}
	}
	if ctrdMetrics.Memory != nil && ctrdMetrics.Memory.Usage != nil {
		metrics.Memory = &types.MemoryStats{
			Used:  ctrdMetrics.Memory.Usage.Usage - ctrdMetrics.Memory.TotalInactiveFile,
			Total: getMemoryLimit(ctrdMetrics.Memory.Usage.Limit),
		}
	}
	return metrics
}

func toMetricsV2(ctrdMetrics *statsV2.Metrics) *types.Metrics {
	metrics := &types.Metrics{
		IO: calculateIO(ctrdMetrics.Io),
	}
	if ctrdMetrics.Pids != nil {
		metrics.PIDs = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil {
		if systemTotal, err := getSystemCPUUsage(); err == nil {
			metrics.CPU = &types.CPUStats{
				Total:       ctrdMetrics.CPU.UsageUsec * 1000,
				SystemTotal: systemTotal,
			}
		} else {
			log.WarnErr(err, "could not get system CPU usage")
		}
	}
	if ctrdMetrics.Memory != nil {
		metrics.Memory = &types.MemoryStats{
			Used:  ctrdMetrics.Memory.Usage - ctrdMetrics.Memory.InactiveFile,
			Total: getMemoryLimit(ctrdMetrics.Memory.UsageLimit),
		}
	}
	return metrics
}

var (
	machineMemory       uint64
	detectMachineMemory sync.Once
)

func getMachineMemory() (memory uint64, err error) {
	detectMachineMemory.Do(func() {
		var vm *system.MemInfo
		if vm, err = system.ReadMemInfo(); err == nil {
			if vm.MemTotal > 0 {
				memory = uint64(vm.MemTotal)
			} else {
				err = log.NewErrorf("unexpected value for machine memory: %d", vm.MemTotal)
			}
		}
	})
	return
}

func getMemoryLimit(limit uint64) uint64 {
	if limit > machineMemory && machineMemory > 0 {
		return machineMemory
	}
	return limit
}

func calculateIO(io *statsV2.IOStat) *types.IOStats {
	if io == nil || io.Usage == nil {
		return nil
	}
	var read, write uint64
	for _, entry := range io.Usage {
		read = read + entry.Rbytes
		write = write + entry.Wbytes
	}
	return &types.IOStats{Read: read, Write: write}
}

// Copyright The PouchContainer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// getSystemCPUUsage returns the host system's cpu usage in
// nanoseconds. An error is returned if the format of the underlying
// file does not match.
//
// Uses /proc/stat defined by POSIX. Looks for the cpu
// statistics line and then sums up the first seven fields
// provided. See `man 5 proc` for details on specific field
// information.
// Package name changed also removed not needed logic and added custom code to handle the specific use case, Eclipse Kanto contributors, 2022
func getSystemCPUUsage() (uint64, error) {
	const nanoSecondsPerSecond = 1e9

	// C.sysconf(C._SC_CLK_TCK), on Linux it's a constant which is safe to be hard coded
	// See https://github.com/containerd/cgroups/pull/12
	const clockTicks = 100

	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	bufReader := bufio.NewReaderSize(nil, 128)
	defer func() {
		bufReader.Reset(nil)
		f.Close()
	}()
	bufReader.Reset(f)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.Fields(line)
		switch parts[0] {
		case "cpu":
			if len(parts) < 8 {
				return 0, log.NewError("invalid number of cpu fields")
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return 0, log.NewErrorf("unable to convert value %s to int: %s", i, err)
				}
				totalClockTicks += v
			}
			return (totalClockTicks * nanoSecondsPerSecond) /
				clockTicks, nil
		}
	}
	return 0, log.NewErrorf("invalid stat format, fail to parse the '/proc/stat' file")
}

func calculateBlkIO(blkio *statsV1.BlkIOStat) *types.IOStats {
	if blkio == nil || blkio.IoServiceBytesRecursive == nil {
		return nil
	}
	var read, write uint64
	for _, entry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(entry.Op) {
		case "read":
			read += entry.Value
		case "write":
			write += entry.Value
		}
	}
	return &types.IOStats{Read: read, Write: write}
}
