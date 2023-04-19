// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package deployment

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const testMode = InitialDeployMode
const testMetaPath = "testMetaPath"
const testCtrPath = "testCtrPath"

func TestApplyDeploymentOpts(t *testing.T) {
	tests := map[string]struct {
		testOpts      []Opt
		expectedError error
	}{
		"test_apply_without_error": {
			testOpts: []Opt{
				WithMode(string(testMode)),
				WithMetaPath(testMetaPath),
				WithCtrPath(testCtrPath),
			},
		},
		"test_apply_with_error": {
			testOpts: []Opt{func() Opt {
				return func(deploymentOptions *opts) error {
					return log.NewError("test error")
				}
			}()},
			expectedError: log.NewError("test error"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			resultOpts := &opts{}
			err := applyOpts(resultOpts, testCase.testOpts...)
			testutil.AssertError(t, testCase.expectedError, err)
		})
	}
}

func TestDeploymentOpts(t *testing.T) {
	tests := map[string]struct {
		testOpt      Opt
		expectedOpts *opts
	}{
		"test_deployment_mode": {
			testOpt: WithMode(string(testMode)),
			expectedOpts: &opts{
				mode: testMode,
			},
		},
		"test_deployment_meta_path": {
			testOpt: WithMetaPath(testMetaPath),
			expectedOpts: &opts{
				metaPath: testMetaPath,
			},
		},
		"test_deployment_initial_deploy_path": {
			testOpt: WithCtrPath(testCtrPath),
			expectedOpts: &opts{
				ctrPath: testCtrPath,
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			resultOpts := &opts{}
			applyOpts(resultOpts, testCase.testOpt)
			testutil.AssertEqual(t, testCase.expectedOpts, resultOpts)
		})
	}
}
