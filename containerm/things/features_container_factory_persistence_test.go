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
	"os"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

var (
	ctr1 = &types.Container{ID: "d90e86c2-9ea5-4e38-8cad-7e572faa44a4"}
	ctr2 = &types.Container{ID: "9ed4eba9-f8a8-483d-9379-d551b24340bf"}
)

const testStorageRoot = "../pkg/testutil/things/storage"

func TestContainerFactoryStorageRestore(t *testing.T) {
	tests := map[string]struct {
		storageRootPath            string
		storageThingID             string
		expectedRestoredContainers map[string]string
		expectedError              bool
	}{
		"test_restore_no_error": {
			storageRootPath:            testStorageRoot,
			storageThingID:             "com.test:device:edge:containers",
			expectedRestoredContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
			expectedError:              false,
		},
		"test_restore_read_not_existing": {
			storageRootPath: testStorageRoot,
			storageThingID:  "com.test:not-existing:edge:containers",
			expectedError:   false,
		},
		"test_restore_read_empty": {
			storageRootPath: testStorageRoot,
			storageThingID:  "com.test:empty:edge:containers",
			expectedError:   false,
		},
		"test_restore_read_broken_path": {
			storageRootPath: "\000",
			storageThingID:  "com.test:not-existing:edge:containers",
			expectedError:   true,
		},
		"test_restore_unmarshal_error": {
			storageRootPath: testStorageRoot,
			storageThingID:  "com.test:corrupted:edge:containers",
			expectedError:   true,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testCtrsStorage := newContainerFactoryStorage(testCase.storageRootPath, testCase.storageThingID)
			res, err := testCtrsStorage.Restore()
			if testCase.expectedError {
				testutil.AssertNotNil(t, err)
				testutil.AssertNil(t, res)
			} else {
				testutil.AssertNil(t, err)
				testutil.AssertEqual(t, testCase.expectedRestoredContainers, res)
			}
		})
	}
}

func TestUpdateContainerInfo(t *testing.T) {
	tests := map[string]struct {
		storageRootPath           string
		storageThingID            string
		expectedUpdatedContainers map[string]string
		expectedStoreErr          bool
	}{
		"test_update_no_error": {
			storageRootPath:           testStorageRoot,
			storageThingID:            "com.test:device-update:edge:containers",
			expectedUpdatedContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID), ctr2.ID: generateContainerFeatureID(ctr2.ID)},
		},
		"test_update_store_error": {
			storageRootPath:           "/root",
			expectedUpdatedContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID), ctr2.ID: generateContainerFeatureID(ctr2.ID)},
			storageThingID:            "com.test:not-existing:edge:containers",
			expectedStoreErr:          true,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testCtrsStorage := newContainerFactoryStorage(testCase.storageRootPath, testCase.storageThingID)
			defer func() {
				if err := os.Remove(testCtrsStorage.(*containerFactoryStorage).generateFileName()); err != nil {
					t.Log("failed to delete file")
				}
			}()

			testCtrsStorage.UpdateContainersInfo(map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID), ctr2.ID: generateContainerFeatureID(ctr2.ID)})

			testutil.AssertEqual(t, testCase.expectedUpdatedContainers, testCtrsStorage.(*containerFactoryStorage).managedContainerFeatures)
			if !testCase.expectedStoreErr {
				bytes, err := os.ReadFile(testCtrsStorage.(*containerFactoryStorage).generateFileName())
				testutil.AssertNil(t, err)
				actual := map[string]string{}
				err = json.Unmarshal(bytes, &actual)
				testutil.AssertNil(t, err)
				testutil.AssertEqual(t, testCase.expectedUpdatedContainers, actual)
			}
		})
	}
}

func TestStoreContainerInfo(t *testing.T) {
	tests := map[string]struct {
		storageRootPath          string
		storageThingID           string
		expectedStoredContainers map[string]string
		withNilMap               bool
		expectedStoreError       bool
	}{
		"test_store_no_error": {
			storageRootPath:          testStorageRoot,
			storageThingID:           "com.test:device-store:edge:containers",
			expectedStoredContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
		},
		"test_store_no_error_no_map": {
			storageRootPath:          testStorageRoot,
			storageThingID:           "com.test:device-store:edge:containers",
			expectedStoredContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
			withNilMap:               true,
		},
		"test_store_store_error": {
			storageRootPath:          "/root",
			expectedStoredContainers: map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
			storageThingID:           "com.test:not-existing:edge:containers",
			expectedStoreError:       true,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testCtrsStorage := newContainerFactoryStorage(testCase.storageRootPath, testCase.storageThingID)
			defer func() {
				if err := os.Remove(testCtrsStorage.(*containerFactoryStorage).generateFileName()); err != nil {
					t.Log("failed to delete file")
				}
			}()
			if testCase.withNilMap {
				testCtrsStorage.(*containerFactoryStorage).managedContainerFeatures = nil
			}

			testCtrsStorage.StoreContainerInfo(ctr1.ID)

			testutil.AssertEqual(t, testCase.expectedStoredContainers, testCtrsStorage.(*containerFactoryStorage).managedContainerFeatures)
			if !testCase.expectedStoreError {
				bytes, err := os.ReadFile(testCtrsStorage.(*containerFactoryStorage).generateFileName())
				testutil.AssertNil(t, err)
				actual := map[string]string{}
				err = json.Unmarshal(bytes, &actual)
				testutil.AssertNil(t, err)
				testutil.AssertEqual(t, testCase.expectedStoredContainers, actual)
			}
		})
	}
}

func TestDeleteContainerInfo(t *testing.T) {
	tests := map[string]struct {
		storageRootPath         string
		storageThingID          string
		expectedContainersAfter map[string]string
		currentContainers       map[string]string
		expectedStoreError      bool
	}{
		"test_delete_no_error": {
			storageRootPath:         testStorageRoot,
			storageThingID:          "com.test:device-del:edge:containers",
			currentContainers:       map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
			expectedContainersAfter: map[string]string{},
		},
		"test_delete_no_error_no_map": {
			storageRootPath:         testStorageRoot,
			storageThingID:          "com.test:device-del:edge:containers",
			currentContainers:       nil,
			expectedContainersAfter: nil,
		},
		"test_delete_store_error": {
			storageRootPath:         "/root",
			currentContainers:       map[string]string{ctr1.ID: generateContainerFeatureID(ctr1.ID)},
			expectedContainersAfter: map[string]string{},
			storageThingID:          "com.test:not-existing:edge:containers",
			expectedStoreError:      true,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testCtrsStorage := newContainerFactoryStorage(testCase.storageRootPath, testCase.storageThingID)
			defer func() {
				if err := os.Remove(testCtrsStorage.(*containerFactoryStorage).generateFileName()); err != nil {
					t.Log("failed to delete file")
				}
			}()

			if !testCase.expectedStoreError {
				testCtrsStorage.UpdateContainersInfo(testCase.currentContainers)
			} else {
				testCtrsStorage.(*containerFactoryStorage).managedContainerFeatures = testCase.currentContainers
			}

			testCtrsStorage.DeleteContainerInfo(ctr1.ID)

			testutil.AssertEqual(t, testCase.expectedContainersAfter, testCtrsStorage.(*containerFactoryStorage).managedContainerFeatures)
			if !testCase.expectedStoreError {
				bytes, err := os.ReadFile(testCtrsStorage.(*containerFactoryStorage).generateFileName())
				testutil.AssertNil(t, err)
				actual := map[string]string{}
				err = json.Unmarshal(bytes, &actual)
				testutil.AssertNil(t, err)
				testutil.AssertEqual(t, testCase.expectedContainersAfter, actual)
			}
		})
	}
}
