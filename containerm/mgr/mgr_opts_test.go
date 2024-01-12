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

package mgr

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestApplyMgrOpts(t *testing.T) {
	tests := map[string]struct {
		testOpts      []ContainerManagerOpt
		expectedError error
	}{
		"test_apply_without_error": {
			testOpts: []ContainerManagerOpt{WithMgrMetaPath(testMetaPath)},
		},
		"test_apply_with_error": {
			testOpts: []ContainerManagerOpt{func() ContainerManagerOpt {
				return func(mgrOptions *mgrOpts) error {
					return log.NewError("test error")
				}
			}()},
			expectedError: log.NewError("test error"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			resultOpts := &mgrOpts{}
			err := applyOptsMgr(resultOpts, testCase.testOpts...)
			testutil.AssertError(t, testCase.expectedError, err)
		})
	}
}

func TestMgrOpts(t *testing.T) {
	tests := map[string]struct {
		testOpt      ContainerManagerOpt
		expectedOpts *mgrOpts
	}{
		"test_mgr_meta_path": {
			testOpt: WithMgrMetaPath(testMetaPath),
			expectedOpts: &mgrOpts{
				metaPath: testMetaPath,
			},
		},
		"test_mgr_root_exec": {
			testOpt: WithMgrRootExec(testRootExec),
			expectedOpts: &mgrOpts{
				rootExec: testRootExec,
			},
		},
		"test_mgr_container_client_service_id": {
			testOpt: WithMgrContainerClientServiceID(testContainerClientServiceID),
			expectedOpts: &mgrOpts{
				containerClientServiceID: testContainerClientServiceID,
			},
		},
		"test_mgr_network_manager_service_id": {
			testOpt: WithMgrNetworkManagerServiceID(testNetworkManagerServiceID),
			expectedOpts: &mgrOpts{
				networkManagerServiceID: testNetworkManagerServiceID,
			},
		},
		"test_mgr_default_container_stop_timeout": {
			testOpt: WithMgrDefaultContainerStopTimeout(testContainerStopTimeout),
			expectedOpts: &mgrOpts{
				defaultCtrsStopTimeout: testContainerStopTimeout,
			},
		},
		"test_mgr_default_container_stop_timeout_int64": {
			testOpt: WithMgrDefaultContainerStopTimeout(int64(10)),
			expectedOpts: &mgrOpts{
				defaultCtrsStopTimeout: testContainerStopTimeout,
			},
		},
		"test_mgr_default_container_stop_timeout_string": {
			testOpt: WithMgrDefaultContainerStopTimeout("10"),
			expectedOpts: &mgrOpts{
				defaultCtrsStopTimeout: 0,
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			resultOpts := &mgrOpts{}
			applyOptsMgr(resultOpts, testCase.testOpt)
			testutil.AssertEqual(t, testCase.expectedOpts, resultOpts)
		})
	}
}
