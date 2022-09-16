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
	"io"
	"testing"
	"time"

	"github.com/containerd/containerd/cio"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocksCtrdCio "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	mocksio "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/io"
	mocksLogger "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/logger"
	mocksStreams "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/streams"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	"github.com/golang/mock/gomock"
)

// Test newIO
func TestNewIO(t *testing.T) {
	const testCIOId = "test-cio-id"
	tests := map[string]struct {
		withStdin bool
	}{
		"test_with_stdin": {
			withStdin: true,
		},
		"test_no_stdin": {
			withStdin: false,
		},
	}

	// run tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			io := newIO(testCIOId, testCase.withStdin)
			testutil.AssertNotNil(t, io)
			defer func() {
				_ = io.Close()
			}()
			testIo := io.(*containerIO)
			testutil.AssertEqual(t, testCIOId, testIo.id)
			testutil.AssertNotNil(t, testIo.stream)
			testutil.AssertEqual(t, testCase.withStdin, testIo.useStdin)
			testutil.AssertNotNil(t, testIo.stream.Stdout())
			testutil.AssertNotNil(t, testIo.stream.Stderr())
			testutil.AssertNotNil(t, testIo.stream.StdinPipe())

			if testCase.withStdin {
				testutil.AssertNotNil(t, testIo.stream.Stdin())
			} else {
				testutil.AssertNil(t, testIo.stream.Stdin())
			}
		})
	}
}

// Test SetLogDriver
func TestLogDriver(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockLogDriver := mocksLogger.NewMockLogDriver(mockCtrl)
	expIO := newIO("testID", false)

	// test set
	expIO.SetLogDriver(mockLogDriver)
	testutil.AssertEqual(t, mockLogDriver, expIO.(*containerIO).logDriver)

}

// Test SetMaxBufferSize
func TestMaxBufferSize(t *testing.T) {
	var buffSize int64 = 10
	expIO := newIO("testID", false)

	// test set
	expIO.SetMaxBufferSize(buffSize)
	testutil.AssertEqual(t, buffSize, expIO.(*containerIO).maxBufferSize)
}

// Test SetNonBlock
func TestNonBlock(t *testing.T) {
	expIO := newIO("testID", false)

	// assert default preconditions for a change to take place
	testutil.AssertFalse(t, expIO.(*containerIO).nonBlock)
	// test set
	expIO.SetNonBlock(true)
	testutil.AssertTrue(t, expIO.(*containerIO).nonBlock)
}

// Test Stream
func TestStream(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockStream := mocksStreams.NewMockStream(mockCtrl)
	expIO := &containerIO{
		stream: mockStream,
	}
	testutil.AssertEqual(t, mockStream, expIO.Stream())
}

// Test wait
func TestWait(t *testing.T) {
	type mockExecTestWait func(stream *mocksStreams.MockStream)
	tests := map[string]struct {
		timeout  time.Duration
		mockExec mockExecTestWait
	}{
		"test_io_wait_no_timeout": {
			timeout: 500 * time.Millisecond,
			mockExec: func(stream *mocksStreams.MockStream) {
				stream.EXPECT().Wait().Times(1)
			},
		},
		"test_io_wait_timeout": {
			timeout: 500 * time.Millisecond,
			mockExec: func(stream *mocksStreams.MockStream) {
				stream.EXPECT().Wait().Times(1).Do(func() {
					time.Sleep(600 * time.Millisecond)
				})
			},
		},
	}

	// run tests
	for testName, testCase := range tests {
		backWaitTimeout := streamCloseTimeout

		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			mockCtrl := gomock.NewController(t)
			mockStream := mocksStreams.NewMockStream(mockCtrl)

			defer func() {
				streamCloseTimeout = backWaitTimeout
				mockCtrl.Finish()
			}()

			// mock execution
			testCase.mockExec(mockStream)
			streamCloseTimeout = testCase.timeout
			testIO := &containerIO{
				stream: mockStream,
			}
			testIO.Wait()
		})
	}
}

// Test IO.Close()
func TestClose(t *testing.T) {
	type mockExecTestClose func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) *errorUtil.CompoundError

	tests := map[string]struct {
		logCopyTimeout time.Duration
		mockExec       mockExecTestClose
	}{
		"test_io_close_no_err": {
			logCopyTimeout: logcopierCloseTimeout,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) *errorUtil.CompoundError {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(nil)
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(nil)
				return nil
			},
		},
		"test_io_close_no_err_handler_timeout": {
			logCopyTimeout: 1 * time.Millisecond,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) *errorUtil.CompoundError {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(nil)
				logHandler.EXPECT().Wait().Times(1).Do(func() {
					time.Sleep(2 * time.Millisecond)
				})
				logDriver.EXPECT().Close().Times(1).Return(nil)
				return nil
			},
		},
		"test_io_close_stream_close_err": {
			logCopyTimeout: logcopierCloseTimeout,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) *errorUtil.CompoundError {
				cmpdError := new(errorUtil.CompoundError)

				stream.EXPECT().Wait().Times(1)
				closeErr := log.NewError("test close stream err")
				cmpdError.Append(closeErr)
				stream.EXPECT().Close().Times(1).Return(closeErr)
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(nil)
				return cmpdError
			},
		},
		"test_io_close_stream_and_driver_close_err": {
			logCopyTimeout: logcopierCloseTimeout,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) *errorUtil.CompoundError {
				cmpdError := new(errorUtil.CompoundError)

				stream.EXPECT().Wait().Times(1)
				closeErr := log.NewError("test close stream err")
				cmpdError.Append(closeErr)
				stream.EXPECT().Close().Times(1).Return(closeErr)
				logHandler.EXPECT().Wait().Times(1)

				driverCloseErr := log.NewError("test close driver err")
				cmpdError.Append(driverCloseErr)
				logDriver.EXPECT().Close().Times(1).Return(driverCloseErr)
				return cmpdError
			},
		},
	}

	// run tests
	for testName, testCase := range tests {
		backWaitTimeout := logcopierCloseTimeout

		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			mockCtrl := gomock.NewController(t)
			mockStream := mocksStreams.NewMockStream(mockCtrl)
			mockLogDriver := mocksLogger.NewMockLogDriver(mockCtrl)
			mockLogHandler := mocksLogger.NewMockLogHandler(mockCtrl)

			defer func() {
				logcopierCloseTimeout = backWaitTimeout
				mockCtrl.Finish()
			}()

			logcopierCloseTimeout = testCase.logCopyTimeout

			cmpdErr := testCase.mockExec(mockStream, mockLogDriver, mockLogHandler)
			testIO := &containerIO{
				stream:     mockStream,
				logDriver:  mockLogDriver,
				logHandler: mockLogHandler,
			}
			err := testIO.Close()

			if cmpdErr != nil {
				testutil.AssertNotNil(t, err)
				testutil.AssertEqual(t, cmpdErr.Size(), err.(*errorUtil.CompoundError).Size())
			} else {
				testutil.AssertNil(t, err)
			}
		})
	}
}

// Test IO.Reset()
func TestReset(t *testing.T) {
	type mockExecTestReset func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler)

	tests := map[string]struct {
		useStdin bool
		mockExec mockExecTestReset
	}{
		"test_io_reset_no_err": {
			useStdin: false,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(nil)
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(nil)
				stream.EXPECT().NewDiscardStdinInput().Times(1)
			},
		},
		"test_io_reset_close_err": {
			useStdin: false,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(log.NewError("test close stream error"))
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(nil)
				stream.EXPECT().NewDiscardStdinInput().Times(1)
			},
		},
		"test_io_reset_no_err_with_stdin": {
			useStdin: true,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(nil)
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(nil)
				stream.EXPECT().NewStdinInput().Times(1)
			},
		},
		"test_io_reset_driver_close_err_with_stdin": {
			useStdin: true,
			mockExec: func(stream *mocksStreams.MockStream, logDriver *mocksLogger.MockLogDriver, logHandler *mocksLogger.MockLogHandler) {
				stream.EXPECT().Wait().Times(1)
				stream.EXPECT().Close().Times(1).Return(nil)
				logHandler.EXPECT().Wait().Times(1)
				logDriver.EXPECT().Close().Times(1).Return(log.NewError("test close driver error"))
				stream.EXPECT().NewStdinInput().Times(1)
			},
		},
	}

	// run tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			mockCtrl := gomock.NewController(t)
			mockStream := mocksStreams.NewMockStream(mockCtrl)
			mockLogDriver := mocksLogger.NewMockLogDriver(mockCtrl)
			mockLogHandler := mocksLogger.NewMockLogHandler(mockCtrl)

			defer mockCtrl.Finish()

			testCase.mockExec(mockStream, mockLogDriver, mockLogHandler)
			testIO := &containerIO{
				stream:     mockStream,
				logDriver:  mockLogDriver,
				logHandler: mockLogHandler,
				useStdin:   testCase.useStdin,
			}

			testIO.Reset()
			testutil.AssertNil(t, testIO.logDriver)
			testutil.AssertNil(t, testIO.logHandler)
		})
	}
}

// Test InitContainerIO
func TestInitContainerIO(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockStream := mocksStreams.NewMockStream(mockCtrl)
	mockDIOStdin := mocksio.NewMockWriteCloser(mockCtrl)
	mockDIOStdout := mocksio.NewMockReadCloser(mockCtrl)
	mockDIOStderr := mocksio.NewMockReadCloser(mockCtrl)

	ctrIO := &containerIO{
		stream: mockStream,
	}
	defer func() {
		_ = ctrIO.Close()
		mockCtrl.Finish()
	}()

	dIO := &cio.DirectIO{}
	dIO.Stdin = mockDIOStdin
	dIO.Stdout = mockDIOStdout
	dIO.Stderr = mockDIOStderr

	mockStream.EXPECT().CopyPipes(gomock.Any()).Do(func(pipes streams.Pipes) {
		testutil.AssertEqual(t, pipes.Stdin, dIO.Stdin)
		testutil.AssertEqual(t, pipes.Stdout, dIO.Stdout)
		testutil.AssertEqual(t, pipes.Stderr, dIO.Stderr)
	}).Times(1)
	// mock close
	mockStream.EXPECT().Wait().Times(1)
	mockStream.EXPECT().Close().Times(1).Return(nil)

	actualCIO, _ := ctrIO.InitContainerIO(dIO)
	wrappedCIO, ok := actualCIO.(*wrapcio)
	testutil.AssertTrue(t, ok)
	testutil.AssertEqual(t, dIO, wrappedCIO.IO)
	testutil.AssertEqual(t, ctrIO, wrappedCIO.ctrio)
}

// Test IO.startLogging()
func TestIOStartLogging(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockStream := mocksStreams.NewMockStream(mockCtrl)
	mockLogDriver := mocksLogger.NewMockLogDriver(mockCtrl)
	mockStd := mocksio.NewMockReadCloser(mockCtrl)

	ctrIO := &containerIO{
		stream:    mockStream,
		logDriver: mockLogDriver,
		nonBlock:  true,
	}
	defer mockCtrl.Finish()

	mockStream.EXPECT().NewStdoutPipe().Times(1).Return(mockStd)
	mockStream.EXPECT().NewStderrPipe().Times(1).Return(mockStd)
	mockStd.EXPECT().Read(gomock.Any()).Times(2).Return(0, io.EOF)
	mockLogDriver.EXPECT().Type().Return(logger.LogDriverType("test-log-driver-type")).Times(2)
	// mock close
	mockStream.EXPECT().Wait().Times(1)
	mockStream.EXPECT().Close().Times(1).Return(nil)
	mockLogDriver.EXPECT().Close().Times(1).Return(nil)

	defer func() {
		_ = ctrIO.Close()
	}()
	_ = ctrIO.startLogging()
}

// Test InitContainerIO
func TestWrappedIO(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockStream := mocksStreams.NewMockStream(mockCtrl)
	mockCIO := mocksCtrdCio.NewMockIO(mockCtrl)

	ctrIO := &containerIO{
		stream: mockStream,
	}
	testWrapCIO := &wrapcio{
		IO:    mockCIO,
		ctrio: ctrIO,
	}

	defer mockCtrl.Finish()

	// mock wait
	mockStream.EXPECT().Wait().Times(1)
	mockCIO.EXPECT().Wait()
	// mock close
	mockStream.EXPECT().Wait().Times(1)
	mockStream.EXPECT().Close().Times(1).Return(nil)
	mockCIO.EXPECT().Close().Times(1).Return(nil)

	defer func() {
		_ = testWrapCIO.Close()
	}()
	testWrapCIO.Wait()

}
