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
	"os"
	"testing"

	"github.com/containerd/containerd/cio"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	loggerMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/logger"
	"github.com/golang/mock/gomock"
)

const (
	testContainerIDWithIO    = "testContainerIDWithIO"
	testContainerIDWithoutIO = "testContainerIDWithoutIO"
)

func TestNewContainerIOManager(t *testing.T) {
	ioCache := newCache()

	want := &cioMgr{
		fifoRootDir: "",
		ioCache:     ioCache,
	}

	got := newContainerIOManager("", ioCache)

	testutil.AssertEqual(t, want, got)
}

func TestExistsIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		exists      bool
	}{
		"test_with_io": {
			containerID: testContainerIDWithIO,
			exists:      true,
		},
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockIO := NewMockIO(mockCtrl)

			ioCache := newCache()
			ioCache.Put(testContainerIDWithIO, mockIO)
			testMgr := &cioMgr{
				ioCache: ioCache,
			}
			testutil.AssertEqual(t, testCase.exists, testMgr.ExistsIO(testCase.containerID))
		})
	}
}

func TestGetIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		exists      bool
	}{
		"test_with_io": {
			containerID: testContainerIDWithIO,
			exists:      true,
		},
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, _ := createTestMgr(mockCtrl)

			if testCase.exists {
				testutil.AssertNotNil(t, testMgr.GetIO(testCase.containerID))
			} else {
				testutil.AssertNil(t, testMgr.GetIO(testCase.containerID))
			}
		})
	}
}

func TestInitIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func() error
	}{
		"test_with_io": {
			containerID: testContainerIDWithIO,
			mockExec: func() error {
				return log.NewErrorf("failed to create containerIO")
			},
		},
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
			mockExec: func() error {
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, _ := createTestMgr(mockCtrl)

			expectedError := testCase.mockExec()

			_, actualError := testMgr.InitIO(testCase.containerID, true)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestConfigureIO(t *testing.T) {
	testCases := map[string]struct {
		containerID string
		logModeCfg  *types.LogModeConfiguration
		mapExec     func(*MockIO, *loggerMocks.MockLogDriver) error
	}{
		"test_log_mode_blocking": {
			containerID: testContainerIDWithIO,
			logModeCfg: &types.LogModeConfiguration{
				Mode: types.LogModeBlocking,
			},
			mapExec: func(mockIO *MockIO, mockLogDriver *loggerMocks.MockLogDriver) error {
				mockIO.EXPECT().SetLogDriver(mockLogDriver).Times(1)
				return nil
			},
		},
		"test_log_mode_non_blocking": {
			containerID: testContainerIDWithIO,
			logModeCfg: &types.LogModeConfiguration{
				Mode:          types.LogModeNonBlocking,
				MaxBufferSize: "1K",
			},
			mapExec: func(mockIO *MockIO, mockLogDriver *loggerMocks.MockLogDriver) error {
				mockIO.EXPECT().SetLogDriver(mockLogDriver).Times(1)
				mockIO.EXPECT().SetMaxBufferSize(int64(1024)).Times(1)
				mockIO.EXPECT().SetNonBlock(true).Times(1)
				return nil
			},
		},
		"test_invalid_max_buffer_size_error": {
			containerID: testContainerIDWithIO,
			logModeCfg: &types.LogModeConfiguration{
				Mode:          types.LogModeNonBlocking,
				MaxBufferSize: "invalid",
			},
			mapExec: func(_ *MockIO, _ *loggerMocks.MockLogDriver) error {
				return log.NewErrorf("invalid size provided invalid")
			},
		},
		"test_no_IO_err": {
			containerID: testContainerIDWithoutIO,
			mapExec: func(_ *MockIO, _ *loggerMocks.MockLogDriver) error {
				return log.NewErrorf("no IO resources allocated for id = %s", testContainerIDWithoutIO)
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)

			testMgr, mockIO := createTestMgr(mockCtrl)

			expectedErr := testData.mapExec(mockIO, mockLogDriver)

			actualErr := testMgr.ConfigureIO(testData.containerID, mockLogDriver, testData.logModeCfg)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestResetIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func(*MockIO)
	}{
		"test_with_io": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO) {
				mockIO.EXPECT().Reset().Times(1)
			},
		},
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
			mockExec: func(mockIO *MockIO) {
				mockIO.EXPECT().Reset().Times(0)
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, mockIO := createTestMgr(mockCtrl)

			testCase.mockExec(mockIO)

			testMgr.ResetIO(testCase.containerID)
		})
	}
}

func TestCloseIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func(*MockIO) error
	}{
		"test_with_io_without_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO) error {
				mockIO.EXPECT().Close().Times(1)
				return nil
			},
		},
		"test_with_io_with_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO) error {
				err := log.NewErrorf("test error")
				mockIO.EXPECT().Close().Return(err)
				return err
			},
		},
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
			mockExec: func(mockIO *MockIO) error {
				mockIO.EXPECT().Close().Times(0)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, mockIO := createTestMgr(mockCtrl)

			expectedError := testCase.mockExec(mockIO)

			actualError := testMgr.CloseIO(testCase.containerID)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestClearIO(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func(*MockIO) error
	}{
		"test_without_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO) error {
				mockIO.EXPECT().Close().Return(nil)
				return nil
			},
		},
		"test_with_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO) error {
				err := log.NewErrorf("test error")
				mockIO.EXPECT().Close().Return(err)
				return err
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, mockIO := createTestMgr(mockCtrl)

			expectedError := testCase.mockExec(mockIO)

			actualError := testMgr.ClearIO(testCase.containerID)
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestNewCioCreator(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func(*MockIO, *cioMgr) cio.Creator
	}{
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
			mockExec: func(_ *MockIO, _ *cioMgr) cio.Creator {
				return func(id string) (cio.IO, error) {
					return nil, log.NewErrorf("no IO resources allocated for id = %s", id)
				}
			},
		},
		"test_with_io_with_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO, _ *cioMgr) cio.Creator {
				err := log.NewErrorf("mkdir : no such file or directory")
				mockIO.EXPECT().UseStdin().Times(2)
				return func(id string) (cio.IO, error) {
					return nil, err
				}
			},
		},
		"test_with_io_without_error": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIO *MockIO, testMgr *cioMgr) cio.Creator {
				testMgr.fifoRootDir = "test.dir"
				mockIO.EXPECT().UseStdin().Times(2).Return(true)
				mockIO.EXPECT().InitContainerIO(gomock.Any()).Return(nil, nil)
				return func(id string) (cio.IO, error) {
					return nil, nil
				}
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, mockIO := createTestMgr(mockCtrl)

			_, expectError := testCase.mockExec(mockIO, testMgr)(testContainerIDWithoutIO)
			_, actualError := testMgr.NewCioCreator(true)(testCase.containerID)

			os.RemoveAll(testMgr.fifoRootDir)

			testutil.AssertError(t, expectError, actualError)
		})
	}
}

func TestNewCioAttach(t *testing.T) {
	tests := map[string]struct {
		containerID string
		mockExec    func(*MockIO) cio.Attach
	}{
		"test_without_io": {
			containerID: testContainerIDWithoutIO,
			mockExec: func(mockIo *MockIO) cio.Attach {
				return func(f *cio.FIFOSet) (cio.IO, error) {
					return nil, log.NewErrorf("no IO resources allocated for id = %s", testContainerIDWithoutIO)
				}
			},
		},
		"test_with_io": {
			containerID: testContainerIDWithIO,
			mockExec: func(mockIo *MockIO) cio.Attach {
				mockIo.EXPECT().InitContainerIO(gomock.Any())
				return func(f *cio.FIFOSet) (cio.IO, error) {
					io, _ := cio.NewCreator(cio.WithFIFODir(""))(testContainerIDWithIO)
					return io, nil
				}
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			testMgr, mockIO := createTestMgr(mockCtrl)

			_, expectError := testCase.mockExec(mockIO)(&cio.FIFOSet{})
			_, actualError := testMgr.NewCioAttach(testCase.containerID)(&cio.FIFOSet{})

			testutil.AssertError(t, expectError, actualError)
		})
	}
}

func createTestMgr(mockCtrl *gomock.Controller) (*cioMgr, *MockIO) {
	mockIO := NewMockIO(mockCtrl)

	ioCache := newCache()
	ioCache.Put(testContainerIDWithIO, mockIO)
	testMgr := &cioMgr{
		ioCache: ioCache,
	}

	return testMgr, mockIO
}
