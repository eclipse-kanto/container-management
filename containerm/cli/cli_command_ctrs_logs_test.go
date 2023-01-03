// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	logsCmdFlagName = "name"
	logsCmdFlagTail = "tail"

	// test input constants
	logsContainerID   = "test-ctr"
	logsContainerName = "test-ctr-name"
	logsTailStr       = "42"
)

var (
	// command args ---------------
	logsCmdArgs = []string{logsContainerID}
)

// Tests ------------------------------
func TestLogsCmdInit(t *testing.T) {
	logsCommandTest := &logsCommandTest{}
	logsCommandTest.init()

	execTestInit(t, logsCommandTest)
}

func TestLogsCmdFlags(t *testing.T) {
	logsCommandTest := &logsCommandTest{}
	logsCommandTest.init()

	expectedCfg := logsConfig{
		name: logsContainerName,
		tail: 42,
	}

	flagsToApply := map[string]string{
		logsCmdFlagName: expectedCfg.name,
		logsCmdFlagTail: fmt.Sprintf("%d", 42),
	}

	execTestSetupFlags(t, logsCommandTest, flagsToApply, expectedCfg)
}

func TestLogsCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	logsCommandTest := &logsCommandTest{}
	logsCommandTest.initWithCtrl(controller)

	execTestsRun(t, logsCommandTest)
}

type logsCommandTest struct {
	cliCommandTestBase
	logsCmd *logsCmd
}

func (l *logsCommandTest) commandConfig() interface{} {
	return l.logsCmd.config
}

func (l *logsCommandTest) commandConfigDefault() interface{} {
	return logsConfig{
		name: "",
		tail: 100,
	}
}

func (l *logsCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &logsCmd{}
	l.logsCmd, l.baseCmd = cmd, cmd

	l.logsCmd.init(l.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, l.logsCmd.cmd)
}

func (l *logsCommandTest) runCommand(args []string) error {
	return l.logsCmd.run(args)
}

func (l *logsCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_get_logs": {
			flags: map[string]string{
				logsCmdFlagName: logsContainerName,
				logsCmdFlagTail: logsTailStr,
			},
			mockExecution: l.mockExecutionGetLogs,
		},
		"test_logs_by_name_more_than_one_ctrs": {
			flags: map[string]string{
				logsCmdFlagName: logsContainerName,
				logsCmdFlagTail: logsTailStr,
			},
			mockExecution: l.mockExecLogsByNameMoreThanOneCtrs,
		},
		"test_logs_no_id_or_name_provided": {
			mockExecution: l.mockExecLogsNoIDorName,
		},
		"test_logs_by_id_default": {
			args:          logsCmdArgs,
			mockExecution: l.mockExecLogsNoErrors,
		},
		"test_logs_by_id_ctr_nil": {
			args:          logsCmdArgs,
			mockExecution: l.mockExecLogsByIDNilCtr,
		},
		"test_logs_by_id_err": {
			args:          logsCmdArgs,
			mockExecution: l.mockExecLogsErrors,
		},
		"test_logs_by_name_err": {
			flags: map[string]string{
				logsCmdFlagName: logsContainerName,
			},
			mockExecution: l.mockExecLogsByNameErr,
		},
		"test_logs_by_name_nil_ctr": {
			flags: map[string]string{
				logsCmdFlagName: logsContainerName,
			},
			mockExecution: l.mockExecLogsByNameNilCtr,
		},
		"test_logs_by_name_zero_ctrs": {
			flags: map[string]string{
				logsCmdFlagName: logsContainerName,
			},
			mockExecution: l.mockExecLogsByNameZeroCtrs,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (l *logsCommandTest) mockExecLogsIDAndName(args []string) error {
	l.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}

func (l *logsCommandTest) mockExecLogsNoIDorName(args []string) error {
	l.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	return log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
}

func (l *logsCommandTest) mockExecLogsNoErrors(args []string) error {
	ctr := &types.Container{ID: args[0]}
	l.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	l.mockClient.EXPECT().Logs(context.Background(), args[0], int32(100)).Times(1).Return(nil)
	return nil
}

func (l *logsCommandTest) mockExecLogsByIDNilCtr(args []string) error {
	l.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, nil)
	return log.NewErrorf("The requested container with ID = %s was not found.", args[0])
}

func (l *logsCommandTest) mockExecLogsErrors(args []string) error {
	err := log.NewError("error getting container")
	l.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, err)
	return err
}

func (l *logsCommandTest) mockExecLogsByNameErr(args []string) error {
	err := log.NewError("error listing containers")
	l.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(logsContainerName))).Times(1).Return(nil, err)
	return err
}

func (l *logsCommandTest) mockExecLogsByNameNilCtr(args []string) error {
	l.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(logsContainerName))).Times(1).Return(nil, nil)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", logsContainerName)
}

func (l *logsCommandTest) mockExecLogsByNameZeroCtrs(args []string) error {
	l.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(logsContainerName))).Times(1).Return([]*types.Container{}, nil)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", logsContainerName)
}

func (l *logsCommandTest) mockExecutionGetLogs(args []string) error {
	res := []*types.Container{{
		ID:   logsContainerID,
		Name: logsContainerName,
	}}
	l.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(logsContainerName))).Times(1).Return(res, nil)
	l.mockClient.EXPECT().Logs(context.Background(), logsContainerID, int32(42)).Times(1).Return(nil)
	return nil
}

func (l *logsCommandTest) mockExecLogsByNameMoreThanOneCtrs(args []string) error {
	res := []*types.Container{{Name: logsContainerName}, {Name: logsContainerName}}
	l.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(logsContainerName))).Times(1).Return(res, nil)
	return log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", logsContainerName)
}
