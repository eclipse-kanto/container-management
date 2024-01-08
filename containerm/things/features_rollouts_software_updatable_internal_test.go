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
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/rollouts/api/features"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/golang/mock/gomock"
)

const (
	testHTTPServerImageURLPathValid   = "/image/valid"
	testHTTPServerImageURLPathInvalid = "/image/invalid"
	testSoftwareName                  = testContainerImage
	testSoftwareVersion               = "1.0"
	testLastOperationProperty         = "status/lastOperation"
	testLastFailedOperationProperty   = "status/lastFailedOperation"
	testCorrelationID                 = "1000"
)

var (
	testSoftwareUpdatable managedFeature
	testSUFeature         model.Feature
	mockHTTPServer        *httptest.Server

	testCtrs = []*types.Container{
		{
			Image: types.Image{
				Name: testContainerImage,
			},
			ID: "id1"},
		{
			Image: types.Image{
				Name: testContainerImage,
			},
			ID: "id2"},
	}
)

func TestCreateSUPFeature(t *testing.T) {
	defer os.RemoveAll(testThingsStoragePath)
	setupSUFeature(t)
	testSUFeature = testSoftwareUpdatable.(*softwareUpdatable).createFeature()
	testutil.AssertEqual(t, SoftwareUpdatableFeatureID, testSUFeature.GetID())
	testutil.AssertEqual(t, 1, len(testSUFeature.GetDefinition()))
	testutil.AssertEqual(t, client.NewDefinitionID(SoftwareUpdatableDefinitionNamespace, SoftwareUpdatableDefinitionName, SoftwareUpdatableDefinitionVersion).String(), testSUFeature.GetDefinition()[0].String())
	testutil.AssertEqual(t, 1, len(testSUFeature.GetProperties()))
	testutil.AssertNotNil(t, testSUFeature.GetProperties()[softwareUpdatablePropertyNameStatus])
	testutil.AssertEqual(t, reflect.TypeOf(&features.SoftwareUpdatableStatus{}), reflect.TypeOf(testSUFeature.GetProperties()[softwareUpdatablePropertyNameStatus]))
	testSUPStatus := testSUFeature.GetProperties()[softwareUpdatablePropertyNameStatus].(*features.SoftwareUpdatableStatus)
	testutil.AssertEqual(t, testSUPStatus.SoftwareModuleType, containersSoftwareUpdatableAgentType)
}

var (
	testSoftwareModule = &datatypes.SoftwareModuleID{
		Name:    testSoftwareName,
		Version: testSoftwareVersion,
	}
	testDependencyDescription = &datatypes.DependencyDescription{
		Name: testSoftwareName,
	}
)

func TestSUFeatureOperationsHandler(t *testing.T) {
	defer os.RemoveAll(testThingsStoragePath)
	tests := map[string]struct {
		operation     string
		opts          interface{}
		mockExecution mockExecSUOperation
	}{
		"test_su_feature_operations_handler_install": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: &datatypes.SoftwareModuleID{
							Name:    testSoftwareName,
							Version: testSoftwareVersion,
						},
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testImageJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: mockExecInstallNoErrors,
		},
		"test_su_feature_operations_handler_install_invalid_hash": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: &datatypes.SoftwareModuleID{
							Name:    testSoftwareName,
							Version: testSoftwareVersion,
						},
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: "invalid-hash",
								},
							},
						},
					},
				},
			},
			mockExecution: mockExecInstallNoErrorInvalidChecksum,
		},
		"test_su_feature_operations_handler_install_no_artifacts": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: &datatypes.SoftwareModuleID{
							Name:    testSoftwareName,
							Version: testSoftwareVersion,
						},
						Artifacts: []*datatypes.SoftwareArtifactAction{},
					},
				}},
			mockExecution: mockExecInstallErrorNoAtrifacts,
		},
		"test_su_feature_operations_handler_install_no_modules": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				SoftwareModules: []*datatypes.SoftwareModuleAction{},
			},
			mockExecution: mockExecInstallErrorNoModules,
		},
		"test_su_feature_operations_handler_remove_no_dep_descr": {
			operation: softwareUpdatableOperationRemove,
			opts: datatypes.RemoveAction{
				Software: []*datatypes.DependencyDescription{},
			},
			mockExecution: mockExecRemoveNoDependencyDescription,
		},
		"test_su_feature_operations_handler_install_error_creating": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: &datatypes.SoftwareModuleID{
							Name:    testSoftwareName,
							Version: testSoftwareVersion,
						},
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testImageJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: mockExecInstallErrorWhileCreatingContainer,
		},
		"test_su_feature_operations_handler_install_error_starting": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: &datatypes.SoftwareModuleID{
							Name:    testSoftwareName,
							Version: testSoftwareVersion,
						},
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testImageJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: mockExecInstallErrorWhileCreatingContainer,
		},
		"test_su_feature_operations_handler_remove_forced": {
			operation: softwareUpdatableOperationRemove,
			opts: datatypes.RemoveAction{
				CorrelationID: testCorrelationID,
				Forced:        true,
				Software: []*datatypes.DependencyDescription{
					testDependencyDescription,
					testDependencyDescription,
					testDependencyDescription,
				},
			},
			mockExecution: mockExecSURemoveForced,
		},
		"test_su_feature_operations_handler_remove_no_such_container": {
			operation: softwareUpdatableOperationRemove,
			opts: datatypes.RemoveAction{
				CorrelationID: testCorrelationID,
				Software: []*datatypes.DependencyDescription{
					testDependencyDescription,
				},
			},
			mockExecution: mockExecSURemoveNoSuchContainer,
		},
		"test_su_feature_operations_handler_default": {
			operation:     "unsupportedOperation",
			mockExecution: mockExecSUDefault,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			defer os.RemoveAll(testThingsStoragePath)
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(t)

			result, resultErr := testSoftwareUpdatable.(*softwareUpdatable).operationsHandler(testCase.operation, testCase.opts)
			time.Sleep(1 * time.Second)
			testutil.AssertEqual(t, result, nil)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

var (
	// operation install
	testOperationInstallStatusStarted = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.Started,
	}
	testOperationInstallStatusDownloading = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.Downloading,
	}
	testOperationInstallStatusDownloaded = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.Downloaded,
	}
	testOperationInstallStatusInstalling = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.Installing,
	}
	testOperationInstallStatusInstalled = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.Installed,
	}
	testOperationInstallStatusFinishedSuccess = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.FinishedSuccess,
	}
	testOperationInstallErrorWhileCreatingMessage        = "error while creating"
	testOperationInstallStatusFinishedErrorWhileCreating = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.FinishedError,
		Message:        testOperationInstallErrorWhileCreatingMessage,
	}
	testOperationInstallErrorWhileStartingMessage        = "error while starting"
	testOperationInstallStatusFinishedErrorWhileStarting = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.FinishedError,
		Message:        testOperationInstallErrorWhileStartingMessage,
	}
	testOperationInstallStatusFinishedRejectedInvalidHash = &datatypes.OperationStatus{
		CorrelationID:  testCorrelationID,
		SoftwareModule: testSoftwareModule,
		Status:         datatypes.FinishedRejected,
		Message:        "the provided input hash is either invalid, not a hex string or the length exceeds 16 bytes",
	}
	// operation remove
	testOperationRemoveStatusRemoving = &datatypes.OperationStatus{
		CorrelationID: testCorrelationID,
		Software:      []*datatypes.DependencyDescription{testDependencyDescription},
		Status:        datatypes.Removing,
	}
	testOperationRemovingStatusRemoved = &datatypes.OperationStatus{
		CorrelationID: testCorrelationID,
		Software:      []*datatypes.DependencyDescription{testDependencyDescription},
		Status:        datatypes.Removed,
	}
	testOperationRemoveErrorWhileRemovingMessage   = "error while removing"
	testOperationRemoveStatusRemovingFinishedError = &datatypes.OperationStatus{
		CorrelationID: testCorrelationID,
		Software:      []*datatypes.DependencyDescription{testDependencyDescription},
		Status:        datatypes.FinishedError,
		Message:       testOperationRemoveErrorWhileRemovingMessage,
	}
	testOperationRemoveErrorNoSuchContainerMessage                = log.NewErrorf("container with ID = %s does not exist", testSoftwareName).Error()
	testOperationRemoveStatusRemovingFinishedErrorNoSuchContainer = &datatypes.OperationStatus{
		CorrelationID: testCorrelationID,
		Software:      []*datatypes.DependencyDescription{testDependencyDescription},
		Status:        datatypes.FinishedError,
		Message:       testOperationRemoveErrorNoSuchContainerMessage,
	}
	testOperationRemoveStatusRemovingFinishedSuccess = &datatypes.OperationStatus{
		CorrelationID: testCorrelationID,
		Software:      []*datatypes.DependencyDescription{testDependencyDescription, testDependencyDescription},
		Status:        datatypes.FinishedSuccess,
	}
)

type mockExecSUOperation func(t *testing.T) error

func mockExecInstallNoErrors(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(nil, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	gomock.InOrder(
		// Started
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusStarted).Times(1).Return(nil),
		// Downloading
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloading).Times(1).Return(nil),
		// Downloaded
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloaded).Times(1).Return(nil),
		// Installing
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusInstalling).Times(1).Return(nil),
		// Installed
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusInstalled).Times(1).Return(nil),
		// Finished Success
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusFinishedSuccess).Times(1).Return(nil),
	)

	return nil
}

func mockExecInstallNoErrorInvalidChecksum(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0).Return(nil, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0).Return(nil)
	gomock.InOrder(
		// Started
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusStarted).Times(1).Return(nil),
		// Downloading
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloading).Times(1).Return(nil),
		// Failed
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastFailedOperationProperty, testOperationInstallStatusFinishedRejectedInvalidHash).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusFinishedRejectedInvalidHash).Times(1).Return(nil),
	)

	return nil
}

func mockExecInstallErrorNoModules(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0).Return(nil, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0).Return(nil)
	mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), gomock.Any(), gomock.Any()).Times(0).Return(nil)
	return client.NewMessagesParameterInvalidError("there are no SoftwareModules to be installed")
}

func mockExecInstallErrorNoAtrifacts(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0).Return(nil, nil)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0).Return(nil)
	mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), gomock.Any(), gomock.Any()).Times(0).Return(nil)
	return client.NewMessagesParameterInvalidError("there are no SoftwareArtifacts referenced for SoftwareModule [Name.version] = [%s.%s]", testSoftwareName, testSoftwareVersion)
}

func mockExecRemoveNoDependencyDescription(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Remove(gomock.Any(), gomock.Any(), gomock.Any(), nil).Times(0).Return(nil)
	mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), gomock.Any(), gomock.Any()).Times(0).Return(nil)
	return client.NewMessagesParameterInvalidError("there are no DependencyDescriptions to be removed")
}

func mockExecInstallErrorWhileCreatingContainer(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(nil, log.NewError(testOperationInstallErrorWhileCreatingMessage))
	//mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	gomock.InOrder(
		// Started
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusStarted).Times(1).Return(nil),
		// Downloading
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloading).Times(1).Return(nil),
		// Downloaded
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloaded).Times(1).Return(nil),
		// Installing
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusInstalling).Times(1).Return(nil),
		// Finished Error
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastFailedOperationProperty, testOperationInstallStatusFinishedErrorWhileCreating).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusFinishedErrorWhileCreating).Times(1).Return(nil),
	)

	return nil
}

func mockExecInstallErrorWhileStartingContainer(t *testing.T) error {
	setupSUFeature(t)
	mockContainerManager.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1).Return(log.NewError(testOperationInstallErrorWhileStartingMessage))
	gomock.InOrder(
		// Started
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusStarted).Times(1).Return(nil),
		// Downloading
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloading).Times(1).Return(nil),
		// Downloaded
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusDownloaded).Times(1).Return(nil),
		// Installing
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusInstalling).Times(1).Return(nil),
		// Finished Error
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastFailedOperationProperty, testOperationInstallStatusFinishedErrorWhileStarting).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationInstallStatusFinishedErrorWhileStarting).Times(1).Return(nil),
	)

	return nil
}

// Simulate - 1 of each in order: success, error, success after error
func mockExecSURemoveForced(t *testing.T) error {
	setupSUFeature(t)
	gomock.InOrder(
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemoving).Times(1).Return(nil),
		mockContainerManager.EXPECT().Get(gomock.Any(), testSoftwareName).Return(&types.Container{}, nil),
		mockContainerManager.EXPECT().Remove(gomock.Any(), testSoftwareName, true, nil).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemovingStatusRemoved).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemoving).Times(1).Return(nil),
		mockContainerManager.EXPECT().Get(gomock.Any(), testSoftwareName).Return(&types.Container{}, nil),
		mockContainerManager.EXPECT().Remove(gomock.Any(), testSoftwareName, true, nil).Return(log.NewErrorf(testOperationRemoveErrorWhileRemovingMessage)),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemoving).Times(1).Return(nil),
		mockContainerManager.EXPECT().Get(gomock.Any(), testSoftwareName).Return(&types.Container{}, nil),
		mockContainerManager.EXPECT().Remove(gomock.Any(), testSoftwareName, true, nil).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemovingStatusRemoved).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastFailedOperationProperty, testOperationRemoveStatusRemovingFinishedError).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemovingFinishedError).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemovingFinishedSuccess).Times(1).Return(nil),
	)

	return nil
}

func mockExecSURemoveNoSuchContainer(t *testing.T) error {
	setupSUFeature(t)
	gomock.InOrder(
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemoving).Times(1).Return(nil),
		mockContainerManager.EXPECT().Get(gomock.Any(), testSoftwareName).Return(nil, nil),
		mockContainerManager.EXPECT().Remove(gomock.Any(), testSoftwareName, true, nil).Times(0).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastFailedOperationProperty, testOperationRemoveStatusRemovingFinishedErrorNoSuchContainer).Times(1).Return(nil),
		mockThing.EXPECT().SetFeatureProperty(testSUFeature.GetID(), testLastOperationProperty, testOperationRemoveStatusRemovingFinishedErrorNoSuchContainer).Times(1).Return(nil),
	)
	return nil
}

func mockExecSUDefault(t *testing.T) error {
	setupSUFeature(t)
	return client.NewMessagesSubjectNotFound(log.NewErrorf("unsupported operation called [operationId = %s]", "unsupportedOperation").Error())
}

func setupSUFeature(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupDummyHTTPServer(true)
	testutil.AssertNil(t, setupThingsContainerManager(controller))
	testSoftwareUpdatable = newSoftwareUpdatable(mockThing, mockContainerManager, mockEventsManager)
}

// HTTP Server mock -----------------------------------------------

/*
	Basically the expected outgoing http request could be asserted by either

creating a wrapper http client and mocking it OR by mocking an http server.
The second approach looks cleaner at the moment,
as it does not require changes in the source code.
*/
func setupDummyHTTPServer(plain bool) {
	handler := http.NewServeMux()
	handler.HandleFunc(testHTTPServerImageURLPathValid, validURLHandler)
	handler.HandleFunc(testHTTPServerImageURLPathInvalid, invalidURLHandler)

	if plain {
		mockHTTPServer = httptest.NewServer(handler)
		return
	}
	mockHTTPServer = httptest.NewTLSServer(handler)

}

const testImageJSONHash = "61d996255cb129284fef6ccb1952279c"

func validURLHandler(w http.ResponseWriter, r *http.Request) {
	// hash value of the json is testImageJSONHash
	_, _ = w.Write([]byte(`{"image":{"name":"image:latest"}}`))
}

func invalidURLHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(""))
}
