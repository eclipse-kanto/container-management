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

package log

import (
	"bytes"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocksio "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/io"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"regexp"
	"testing"
)

func TestProcessFormat(t *testing.T) {
	nestedCaller := func(formatOrigin string) string {
		return func() string {
			return processFormat(formatOrigin)
		}()
	}

	const (
		expectedMsgPattern = `published event &{...}$`
	)

	testCases := map[string]struct {
		caller           func(string) string
		msg              string
		expectedMsgMatch bool
	}{
		"test_valid": {
			caller:           nestedCaller,
			msg:              "published event &{...}",
			expectedMsgMatch: true,
		},
		"test_invalid": {
			caller:           nestedCaller,
			msg:              "wrong event that has occured",
			expectedMsgMatch: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.caller(testCase.msg)
			matched, _ := regexp.MatchString(expectedMsgPattern, actual)
			testutil.AssertEqual(t, testCase.expectedMsgMatch, matched)
		})
	}
}

func TestProcessFormatWithError(t *testing.T) {
	nestedCaller := func(formatOrigin string, errMsg string) string {
		return func() string {
			return processFormatWithError(formatOrigin, NewError(errMsg))
		}()
	}

	const (
		expectedMsgPattern = `published event &{...}`
		expectedErrPattern = `connection lost`
	)

	testCases := map[string]struct {
		caller           func(string, string) string
		msg              string
		err              string
		expectedMsgMatch bool
		expectedErrMatch bool
	}{
		"test_valid": {
			caller:           nestedCaller,
			msg:              "published event &{...}",
			err:              "connection lost",
			expectedMsgMatch: true,
			expectedErrMatch: true,
		},
		"test_invalid_msg": {
			caller:           nestedCaller,
			msg:              "wrong event that has occured",
			err:              "connection lost",
			expectedMsgMatch: false,
			expectedErrMatch: true,
		},
		"test_invalid_err": {
			caller:           nestedCaller,
			msg:              "published event &{...}",
			err:              "wrong error",
			expectedMsgMatch: true,
			expectedErrMatch: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.caller(testCase.msg, testCase.err)
			matchedMsg, _ := regexp.MatchString(expectedMsgPattern, actual)
			matchedErr, _ := regexp.MatchString(expectedErrPattern, actual)
			testutil.AssertEqual(t, testCase.expectedMsgMatch, matchedMsg)
			testutil.AssertEqual(t, testCase.expectedErrMatch, matchedErr)
		})
	}
}

func TestPreparePrefix(t *testing.T) {
	nestedCaller := func() string {
		return func() string {
			return preparePrefix()
		}()
	}

	testCases := map[string]struct {
		caller        func() string
		pattern       string
		expectedMatch bool
	}{
		"test_valid": {
			caller:        nestedCaller,
			pattern:       `^\[container-management\]\[logger_util_test.go:[0-9]+\]\[pkg:github\]\[func:com/eclipse-kanto/container-management/containerm/log\] $`,
			expectedMatch: true,
		},
		"test_invalid": {
			caller:        nestedCaller,
			pattern:       `Definitely a wrong pattern`,
			expectedMatch: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.caller()
			matched, _ := regexp.MatchString(testCase.pattern, actual)
			testutil.AssertEqual(t, testCase.expectedMatch, matched)
		})
	}
}

func TestClear(t *testing.T) {
	testCases := map[string]struct {
		setUp                 func(mockCtrl *gomock.Controller) io.WriteCloser
		expectedErrMsgPattern string
	}{
		"test_valid": {
			setUp: func(mockCtrl *gomock.Controller) io.WriteCloser {
				mockWriterCloser := mocksio.NewMockWriteCloser(mockCtrl)
				mockWriterCloser.EXPECT().Close().Return(nil)
				return mockWriterCloser
			},
			expectedErrMsgPattern: `^$`,
		},
		"test_invalid": {
			setUp: func(mockCtrl *gomock.Controller) io.WriteCloser {
				logrus.SetLevel(logrus.DebugLevel)
				mockWriterCloser := mocksio.NewMockWriteCloser(mockCtrl)
				mockWriterCloser.EXPECT().Close().Return(NewError("test error"))
				return mockWriterCloser
			},
			expectedErrMsgPattern: `\btest error\b`,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			logFileWriteCloser = testCase.setUp(mockCtrl)
			out := bytes.NewBuffer([]byte{})
			logrus.SetOutput(out)
			defer logrus.SetOutput(os.Stdout)

			clear()

			matchedErr, _ := regexp.MatchString(testCase.expectedErrMsgPattern, out.String())
			testutil.AssertTrue(t, matchedErr)
		})
	}
}
