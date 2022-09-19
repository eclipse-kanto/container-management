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
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"testing"
	"time"
)

func TestMetricsRequestHasFilterFor(t *testing.T) {
	tests := map[string]struct {
		filter     []Filter
		originator string
		want       bool
	}{
		"test_system_originator": {
			filter: []Filter{{
				ID:         []string{},
				Originator: "SYSTEM",
			}},
			originator: "SYSTEM",
			want:       true,
		},
		"test_empty_originator": {
			filter: []Filter{{
				ID: []string{},
			}},
			originator: "SYSTEM",
			want:       true,
		},
		"test_non_matching_originator": {
			filter: []Filter{{
				ID:         []string{},
				Originator: "N/A",
			}},
			originator: "SYSTEM",
			want:       false,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mr := &Request{
				Frequency: Duration{5 * time.Second},
				Filter:    testCase.filter,
			}
			got := mr.HasFilterFor(testCase.originator)
			testutil.AssertEqual(t, testCase.want, got)
		})
	}
}

func TestMetricsRequestHasFilterForItem(t *testing.T) {
	tests := map[string]struct {
		filter         []Filter
		dataID         string
		dataOriginator string
		want           bool
	}{
		"test_same_originator_single_fetureID": {
			filter: []Filter{{
				ID:         []string{CPUUtilization},
				Originator: "SYSTEM",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_same_originator_multiple_fetureID": {
			filter: []Filter{{
				ID:         []string{CPUUtilization, MemoryUtilization},
				Originator: "SYSTEM",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_empty_originator": {
			filter: []Filter{{
				ID:         []string{CPUUtilization, MemoryUtilization},
				Originator: "",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_empty_ID": {
			filter: []Filter{{
				ID:         []string{},
				Originator: "",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_without_ID": {
			filter: []Filter{{
				Originator: "",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_wildcard_single": {
			filter: []Filter{{
				ID:         []string{"cpu.*"},
				Originator: "SYSTEM",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_wildcard_multiple": {
			filter: []Filter{{
				ID:         []string{"memory.*"},
				Originator: "SYSTEM",
			}},
			dataID:         "memory.",
			dataOriginator: "SYSTEM",
			want:           true,
		},
		"test_wildcard_process": {
			filter: []Filter{{
				ID:         []string{"io.*"},
				Originator: "test-process",
			}},
			dataID:         "io.",
			dataOriginator: "test-process",
			want:           true,
		},
		"test_wildcard_error": {
			filter: []Filter{{
				ID:         []string{"cpu.unknown"},
				Originator: "SYSTEM",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           false,
		},
		"test_originator_error": {
			filter: []Filter{{
				ID:         []string{},
				Originator: "N/A",
			}},
			dataID:         CPUUtilization,
			dataOriginator: "SYSTEM",
			want:           false,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mr := &Request{
				Frequency: Duration{5 * time.Second},
				Filter:    testCase.filter,
			}
			got := mr.HasFilterForItem(testCase.dataID, testCase.dataOriginator)
			testutil.AssertEqual(t, testCase.want, got)
		})
	}
}
