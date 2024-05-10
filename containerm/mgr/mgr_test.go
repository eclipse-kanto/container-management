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
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/network"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	ctrMock "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctr"
	eventsMock "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/events"
	mgrMock "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	networkMock "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/network"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus/hooks/test"
)

const (
	testMetaPath                    = "testMetaPath"
	testRootExec                    = "testRootExec"
	testContainerClientServiceID    = "testContainerClientServiceID"
	testNetworkManagerServiceID     = "testNetworkManagerServiceID"
	testContainerStopTimeout        = time.Duration(10) * time.Second
	testDefaultContainerStopTimeout = time.Duration(30) * time.Second
)

func TestGetContainer(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metapath := "../pkg/testutil/metapath/valid"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	ctrID, container := getDefaultContainer()
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	containerUnderTestBefore, err := unitUnderTest.Get(ctx, ctrID)

	// Assert
	testutil.AssertNil(t, err)
	testutil.AssertNotNil(t, containerUnderTestBefore)
	testutil.AssertEqual(t, container, containerUnderTestBefore)

}

func TestRenameContainer(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	name := "awesome"
	ctrID, container := getDefaultContainer()
	metapath := "../pkg/testutil/metapath/valid"
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockEventsManager.EXPECT().Publish(
		gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersRenamed,
		matchers.MatchesContainerName(name)).Times(1)

	mockRepository.EXPECT().Save(
		matchers.MatchesContainerName(name)).Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	err := unitUnderTest.Rename(ctx, ctrID, name)

	// Assert
	testutil.AssertError(t, nil, err)
}

func TestRenameContainerInvalid(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	name := "!awesome"
	ctrID, container := getDefaultContainer()
	metapath := "../pkg/testutil/metapath/valid"
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	err := unitUnderTest.Rename(ctx, ctrID, name)

	// Assert
	testutil.AssertError(t, log.NewErrorf("invalid container name format : %s", name), err)
}

func TestRenameContainerWithSameName(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	ctrID, container := getDefaultContainer()
	metapath := "../pkg/testutil/metapath/valid"
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	err := unitUnderTest.Rename(ctx, ctrID, container.Name)

	// Assert
	testutil.AssertError(t, nil, err)
}

func TestUpdateRunningContainer(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	opts := &types.UpdateOpts{
		RestartPolicy: &types.RestartPolicy{
			Type: types.No,
		},
		Resources: &types.Resources{
			Memory: "500M",
		},
	}
	ctrID, container := getDefaultContainer()
	metapath := "../pkg/testutil/metapath/valid"
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockCtrClient.EXPECT().UpdateContainer(gomock.Any(), container, opts.Resources)
	mockEventsManager.EXPECT().Publish(
		gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersUpdated,
		matchers.MatchesContainerUpdate(opts.RestartPolicy, opts.Resources)).Times(1)

	mockRepository.EXPECT().Save(
		matchers.MatchesContainerUpdate(opts.RestartPolicy, opts.Resources)).Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	err := unitUnderTest.Update(ctx, ctrID, opts)

	// Assert
	testutil.AssertNil(t, err)
}

func TestUpdateContainerWithInvalidOpts(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()

	ctrID, container := getDefaultContainer()
	metapath := "../pkg/testutil/metapath/valid"
	cache := map[string]*types.Container{}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath, mockCtrClient,
		mockNetworkManager, mockEventsManager,
		mockRepository, cache)
	unitUnderTest.Load(ctx)

	invalidPolicyType := types.PolicyType("invalid")
	invalidMemory := "invalid"
	tests := map[string]struct {
		opts        *types.UpdateOpts
		expectedErr error
	}{
		"test_with_invalid_restart_policy": {
			opts: &types.UpdateOpts{
				RestartPolicy: &types.RestartPolicy{Type: invalidPolicyType},
			},
			expectedErr: log.NewErrorf("unsupported restart policy type %s", invalidPolicyType),
		},
		"test_with_invalid_resource": {
			opts: &types.UpdateOpts{
				Resources: &types.Resources{Memory: invalidMemory},
			},
			expectedErr: log.NewErrorf("invalid format of memory - %s", invalidMemory),
		},
		"test_with_nil_opts": {
			opts:        nil,
			expectedErr: nil,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			err := unitUnderTest.Update(ctx, ctrID, testCase.opts)
			testutil.AssertError(t, testCase.expectedErr, err)
		})
	}
}

func TestRestartContainer(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metapath := "../pkg/testutil/metapath/valid"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	ctrID, container := getDefaultContainer()
	cache := map[string]*types.Container{}

	timeout := int64(60)

	mockRepository.EXPECT().Prune().Return(nil).Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)

	unitUnderTest.Load(ctx)
	err := unitUnderTest.Restart(ctx, ctrID, timeout)

	// Assert
	testutil.AssertNotNil(t, err)
	testutil.AssertError(t, log.NewErrorf("restart not supported"), err)
}

func TestCreateValidContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metapath := "../pkg/testutil/metapath/empty"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	_, container := getDefaultContainer()
	cache := map[string]*types.Container{}

	mockRepository.
		EXPECT().
		Save(container).
		Times(1)

	mockCtrClient.
		EXPECT().
		CreateContainer(gomock.Any(), gomock.Eq(container), gomock.Any()).
		Return(nil)

	mockEventsManager.
		EXPECT().
		Publish(gomock.Any(),
			types.EventTypeContainers,
			types.EventActionContainersCreated,
			gomock.Any()).
		Return(nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)

	containerCheckBefore, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 0, len(containerCheckBefore))

	_, err := unitUnderTest.Create(context.Background(), container)
	testutil.AssertNil(t, err)

	containerCheckAfter, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 1, len(containerCheckAfter))
}

func TestCreateExistingContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctrID, container := getDefaultContainer()
	expectedErr := log.NewErrorf("container with id = %s already exists", ctrID)
	cache := map[string]*types.Container{}

	metapath := "../pkg/testutil/metapath/valid"

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache,
	)
	unitUnderTest.Load(context.Background())

	_, err := unitUnderTest.Create(context.Background(), container)

	testutil.AssertNotNil(t, err)
	testutil.AssertError(t, expectedErr, err)
}

func TestValidateWrongContainerName(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metapath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	cache := map[string]*types.Container{}

	_, container := getDefaultContainer()
	container.Name = "@abs"
	expectedErr := log.NewErrorf("invalid container name format : %s", container.Name)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metapath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)

	_, err := unitUnderTest.Create(context.Background(), container)

	testutil.AssertNotNil(t, err)
	testutil.AssertError(t, expectedErr, err)
}

func TestDeleteContainerFromManager(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()
	expectedState := types.State{
		Pid:        container.State.Pid,
		Status:     types.Dead,
		Dead:       true,
		Paused:     false,
		Running:    false,
		Restarting: false,
		Exited:     false,
		ExitCode:   container.State.ExitCode,
		Error:      container.State.Error,
		StartedAt:  container.State.StartedAt,
		FinishedAt: container.State.FinishedAt,
	}

	mockRepository.EXPECT().Prune().Times(1)

	mockCtrClient.EXPECT().
		DestroyContainer(gomock.Any(), gomock.Eq(container), gomock.Any(), gomock.Any()).
		Times(1)

	mockCtrClient.EXPECT().
		ReleaseContainerResources(gomock.Any(), gomock.Eq(container)).
		Times(1)

	mockNetworkManager.EXPECT().
		ReleaseNetworkResources(gomock.Any(), gomock.Eq(container)).
		Times(1)

	mockEventsManager.EXPECT().Publish(
		gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersRemoved,
		matchers.MatchesContainerID(ctrID),
	)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerState(ctrID, expectedState)).
		Times(1)

	mockRepository.
		EXPECT().
		Delete(ctrID).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)

	unitUnderTest.Load(ctx)
	containerCheck, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 1, len(containerCheck))

	// Act
	unitUnderTest.Remove(context.Background(), containerCheck[0].ID, true, &types.StopOpts{Force: true})

	containerCheckAfter, err := unitUnderTest.List(context.Background())
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, 0, len(containerCheckAfter))

	//Check that the folder is empty
	testutil.AssertEqual(t, 0, len(cache))
}

func TestContainerPause(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()

	expectedState := types.State{
		Pid:        container.State.Pid,
		Status:     types.Paused,
		Dead:       false,
		Paused:     true,
		Running:    false,
		Restarting: false,
		Exited:     false,
		ExitCode:   container.State.ExitCode,
		Error:      container.State.Error,
		StartedAt:  container.State.StartedAt,
		FinishedAt: container.State.FinishedAt,
	}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerState(ctrID, expectedState)).
		Times(1)

	mockCtrClient.EXPECT().
		PauseContainer(gomock.Any(), gomock.Eq(container)).
		Return(nil).
		Times(1)

	mockEventsManager.EXPECT().Publish(gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersPaused,
		gomock.Any()).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerUnderTestBefore, _ := unitUnderTest.Get(ctx, ctrID)
	testutil.AssertFalse(t, containerUnderTestBefore.State.Paused)
	// Perform
	err := unitUnderTest.Pause(ctx, ctrID)
	containerUnderTestAfter, _ := unitUnderTest.Get(ctx, ctrID)

	// Assert
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, containerUnderTestAfter.State.Paused)

}

func TestContainerUnpause(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getPausedContainer()

	expectedState := types.State{
		Pid:        container.State.Pid,
		Status:     types.Running,
		Dead:       false,
		Paused:     false,
		Running:    true,
		Restarting: false,
		Exited:     false,
		ExitCode:   container.State.ExitCode,
		Error:      container.State.Error,
		StartedAt:  container.State.StartedAt,
		FinishedAt: container.State.FinishedAt,
	}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerState(ctrID, expectedState)).
		Times(1)

	mockCtrClient.EXPECT().
		UnpauseContainer(gomock.Any(), gomock.Eq(container)).
		Return(nil).
		Times(1)

	mockEventsManager.EXPECT().Publish(gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersResumed,
		gomock.Any()).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerUnderTestBefore, _ := unitUnderTest.Get(ctx, ctrID)
	testutil.AssertTrue(t, containerUnderTestBefore.State.Paused)
	// Perform
	err := unitUnderTest.Unpause(ctx, ctrID)
	containerUnderTestAfter, _ := unitUnderTest.Get(ctx, ctrID)

	// Assert
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, containerUnderTestAfter.State.Running)
}

func TestContainerStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getStoppedContainer()

	expectedState := types.State{
		Pid:        container.State.Pid,
		Status:     types.Running,
		Dead:       false,
		Paused:     false,
		Running:    true,
		Restarting: false,
		Exited:     false,
		ExitCode:   container.State.ExitCode,
		Error:      container.State.Error,
		StartedAt:  container.State.StartedAt,
		FinishedAt: container.State.FinishedAt,
	}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(gomock.Any()).
		Times(1)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerState(ctrID, expectedState)).
		Times(1)

	mockCtrClient.EXPECT().
		StartContainer(gomock.Any(), gomock.Eq(container), gomock.Any()).
		Return(int64(1000), nil).
		Times(1)

	mockEventsManager.EXPECT().Publish(gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersRunning,
		matchers.MatchesContainerID(ctrID)).
		Times(1)

	mockNetworkManager.EXPECT().
		Manage(gomock.Any(), gomock.Eq(container)).
		Times(1)

	mockNetworkManager.EXPECT().
		Connect(gomock.Any(), gomock.Eq(container)).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerUnderTestBefore, _ := unitUnderTest.Get(ctx, ctrID)
	testutil.AssertEqual(t, types.Stopped, containerUnderTestBefore.State.Status)

	// Start
	startErr := unitUnderTest.Start(ctx, ctrID)
	containerUnderTestAfter, _ := unitUnderTest.Get(ctx, ctrID)

	testutil.AssertNil(t, startErr)
	testutil.AssertTrue(t, containerUnderTestAfter.State.Running)
}

func TestStopContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()
	expectedState := types.State{
		Pid:        container.State.Pid,
		Status:     types.Stopped,
		Dead:       false,
		Paused:     false,
		Running:    false,
		Restarting: false,
		Exited:     false,
		ExitCode:   container.State.ExitCode,
		Error:      container.State.Error,
		StartedAt:  container.State.StartedAt,
		FinishedAt: container.State.FinishedAt,
	}
	destOpts := types.StopOpts{Force: true}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerState(ctrID, expectedState)).
		Times(1)

	mockEventsManager.EXPECT().Publish(gomock.Any(),
		types.EventTypeContainers,
		types.EventActionContainersStopped,
		matchers.MatchesContainerID(ctrID)).
		Times(1)

	mockCtrClient.EXPECT().DestroyContainer(
		gomock.Any(),
		gomock.Eq(container),
		gomock.Any(),
		gomock.Any()).
		MaxTimes(2)

	mockCtrClient.EXPECT().
		ReleaseContainerResources(gomock.Any(), matchers.MatchesContainerID(ctrID)).
		MinTimes(1).
		MaxTimes(1)

	mockNetworkManager.EXPECT().
		ReleaseNetworkResources(gomock.Any(), matchers.MatchesContainerID(ctrID)).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerUnderTestBefore, _ := unitUnderTest.Get(ctx, ctrID)
	testutil.AssertEqual(t, types.Running, containerUnderTestBefore.State.Status)

	stopErr := unitUnderTest.Stop(ctx, ctrID, &destOpts)
	containerAfterStop, _ := unitUnderTest.Get(ctx, ctrID)

	testutil.AssertNil(t, stopErr)
	testutil.AssertEqual(t, types.Stopped, containerAfterStop.State.Status)
}

func TestStopOpsTimeoutValidation(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/valid"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()
	timeOut := -100
	expectedError := log.NewErrorf("the timeout = %d shouldn't be negative", timeOut)
	destOpts := types.StopOpts{Force: true, Timeout: int64(timeOut)}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	stopErr := unitUnderTest.Stop(ctx, ctrID, &destOpts)

	testutil.AssertNotNil(t, stopErr)
	testutil.AssertError(t, expectedError, stopErr)
}

func TestStopOpsSignalValidation(t *testing.T) {
	// Set UP
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/valid"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	ctx := context.Background()
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()
	signal := "256"
	expectedError := log.NewErrorf("invalid signal = %s", signal)
	destOpts := types.StopOpts{Force: true, Signal: signal}

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	stopErr := unitUnderTest.Stop(ctx, ctrID, &destOpts)

	testutil.AssertNotNil(t, stopErr)
	testutil.AssertError(t, expectedError, stopErr)
}

func TestRestore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDefaultContainer()

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockRepository.
		EXPECT().
		Save(matchers.MatchesContainerID(ctrID)).
		Times(1)

	mockCtrClient.EXPECT().
		RestoreContainer(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	mockNetworkManager.EXPECT().
		Restore(gomock.Any(), gomock.Any()).
		Times(1)

	mockNetworkManager.EXPECT().
		Initialize(gomock.Any()).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerCheck, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 1, len(containerCheck))

	// Act
	unitUnderTest.Restore(context.Background())
}

func TestRestoreOnDeadContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	hook := test.NewGlobal()
	ctx := context.Background()
	cache := map[string]*types.Container{}
	_, container := getDeadContainer()

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)

	mockNetworkManager.EXPECT().
		Restore(gomock.Any(), gomock.Any()).
		Times(1)

	mockNetworkManager.EXPECT().
		Initialize(gomock.Any()).
		Times(1)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	containerCheck, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 1, len(containerCheck))

	// Act
	unitUnderTest.Restore(context.Background())

	// Assert that a log was added for dead container
	testutil.AssertEqual(t, 1, len(hook.Entries))

	containerCheckAfter, _ := unitUnderTest.List(context.Background())
	testutil.AssertEqual(t, 0, len(containerCheckAfter))
}

func TestAttach(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	metaPath := "../pkg/testutil/metapath/tmp"
	mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
	mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)
	mockEventsManager := eventsMock.NewMockContainerEventsManager(mockCtrl)
	mockRepository := mgrMock.NewMockcontainerRepository(mockCtrl)
	ctx := context.Background()
	cache := map[string]*types.Container{}
	ctrID, container := getDeadContainer()

	mockRepository.EXPECT().Prune().Times(1)

	mockRepository.EXPECT().
		ReadAll().
		Return([]*types.Container{container}, nil)
	attachConfig := &streams.AttachConfig{}

	mockCtrClient.EXPECT().
		AttachContainer(gomock.Any(),
			matchers.MatchesContainerID(ctrID),
			attachConfig)

	unitUnderTest := createContainerManagerWithCustomMocks(
		metaPath,
		mockCtrClient,
		mockNetworkManager,
		mockEventsManager,
		mockRepository,
		cache)
	unitUnderTest.Load(ctx)

	// Act
	err := unitUnderTest.Attach(context.Background(), ctrID, attachConfig)

	testutil.AssertNil(t, err)
}

func TestMetrics(t *testing.T) {
	const testCtrID = "test-ctr-id"
	metricsReportFull := &types.Metrics{
		CPU: &types.CPUStats{
			Used:  15000,
			Total: 150000,
		},
		Memory: &types.MemoryStats{
			Used:  1024 * 1024 * 1024,
			Total: 8 * 1024 * 1024 * 1024,
		},
		IO: &types.IOStats{
			Read:  1024,
			Write: 2028,
		},
		Network: &types.IOStats{
			Read:  2048,
			Write: 4096,
		},
		PIDs:      5,
		Timestamp: time.Now(),
	}

	tests := map[string]struct {
		ctr              *types.Container
		ctrStatsFailOnly bool
		addCtrToCache    bool
		mockExec         func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error)
	}{
		"test_missing_in_cache": {
			ctr: &types.Container{
				ID: testCtrID,
			},
			addCtrToCache: false,
			mockExec: func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error) {
				return nil, log.NewErrorf(noSuchContainerErrorMsg, testCtrID)
			},
		},
		"test_exited_container": {
			ctr: &types.Container{
				ID: testCtrID,
				State: &types.State{
					Exited:  true,
					Paused:  false,
					Running: false,
				},
			},
			addCtrToCache: true,
			mockExec: func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error) {
				return nil, nil
			},
		},
		"test_ctr_stats_error": {
			ctrStatsFailOnly: true,
			addCtrToCache:    true,
			ctr: &types.Container{
				ID: testCtrID,
				State: &types.State{
					Running: true,
				},
			},
			mockExec: func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error) {
				err := log.NewError("test error")
				client.EXPECT().GetContainerStats(ctx, ctr).Return(nil, nil, nil, uint64(0), time.Time{}, err)
				manager.EXPECT().Stats(ctx, ctr).Return(metricsReportFull.Network, nil)
				return &types.Metrics{
					Network:   metricsReportFull.Network,
					Timestamp: time.Now(),
				}, nil
			},
		},
		"test_net_stats_error": {
			ctr: &types.Container{
				ID: testCtrID,
				State: &types.State{
					Running: true,
				},
			},
			addCtrToCache: true,
			mockExec: func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error) {
				err := log.NewError("test error")
				client.EXPECT().GetContainerStats(ctx, ctr).Return(metricsReportFull.CPU, metricsReportFull.Memory, metricsReportFull.IO, uint64(metricsReportFull.PIDs), metricsReportFull.Timestamp, nil)
				manager.EXPECT().Stats(ctx, ctr).Return(nil, err)
				return &types.Metrics{
					CPU:       metricsReportFull.CPU,
					Memory:    metricsReportFull.Memory,
					IO:        metricsReportFull.IO,
					Network:   nil,
					Timestamp: metricsReportFull.Timestamp,
					PIDs:      metricsReportFull.PIDs,
				}, nil
			},
		},
		"test_ctr_net_stats_error": {
			ctr: &types.Container{
				ID: testCtrID,
				State: &types.State{
					Running: true,
				},
			},
			addCtrToCache: true,
			mockExec: func(ctx context.Context, ctr *types.Container, client *ctrMock.MockContainerAPIClient, manager *networkMock.MockContainerNetworkManager) (*types.Metrics, error) {
				err := log.NewError("test error")
				client.EXPECT().GetContainerStats(ctx, ctr).Return(nil, nil, nil, uint64(0), time.Time{}, err)
				manager.EXPECT().Stats(ctx, ctr).Return(nil, err)
				errs := &errorUtil.CompoundError{}
				errs.Append(err, err)
				return nil, errs
			},
		},
	}
	// run tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// init mocks
			mockCtrClient := ctrMock.NewMockContainerAPIClient(mockCtrl)
			mockNetworkManager := networkMock.NewMockContainerNetworkManager(mockCtrl)

			testMgr := &containerMgr{
				ctrClient:  mockCtrClient,
				netMgr:     mockNetworkManager,
				containers: make(map[string]*types.Container),
			}
			if testCase.addCtrToCache {
				testMgr.containers[testCase.ctr.ID] = testCase.ctr
			}
			ctx := context.Background()
			expectedMetrics, expectedErr := testCase.mockExec(ctx, testCase.ctr, mockCtrClient, mockNetworkManager)
			metrics, err := testMgr.Metrics(ctx, testCase.ctr.ID)

			testutil.AssertError(t, expectedErr, err)
			if expectedErr != nil {
				testutil.AssertNil(t, metrics)
			} else if expectedMetrics == nil {
				testutil.AssertNil(t, metrics)
			} else {
				testutil.AssertEqual(t, expectedMetrics.CPU, metrics.CPU)
				testutil.AssertEqual(t, expectedMetrics.Memory, metrics.Memory)
				testutil.AssertEqual(t, expectedMetrics.IO, metrics.IO)
				testutil.AssertEqual(t, expectedMetrics.PIDs, metrics.PIDs)
				testutil.AssertEqual(t, expectedMetrics.Network, metrics.Network)

				if testCase.ctrStatsFailOnly {
					testutil.AssertNotNil(t, metrics.Timestamp)
				} else {
					testutil.AssertEqual(t, expectedMetrics.Timestamp, metrics.Timestamp)
				}
			}
		})
	}
}

func getDeadContainer() (string, *types.Container) {
	containerID := "dead-container"
	pathToContatiner := filepath.Join("../pkg/testutil/metapath/valid/containers/", containerID, "/config.json")
	container := readContainerFormFS(pathToContatiner)

	return containerID, container
}

func getDefaultContainer() (string, *types.Container) {
	containerID := "61aff3dc-1f31-420b-883a-686165e1b06b"
	pathToContatiner := filepath.Join("../pkg/testutil/metapath/valid/containers/", containerID, "/config.json")
	container := readContainerFormFS(pathToContatiner)

	return containerID, container
}

func getPausedContainer() (string, *types.Container) {
	containerID := "paused-container"
	pathToContatiner := filepath.Join("../pkg/testutil/metapath/valid/containers/", containerID, "/config.json")
	container := readContainerFormFS(pathToContatiner)

	return containerID, container
}

func getStoppedContainer() (string, *types.Container) {
	containerID := "stopped-container"
	pathToContatiner := filepath.Join("../pkg/testutil/metapath/valid/containers/", containerID, "/config.json")
	container := readContainerFormFS(pathToContatiner)

	return containerID, container
}

func createContainerManagerWithCustomMocks(
	metaPath string,
	mockCtrClient ctr.ContainerAPIClient,
	mockNetworkManager network.ContainerNetworkManager,
	mockEventsManager events.ContainerEventsManager,
	mockRepository containerRepository,
	containersCache map[string]*types.Container,
) containerMgr {
	return containerMgr{
		metaPath:               metaPath,
		execPath:               testRootExec,
		defaultCtrsStopTimeout: testDefaultContainerStopTimeout,
		ctrClient:              mockCtrClient,
		netMgr:                 mockNetworkManager,
		eventsMgr:              mockEventsManager,
		containers:             containersCache,
		containersLock:         sync.RWMutex{},
		restartCtrsMgrCache:    newRestartMgrCache(),
		containerRepository:    mockRepository,
	}
}
