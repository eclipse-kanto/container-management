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

package services

import (
	"context"
	"errors"
	"testing"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	pbsysinfo "github.com/eclipse-kanto/container-management/containerm/api/services/sysinfo"
	pbcontainerstypes "github.com/eclipse-kanto/container-management/containerm/api/types/containers"
	pbsysinfotypes "github.com/eclipse-kanto/container-management/containerm/api/types/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocksmgrspb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	mockssysinfopb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/util/protobuf"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
)

const (
	projectVersion    = "test-version"
	apiVersion        = "test-api-version"
	containerID       = "test-ctr"
	containerImageID  = "host/group/image:tag"
	containerName     = "test-ctr-name"
	containerID2      = "test-ctr-2"
	containerImageID2 = "host/group/image2:tag"
	containerName2    = "test-ctr-name-2"
)

var (
	mockContainerManager  *mocksmgrspb.MockContainerManager
	mockSystemInfoManager *mockssysinfopb.MockSystemInfoManager
	testCtrsService       containers
	testSysInfoService    systemInfo
	testCtx               context.Context
)

func setup(controller *gomock.Controller) {
	mockContainerManager = mocksmgrspb.NewMockContainerManager(controller)
	testCtrsService = containers{
		mgr: mockContainerManager,
	}
	mockSystemInfoManager = mockssysinfopb.NewMockSystemInfoManager(controller)
	testSysInfoService = systemInfo{
		sysInfoMgr: mockSystemInfoManager,
	}
	testCtx = context.Background()
}

// SystemInfo -------------------------------------------------------------
type testProjectInfoArgs struct {
	ctx     context.Context
	request *empty.Empty
}
type mockExecProjectInfo func(args testProjectInfoArgs) (*pbsysinfo.ProjectInfoResponse, error)

func TestProjectInfo(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testProjectInfoArgs
		mockExecution mockExecProjectInfo
	}{
		"test_project_info_no_errs": {
			args: testProjectInfoArgs{
				ctx:     testCtx,
				request: &empty.Empty{},
			},
			mockExecution: mockExecProjectInfoNoErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedProjectInfo, expectedRunErr := testCase.mockExecution(testCase.args)

			projectInfo, resultErr := testSysInfoService.ProjectInfo(testCase.args.ctx, testCase.args.request)
			// assert project info
			testutil.AssertEqual(t, expectedProjectInfo, projectInfo)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Containers -------------------------------------------------------------
type testCreateArgs struct {
	ctx     context.Context
	request *pbcontainers.CreateContainerRequest
}
type mockExecCreate func(args testCreateArgs) (*pbcontainers.CreateContainerResponse, error)

func TestCreate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testCreateArgs
		mockExecution mockExecCreate
	}{
		"test_create_no_errs": {
			args: testCreateArgs{
				ctx: testCtx,
				request: &pbcontainers.CreateContainerRequest{
					Container: &pbcontainerstypes.Container{
						Id:   containerID,
						Name: containerName,
						Image: &pbcontainerstypes.Image{
							Name: containerImageID,
						},
					},
				},
			},
			mockExecution: mockExecCreateNoErrors,
		},
		"test_create_errs": {
			args: testCreateArgs{
				ctx: testCtx,
				request: &pbcontainers.CreateContainerRequest{
					Container: &pbcontainerstypes.Container{
						Id:   containerID,
						Name: containerName,
						Image: &pbcontainerstypes.Image{
							Name: containerImageID,
						},
					},
				},
			},
			mockExecution: mockExecCreateErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Create(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testGetArgs struct {
	ctx     context.Context
	request *pbcontainers.GetContainerRequest
}
type mockExecGet func(args testGetArgs) (*pbcontainers.GetContainerResponse, error)

func TestGet(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testGetArgs
		mockExecution mockExecGet
	}{
		"test_get_no_errs": {
			args: testGetArgs{
				ctx: testCtx,
				request: &pbcontainers.GetContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecGetNoErrors,
		},
		"test_get_errs": {
			args: testGetArgs{
				ctx: testCtx,
				request: &pbcontainers.GetContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecGetErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Get(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testListArgs struct {
	ctx     context.Context
	request *pbcontainers.ListContainersRequest
}
type mockExecList func(args testListArgs) (*pbcontainers.ListContainersResponse, error)

func TestList(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testListArgs
		mockExecution mockExecList
	}{
		"test_list_no_errs": {
			args: testListArgs{
				ctx:     testCtx,
				request: &pbcontainers.ListContainersRequest{},
			},
			mockExecution: mockExecListNoErrors,
		},
		"test_list_errs": {
			args: testListArgs{
				ctx:     testCtx,
				request: &pbcontainers.ListContainersRequest{},
			},
			mockExecution: mockExecListErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.List(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testStartArgs struct {
	ctx     context.Context
	request *pbcontainers.StartContainerRequest
}
type mockExecStart func(args testStartArgs) (*empty.Empty, error)

func TestStart(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testStartArgs
		mockExecution mockExecStart
	}{
		"test_start_no_errs": {
			args: testStartArgs{
				ctx: testCtx,
				request: &pbcontainers.StartContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecStartNoErrors,
		},
		"test_start_errs": {
			args: testStartArgs{
				ctx: testCtx,
				request: &pbcontainers.StartContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecStartErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Start(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testStopArgs struct {
	ctx     context.Context
	request *pbcontainers.StopContainerRequest
}
type mockExecStop func(args testStopArgs) (*empty.Empty, error)

func TestStop(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testStopArgs
		mockExecution mockExecStop
	}{
		"test_stop_no_errs": {
			args: testStopArgs{
				ctx: testCtx,
				request: &pbcontainers.StopContainerRequest{
					Id: containerID,
					StopOptions: &pbcontainerstypes.StopOptions{
						Timeout: 20,
						Force:   true,
					},
				},
			},
			mockExecution: mockExecStopNoErrors,
		},
		"test_stop_default_opts": {
			args: testStopArgs{
				ctx: testCtx,
				request: &pbcontainers.StopContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecStopDefaultOpts,
		},
		"test_stop_errs": {
			args: testStopArgs{
				ctx: testCtx,
				request: &pbcontainers.StopContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecStopErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Stop(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testPauseArgs struct {
	ctx     context.Context
	request *pbcontainers.PauseContainerRequest
}
type mockExecPause func(args testPauseArgs) (*empty.Empty, error)

func TestPause(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testPauseArgs
		mockExecution mockExecPause
	}{
		"test_pause_no_errs": {
			args: testPauseArgs{
				ctx: testCtx,
				request: &pbcontainers.PauseContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecPauseNoErrors,
		},
		"test_pause_errs": {
			args: testPauseArgs{
				ctx: testCtx,
				request: &pbcontainers.PauseContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecPauseErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Pause(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testUnpauseArgs struct {
	ctx     context.Context
	request *pbcontainers.UnpauseContainerRequest
}
type mockExecUnpause func(args testUnpauseArgs) (*empty.Empty, error)

func TestUnpause(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testUnpauseArgs
		mockExecution mockExecUnpause
	}{
		"test_unpause_no_errs": {
			args: testUnpauseArgs{
				ctx: testCtx,
				request: &pbcontainers.UnpauseContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecUnpauseNoErrors,
		},
		"test_unpause_errs": {
			args: testUnpauseArgs{
				ctx: testCtx,
				request: &pbcontainers.UnpauseContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecUnpauseErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Unpause(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testRemoveArgs struct {
	ctx     context.Context
	request *pbcontainers.RemoveContainerRequest
}
type mockExecRemove func(args testRemoveArgs) (*empty.Empty, error)

func TestRemove(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testRemoveArgs
		mockExecution mockExecRemove
	}{
		"test_remove_no_errs": {
			args: testRemoveArgs{
				ctx: testCtx,
				request: &pbcontainers.RemoveContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecRemoveNoErrors,
		},
		"test_remove_force": {
			args: testRemoveArgs{
				ctx: testCtx,
				request: &pbcontainers.RemoveContainerRequest{
					Id:    containerID,
					Force: true,
				},
			},
			mockExecution: mockExecRemoveNoErrors,
		},
		"test_remove_errs": {
			args: testRemoveArgs{
				ctx: testCtx,
				request: &pbcontainers.RemoveContainerRequest{
					Id: containerID,
				},
			},
			mockExecution: mockExecRemoveErrors,
		},
		"test_remove_timeout": {
			args: testRemoveArgs{
				ctx: testCtx,
				request: &pbcontainers.RemoveContainerRequest{
					Id:    containerID,
					Force: true,
					StopOptions: &pbcontainerstypes.StopOptions{
						Timeout: 20,
						Force:   true,
					},
				},
			},
			mockExecution: mockExecRemoveNoErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Remove(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testUpdateArgs struct {
	ctx     context.Context
	request *pbcontainers.UpdateContainerRequest
}
type mockExecUpdate func(args testUpdateArgs) (*empty.Empty, error)

func TestUpdate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testUpdateArgs
		mockExecution mockExecUpdate
	}{
		"test_update_no_errs": {
			args: testUpdateArgs{
				ctx: testCtx,
				request: &pbcontainers.UpdateContainerRequest{
					Id: containerID,
					UpdateOptions: &pbcontainerstypes.UpdateOptions{
						RestartPolicy: &pbcontainerstypes.RestartPolicy{Type: string(types.Always)},
						Resources:     &pbcontainerstypes.Resources{Memory: "100m"},
					},
				},
			},
			mockExecution: mockExecUpdateNoErrors,
		},
		"test_update_errs": {
			args: testUpdateArgs{
				ctx: testCtx,
				request: &pbcontainers.UpdateContainerRequest{
					Id:            containerID,
					UpdateOptions: &pbcontainerstypes.UpdateOptions{},
				},
			},
			mockExecution: mockExecUpdateErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Update(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testRenameArgs struct {
	ctx     context.Context
	request *pbcontainers.RenameContainerRequest
}
type mockExecRename func(args testRenameArgs) (*empty.Empty, error)

func TestRename(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testRenameArgs
		mockExecution mockExecRename
	}{
		"test_rename_no_errs": {
			args: testRenameArgs{
				ctx: testCtx,
				request: &pbcontainers.RenameContainerRequest{
					Id:   containerID,
					Name: containerName2,
				},
			},
			mockExecution: mockExecRenameNoErrors,
		},
		"test_rename_errs": {
			args: testRenameArgs{
				ctx: testCtx,
				request: &pbcontainers.RenameContainerRequest{
					Id:   containerID,
					Name: containerName2,
				},
			},
			mockExecution: mockExecRenameErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRsp, expectedRunErr := testCase.mockExecution(testCase.args)

			rsp, resultErr := testCtrsService.Rename(testCase.args.ctx, testCase.args.request)
			// assert response
			testutil.AssertEqual(t, expectedRsp, rsp)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testLogsArgs struct {
	request *pbcontainers.GetLogsRequest
	srv     *fakeLogsServer
}
type mockExecLogs func(args testLogsArgs) error

type fakeLogsServer struct {
	pbcontainers.Containers_LogsServer
}

func newFakeClient() *fakeLogsServer {
	return &fakeLogsServer{}
}

func (f fakeLogsServer) Send(m *pbcontainers.GetLogsResponse) error {
	return nil
}

func TestLogs(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testLogsArgs
		mockExecution mockExecLogs
	}{
		"test_logs_no_errs": {
			args: testLogsArgs{
				request: &pbcontainers.GetLogsRequest{
					Id:   containerID,
					Tail: 10,
				},
				srv: newFakeClient(),
			},
			mockExecution: mockExecLogsNoErrors,
		},
		"test_logs_no_host_cfg_errs": {
			args: testLogsArgs{
				request: &pbcontainers.GetLogsRequest{
					Id:   containerID,
					Tail: 10,
				},
				srv: newFakeClient(),
			},
			mockExecution: mockExecLogsNoHostConfig,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testCtrsService.Logs(testCase.args.request, *testCase.args.srv)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Mock executions -------------------------------------------------------------
// SystemInfo -------------------------------------------------------------
// ProjectInfo -------------------------------------------------------------
func mockExecProjectInfoNoErrors(args testProjectInfoArgs) (*pbsysinfo.ProjectInfoResponse, error) {
	pbProjectInfo := &pbsysinfotypes.ProjectInfo{
		ProjectVersion: projectVersion,
		ApiVersion:     apiVersion,
	}
	pbResponse := &pbsysinfo.ProjectInfoResponse{
		ProjectInfo: pbProjectInfo,
	}
	projectInfo := protobuf.ToInternalProjectInfo(pbProjectInfo)
	mockSystemInfoManager.EXPECT().GetProjectInfo().Times(1).Return(projectInfo)
	return pbResponse, nil
}

// Containers -------------------------------------------------------------
// Create -------------------------------------------------------------
func mockExecCreateNoErrors(args testCreateArgs) (*pbcontainers.CreateContainerResponse, error) {
	pbCtr := &pbcontainerstypes.Container{
		Id:    containerID,
		Name:  containerName,
		Image: &pbcontainerstypes.Image{Name: containerImageID},
	}
	pbResponse := &pbcontainers.CreateContainerResponse{
		Container: pbCtr,
	}
	ctr := protobuf.ToInternalContainer(pbCtr)
	mockContainerManager.EXPECT().Create(args.ctx, gomock.Eq(protobuf.ToInternalContainer(args.request.Container))).Times(1).Return(ctr, nil)
	return pbResponse, nil
}

func mockExecCreateErrors(args testCreateArgs) (*pbcontainers.CreateContainerResponse, error) {
	err := errors.New("failed to create container")
	mockContainerManager.EXPECT().Create(args.ctx, gomock.Eq(protobuf.ToInternalContainer(args.request.Container))).Times(1).Return(nil, err)
	return nil, err
}

// Get -------------------------------------------------------------
func mockExecGetNoErrors(args testGetArgs) (*pbcontainers.GetContainerResponse, error) {
	pbCtr := &pbcontainerstypes.Container{
		Id:    containerID,
		Name:  containerName,
		Image: &pbcontainerstypes.Image{Name: containerImageID},
	}
	pbResponse := &pbcontainers.GetContainerResponse{
		Container: pbCtr,
	}
	mockContainerManager.EXPECT().Get(args.ctx, args.request.Id).Times(1).Return(protobuf.ToInternalContainer(pbCtr), nil)
	return pbResponse, nil
}

func mockExecGetErrors(args testGetArgs) (*pbcontainers.GetContainerResponse, error) {
	err := errors.New("failed to get container")
	mockContainerManager.EXPECT().Get(args.ctx, args.request.Id).Times(1).Return(nil, err)
	return nil, err
}

// List -------------------------------------------------------------
func mockExecListNoErrors(args testListArgs) (*pbcontainers.ListContainersResponse, error) {
	pbCtr := &pbcontainerstypes.Container{
		Id:    containerID,
		Name:  containerName,
		Image: &pbcontainerstypes.Image{Name: containerImageID},
	}
	pbCtr2 := &pbcontainerstypes.Container{
		Id:    containerID2,
		Name:  containerName2,
		Image: &pbcontainerstypes.Image{Name: containerImageID2},
	}
	pbResponse := &pbcontainers.ListContainersResponse{
		Containers: []*pbcontainerstypes.Container{pbCtr, pbCtr2},
	}
	ctrs := []*types.Container{protobuf.ToInternalContainer(pbCtr), protobuf.ToInternalContainer(pbCtr2)}
	mockContainerManager.EXPECT().List(args.ctx).Times(1).Return(ctrs, nil)
	return pbResponse, nil
}

func mockExecListErrors(args testListArgs) (*pbcontainers.ListContainersResponse, error) {
	err := errors.New("failed to list container")
	mockContainerManager.EXPECT().List(args.ctx).Times(1).Return(nil, err)
	return &pbcontainers.ListContainersResponse{Containers: []*pbcontainerstypes.Container{}}, err
}

// Start -------------------------------------------------------------
func mockExecStartNoErrors(args testStartArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Start(args.ctx, args.request.Id).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecStartErrors(args testStartArgs) (*empty.Empty, error) {
	err := errors.New("failed to start container")
	mockContainerManager.EXPECT().Start(args.ctx, args.request.Id).Times(1).Return(err)
	return nil, err
}

// Stop -------------------------------------------------------------
func mockExecStopNoErrors(args testStopArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Stop(args.ctx, args.request.Id, gomock.Eq(protobuf.ToInternalStopOptions(args.request.StopOptions))).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecStopDefaultOpts(args testStopArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Stop(args.ctx, args.request.Id, nil).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecStopErrors(args testStopArgs) (*empty.Empty, error) {
	err := errors.New("failed to stop container")
	mockContainerManager.EXPECT().Stop(args.ctx, args.request.Id, gomock.Eq(protobuf.ToInternalStopOptions(args.request.StopOptions))).Times(1).Return(err)
	return nil, err
}

// Pause -------------------------------------------------------------
func mockExecPauseNoErrors(args testPauseArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Pause(args.ctx, args.request.Id).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecPauseErrors(args testPauseArgs) (*empty.Empty, error) {
	err := errors.New("failed to pause container")
	mockContainerManager.EXPECT().Pause(args.ctx, args.request.Id).Times(1).Return(err)
	return nil, err
}

// Unpause -------------------------------------------------------------
func mockExecUnpauseNoErrors(args testUnpauseArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Unpause(args.ctx, args.request.Id).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecUnpauseErrors(args testUnpauseArgs) (*empty.Empty, error) {
	err := errors.New("failed to unpause container")
	mockContainerManager.EXPECT().Unpause(args.ctx, args.request.Id).Times(1).Return(err)
	return nil, err
}

// Remove -------------------------------------------------------------
func mockExecRemoveNoErrors(args testRemoveArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Remove(args.ctx, args.request.Id, args.request.Force, args.request.Removecache, gomock.Eq(protobuf.ToInternalStopOptions(args.request.StopOptions))).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecRemoveForce(args testRemoveArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Remove(args.ctx, args.request.Id, true, false, gomock.Eq(protobuf.ToInternalStopOptions(args.request.StopOptions))).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

//TODO add tests for timeout

func mockExecRemoveErrors(args testRemoveArgs) (*empty.Empty, error) {
	err := errors.New("failed to remove container")
	mockContainerManager.EXPECT().Remove(args.ctx, args.request.Id, args.request.Force, args.request.Removecache, gomock.Eq(protobuf.ToInternalStopOptions(args.request.StopOptions))).Times(1).Return(err)
	return nil, err
}

// Update -------------------------------------------------------------
func mockExecUpdateNoErrors(args testUpdateArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Update(args.ctx, args.request.Id, gomock.Eq(protobuf.ToInternalUpdateOptions(args.request.UpdateOptions))).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecUpdateErrors(args testUpdateArgs) (*empty.Empty, error) {
	err := errors.New("failed to update container")
	mockContainerManager.EXPECT().Update(args.ctx, args.request.Id, gomock.Eq(protobuf.ToInternalUpdateOptions(args.request.UpdateOptions))).Times(1).Return(err)
	return nil, err
}

// Rename -------------------------------------------------------------
func mockExecRenameNoErrors(args testRenameArgs) (*empty.Empty, error) {
	mockContainerManager.EXPECT().Rename(args.ctx, args.request.Id, gomock.Eq(args.request.Name)).Times(1).Return(nil)
	return &empty.Empty{}, nil
}

func mockExecRenameErrors(args testRenameArgs) (*empty.Empty, error) {
	err := errors.New("failed to rename container")
	mockContainerManager.EXPECT().Rename(args.ctx, args.request.Id, gomock.Eq(args.request.Name)).Times(1).Return(err)
	return nil, err
}

// Logs -------------------------------------------------------------
func mockExecLogsNoErrors(args testLogsArgs) error {
	pbCtr := &pbcontainerstypes.Container{
		Id:    containerID,
		Name:  containerName,
		Image: &pbcontainerstypes.Image{Name: containerImageID},
		HostConfig: &pbcontainerstypes.HostConfig{
			LogConfig: &pbcontainerstypes.LogConfiguration{
				DriverConfig: &pbcontainerstypes.LogDriverConfiguration{
					Type:    "json-file",
					RootDir: "../pkg/testutil/logs",
				},
			},
		},
	}
	mockContainerManager.EXPECT().Get(context.Background(), args.request.Id).Times(1).Return(protobuf.ToInternalContainer(pbCtr), nil)
	return nil
}

func mockExecLogsNoHostConfig(args testLogsArgs) error {
	err := errors.New("no host config for container test-ctr")
	pbCtr := &pbcontainerstypes.Container{
		Id:    containerID,
		Name:  containerName,
		Image: &pbcontainerstypes.Image{Name: containerImageID},
	}

	mockContainerManager.EXPECT().Get(context.Background(), args.request.Id).Times(1).Return(protobuf.ToInternalContainer(pbCtr), nil)
	return err
}
