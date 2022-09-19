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
	"encoding/json"
	statsV1 "github.com/containerd/cgroups/stats/v1"
	statsV2 "github.com/containerd/cgroups/v2/stats"
	ctrdTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"strings"
	"time"
)

func toMetrics(ctrdMetrics *ctrdTypes.Metric, ctrID string) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
	var (
		metricsData interface{}
		err         error
	)

	if metricsData, err = typeurl.UnmarshalAny(ctrdMetrics.Data); err != nil {
		return nil, nil, nil, 0, time.Time{}, err
	}

	switch metricsData.(type) {
	case *statsV1.Metrics:
		m := metricsData.(*statsV1.Metrics)
		data, _ := json.Marshal(m) // type is checked and error is not expected, an error is not critical as it is used only for the debug log
		log.Debug("metrics of a container with ID = %s: %s", ctrID, string(data))
		cpu, mem, io, pids := toMetricsV1(m, ctrID)
		return cpu, mem, io, pids, ctrdMetrics.Timestamp, nil
	case *statsV2.Metrics:
		m := metricsData.(*statsV2.Metrics)
		data, _ := json.Marshal(m) // type is checked and error is not expected, an error is not critical as it is used only for the debug log
		log.Debug("metrics of a container with ID = %s: %s", ctrID, string(data))
		cpu, mem, io, pids := toMetricsV2(m, ctrID)
		return cpu, mem, io, pids, ctrdMetrics.Timestamp, nil
	default:
		return nil, nil, nil, 0, time.Time{}, log.NewErrorf("unexpected metrics type = %T for container with ID = %s", metricsData, ctrID)
	}
}

func toMetricsV1(ctrdMetrics *statsV1.Metrics, ctrID string) (cpu *types.CPUStats, mem *types.MemoryStats, io *types.IOStats, pids uint64) {
	io = calculateBlkIO(ctrdMetrics.Blkio)
	if ctrdMetrics.Pids != nil {
		pids = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil && ctrdMetrics.CPU.Usage != nil {
		cpu = &types.CPUStats{
			Used: ctrdMetrics.CPU.Usage.Total,
		}
		var err error
		if cpu.Total, err = util.GetSystemCPUUsage(); err != nil {
			log.WarnErr(err, "could not get system CPU usage for metrics of a container with ID = %s", ctrID)
		}
	}
	if ctrdMetrics.Memory != nil && ctrdMetrics.Memory.Usage != nil {
		mem = &types.MemoryStats{
			Used:  ctrdMetrics.Memory.Usage.Usage - ctrdMetrics.Memory.TotalInactiveFile,
			Total: util.GetMemoryLimit(ctrdMetrics.Memory.Usage.Limit),
		}
	}
	return
}

func toMetricsV2(ctrdMetrics *statsV2.Metrics, ctrID string) (cpu *types.CPUStats, mem *types.MemoryStats, io *types.IOStats, pids uint64) {
	io = calculateIO(ctrdMetrics.Io)
	if ctrdMetrics.Pids != nil {
		pids = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil {
		cpu = &types.CPUStats{
			Used: uint64((time.Duration(ctrdMetrics.CPU.UsageUsec) * time.Microsecond).Nanoseconds()),
		}
		var err error
		if cpu.Total, err = util.GetSystemCPUUsage(); err != nil {
			log.WarnErr(err, "could not get system CPU usage for metrics of a container with ID = %s", ctrID)
		}
	}
	if ctrdMetrics.Memory != nil {
		mem = &types.MemoryStats{
			Used:  ctrdMetrics.Memory.Usage - ctrdMetrics.Memory.InactiveFile,
			Total: util.GetMemoryLimit(ctrdMetrics.Memory.UsageLimit),
		}
	}
	return
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
