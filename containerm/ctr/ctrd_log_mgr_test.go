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

package ctr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testPath = "../pkg/testutil/ctr/"
	testID   = "test_container_id"
)

var (
	testDir = filepath.Join(testPath, testID)
)

func TestNewConainerLogsManager(t *testing.T) {
	containerLogsManager := newContainerLogsManager(testPath)
	testutil.AssertEqual(t, testPath, containerLogsManager.(*ctrLogsMgr).containerLogsDirRoot)
}

func TestGetLogDriver(t *testing.T) {
	createTestDir(t)
	defer deleteTestDir(t)

	testCases := map[string]struct {
		container      *types.Container
		expectedDriver bool
		expectedError  error
	}{
		"test_nil_config": {
			container: &types.Container{
				HostConfig: &types.HostConfig{},
			},
			expectedDriver: false,
			expectedError:  nil,
		},
		"test_config_type_log_none": {
			container: &types.Container{
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							Type: types.LogConfigDriverNone,
						},
					},
				},
			},
			expectedDriver: false,
			expectedError:  nil,
		},
		"test_error_init_log_dir": {
			container: &types.Container{
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "some_not_absolute_dir",
						},
					},
				},
			},
			expectedDriver: false,
			expectedError:  errors.New("root dir for container log, some_not_absolute_dir should be an absolute path"),
		},
		"test_error_in_json_file": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "",
							Type:    types.LogConfigDriverJSONFile,
							MaxSize: "invalid_size",
						},
					},
				},
			},
			expectedDriver: false,
			expectedError:  errors.New("invalid size provided invalid_size"),
		},
		"test_normal_json_file": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "",
							Type:    types.LogConfigDriverJSONFile,
						},
					},
				},
			},
			expectedDriver: true,
			expectedError:  nil,
		},
		"test_not_supported_file_format": {
			container: &types.Container{
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "",
							Type:    types.LogDriver("not_json"),
						},
					},
				},
			},
			expectedDriver: false,
			expectedError:  nil,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			containerLogsManager := newContainerLogsManager(testPath)
			driver, err := containerLogsManager.GetLogDriver(testCase.container)

			testutil.AssertError(t, testCase.expectedError, err)
			if testCase.expectedDriver {
				testutil.AssertNotNil(t, driver)
			} else {
				testutil.AssertNil(t, driver)
			}
		})
	}
}

func TestPrepareLogDriverConfig(t *testing.T) {
	testCases := map[string]struct {
		driverConfig       *types.LogDriverConfiguration
		expectedOptionSize int
		expectedError      error
	}{
		"test_max_file": {
			driverConfig: &types.LogDriverConfiguration{
				MaxFiles: 2,
				RootDir:  "",
				Type:     types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 1,
			expectedError:      nil,
		},
		"test_zero_max_file": {
			driverConfig: &types.LogDriverConfiguration{
				MaxFiles: 0,
				RootDir:  "",
				Type:     types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 0,
			expectedError:      nil,
		},
		"test_max_size_missing": {
			driverConfig: &types.LogDriverConfiguration{
				RootDir: "",
				Type:    types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 0,
			expectedError:      nil,
		},
		"test_max_size_incorrect": {
			driverConfig: &types.LogDriverConfiguration{
				MaxSize: "incorrect_size",
				RootDir: "",
				Type:    types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 0,
			expectedError:      errors.New("invalid size provided incorrect_size"),
		},
		"test_max_size_correct": {
			driverConfig: &types.LogDriverConfiguration{
				MaxSize: "1 K",
				RootDir: "",
				Type:    types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 1,
			expectedError:      nil,
		},
		"test_both_max_size_max_file": {
			driverConfig: &types.LogDriverConfiguration{
				MaxSize:  "4 m",
				MaxFiles: 4,
				RootDir:  "",
				Type:     types.LogConfigDriverJSONFile,
			},
			expectedOptionSize: 2,
			expectedError:      nil,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			containerLogsManager := &ctrLogsMgr{containerLogsDirRoot: testPath}
			option, err := containerLogsManager.prepareLogDriverConfig(testCase.driverConfig)

			testutil.AssertError(t, testCase.expectedError, err)
			testutil.AssertEqual(t, testCase.expectedOptionSize, len(option))
		})
	}
}

func TestInitContainerLogsRootDir(t *testing.T) {
	testPathAbsolute, _ := filepath.Abs(testPath)
	testDirAbsolute, _ := filepath.Abs(testDir)

	createTestDir(t)
	defer deleteTestDir(t)

	testCases := map[string]struct {
		container     *types.Container
		expectedName  string
		expectedError error
	}{
		"test_empty_root_path": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "",
						},
					},
				},
			},
			expectedName:  testDir,
			expectedError: nil,
		},
		"test_non_absolute_path_dir": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "some_non_absolute_dir",
						},
					},
				},
			},
			expectedName:  "",
			expectedError: errors.New("root dir for container log, some_non_absolute_dir should be an absolute path"),
		},
		"test_specific_path_error": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: "/\000",
						},
					},
				},
			},
			expectedName:  "",
			expectedError: errors.New("failed to create root log dir /\000/test_container_id: mkdir /\000: invalid argument"),
		},
		"test_specific_path_success": {
			container: &types.Container{
				ID: testID,
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						DriverConfig: &types.LogDriverConfiguration{
							RootDir: testPathAbsolute,
						},
					},
				},
			},
			expectedName:  testDirAbsolute,
			expectedError: nil,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			containerLogsManager := &ctrLogsMgr{containerLogsDirRoot: testPath}
			rootDirName, err := containerLogsManager.initContainerLogsRootDir(testCase.container)

			testutil.AssertError(t, testCase.expectedError, err)
			testutil.AssertEqual(t, testCase.expectedName, rootDirName)
		})
	}
}

func createTestDir(t *testing.T) {
	if os.MkdirAll(testDir, os.ModePerm) != nil {
		t.Fatalf("The test couldn't create the testResource %s!", testPath)
	}
}

func deleteTestDir(t *testing.T) {
	if os.RemoveAll(testPath) != nil {
		t.Logf("The test couldn't detele the testResource %s!", testPath)
	}
}
