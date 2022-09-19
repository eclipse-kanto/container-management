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

func TestFromAPIRPType(t *testing.T) {
	tests := map[string]struct {
		apiRPType types.PolicyType
		expected  restartPolicyType
	}{
		"test_from_api_rp_type_always": {
			apiRPType: types.Always,
			expected:  always,
		},
		"test_from_api_rp_type_on_failure": {
			apiRPType: types.OnFailure,
			expected:  onFailure,
		},
		"test_from_api_rp_type_unless_stopped": {
			apiRPType: types.UnlessStopped,
			expected:  unlessStopped,
		},
		"test_from_api_rp_type_no": {
			apiRPType: types.No,
			expected:  no,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := fromAPIRPType(testCase.apiRPType)
			testutil.AssertEqual(t, testCase.expected, actual)
		})
	}
}

func TestToAPIRPType(t *testing.T) {
	tests := map[string]struct {
		rpType   restartPolicyType
		expected types.PolicyType
	}{
		"test_to_api_rp_type_always": {
			rpType:   always,
			expected: types.Always,
		},
		"test_to_api_rp_type_on_failure": {
			rpType:   onFailure,
			expected: types.OnFailure,
		},
		"test_to_api_rp_type_unless_stopped": {
			rpType:   unlessStopped,
			expected: types.UnlessStopped,
		},
		"test_to_api_rp_type_no": {
			rpType:   no,
			expected: types.No,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := toAPIRPType(testCase.rpType)
			testutil.AssertEqual(t, testCase.expected, actual)
		})
	}
}
