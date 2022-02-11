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
	"context"
	"reflect"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/snapshots"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	loggerMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/logger"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/golang/mock/gomock"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var testContainerID = "test-container-id"

func TestCtrdClientCreateContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrd.NewMockcontainerLogsManager(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockSpi := ctrd.NewMockcontainerdSpi(mockCtrl)
	mockImage := containerdMocks.NewMockImage(mockCtrl)

	testClient := &containerdClient{
		ioMgr:   mockIoMgr,
		logsMgr: mockLogMgr,
		spi:     mockSpi,
	}

	testCtr := &types.Container{
		ID: testContainerID,
		Image: types.Image{
			Name: "test.host/name:latest",
		},
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
		IOConfig: &types.IOConfig{
			OpenStdin: false,
			Tty:       false,
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		mockExec func() error
	}{
		"test_error_initialise_IO": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialise IO for container ID = %s", testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_error_initialize_logger": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_error_get_image": {
			mockExec: func() error {
				err := log.NewErrorf("error while trying to get container image with ID = %s for container ID = %s ", testCtr.Image.Name, testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(nil, err)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_create_container_without_error": {
			mockExec: func() error {
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().PrepareSnapshot(ctx, testCtr.ID, mockImage).Return(nil)
				mockSpi.EXPECT().MountSnapshot(ctx, testCtr.ID, rootFSPathDefault)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec()
			actualError := testClient.CreateContainer(ctx, testCtr, "")
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestCtrdClientDestroyContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockSpi := ctrd.NewMockcontainerdSpi(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)

	ctx := context.Background()
	stopOpts := &types.StopOpts{}

	notExistingContainerID := "test-id"

	tests := map[string]struct {
		testClient *containerdClient
		testCtr    *types.Container
		mockExec   func() error
	}{
		"test_error_not_existing_container": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
				ioMgr: mockIoMgr,
				spi:   mockSpi,
			},
			testCtr: &types.Container{
				ID: notExistingContainerID,
			},
			mockExec: func() error {
				err := log.NewErrorf("container with ID = test-id does not exist")
				mockIoMgr.EXPECT().ClearIO(notExistingContainerID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, notExistingContainerID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, notExistingContainerID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_error_kill_task": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
							task: mockTask,
						},
					},
				},
				ioMgr: mockIoMgr,
				spi:   mockSpi,
			},
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func() error {
				err := log.NewErrorf("Kill Error")
				mockTask.EXPECT().Kill(ctx, util.ToSignal(stopOpts.Signal)).Return(err)
				return err
			},
		},
		"test_destroy_container_without_error": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
				ioMgr: mockIoMgr,
				spi:   mockSpi,
			},
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func() error {
				mockIoMgr.EXPECT().ClearIO(testContainerID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testContainerID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testContainerID, rootFSPathDefault).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec()
			if expectedError == nil {
				testCase.testClient.DestroyContainer(ctx, testCase.testCtr, stopOpts, true)
				testutil.AssertTrue(t, len(testCase.testClient.ctrdCache.cache) == 0)
			} else {
				_, _, actualError := testCase.testClient.DestroyContainer(ctx, testCase.testCtr, stopOpts, true)
				testutil.AssertError(t, expectedError, actualError)
			}
		})
	}
}

func TestCtrdClientStartContainer(t *testing.T) {
	const (
		testCtrTaskPID uint32 = 123
	)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSpi := ctrd.NewMockcontainerdSpi(mockCtrl)
	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrd.NewMockcontainerLogsManager(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockImage := containerdMocks.NewMockImage(mockCtrl)
	mockContainer := containerdMocks.NewMockContainer(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)
	mockIo := containerdMocks.NewMockIO(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: newContainerInfoCache(),
		ioMgr:     mockIoMgr,
		logsMgr:   mockLogMgr,
		spi:       mockSpi,
	}
	ctx := context.Background()

	tests := map[string]struct {
		testCtr  *types.Container
		mockExec func(testCtr *types.Container) error
	}{
		"test_error_check_existing_container": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error trying to check for existing container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testContainerID).Return(nil, err)
				return err
			},
		},
		"test_error_existing_container": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("container with ID = %s already exists", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(mockContainer, nil)
				return err
			},
		},
		"test_error_snapshot_for_container_not_exist": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("snapshot for container with ID = %s does not exist", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, err)
				return err
			},
		},
		"test_error_missing_container_image": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("missing image ID = %s for container with ID = %s", testCtr.Image.Name, testContainerID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(nil, err)
				return err
			},
		},
		"test_error_creating_new_container": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error creating new container with ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(nil, err)
				return err
			},
		},
		"test_error_initialise_IO": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("failed to initialise IO for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_error_initialize_logger": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockContainer.EXPECT().Delete(ctx)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				return err
			},
		},
		"test_error_creating_task": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error creating task for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any()).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				cioCreator := func(id string) (cio.IO, error) { return mockIo, nil }
				mockIoMgr.EXPECT().NewCioCreator(testCtr.IOConfig.Tty).Return(cioCreator)
				mockSpi.EXPECT().CreateTask(ctx, mockContainer, gomock.Any()).Return(nil, err)
				mockIoMgr.EXPECT().ResetIO(testCtr.ID)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_create_container_without_error": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				cioCreator := func(id string) (cio.IO, error) { return mockIo, nil }
				mockIoMgr.EXPECT().NewCioCreator(testCtr.IOConfig.Tty).Return(cioCreator)
				mockSpi.EXPECT().CreateTask(ctx, mockContainer, gomock.Any() /*for now..*/).Return(mockTask, nil)
				mockContainer.EXPECT().ID().Return(testCtr.ID)
				resChan := make(<-chan containerd.ExitStatus)
				mockTask.EXPECT().Wait(context.TODO()).Return(resChan, nil)
				mockTask.EXPECT().Start(ctx).Return(nil)
				mockTask.EXPECT().Pid().Return(testCtrTaskPID).Times(2)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec(testCase.testCtr)
			_, actualError := testClient.StartContainer(ctx, testCase.testCtr, "")
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestAttachContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)

	testClient := &containerdClient{
		ioMgr: mockIoMgr,
	}

	ctx := context.Background()
	testCtr := &types.Container{
		ID: "test-id",
		IOConfig: &types.IOConfig{
			OpenStdin: true,
			Tty:       true,
		},
	}

	expectedError := log.NewErrorf("failed to initialise IO for container ID = test-id")

	mockIoMgr.EXPECT().GetIO(testCtr.ID).Return(nil)
	mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, expectedError)

	actualError := testClient.AttachContainer(ctx, testCtr, &streams.AttachConfig{})
	testutil.AssertError(t, expectedError, actualError)
}

func TestPauseContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
					task: mockTask,
				},
			},
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		arg      *types.Container
		mockExec func(context context.Context, mockTask *containerdMocks.MockTask) error
	}{
		"test_error_missing_container_to_pause": {
			arg: &types.Container{
				ID: "test-container",
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				return log.NewErrorf("missing container to pause")
			},
		},
		"test_pause_container_with_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				err := log.NewErrorf("test pause task error")
				mockTask.EXPECT().Pause(context).Return(err)
				return err
			},
		},
		"test_pause_container_without_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				mockTask.EXPECT().Pause(context).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec(ctx, mockTask)
			actualError := testClient.PauseContainer(ctx, testCase.arg)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestUnpauseContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
					task: mockTask,
				},
			},
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		arg      *types.Container
		mockExec func(context context.Context, mockTask *containerdMocks.MockTask) error
	}{
		"test_error_missing_container_to_unpause": {
			arg: &types.Container{
				ID: "test-container",
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				return log.NewErrorf("missing container to unpause")
			},
		},
		"test_unpause_container_with_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				err := log.NewErrorf("test unpause task error")
				mockTask.EXPECT().Resume(context).Return(err)
				return err
			},
		},
		"test_unpause_container_without_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				mockTask.EXPECT().Resume(context).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec(ctx, mockTask)
			actualError := testClient.UnpauseContainer(ctx, testCase.arg)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestListContainers(t *testing.T) {
	tests := map[string]struct {
		testClient *containerdClient
		want       []*types.Container
	}{
		"test_without_containers": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{},
			},
			want: make([]*types.Container, 0),
		},
		"test_with_containers": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
			},
			want: []*types.Container{
				{
					ID: testContainerID,
				},
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, _ := testCase.testClient.ListContainers(context.Background())
			testutil.AssertEqual(t, got, testCase.want)
		})
	}
}

func TestGetContainerInfo(t *testing.T) {
	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
				},
			},
		},
	}

	tests := map[string]struct {
		arg       string
		want      *types.Container
		wantError error
	}{
		"test_existing_container": {
			arg: testContainerID,
			want: &types.Container{
				ID: testContainerID,
			},
			wantError: nil,
		},
		"test_not_existing_container": {
			arg:       "not-existing-container-id",
			want:      nil,
			wantError: log.NewErrorf("missing container with ID = not-existing-container-id"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := testClient.GetContainerInfo(context.Background(), testCase.arg)
			if testCase.wantError != nil {
				testutil.AssertEqual(t, testCase.want, got)
			} else {
				testutil.AssertError(t, testCase.wantError, err)
			}
		})
	}
}

func TestRestoreContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSpi := ctrd.NewMockcontainerdSpi(mockCtrl)
	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrd.NewMockcontainerLogsManager(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockContainer := containerdMocks.NewMockContainer(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: newContainerInfoCache(),
		ioMgr:     mockIoMgr,
		logsMgr:   mockLogMgr,
		spi:       mockSpi,
	}
	ctx := context.Background()
	testCtr := &types.Container{
		ID: testContainerID,
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
		IOConfig: &types.IOConfig{
			OpenStdin: false,
		},
	}

	tests := map[string]struct {
		mockExec func() error
	}{
		"test_error_snapshot_not_exist": {
			mockExec: func() error {
				err := log.NewErrorf("snapshot for container with ID = %s does not exist", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, err)
				return err
			},
		},
		"test_error_retrieve_container": {
			mockExec: func() error {
				err := log.NewErrorf("failed to retrieve container ID = %s from containerd while restoring", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(nil, err)
				return err
			},
		},
		"test_error_initialising IO": {
			mockExec: func() error {
				err := log.NewErrorf("error while initialising IO for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				return err
			},
		},
		"test_error_init_log_driver": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				return err
			},
		},
		"test_error_load_task": {
			mockExec: func() error {
				err := log.NewErrorf("error loading task for container ID = %s while restoring", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockIoMgr.EXPECT().NewCioAttach(testCtr.ID)
				mockSpi.EXPECT().LoadTask(gomock.Any(), mockContainer, gomock.Any()).Return(nil, err)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_restore_container_without_error": {
			mockExec: func() error {
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockIoMgr.EXPECT().NewCioAttach(testCtr.ID)
				mockSpi.EXPECT().LoadTask(gomock.Any(), mockContainer, gomock.Any()).Return(mockTask, nil)
				mockContainer.EXPECT().ID().Return(testCtr.ID)
				resChan := make(<-chan containerd.ExitStatus)
				mockTask.EXPECT().Wait(context.TODO()).Return(resChan, nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec()
			actualError := testClient.RestoreContainer(ctx, testCtr)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestReleaseContainerResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)

	testClient := &containerdClient{
		ioMgr: mockIoMgr,
	}

	testCtr := &types.Container{
		ID: "test-id",
	}

	mockIoMgr.EXPECT().ResetIO(testCtr.ID).MaxTimes(1)

	testClient.ReleaseContainerResources(context.Background(), testCtr)
}

func TestSetContainerExitHooks(t *testing.T) {
	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{},
	}

	arg := func(container *types.Container, code int64, err error, oomKilled bool, cleanup func() error) error {
		return nil
	}

	testClient.SetContainerExitHooks(arg)

	got := testClient.ctrdCache.containerExitHooks
	testutil.AssertEqual(t, 1, len(got))

	if reflect.ValueOf(got[0]).Pointer() != reflect.ValueOf(arg).Pointer() {
		t.Errorf("SetContainerExitHooks() = %v, want %v", reflect.ValueOf(got[0]), reflect.ValueOf(arg))
	}
}

func TestCtrdClientDispose(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(mockCtrdWrapper *ctrd.MockcontainerdSpi) error
	}{
		"test_no_err": {
			mapExec: func(mockCtrdWrapper *ctrd.MockcontainerdSpi) error {
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(nil)
				return nil
			},
		},
		"test_err": {
			mapExec: func(mockCtrdWrapper *ctrd.MockcontainerdSpi) error {
				err := log.NewError("test error")
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(err)
				return err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockSpi := ctrd.NewMockcontainerdSpi(mockCtrl)
			// mock exec
			expectedErr := testData.mapExec(mockSpi)
			// init spi under test
			testClient := &containerdClient{
				spi:       mockSpi,
				ctrdCache: newContainerInfoCache(),
			}
			// test
			actualErr := testClient.Dispose(context.Background())
			testutil.AssertTrue(t, testClient.ctrdCache.isContainerdDead())
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestCtrdClientUpdateContainer(t *testing.T) {
	var (
		mockContainer *containerdMocks.MockContainer
		mockTask      *containerdMocks.MockTask
	)

	ctx := namespaces.WithNamespace(context.Background(), "test")
	spec, _ := oci.GenerateSpec(ctx, nil, &containers.Container{}) // populates default unix spec
	testCtrID := "test-update-id"
	unlimited := int64(-1)

	type mockExec func() error

	var testClient *containerdClient
	tests := map[string]struct {
		ctr       *types.Container
		resources *types.Resources
		exec      mockExec
	}{
		"test_with_initial_limits": {
			ctr: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					Resources: &types.Resources{
						Memory:     "200M",
						MemorySwap: "500M",
					},
				},
			},
			resources: &types.Resources{
				Memory: "200M",
			},
			exec: func() error {
				limit := int64(200 * 1024 * 1024)
				resources := &specs.LinuxResources{
					Devices: spec.Linux.Resources.Devices,
					Memory: &specs.LinuxMemory{
						Limit: &limit,
						Swap:  &unlimited,
					},
				}
				mockContainer.EXPECT().Spec(ctx).Return(spec, nil)
				mockTask.EXPECT().Update(ctx, matchers.MatchesUpdateTaskOpts(containerd.WithResources(resources))).Return(nil)
				return nil
			},
		},
		"test_with_initially_missing_limits": {
			ctr: &types.Container{
				ID:         testCtrID,
				HostConfig: &types.HostConfig{},
			},
			resources: &types.Resources{
				Memory:            "200M",
				MemoryReservation: "100M",
			},
			exec: func() error {
				limit := int64(200 * 1024 * 1024)
				reservation := int64(100 * 1024 * 1024)
				resources := &specs.LinuxResources{
					Devices: spec.Linux.Resources.Devices,
					Memory: &specs.LinuxMemory{
						Limit:       &limit,
						Reservation: &reservation,
					},
				}
				mockContainer.EXPECT().Spec(ctx).Return(spec, nil)
				mockTask.EXPECT().Update(ctx, matchers.MatchesUpdateTaskOpts(containerd.WithResources(resources))).Return(nil)
				return nil
			},
		},
		"test_no_error_nil_resources": {
			ctr: &types.Container{
				ID:         testCtrID,
				HostConfig: &types.HostConfig{},
			},
			exec: func() error {
				return nil
			},
		},
		"test_get_spec_err": {
			ctr: &types.Container{
				ID: testCtrID,
			},
			resources: &types.Resources{},
			exec: func() error {
				err := log.NewError("test error")
				mockContainer.EXPECT().Spec(ctx).Return(nil, err)
				return err
			},
		},
		"test_missing_ctr_err": {
			ctr: &types.Container{
				ID: testCtrID,
			},
			resources: &types.Resources{},
			exec: func() error {
				testClient.ctrdCache.cache[testCtrID] = nil
				return log.NewErrorf("missing container to update with ID = %s", testCtrID)
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {

			// init mock ctrl
			mockCtrl := gomock.NewController(t)

			// init mocks
			mockContainer = containerdMocks.NewMockContainer(mockCtrl)
			mockTask = containerdMocks.NewMockTask(mockCtrl)
			defer mockCtrl.Finish()

			testClient = &containerdClient{
				ctrdCache: newContainerInfoCache(),
			}
			testClient.ctrdCache.cache[testCase.ctr.ID] = &containerInfo{
				container: mockContainer,
				task:      mockTask,
			}

			expectedErr := testCase.exec()
			actualErr := testClient.UpdateContainer(ctx, testCase.ctr, testCase.resources)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
