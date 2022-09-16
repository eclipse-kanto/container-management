// Copyright (c) 2021 Contributors to the Eclipse Foundation
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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testNamespace          = "test-namespace"
	testConnectionPath     = "test-conn-path"
	testRootExec           = "test-root-exec"
	testMetaPath           = "test-meta-path"
	testHost               = "test-host"
	testUser               = "test-user"
	testPass               = "test-pass"
	testImageExpiry        = 31 * 24 * time.Hour
	testImageExpiryDisable = true
	testLeaseID            = "test-lease-id"
)

var (
	regConfig = &RegistryConfig{
		IsInsecure: false,
		Credentials: &AuthCredentials{
			UserID:   testUser,
			Password: testPass,
		},
		Transport: nil,
	}

	testOpt = &ctrOpts{
		namespace:          testNamespace,
		connectionPath:     testConnectionPath,
		registryConfigs:    map[string]*RegistryConfig{testHost: regConfig},
		rootExec:           testRootExec,
		metaPath:           testMetaPath,
		imageDecKeys:       testDecKeys,
		imageDecRecipients: testDecRecipients,
		runcRuntime:        types.RuntimeTypeV2runcV2,
		imageExpiry:        testImageExpiry,
		imageExpiryDisable: testImageExpiryDisable,
		leaseID:            testLeaseID,
	}
)

func TestCtrOpts(t *testing.T) {
	testCases := map[string]struct {
		opts         []ContainerOpts
		expectedOpts *ctrOpts
		expectedErr  error
	}{
		"test_ctr_opts_unexpected_runc_runtime_error": {
			opts: []ContainerOpts{
				WithCtrdRuncRuntime("unknown"),
			},
			expectedOpts: &ctrOpts{},
			expectedErr:  log.NewErrorf("unexpected runc runtime = unknown"),
		},
		"test_ctr_opts_no_error": {
			opts: []ContainerOpts{WithCtrdConnectionPath(testConnectionPath),
				WithCtrdNamespace(testNamespace),
				WithCtrdRootExec(testRootExec),
				WithCtrdMetaPath(testMetaPath),
				WithCtrdRegistryConfigs(map[string]*RegistryConfig{testHost: regConfig}),
				WithCtrdImageDecryptKeys(testDecKeys...),
				WithCtrdImageDecryptRecipients(testDecRecipients...),
				WithCtrdRuncRuntime(string(types.RuntimeTypeV2runcV2)),
				WithCtrdImageExpiry(testImageExpiry),
				WithCtrdImageExpiryDisable(testImageExpiryDisable),
				WithCtrdLeaseID(testLeaseID)},
			expectedOpts: testOpt,
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			opts := &ctrOpts{}
			err := applyOptsCtr(opts, testCaseData.opts...)

			testutil.AssertError(t, testCaseData.expectedErr, err)
			testutil.AssertEqual(t, testCaseData.expectedOpts, opts)
		})
	}
}
