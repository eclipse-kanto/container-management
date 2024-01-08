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
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/client"
)

const testUnmarshableArg = `{\"unmarshable\":\"arg\"}`

var (
	testValidUnmarshaled interface{} = nil
	testInvalidArg                   = make(chan int)
)

type mockExec func() error

// Start -------------------------------------------------------------
func TestFeatureOperationsHandlerStart(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		mockExecution mockExec
	}{
		"test_feature_operations_handler_start_no_errors": {
			mockExecution: func() error {
				mockContainerManager.EXPECT().Start(gomock.Any(), testContainerID).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_start_errors": {
			mockExecution: func() error {
				err := errors.New("failed to start container")
				mockContainerManager.EXPECT().Start(gomock.Any(), testContainerID).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationStart, nil)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Pause -------------------------------------------------------------
func TestFeatureOperationsHandlerPause(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		mockExecution mockExec
	}{
		"test_feature_operations_handler_pause_no_errors": {
			mockExecution: func() error {
				mockContainerManager.EXPECT().Pause(gomock.Any(), testContainerID).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_pause_errors": {
			mockExecution: func() error {
				err := errors.New("failed to pause container")
				mockContainerManager.EXPECT().Pause(gomock.Any(), testContainerID).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationPause, nil)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Resume -------------------------------------------------------------
func TestFeatureOperationsHandlerResume(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		mockExecution mockExec
	}{
		"test_feature_operations_handler_unpause_no_errors": {
			mockExecution: func() error {
				mockContainerManager.EXPECT().Unpause(gomock.Any(), testContainerID).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_unpause_errors": {
			mockExecution: func() error {
				err := errors.New("failed to unpause container")
				mockContainerManager.EXPECT().Unpause(gomock.Any(), testContainerID).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationResume, nil)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Stop -------------------------------------------------------------
func TestFeatureOperationsHandlerStop(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		mockExecution mockExec
	}{
		"test_feature_operations_handler_stop_no_errors": {
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), testContainerID, nil).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_stop_errors": {
			mockExecution: func() error {
				err := errors.New("failed to stop container")
				mockContainerManager.EXPECT().Stop(gomock.Any(), testContainerID, nil).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationStop, nil)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// StopWithOptions -------------------------------------------------------------
func TestFeatureOperationsHandlerStopWithOptions(t *testing.T) {
	_ = json.Unmarshal([]byte(`{"force":true,"timeout":50,"signal":"SIGKILL"}`), &testValidUnmarshaled)

	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		opts          interface{}
		mockExecution mockExec
	}{
		"test_feature_operations_handler_stop_with_options_no_errors": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), testContainerID, gomock.AssignableToTypeOf(&types.StopOpts{})).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_stop_with_options_invalid_args": {
			opts: testInvalidArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
				return err
			},
		},
		"test_feature_operations_handler_stop_with_options_error_unmarshalling": {
			opts: testUnmarshableArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type things.stopOptions")
				return err
			},
		},
		"test_feature_operations_handler_stop_with_options_error_while_stopping": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				err := log.NewError("error while stopping")
				mockContainerManager.EXPECT().Stop(gomock.Any(), testContainerID, gomock.AssignableToTypeOf(&types.StopOpts{})).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationStopWithOptions, testCase.opts)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Remove -------------------------------------------------------------
func TestFeatureOperationsHandlerRemove(t *testing.T) {
	_ = json.Unmarshal([]byte(`true`), &testValidUnmarshaled)

	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		opts          interface{}
		mockExecution mockExec
	}{
		"test_feature_operations_handler_remove_no_errors": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Remove(gomock.Any(), testContainerID, true, nil).Times(1)
				return nil
			},
		},
		"test_feature_operations_handler_remove_invalid_error": {
			opts: testInvalidArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
				return err
			},
		},
		"test_feature_operations_handler_remove_error_unmarshalling": {
			opts: testUnmarshableArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type bool")
				return err
			},
		},
		"test_feature_operations_handler-remove_error_while_removing": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				err := log.NewError("error while removing")
				mockContainerManager.EXPECT().Remove(gomock.Any(), testContainerID, true, nil).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationRemove, testCase.opts)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Rename -------------------------------------------------------------
func TestFeatureOperationsHandlerRename(t *testing.T) {
	const (
		testRenameArg            = "new-name"
		testRenameUnmarshableArg = 1
	)

	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		arg           interface{}
		mockExecution mockExec
	}{
		"test_feature_operations_handler_rename_no_errors": {
			arg: testRenameArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Rename(gomock.Any(), testContainerID, testRenameArg).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_rename_invalid_error": {
			arg: testInvalidArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Rename(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
				return err
			},
		},
		"test_feature_operations_handler_rename_error_unmarshalling": {
			arg: testRenameUnmarshableArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Rename(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: cannot unmarshal number into Go value of type string")
				return err
			},
		},
		"test_feature_operations_handler-rename_error_while_renaming": {
			arg: testRenameArg,
			mockExecution: func() error {
				err := log.NewError("error while renaming")
				mockContainerManager.EXPECT().Rename(gomock.Any(), testContainerID, testRenameArg).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationRename, testCase.arg)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Update -------------------------------------------------------------
func TestFeatureOperationsHandlerUpdate(t *testing.T) {
	_ = json.Unmarshal([]byte(`{"restartPolicy":{"type":"UNLESS_STOPPED","RetryTimeout":10,"MaxRetryCount":3}}`), &testValidUnmarshaled)

	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)

	tests := map[string]struct {
		opts          interface{}
		mockExecution mockExec
	}{
		"test_feature_operations_handler_update_no_errors": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Update(gomock.Any(), testContainerID, gomock.AssignableToTypeOf(&types.UpdateOpts{})).Times(1).Return(nil)
				return nil
			},
		},
		"test_feature_operations_handler_update_invalid_args": {
			opts: testInvalidArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
				return err
			},
		},
		"test_feature_operations_handler_update_error_unmarshalling": {
			opts: testUnmarshableArg,
			mockExecution: func() error {
				mockContainerManager.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				err := client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type things.updateOptions")
				return err
			},
		},
		"test_feature_operations_handler_update_error_while_updating": {
			opts: testValidUnmarshaled,
			mockExecution: func() error {
				err := log.NewError("error while updating")
				mockContainerManager.EXPECT().Update(gomock.Any(), testContainerID, gomock.AssignableToTypeOf(&types.UpdateOpts{})).Times(1).Return(err)
				return err
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution()

			result, resultErr := containerFeature.featureOperationsHandler(containerFeatureOperationUpdate, testCase.opts)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// UnsupportedOperation -------------------------------------------------------------
func TestFeatureOperationsUnsupportedOperation(t *testing.T) {
	const testUnsupportedOperationName = "doSomethingUnsupported"

	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)

	containerFeature := newContainerFeature(testContainerImage, testContainerName, testContainer, mockContainerManager)
	mockContainerManager.EXPECT().Start(gomock.Any(), testContainerID).Times(0)
	mockContainerManager.EXPECT().Stop(gomock.Any(), testContainerID, gomock.Any()).Times(0)
	mockContainerManager.EXPECT().Pause(gomock.Any(), testContainerID).Times(0)
	mockContainerManager.EXPECT().Unpause(gomock.Any(), testContainerID).Times(0)
	mockContainerManager.EXPECT().Remove(gomock.Any(), testContainerID, true, nil).Times(0)

	t.Run("test_feature_operations_handler_unsupported_operation", func(t *testing.T) {
		result, resultErr := containerFeature.featureOperationsHandler(testUnsupportedOperationName, nil)
		testutil.AssertEqual(t, result, nil)
		err := client.NewMessagesSubjectNotFound("unsupported operation " + testUnsupportedOperationName)
		testutil.AssertError(t, err, resultErr)
	})

}
