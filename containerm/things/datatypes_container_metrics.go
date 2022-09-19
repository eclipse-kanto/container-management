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

package things

import (
	"encoding/json"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"strings"
	"time"
)

const (
	// CPUUtilization measurements id
	CPUUtilization = "cpu.utilization"

	// MemoryUtilization measurements id
	MemoryUtilization = "memory.utilization"

	// MemoryTotal measurements id
	MemoryTotal = "memory.total"

	// MemoryUsed measurements id
	MemoryUsed = "memory.used"

	// IOReadBytes measurements id
	IOReadBytes = "io.readBytes"

	// IOWriteBytes measurements id
	IOWriteBytes = "io.writeBytes"

	// NetReadBytes measurements id
	NetReadBytes = "net.readBytes"

	// NetWriteBytes measurements id
	NetWriteBytes = "net.writeBytes"

	// PIDs number of pid
	PIDs = "pids"
)

// Filter defines the type of metric data to be reported.
type Filter struct {
	ID         []string `json:"id"`
	Originator string   `json:"originator"`
}

// Duration is used to support duration string un-marshalling to time.Duration.
type Duration struct {
	time.Duration
}

// UnmarshalJSON supports '50s' string format.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value) * time.Second
		return nil

	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil

	default:
		return log.NewErrorf("invalid duration: %v", v)
	}
}

// MarshalJSON supports marshalling to '50s' string format.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// MetricData contains a snapshot with all originators' measurements collected at a concrete time.
type MetricData struct {
	Snapshot  []OriginatorMeasurements `json:"snapshot"`
	Timestamp int64                    `json:"timestamp"`
}

// OriginatorMeasurements represents all the measurements collected per originator.
type OriginatorMeasurements struct {
	Originator   string        `json:"originator"`
	Measurements []Measurement `json:"measurements"`
}

// Measurement represents a measured value per metric ID.
type Measurement struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
}

// Request defines the metric data request with defined frequency.
type Request struct {
	Frequency Duration `json:"frequency"`
	Filter    []Filter `json:"filter"`
}

// HasFilterFor returns true if there is filter for the provided originator.
func (mr *Request) HasFilterFor(originator string) bool {
	for _, f := range mr.Filter {
		if f.Originator == originator || len(f.Originator) == 0 {
			return true
		}
	}
	return len(mr.Filter) == 0
}

// HasFilterForItem returns true if there is filter for same originator and filter's ID
// that is the same or with wildcard for the last ID segment,
// i.e. for provided "cpu.utilization" will return true if there is "cpu.utilization", "cpu.*" filter ID or if no filters are set.
func (mr *Request) HasFilterForItem(dataID, dataOriginator string) bool {
	for _, f := range mr.Filter {
		if f.Originator == dataOriginator || len(f.Originator) == 0 {
			if len(f.ID) == 0 {
				return true
			}
			for _, fid := range f.ID {
				if fid == dataID {
					return true
				} else if strings.HasPrefix(dataID, strings.Trim(fid, "*")) {
					return true
				}
			}
		}
	}
	return len(mr.Filter) == 0
}
