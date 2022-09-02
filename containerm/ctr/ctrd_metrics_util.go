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
	"encoding/json"
	statsV1 "github.com/containerd/cgroups/stats/v1"
	statsV2 "github.com/containerd/cgroups/v2/stats"
	ctrdTypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/typeurl"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"strings"
)

func toMetrics(ctrdMetrics *ctrdTypes.Metric, ctrID string) (*types.Metrics, error) {
	var (
		metrics     *types.Metrics
		metricsData interface{}
		err         error
	)

	if metricsData, err = typeurl.UnmarshalAny(ctrdMetrics.Data); err != nil {
		return nil, err
	}

	debugArgsFunc := func(m interface{}) log.ArgsFunction {
		data, _ := json.Marshal(m)
		return func() []interface{} {
			return []interface{}{ctrID, data}
		}
	}
	switch metricsData.(type) {
	case *statsV1.Metrics:
		m := metricsData.(*statsV1.Metrics)
		log.DebugFn("metrics of a container with ID = %s: %s", debugArgsFunc(m))
		metrics = toMetricsV1(m, ctrID)
	case *statsV2.Metrics:
		m := metricsData.(*statsV2.Metrics)
		log.DebugFn("metrics of a container with ID = %s: %s", debugArgsFunc(m))
		metrics = toMetricsV2(m, ctrID)
	default:
		return nil, log.NewErrorf("unexpected metrics type = %T for container with ID = %s", metricsData, ctrID)
	}

	metrics.Timestamp = ctrdMetrics.Timestamp
	return metrics, nil
}

func toMetricsV1(ctrdMetrics *statsV1.Metrics, ctrID string) *types.Metrics {
	metrics := &types.Metrics{
		IO: calculateBlkIO(ctrdMetrics.Blkio),
	}
	if ctrdMetrics.Pids != nil {
		metrics.PIDs = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil && ctrdMetrics.CPU.Usage != nil {
		metrics.CPU = &types.CPUMetrics{
			Used: ctrdMetrics.CPU.Usage.Total,
		}
		var err error
		if metrics.CPU.Total, err = util.GetSystemCPUUsage(); err != nil {
			log.WarnErr(err, "could not get system CPU usage for metrics of a container with ID = %s", ctrID)
		}
	}
	if ctrdMetrics.Memory != nil && ctrdMetrics.Memory.Usage != nil {
		metrics.Memory = &types.MemoryMetrics{
			Used:  ctrdMetrics.Memory.Usage.Usage - ctrdMetrics.Memory.TotalInactiveFile,
			Total: util.GetMemoryLimit(ctrdMetrics.Memory.Usage.Limit),
		}
	}
	return metrics
}

func toMetricsV2(ctrdMetrics *statsV2.Metrics, ctrID string) *types.Metrics {
	data, _ := json.Marshal(ctrdMetrics)
	log.Debug("metrics of a container with ID = %s: %s", ctrID, string(data))
	metrics := &types.Metrics{
		IO: calculateIO(ctrdMetrics.Io),
	}
	if ctrdMetrics.Pids != nil {
		metrics.PIDs = ctrdMetrics.Pids.Current
	}
	if ctrdMetrics.CPU != nil {
		metrics.CPU = &types.CPUMetrics{
			Used: ctrdMetrics.CPU.UsageUsec * 1000,
		}
		var err error
		if metrics.CPU.Total, err = util.GetSystemCPUUsage(); err != nil {
			log.WarnErr(err, "could not get system CPU usage for metrics of a container with ID = %s", ctrID)
		}
	}
	if ctrdMetrics.Memory != nil {
		metrics.Memory = &types.MemoryMetrics{
			Used:  ctrdMetrics.Memory.Usage - ctrdMetrics.Memory.InactiveFile,
			Total: util.GetMemoryLimit(ctrdMetrics.Memory.UsageLimit),
		}
	}
	return metrics
}

func calculateIO(io *statsV2.IOStat) *types.IOMetrics {
	if io == nil || io.Usage == nil {
		return nil
	}
	var read, write uint64
	for _, entry := range io.Usage {
		read = read + entry.Rbytes
		write = write + entry.Wbytes
	}
	return &types.IOMetrics{Read: read, Write: write}
}

func calculateBlkIO(blkio *statsV1.BlkIOStat) *types.IOMetrics {
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
	return &types.IOMetrics{Read: read, Write: write}
}
