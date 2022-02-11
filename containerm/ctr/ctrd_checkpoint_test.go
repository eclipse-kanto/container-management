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
	"path/filepath"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testContainerName    = "test-container-name"
	testCheckDir         = "test-check-dir"
	testCheckpointID     = "test-checkpoint-id"
	testCtrCheckpointDir = "test-ctr-checkpoint-dir"
	testUtilCtrdDir      = "../pkg/testutil/test-dir"
)

var (
	testContainer = &types.Container{
		ID:   testCtrID1,
		Name: testContainerName,
	}
)

type testGetCheckpointDirArgs struct {
	checkDir            string
	checkpointID        string
	ctrName             string
	ctrID               string
	ctrCheckpointDir    string
	create              bool
	dirToCreate         string
	dirToRemove         string
	fileToCreate        string
	expCheckpointAbsDir string
	expErr              error
}

func TestGetCheckpointDir(t *testing.T) {
	tests := map[string]struct {
		args testGetCheckpointDirArgs
	}{
		"test_get_container_with_create": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              true,
				dirToCreate:         "",
				dirToRemove:         "",
				fileToCreate:        "",
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              nil,
			},
		},
		"test_get_container": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              false,
				dirToCreate:         "",
				dirToRemove:         "",
				fileToCreate:        "",
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              log.NewErrorf("checkpoint %s does not exist for container %s", testCheckpointID, testContainerName),
			},
		},
		"test_get_container_empty": {
			args: testGetCheckpointDirArgs{
				checkDir:            "",
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              true,
				dirToCreate:         "",
				dirToRemove:         testCtrCheckpointDir,
				fileToCreate:        "",
				expCheckpointAbsDir: filepath.Join(testCtrCheckpointDir, testCheckpointID),
				expErr:              nil,
			},
		},
		"test_get_container_existing_with_create": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              true,
				dirToCreate:         filepath.Join(testUtilCtrdDir, testCheckpointID),
				dirToRemove:         "",
				fileToCreate:        "",
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              log.NewErrorf("checkpoint with name %s already exists for container %s", testCheckpointID, testContainerName),
			},
		},
		"test_get_container_existing_without_create": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              false,
				dirToCreate:         filepath.Join(testUtilCtrdDir, testCheckpointID),
				dirToRemove:         "",
				fileToCreate:        "",
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              nil,
			},
		},
		"test_get_container_non_dir_with_create": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              true,
				dirToCreate:         testUtilCtrdDir,
				dirToRemove:         "",
				fileToCreate:        filepath.Join(testUtilCtrdDir, testCheckpointID),
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              log.NewErrorf("%s exists and is not a directory", filepath.Join(testUtilCtrdDir, testCheckpointID)),
			},
		},
		"test_get_container_non_dir_without_create": {
			args: testGetCheckpointDirArgs{
				checkDir:            testUtilCtrdDir,
				checkpointID:        testCheckpointID,
				ctrName:             testContainerName,
				ctrID:               testCtrID1,
				ctrCheckpointDir:    testCtrCheckpointDir,
				create:              false,
				dirToCreate:         testUtilCtrdDir,
				dirToRemove:         "",
				fileToCreate:        filepath.Join(testUtilCtrdDir, testCheckpointID),
				expCheckpointAbsDir: filepath.Join(testUtilCtrdDir, testCheckpointID),
				expErr:              log.NewErrorf("%s exists and is not a directory", filepath.Join(testUtilCtrdDir, testCheckpointID)),
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			defer os.RemoveAll(testUtilCtrdDir)
			defer os.RemoveAll(testCase.args.dirToRemove)

			if testCase.args.dirToCreate != "" {
				os.MkdirAll(testCase.args.dirToCreate, 0777)
			}

			if testCase.args.fileToCreate != "" {
				os.Create(testCase.args.fileToCreate)
			}

			result, resultErr := getCheckpointDir(testCase.args.checkDir, testCase.args.checkpointID, testCase.args.ctrName, testCase.args.ctrID, testCtrCheckpointDir, testCase.args.create)
			testutil.AssertEqual(t, testCase.args.expCheckpointAbsDir, result)
			testutil.AssertError(t, testCase.args.expErr, resultErr)
		})
	}
}
