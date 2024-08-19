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

package updateagent

import (
	"context"
	"reflect"
	"testing"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	mgrmocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/eclipse-kanto/update-manager/api/types"
	ummocks "github.com/eclipse-kanto/update-manager/test/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

var (
	stopOpts = &ctrtypes.StopOpts{
		Force:  true,
		Signal: "SIGTERM",
	}
)

type testContext struct {
	activityID        string
	baseline          string
	desiredState      *types.DesiredState
	currentContainers []*ctrtypes.Container
	desiredContainers []*ctrtypes.Container
	actions           []*types.Action
	matchers          []gomock.Matcher
}

type testStep struct {
	command types.CommandType
	expect  func(*testing.T, *mgrmocks.MockContainerManager, *ummocks.MockUpdateManagerCallback, *testContext)
}

func TestExecute(t *testing.T) {
	testCases := map[string]struct {
		baseline string
		steps    []testStep
	}{
		"test-execute-do-without-baseline": {
			steps: []testStep{{command: types.CommandType("DO"), expect: nil}},
		},
		"test-execute-do-for-baseline": {
			baseline: "test-baseline-0",
			steps:    []testStep{{command: types.CommandType("DO"), expect: nil}},
		},

		"test-execute-without-baseline-no-errors": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect:  expectActivationOK,
				},
				{
					command: types.CommandCleanup,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 4, types.ActionStatusRemovalSuccess, "Old container instance is removed.")
						gomock.InOrder(
							// call to ContainerManager to remove old instance of test-container-2
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[1].ID, true, false, nil).Return(nil),
							// call to ContainerManager to remove instance of test-container-5
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[3].ID, true, false, nil).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, expActions1),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.StatusCompleted, expActions1),
						)
						testctx.actions = expActions1
					},
				},
			},
		},

		"test-execute-for-baselineA-no-errors": {
			baseline: "test-baseline-a",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadSuccess, "New container created.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusUpdateSuccess, "Old container instance is stopped.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to stop test-container-2 is expected
							mockStopContainer(mockContainerManager, testctx.currentContainers[1]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationSuccess, "Existing container instance is running.")
						expActions2[1].Status = types.ActionStatusActivating
						expActions3 := copyAndUpdateActions(expActions2, 1, types.ActionStatusActivationSuccess, "New container instance is started.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-1 (state paused)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
							// call to ContainerManager to unpause test-container-1
							mockUnpauseContainer(mockContainerManager, testctx.currentContainers[0]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
							// call to ContainerManager to start new test-container-2
							mockContainerManager.EXPECT().Start(context.Background(), gomock.Not(testctx.currentContainers[1].ID)).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationSuccess, expActions3),
						)
						testctx.actions = expActions3
					},
				},
				{
					command: types.CommandCleanup,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						gomock.InOrder(
							// call to ContainerManager to remove old instance of test-container-2
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[1].ID, true, false, nil).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, testctx.actions),
						)
					},
				},
			},
		},

		"test-execute-for-baselineB-no-errors": {
			baseline: "test-baseline-b",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 3, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 3, types.ActionStatusDownloadSuccess, "New container created.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[3]).Return(nil, nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 2, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 2, types.ActionStatusUpdateSuccess, "Container instance is updated with new configuration.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to update test-container-3 with new restart policy
							mockContainerManager.EXPECT().Update(context.Background(), testctx.currentContainers[2].ID, gomock.Any()).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 2, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 2, types.ActionStatusActivationSuccess, "")
						expActions2[3].Status = types.ActionStatusActivating
						expActions3 := copyAndUpdateActions(expActions2, 3, types.ActionStatusActivationSuccess, "New container instance is started.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-3 (state running)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[2].ID).Return(testctx.currentContainers[2], nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
							// call to ContainerManager to start test-container-4
							mockContainerManager.EXPECT().Start(context.Background(), testctx.desiredContainers[3].ID).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationSuccess, expActions3),
						)
						testctx.actions = expActions3
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupNoActions,
				},
			},
		},

		"test-execute-for-baseline-destroy-no-errors": {
			baseline: "containers:remove-components",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, testctx.actions)
					},
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 4, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 4, types.ActionStatusUpdateSuccess, "Old container instance is stopped.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// no stop call to ContainerManager as test-container-4 is not running (test setup)
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationSuccess, testctx.actions)
					},
				},
				{
					command: types.CommandCleanup,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 4, types.ActionStatusRemovalSuccess, "Old container instance is removed.")
						gomock.InOrder(
							// call to ContainerManager to remove instance of test-container-5
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[3].ID, true, false, nil).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, expActions1),
						)
						testctx.actions = expActions1
					},
				},
			},
		},

		"test-execute-without-baseline-download-error-1": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadFailure, "cannot download container image")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, errors.New("cannot download container image")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-without-baseline-download-error-2": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadSuccess, "New container created.")
						expActions2[3].Status = types.ActionStatusDownloading
						expActions3 := copyAndUpdateActions(expActions2, 3, types.ActionStatusDownloadFailure, "cannot download container image")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions2),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[3]).Return(nil, errors.New("cannot download container image")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadFailure, expActions3),
						)
						testctx.actions = expActions3
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-for-baselineA-download-error": {
			baseline: "test-baseline-a",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadFailure, "cannot download container image")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, errors.New("cannot download container image")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupNoActions,
				},
			},
		},

		"test-execute-without-baseline-update-error-1": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusUpdateFailure, "cannot stop container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to stop test-container-2 is expected
							mockContainerManager.EXPECT().Stop(context.Background(), testctx.currentContainers[1].ID, stopOpts).Return(errors.New("cannot stop container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-without-baseline-update-error-2": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusUpdateSuccess, "Old container instance is stopped.")
						expActions2[2].Status = types.ActionStatusUpdating
						expActions3 := copyAndUpdateActions(expActions2, 2, types.ActionStatusUpdateFailure, "cannot update container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to stop test-container-2 is expected
							mockStopContainer(mockContainerManager, testctx.currentContainers[1]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions2),
							// call to ContainerManager to update test-container-3 with new restart policy
							mockContainerManager.EXPECT().Update(context.Background(), testctx.currentContainers[2].ID, gomock.Any()).Return(errors.New("cannot update container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateFailure, expActions3),
						)
						testctx.actions = expActions3
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-for-baselineA-update-error": {
			baseline: "test-baseline-a",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadSuccess, "New container created.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusUpdateFailure, "cannot stop container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to stop test-container-2 is expected
							mockContainerManager.EXPECT().Stop(context.Background(), testctx.currentContainers[1].ID, stopOpts).Return(errors.New("cannot stop container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupNoActions,
				},
			},
		},
		"test-execute-for-baselineB-update-error": {
			baseline: "test-baseline-b",
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 3, types.ActionStatusDownloading, "")
						expActions2 := copyAndUpdateActions(expActions1, 3, types.ActionStatusDownloadSuccess, "New container created.")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
							mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[3]).Return(nil, nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandUpdate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 2, types.ActionStatusUpdating, "")
						expActions2 := copyAndUpdateActions(expActions1, 2, types.ActionStatusUpdateFailure, "cannot update container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
							// call to ContainerManager to update test-container-3 with new restart policy
							mockContainerManager.EXPECT().Update(context.Background(), testctx.currentContainers[2].ID, gomock.Any()).Return(errors.New("cannot update container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupNoActions,
				},
			},
		},

		"test-execute-without-baseline-activate-error-1": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationFailure, "cannot unpause container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-1 (state paused)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
							// call to ContainerManager to unpause test-container-1
							mockContainerManager.EXPECT().Unpause(context.Background(), testctx.currentContainers[0].ID).Return(errors.New("cannot unpause container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationFailure, expActions2),
						)
						testctx.actions = expActions2
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-without-baseline-activate-error-2": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationSuccess, "Existing container instance is running.")
						expActions2[1].Status = types.ActionStatusActivating
						expActions3 := copyAndUpdateActions(expActions2, 1, types.ActionStatusActivationFailure, "cannot start new container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-1 (state paused)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
							// call to ContainerManager to unpause test-container-1
							mockUnpauseContainer(mockContainerManager, testctx.currentContainers[0]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
							// call to ContainerManager to start new test-container-2
							mockContainerManager.EXPECT().Start(context.Background(), gomock.Not(testctx.currentContainers[1].ID)).Return(errors.New("cannot start new container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationFailure, expActions3),
						)
						testctx.actions = expActions3
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-without-baseline-activate-error-3": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationSuccess, "Existing container instance is running.")
						expActions2[1].Status = types.ActionStatusActivating
						expActions3 := copyAndUpdateActions(expActions2, 1, types.ActionStatusActivationSuccess, "New container instance is started.")
						expActions3[2].Status = types.ActionStatusActivating
						expActions4 := copyAndUpdateActions(expActions3, 2, types.ActionStatusActivationFailure, "cannot get current state for container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-1 (state paused)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
							// call to ContainerManager to unpause test-container-1
							mockUnpauseContainer(mockContainerManager, testctx.currentContainers[0]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
							// call to ContainerManager to start new test-container-2
							mockContainerManager.EXPECT().Start(context.Background(), gomock.Not(testctx.currentContainers[1].ID)).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions3),
							// call to ContainerManager to retrieve test-container-3 (state running)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[2].ID).Return(nil, errors.New("cannot get current state for container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationFailure, expActions4),
						)
						testctx.actions = expActions4
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},
		"test-execute-without-baseline-activate-error-4": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
						expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationSuccess, "Existing container instance is running.")
						expActions2[1].Status = types.ActionStatusActivating
						expActions3 := copyAndUpdateActions(expActions2, 1, types.ActionStatusActivationSuccess, "New container instance is started.")
						expActions3[2].Status = types.ActionStatusActivating
						expActions4 := copyAndUpdateActions(expActions3, 2, types.ActionStatusActivationSuccess, "")
						expActions4[3].Status = types.ActionStatusActivating
						expActions5 := copyAndUpdateActions(expActions4, 3, types.ActionStatusActivationFailure, "cannot start new container instance")
						gomock.InOrder(
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
							// call to ContainerManager to retrieve test-container-1 (state paused)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
							// call to ContainerManager to unpause test-container-1
							mockUnpauseContainer(mockContainerManager, testctx.currentContainers[0]),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
							// call to ContainerManager to start new test-container-2
							mockContainerManager.EXPECT().Start(context.Background(), gomock.Not(testctx.currentContainers[1].ID)).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions3),
							// call to ContainerManager to retrieve test-container-3 (state running)
							mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[2].ID).Return(testctx.currentContainers[2], nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions4),
							// call to ContainerManager to start test-container-4
							mockContainerManager.EXPECT().Start(context.Background(), testctx.desiredContainers[3].ID).Return(errors.New("cannot start new container instance")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationFailure, expActions5),
						)
						testctx.actions = expActions5
					},
				},
				{
					command: types.CommandCleanup,
					expect:  expectCleanupIncomplete,
				},
			},
		},

		"test-execute-without-baseline-cleanup-error-1": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect:  expectActivationOK,
				},
				{
					command: types.CommandCleanup,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 4, types.ActionStatusRemovalSuccess, "Old container instance is removed.")
						gomock.InOrder(
							// call to ContainerManager to remove old instance of test-container-2
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[1].ID, true, false, nil).Return(errors.New("cannot remove old container instance")),
							// call to ContainerManager to remove instance of test-container-5
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[3].ID, true, false, nil).Return(nil),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, expActions1),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.StatusCompleted, expActions1),
						)
						testctx.actions = expActions1
					},
				},
			},
		},
		"test-execute-without-baseline-cleanup-error-2": {
			steps: []testStep{
				{
					command: types.CommandDownload,
					expect:  expectDownloadOK,
				},
				{
					command: types.CommandUpdate,
					expect:  expectUpdateOK,
				},
				{
					command: types.CommandActivate,
					expect:  expectActivationOK,
				},
				{
					command: types.CommandCleanup,
					expect: func(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
						expActions1 := copyAndUpdateActions(testctx.actions, 4, types.ActionStatusRemovalFailure, "old container instance cannot be removed")
						gomock.InOrder(
							// call to ContainerManager to remove old instance of test-container-2
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[1].ID, true, false, nil).Return(nil),
							// call to ContainerManager to remove instance of test-container-5
							mockContainerManager.EXPECT().Remove(context.Background(), testctx.currentContainers[3].ID, true, false, nil).Return(errors.New("old container instance cannot be removed")),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupFailure, expActions1),
							expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.StatusIncomplete, expActions1),
						)
						testctx.actions = expActions1
					},
				},
			},
		},
	}
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	for testActivityID, testCase := range testCases {
		t.Run(testActivityID, func(t *testing.T) {
			mockContainerManager := mgrmocks.NewMockContainerManager(mockCtr)
			updateManager := newUpdateManager(mockContainerManager, nil, domainName, []string{sysContainerName}, false)
			ctrUpdManager := updateManager.(*containersUpdateManager)
			mockCallback := ummocks.NewMockUpdateManagerCallback(mockCtr)
			updateManager.SetCallback(mockCallback)

			// setup mocks
			testctx := setupTestEnvironment(testActivityID, testCase.baseline)
			mockContainerManager.EXPECT().List(gomock.Any()).Return(testctx.currentContainers, nil)
			expectFeedback(t, mockCallback, testActivityID, "", types.StatusIdentifying, nil)
			expectFeedback(t, mockCallback, testActivityID, "", types.StatusIdentified, testctx.actions)

			// perform identification before commands
			updateManager.Apply(context.Background(), testActivityID, testctx.desiredState)
			testutil.AssertNotNil(t, ctrUpdManager.operation)
			operation := ctrUpdManager.operation.(*operation)
			testutil.AssertEqual(t, testActivityID, operation.GetActivityID())
			testctx.desiredContainers = operation.desiredState.containers

			for _, step := range testCase.steps {
				t.Logf("executing command %s for baseline '%s'", step.command, testCase.baseline)
				if step.expect != nil {
					step.expect(t, mockContainerManager, mockCallback, testctx)
				}
				operation.Execute(step.command, testCase.baseline)
			}
		})
	}
}

// test setup:
// current containers: test-container-1 (paused), test-container-2, test-container-3, test-container-5 (stopped). note: test-container-4 is missing.
// desired state:
// - containers: test-container-1 (same), test-container-2 (new image), test-container-3 (config update), test-container-4 (new)
// - baselines: test-baseline-a (test-container-1, test-container-2), test-baseline-b (test-container-3, test-container-5)
// actions: test-container-1 (check), test-container-2 (recreate), test-container-3 (update), test-container-4 (create), test-container-5 (destroy)
func setupTestEnvironment(activityID, baseline string) *testContext {
	component1 := createSimpleDesiredComponent("test-container-1", "1.1.1")
	component2 := createSimpleDesiredComponent("test-container-2", "2.2.22")
	component3 := createSimpleDesiredComponent("test-container-3", "3.3.3")
	component3.Config = append(component3.Config, &types.KeyValuePair{Key: "restartPolicy", Value: "no"})
	component4 := createSimpleDesiredComponent("test-container-4", "4.4.4")

	container1 := createSimpleContainer("test-container-1", "1.1.1")
	util.SetContainerStatusPaused(container1)
	container2 := createSimpleContainer("test-container-2", "2.2.2")
	container3 := createSimpleContainer("test-container-3", "3.3.3")
	container5 := createSimpleContainer("test-container-5", "5.5.5")
	util.SetContainerStatusStopped(container5, 0, "")

	return &testContext{
		activityID: activityID,
		baseline:   baseline,
		desiredState: &types.DesiredState{
			Baselines: []*types.Baseline{
				{Title: "test-baseline-a", Components: []string{domainName + ":" + component1.ID, domainName + ":" + component2.ID}},
				{Title: "test-baseline-b", Components: []string{domainName + ":" + component3.ID, domainName + ":" + component4.ID}},
			},
			Domains: []*types.Domain{{
				ID:         domainName,
				Components: []*types.ComponentWithConfig{component1, component2, component3, component4},
			}},
		},
		currentContainers: []*ctrtypes.Container{container1, container2, container3, container5},
		actions: []*types.Action{
			{
				Component: createActionComponent(component1),
				Status:    types.ActionStatusIdentified,
				Message:   util.GetActionMessage(util.ActionCheck),
			},
			{
				Component: createActionComponent(component2),
				Status:    types.ActionStatusIdentified,
				Message:   util.GetActionMessage(util.ActionRecreate),
			},
			{
				Component: createActionComponent(component3),
				Status:    types.ActionStatusIdentified,
				Message:   util.GetActionMessage(util.ActionUpdate),
			},
			{
				Component: createActionComponent(component4),
				Status:    types.ActionStatusIdentified,
				Message:   util.GetActionMessage(util.ActionCreate),
			},
			{
				Component: &types.Component{ID: domainName + ":" + container5.Name, Version: "5.5.5"},
				Status:    types.ActionStatusIdentified,
				Message:   util.GetActionMessage(util.ActionDestroy),
			}},
		matchers: []gomock.Matcher{
			matchers.MatchesContainerImage(component1.ID, component1.ID+":"+component1.Version),
			matchers.MatchesContainerImage(component2.ID, component2.ID+":"+component2.Version),
			matchers.MatchesContainerImage(component3.ID, component3.ID+":"+component3.Version),
			matchers.MatchesContainerImage(component4.ID, component4.ID+":"+component4.Version),
		},
	}
}

func copyAndUpdateActions(actions []*types.Action, index int, status types.ActionStatusType, message string) []*types.Action {
	copy := make([]*types.Action, len(actions))
	for i, action := range actions {
		copy[i] = &types.Action{
			Component: action.Component,
			Status:    action.Status,
			Progress:  action.Progress,
			Message:   action.Message,
		}
	}
	copy[index].Status = status
	if message != "" {
		copy[index].Message = message
	}
	return copy
}

func expectFeedback(t *testing.T, mockCallback *ummocks.MockUpdateManagerCallback,
	expActivityID, expBaseline string, expStatus types.StatusType, expActions []*types.Action) *gomock.Call {
	return mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, expActivityID, expBaseline, gomock.Any(), "", gomock.Any()).Do(
		func(domain, activityID, baseline string, status types.StatusType, message string, actions []*types.Action) {
			testutil.AssertEqual(t, expStatus, status)
			testutil.AssertEqual(t, len(expActions), len(actions))
			for i, action := range actions {
				expected := expActions[i]
				if !reflect.DeepEqual(expected, action) {
					t.Errorf("feedback action %d: expected (%s:%s, %s, %s, %d), got (%s:%s, %s, %s, %d)", i,
						expected.Component.ID, expected.Component.Version, expected.Status, expected.Message, expected.Progress,
						action.Component.ID, action.Component.Version, action.Status, action.Message, action.Progress)
				}
			}
		})
}

func expectDownloadOK(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
	expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusDownloading, "")
	expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusDownloadSuccess, "New container created.")
	expActions2[3].Status = types.ActionStatusDownloading
	expActions3 := copyAndUpdateActions(expActions2, 3, types.ActionStatusDownloadSuccess, "New container created.")
	gomock.InOrder(
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions1),
		mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[1]).Return(nil, nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloading, expActions2),
		mockContainerManager.EXPECT().Create(context.Background(), testctx.matchers[3]).Return(nil, nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusDownloadSuccess, expActions3),
	)
	testctx.actions = expActions3
}

func expectUpdateOK(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
	expActions1 := copyAndUpdateActions(testctx.actions, 1, types.ActionStatusUpdating, "")
	expActions2 := copyAndUpdateActions(expActions1, 1, types.ActionStatusUpdateSuccess, "Old container instance is stopped.")
	expActions2[2].Status = types.ActionStatusUpdating
	expActions3 := copyAndUpdateActions(expActions2, 2, types.ActionStatusUpdateSuccess, "Container instance is updated with new configuration.")
	expActions3[4].Status = types.ActionStatusUpdating
	expActions4 := copyAndUpdateActions(expActions3, 4, types.ActionStatusUpdateSuccess, "Old container instance is stopped.")
	gomock.InOrder(
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions1),
		// call to ContainerManager to stop test-container-2 is expected
		mockStopContainer(mockContainerManager, testctx.currentContainers[1]),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions2),
		// call to ContainerManager to update test-container-3 with new restart policy
		mockContainerManager.EXPECT().Update(context.Background(), testctx.currentContainers[2].ID, gomock.Any()).Return(nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdating, expActions3),
		// no stop call to ContainerManager as test-container-4 is not running (test setup)
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusUpdateSuccess, expActions4),
	)
	testctx.actions = expActions4
}

func expectActivationOK(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
	expActions1 := copyAndUpdateActions(testctx.actions, 0, types.ActionStatusActivating, "")
	expActions2 := copyAndUpdateActions(expActions1, 0, types.ActionStatusActivationSuccess, "Existing container instance is running.")
	expActions2[1].Status = types.ActionStatusActivating
	expActions3 := copyAndUpdateActions(expActions2, 1, types.ActionStatusActivationSuccess, "New container instance is started.")
	expActions3[2].Status = types.ActionStatusActivating
	expActions4 := copyAndUpdateActions(expActions3, 2, types.ActionStatusActivationSuccess, "")
	expActions4[3].Status = types.ActionStatusActivating
	expActions5 := copyAndUpdateActions(expActions4, 3, types.ActionStatusActivationSuccess, "New container instance is started.")
	gomock.InOrder(
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions1),
		// call to ContainerManager to retrieve test-container-1 (state paused)
		mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[0].ID).Return(testctx.currentContainers[0], nil),
		// call to ContainerManager to unpause test-container-1
		mockUnpauseContainer(mockContainerManager, testctx.currentContainers[0]),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions2),
		// call to ContainerManager to start new test-container-2
		mockContainerManager.EXPECT().Start(context.Background(), gomock.Not(testctx.currentContainers[1].ID)).Return(nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions3),
		// call to ContainerManager to retrieve test-container-3 (state running)
		mockContainerManager.EXPECT().Get(context.Background(), testctx.currentContainers[2].ID).Return(testctx.currentContainers[2], nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivating, expActions4),
		// call to ContainerManager to start test-container-4
		mockContainerManager.EXPECT().Start(context.Background(), testctx.desiredContainers[3].ID).Return(nil),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusActivationSuccess, expActions5),
	)
	testctx.actions = expActions5
}

func mockUnpauseContainer(mockContainerManager *mgrmocks.MockContainerManager, container *ctrtypes.Container) *gomock.Call {
	return mockContainerManager.EXPECT().Unpause(context.Background(), container.ID).DoAndReturn(
		func(ctx context.Context, id string) error {
			util.SetContainerStatusRunning(container, 5678)
			return nil
		})
}

func mockStopContainer(mockContainerManager *mgrmocks.MockContainerManager, container *ctrtypes.Container) *gomock.Call {
	return mockContainerManager.EXPECT().Stop(context.Background(), container.ID, stopOpts).DoAndReturn(
		func(ctx context.Context, id string, stopOpts *ctrtypes.StopOpts) error {
			util.SetContainerStatusStopped(container, 9999, "stopped by update")
			return nil
		})
}

func expectCleanupNoActions(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
	expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, testctx.actions)
}

func expectCleanupIncomplete(t *testing.T, mockContainerManager *mgrmocks.MockContainerManager, mockCallback *ummocks.MockUpdateManagerCallback, testctx *testContext) {
	gomock.InOrder(
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.BaselineStatusCleanupSuccess, testctx.actions),
		expectFeedback(t, mockCallback, testctx.activityID, testctx.baseline, types.StatusIncomplete, testctx.actions),
	)
}
