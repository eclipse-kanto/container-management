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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/golang/mock/gomock"
)

var (
	testContainerToCreate = &types.Container{
		Image: types.Image{
			Name: testContainerImage,
		},
	}
	testContainerCreated = &types.Container{
		Image: types.Image{
			Name: testContainerToCreate.Image.Name,
		},
		ID: testContainerID,
	}
	expectedCtr = toAPIContainerConfig(testContainerConfig)
)

func TestContainerFactoryCreateFeature(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupContainerFactoryStorageMock(controller)
	setupThingMock(controller)
	ctrFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)

	resultFeature := ctrFactory.(*containerFactoryFeature).createFeature()
	testutil.AssertEqual(t, ContainerFactoryFeatureID, resultFeature.GetID())
	testutil.AssertEqual(t, containerFactoryFeatureDefinition, resultFeature.GetDefinition()[0].String())
	testutil.AssertEqual(t, 1, len(resultFeature.GetDefinition()))
}

func TestContainerFactoryOperationsHandlerCreate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	expectedCtr.Image.Name = testContainerImage

	tests := map[string]struct {
		operation     string
		args          interface{}
		mockExecution mockExecCreate
	}{
		// create without config
		"test_container_factory_operations_handler_create_no_errors": {
			operation: containerFactoryFeatureOperationCreate,
			args: &createArgs{
				ImageRef: testContainerImage,
				Start:    false,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateNoErrors,
		},
		"test_container_factory_operations_handler_create_no_image_ref": {
			operation: containerFactoryFeatureOperationCreate,
			args: &createArgs{
				ImageRef: "",
				Start:    false,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateNoRefError,
		},
		"test_container_factory_operations_handler_create_error": {
			operation: containerFactoryFeatureOperationCreate,
			args: &createArgs{
				ImageRef: testContainerImage,
				Start:    false,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateErrorReturned,
		},
		"test_container_factory_operations_handler_create_start_error": {
			operation: containerFactoryFeatureOperationCreate,
			args: &createArgs{
				ImageRef: testContainerImage,
				Start:    true,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateStartErrorReturned,
		},
		"test_container_factory_operations_handler_create_start_no_errors": {
			operation: containerFactoryFeatureOperationCreate,
			args: &createArgs{
				ImageRef: testContainerImage,
				Start:    true,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateStartNoErrors,
		},
		"test_container_factory_operations_handler_create_args_invalid": {
			operation:     containerFactoryFeatureOperationCreate,
			args:          testCreateOperationsHandlerInvalidArgs,
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreategInvalidArgsType,
		},
		"test_container_factory_operations_handler_create_args_type": {
			operation:     containerFactoryFeatureOperationCreate,
			args:          `{invalid : \"invalid\"}`,
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateInvalidCreateArgs,
		},
		// create with config
		"test_container_factory_operations_handler_create_config_no_errors": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    true,
				Config:   testContainerConfig,
			},
			mockExecution: mockExecCreateStartWithConfigNoErrors,
		},
		"test_container_factory_operations_handler_create_config_no_image_ref": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: "",
				Start:    false,
			},
			mockExecution: mockExecCreateStartWithConfigNoImageRefError,
		},

		"test_container_factory_operations_handler_create_config_error": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    false,
				Config:   testContainerConfig,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigErrorReturned,
		},
		"test_container_factory_operations_handler_create_config_start_error": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    true,
				Config:   testContainerConfig,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigStartErrorReturned,
		},
		"test_container_factory_operations_handler_create_config_start_no_errors": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    true,
				Config:   testContainerConfig,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigStartNoErrors,
		},
		"test_container_factory_operations_handler_create_config_nil": {
			operation: containerFactoryFeatureOperationCreateWithConfig,
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    true,
				Config:   nil,
			}, mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigNilConfig,
		},
		"test_container_factory_operations_handler_create_config_args_invalid": {
			operation:     containerFactoryFeatureOperationCreateWithConfig,
			args:          testCreateOperationsHandlerInvalidArgs,
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigInvalidArgsType,
		},
		"test_container_factory_operations_handler_create_config_args_type": {
			operation:     containerFactoryFeatureOperationCreateWithConfig,
			args:          `{invalid : \"invalid\"}`,
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigInvalidCreateArgs,
		},
		// default
		"test_container_factory_operations_handler_default": {
			operation: "unsupported-operation",
			args: &createWithConfigArgs{
				ImageRef: testContainerImage,
				Start:    true,
				Config:   testContainerConfig,
			},
			mockExecution: mockExecContainerFactoryFeatureOperationsHandlerDefault,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupContainerFactoryStorageMock(controller)
			setupThingMock(controller)
			ctrFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)

			expectedRunErr := testCase.mockExecution(t)

			result, resultErr := ctrFactory.(*containerFactoryFeature).featureOperationsHandler(testCase.operation, testCase.args)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

var (
	testCreateOperationsHandlerInvalidArgs = make(chan int)
)

type mockExecCreate func(t *testing.T) error

// create without config mocks

func mockExecContainerFactoryFeatureOperationsHandlerCreateNoErrors(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(1).Return(testContainerCreated, nil)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateNoRefError(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(0)
	return log.NewError("imageRef must be set")
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateErrorReturned(t *testing.T) error {
	err := log.NewError("error while creating")
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(1).Return(nil, err)
	return err
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateStartErrorReturned(t *testing.T) error {
	err := log.NewError("error while starting")
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(err)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateStartNoErrors(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(nil)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreategInvalidArgsType(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateInvalidCreateArgs(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type things.createArgs")
}

// create with config mocks

func mockExecCreateStartWithConfigNoErrors(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), expectedCtr).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(nil)
	return nil
}
func mockExecCreateStartWithConfigNoImageRefError(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return log.NewError("imageRef must be set")
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigErrorReturned(t *testing.T) error {
	err := log.NewError("error while creating")
	mockContainerManager.EXPECT().Create(gomock.Any(), expectedCtr).Times(1).Return(nil, err)
	return err
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigStartErrorReturned(t *testing.T) error {
	err := log.NewError("error while starting")
	mockContainerManager.EXPECT().Create(gomock.Any(), expectedCtr).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(err)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigStartNoErrors(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), expectedCtr).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(nil)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigNilConfig(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), testContainerToCreate).Times(1).Return(testContainerCreated, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerCreated.ID).Times(1).Return(nil)
	return nil
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigInvalidArgsType(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
}

func mockExecContainerFactoryFeatureOperationsHandlerCreateWithConfigInvalidCreateArgs(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type things.createWithConfigArgs")
}

func mockExecContainerFactoryFeatureOperationsHandlerDefault(t *testing.T) error {
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return client.NewMessagesSubjectNotFound(log.NewErrorf("unsupported operation %s", "unsupported-operation").Error())
}
