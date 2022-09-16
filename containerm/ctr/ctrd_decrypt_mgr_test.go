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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"testing"
)

var (
	testDecKeys       = []string{"../pkg/testutil/certs/testImageKey.pem"}
	testDecRecipients = []string{"pkcs7:../pkg/testutil/certs/testImageCert.pem"}
)

func TestNewContainerDecryptManager(t *testing.T) {
	testCases := map[string]struct {
		testDecKeys       []string
		testDecRecipients []string
		expectedError     bool
	}{
		"test_decrypt_config": {
			testDecKeys:       testDecKeys,
			testDecRecipients: testDecRecipients,
			expectedError:     false,
		},
		"test_empty_decrypt_config": {
			testDecKeys:       []string{},
			testDecRecipients: []string{},
			expectedError:     false,
		},
		"test_parse_error": {
			testDecKeys:       []string{"invalid"},
			testDecRecipients: []string{"invalid"},
			expectedError:     true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			containerDecryptManager, err := newContainerDecryptManager(testCase.testDecKeys, testCase.testDecRecipients)
			if testCase.expectedError {
				testutil.AssertNil(t, containerDecryptManager)
				testutil.AssertNotNil(t, err)
			} else {
				testutil.AssertNotNil(t, containerDecryptManager)
				testutil.AssertNil(t, err)
			}
		})
	}
}

func TestGetDecryptInfo(t *testing.T) {
	testCases := map[string]struct {
		dc              *types.DecryptConfig
		expectedDecInfo bool
		expectedError   bool
	}{
		"test_decrypt_config": {
			dc: &types.DecryptConfig{
				Keys:       testDecKeys,
				Recipients: testDecRecipients,
			},
			expectedDecInfo: true,
		},
		"test_empty_decrypt_config": {
			dc:              &types.DecryptConfig{},
			expectedDecInfo: true,
		},
		"test_parse_error": {
			dc: &types.DecryptConfig{
				Keys: []string{"invalid"},
			},
			expectedError: true,
		},
		"test_no_configs": {
			dc:            nil,
			expectedError: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			containerDecryptManager, _ := newContainerDecryptManager(nil, nil)
			decryptCfg, err := containerDecryptManager.GetDecryptConfig(testCase.dc)
			if testCase.expectedError {
				testutil.AssertNotNil(t, err)
			}
			if testCase.expectedDecInfo {
				testutil.AssertNotNil(t, decryptCfg)
			}
		})
	}
}
