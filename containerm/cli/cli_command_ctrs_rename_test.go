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

package main

import (
	"context"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	renameCmdFlagName = "name"

	// test input constants
	renameContainerID   = "test-ctr"
	renameContainerName = "test-ctr-name"

	newRenameContainerName     = "test-ctr-name-renamed"
	invalidRenameContainerName = "@test-ctr-name-renamed"
)

var (
	// command args ---------------
	renameCmdArgsWithID = []string{renameContainerID, newRenameContainerName}

	renameCtr = &types.Container{
		ID:   renameContainerID,
		Name: renameContainerName,
	}
)

// Tests ------------------------------
func TestRenameCmdInit(t *testing.T) {
	renameCliTest := &renameCommandTest{}
	renameCliTest.init()

	execTestInit(t, renameCliTest)
}

func TestRenameCmdFlags(t *testing.T) {
	renameCliTest := &renameCommandTest{}
	renameCliTest.init()

	expectedCfg := renameConfig{
		name: renameContainerName,
	}

	flagsToApply := map[string]string{
		renameCmdFlagName: expectedCfg.name,
	}

	execTestSetupFlags(t, renameCliTest, flagsToApply, expectedCfg)
}

func TestRenameCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	renameCliTest := &renameCommandTest{}
	renameCliTest.initWithCtrl(controller)

	execTestsRun(t, renameCliTest)
}

// EOF Tests --------------------------

type renameCommandTest struct {
	cliCommandTestBase
	renameCmd *renameCtrCmd
}

func (renameTc *renameCommandTest) commandConfig() interface{} {
	return renameTc.renameCmd.config
}

func (renameTc *renameCommandTest) commandConfigDefault() interface{} {
	return renameConfig{}
}

func (renameTc *renameCommandTest) prepareCommand(flagsCfg map[string]string) {
	// setup command to test
	cmd := &renameCtrCmd{}
	renameTc.renameCmd, renameTc.baseCmd = cmd, cmd

	renameTc.renameCmd.init(renameTc.mockRootCommand)
	// setup command flags
	setCmdFlags(flagsCfg, renameTc.renameCmd.cmd)
}

func (renameTc *renameCommandTest) runCommand(args []string) error {
	return renameTc.renameCmd.run(args)
}

func (renameTc *renameCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		// Test name
		"test_rename_error_id_and_name_provided": {
			args: renameCmdArgsWithID,
			flags: map[string]string{
				renameCmdFlagName: renameContainerName,
			},
			mockExecution: renameTc.mockExecRenameErrIDAndName,
		},
		"test_rename_by_id": {
			args:          renameCmdArgsWithID,
			mockExecution: renameTc.mockExecRenameByID,
		},
		"test_rename_by_name": {
			args: []string{newRenameContainerName},
			flags: map[string]string{
				renameCmdFlagName: renameContainerName,
			},
			mockExecution: renameTc.mockExecRenameByName,
		},
		"test_rename_error_same_name": {
			args:          []string{renameContainerID, renameContainerName},
			mockExecution: renameTc.mockExecRenameErrSameName,
		},
		"test_rename_error_invalid_name": {
			args:          []string{renameContainerID, invalidRenameContainerName},
			mockExecution: renameTc.mockExecRenameErrInvalidName,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (renameTc *renameCommandTest) mockExecRenameErrIDAndName(args []string) error {
	renameTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}

func (renameTc *renameCommandTest) mockExecRenameByID(args []string) error {
	renameTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(renameCtr, nil)
	renameTc.mockClient.EXPECT().Rename(context.Background(), args[0], args[1]).Times(1)
	return nil
}

func (renameTc *renameCommandTest) mockExecRenameByName(args []string) error {
	res := []*types.Container{renameCtr}
	renameTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(args[0]))).Times(1).Return(res, nil)
	renameTc.mockClient.EXPECT().Rename(context.Background(), renameCtr.ID, args[0]).Times(1)
	return nil
}

func (renameTc *renameCommandTest) mockExecRenameErrSameName(args []string) error {
	renameTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(renameCtr, nil)
	renameTc.mockClient.EXPECT().Rename(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("the new name = %s shouldn't be the same", testCtr.Name)
}

func (renameTc *renameCommandTest) mockExecRenameErrInvalidName(args []string) error {
	renameTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	renameTc.mockClient.EXPECT().Rename(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("invalid container name format : %s", args[1])
}
