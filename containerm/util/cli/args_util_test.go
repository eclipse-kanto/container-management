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

package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/client"
	"github.com/golang/mock/gomock"
)

func TestValidateContainerByNameArgsSingle(t *testing.T) {

	type mockExec func(mockClient *mocks.MockClient) (*types.Container, error)
	const sContainerName = "ContainerName"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)
	testCtx := context.Background()
	aFilter := client.WithName(sContainerName)
	aFilterWrong := client.WithName("Wrong")

	tests := map[string]struct {
		args                  []string
		providedContainerName string
		exec                  mockExec
	}{
		"test_with_args_and_ctr_name": {
			args:                  []string{"1"},
			providedContainerName: sContainerName,
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				return nil, log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
			},
		},
		"test_with_no_args_and_no_ctr_name": {
			args:                  nil,
			providedContainerName: "",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				return nil, log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
			},
		},
		"test_with_one_arg_and_no_ctr_name": {
			args:                  []string{"1"},
			providedContainerName: "",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				container := &types.Container{ID: "1"}
				mockClient.EXPECT().Get(testCtx, container.ID).Return(container, nil).Times(1)
				return container, nil
			},
		},
		"test_with_args_and_no_ctr_name_get_error": {
			args:                  []string{"1"},
			providedContainerName: "",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				err := log.NewError("An Error")
				var container *types.Container = nil
				mockClient.EXPECT().Get(testCtx, "1").Return(container, err).Times(1)
				return container, err
			},
		},
		"test_with_args_and_no_ctr_name_get_error_no_container": {
			args:                  []string{"2"},
			providedContainerName: "",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				err := log.NewError("The requested container with ID = 2 was not found.")
				mockClient.EXPECT().Get(testCtx, "2").Return(nil, nil).Times(1)
				return nil, err
			},
		},
		"test_with_ctr_name_and_no_args": {
			args:                  nil,
			providedContainerName: sContainerName,
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				cnt := &types.Container{Name: sContainerName}
				containers := []*types.Container{cnt}
				mockClient.EXPECT().List(testCtx, gomock.AssignableToTypeOf(aFilter)).Return(containers, nil).Times(1)
				return containers[0], nil
			},
		},
		"test_with_ctr_name_and_no_args_get_error": {
			args:                  nil,
			providedContainerName: "Wrong",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				err := errors.New("")
				containers := []*types.Container{nil}
				mockClient.EXPECT().List(testCtx, gomock.AssignableToTypeOf(aFilter)).Return(containers, err).Times(1)
				return containers[0], err
			},
		},
		"test_with_ctr_name_and_no_args_get_error_no_ctrs": {
			args:                  nil,
			providedContainerName: "Wrong",
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				err := log.NewError("The requested container with name = Wrong was not found. Try using an ID instead.")
				mockClient.EXPECT().List(testCtx, gomock.AssignableToTypeOf(aFilterWrong)).Return(nil, nil).Times(1)
				return nil, err
			},
		},
		"test_with_ctr_name_and_no_args_get_error_more_ctrs": {
			args:                  nil,
			providedContainerName: sContainerName,
			exec: func(client *mocks.MockClient) (*types.Container, error) {
				cnt := &types.Container{Name: sContainerName}
				containers := []*types.Container{cnt, cnt}
				err := log.NewError("There are more than one containers with name = ContainerName. Try using an ID instead.")
				mockClient.EXPECT().List(testCtx, gomock.AssignableToTypeOf(aFilter)).Return(containers, nil).Times(1)
				return nil, err
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedContainer, expectedError := testCase.exec(mockClient)
			got, err := ValidateContainerByNameArgsSingle(testCtx, testCase.args, testCase.providedContainerName, mockClient)
			testutil.AssertEqual(t, expectedContainer, got)
			testutil.AssertError(t, expectedError, err)
		})
	}
}
