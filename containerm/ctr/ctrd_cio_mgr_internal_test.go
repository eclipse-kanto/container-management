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

package ctr

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

// Test newFIFOSet
func TestNewFIFOSet(t *testing.T) {
	const (
		testRoot      = "../pkg/testutil/execpath/container-management-cio-test-root"
		testProcessID = "test-ctr-process-id"
		regexBase     = "^" + testRoot + "/[0-9]+/" + testProcessID
	)
	var (
		stdinRegex  = regexp.MustCompile(regexBase + "-stdin$")
		stdoutRegex = regexp.MustCompile(regexBase + "-stdout$")
		stderrRegex = regexp.MustCompile(regexBase + "-stderr$")
	)
	tests := map[string]struct {
		rootDir       string
		withStdin     bool
		withTerminal  bool
		expectedError bool
	}{
		"test_with_stdin_no_terminal": {
			rootDir:   testRoot,
			withStdin: true,
		},
		"test_with_stdin_with_terminal": {
			rootDir:      testRoot,
			withStdin:    true,
			withTerminal: true,
		},
		"test_no_stdin_no_terminal": {
			rootDir: testRoot,
		},
		"test_create_err": {
			rootDir:       "\000",
			expectedError: true,
		},
	}

	// run tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			defer func() {
				_ = os.RemoveAll(testCase.rootDir)
			}()
			ioMgr := &cioMgr{
				fifoRootDir: testCase.rootDir,
			}
			cio, err := ioMgr.newFIFOSet(testProcessID, testCase.withStdin, testCase.withTerminal)
			if testCase.expectedError {
				testutil.AssertNotNil(t, err)
				testutil.AssertNil(t, cio)
			} else {
				testutil.AssertNil(t, err)
				testutil.AssertNotNil(t, cio)
				defer func() {
					testutil.AssertNil(t, cio.Close())
				}()
				testutil.AssertNotNil(t, cio.Config)
				testutil.AssertEqual(t, testCase.withTerminal, cio.Terminal)
				testutil.AssertTrue(t, stdoutRegex.MatchString(cio.Stdout))

				// validate temp FIFO dir is created
				fi, err := os.Stat(strings.Replace(cio.Stdout, testProcessID+"-stdout", "", 1))
				testutil.AssertNil(t, err)
				testutil.AssertTrue(t, fi.IsDir())

				if testCase.withStdin {
					testutil.AssertTrue(t, stdinRegex.MatchString(cio.Stdin))
				}
				if !testCase.withTerminal {
					testutil.AssertTrue(t, stderrRegex.MatchString(cio.Stderr))
				}
			}
		})
	}
}
