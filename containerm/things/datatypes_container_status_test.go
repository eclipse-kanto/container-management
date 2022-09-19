// Copyright (c) 2021 Contributors to the Eclipse Foundation
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
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestFromAPIStatus(t *testing.T) {
	tests := map[string]struct {
		apiStatus types.Status
		expected  status
	}{
		"test_from_api_status_creating": {
			apiStatus: types.Creating,
			expected:  creating,
		},
		"test_from_api_status_created": {
			apiStatus: types.Created,
			expected:  created,
		},
		"test_from_api_status_running": {
			apiStatus: types.Running,
			expected:  running,
		},
		"test_from_api_status_stopped": {
			apiStatus: types.Stopped,
			expected:  stopped,
		},
		"test_from_api_status_paused": {
			apiStatus: types.Paused,
			expected:  paused,
		},
		"test_from_api_status_exited": {
			apiStatus: types.Exited,
			expected:  exited,
		},
		"test_from_api_status_dead": {
			apiStatus: types.Dead,
			expected:  dead,
		},
		"test_from_api_status_unknown": {
			apiStatus: types.Unknown,
			expected:  unknown,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := fromAPIStatus(testCase.apiStatus)
			testutil.AssertEqual(t, testCase.expected, actual)
		})
	}
}

func TestToAPIStatus(t *testing.T) {
	tests := map[string]struct {
		status   status
		expected types.Status
	}{
		"test_to_api_status_creating": {
			status:   creating,
			expected: types.Creating,
		},
		"test_to_api_status_created": {
			status:   created,
			expected: types.Created,
		},
		"test_to_api_status_running": {
			status:   running,
			expected: types.Running,
		},
		"test_to_api_status_stopped": {
			status:   stopped,
			expected: types.Stopped,
		},
		"test_to_api_status_paused": {
			status:   paused,
			expected: types.Paused,
		},
		"test_to_api_status_exited": {
			status:   exited,
			expected: types.Exited,
		},
		"test_to_api_status_dead": {
			status:   dead,
			expected: types.Dead,
		},
		"test_to_api_status_unknown": {
			status:   unknown,
			expected: types.Unknown,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := toAPIStatus(testCase.status)
			testutil.AssertEqual(t, testCase.expected, actual)
		})
	}
}
