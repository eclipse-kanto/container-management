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
	statsV1 "github.com/containerd/cgroups/stats/v1"
	statsV2 "github.com/containerd/cgroups/v2/stats"
	ctrdTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"strings"
	"sync"
	"time"
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

func getMemoryLimit(limit uint64) uint64 {
	detectMachineMemory.Do(func() {
		if vm, err := mem.VirtualMemory(); err == nil {
			if vm.Total > 0 {
				machineMemory = vm.Total
			} else {
				err = log.NewErrorf("unexpected value for machine memory: %d", vm.Total)
			}
		}
	})
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

func getSystemCPUUsage() (uint64, error) {
	if times, err := cpu.Times(false); err == nil {
		usage := uint64(total(times[0]) * float64(time.Second))
		return usage, nil
	}
	return 0, log.NewErrorf("could not get system CPU usage")
}

func total(time cpu.TimesStat) float64 {
	return time.User + time.System + time.Idle + time.Nice + time.Iowait + time.Irq +
		time.Softirq + time.Steal
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
