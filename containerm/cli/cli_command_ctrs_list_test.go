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
	listCmdFlagName = "name"

	// test input constants
	listContainerID = "test-ctr"
	listFlagName    = "test-ctr-name"
)

// Tests ------------------------------
func TestListCmdInit(t *testing.T) {
	listCliTest := &listCommandTest{}
	listCliTest.init()

	execTestInit(t, listCliTest)
}

func TestListCmdFlags(t *testing.T) {
	listCliTest := &listCommandTest{}
	listCliTest.init()

	expectedCfg := listConfig{
		name: listFlagName,
	}

	flagsToApply := map[string]string{
		listCmdFlagName: expectedCfg.name,
	}

	execTestSetupFlags(t, listCliTest, flagsToApply, expectedCfg)
}

func TestListCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	listCliTest := &listCommandTest{}
	listCliTest.initWithCtrl(controller)

	execTestsRun(t, listCliTest)
}

// EOF Tests --------------------------

type listCommandTest struct {
	cliCommandTestBase
	listCmd *listCmd
}

func (listTc *listCommandTest) commandConfig() interface{} {
	return listTc.listCmd.config
}

func (listTc *listCommandTest) commandConfigDefault() interface{} {
	return listConfig{
		name: "",
	}
}
func (listTc *listCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &listCmd{}
	listTc.listCmd, listTc.baseCmd = cmd, cmd

	listTc.listCmd.init(listTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, listTc.listCmd.cmd)
}

func (listTc *listCommandTest) runCommand(args []string) error {
	return listTc.listCmd.run(args)
}

func (listTc *listCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_list_default": {
			mockExecution: listTc.mockExecListNoErrors,
		},
		"test_list_no_ctrs": {
			mockExecution: listTc.mockExecListNoCtrs,
		},
		"test_list_err": {
			mockExecution: listTc.mockExecListErrors,
		},
		"test_list_by_name_default": {
			flags: map[string]string{
				listCmdFlagName: listFlagName,
			},
			mockExecution: listTc.mockExecListByNameNoErrors,
		},
		"test_list_by_name_no_ctrs": {
			flags: map[string]string{
				listCmdFlagName: listFlagName,
			},
			mockExecution: listTc.mockExecListByNameNoCtrs,
		},
		"test_list_by_name_err": {
			flags: map[string]string{
				listCmdFlagName: listFlagName,
			},
			mockExecution: listTc.mockExecListByNameErrors,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (listTc *listCommandTest) mockExecListNoErrors(args []string) error {
	// setup expected calls
	ctrs := []*types.Container{{
		ID:    listContainerID,
		Name:  listFlagName,
		State: &types.State{},
	}}
	listTc.mockClient.EXPECT().List(context.Background()).Times(1).Return(ctrs, nil)
	// no error expected
	return nil
}

func (listTc *listCommandTest) mockExecListNoCtrs(args []string) error {
	// setup expected calls
	listTc.mockClient.EXPECT().List(context.Background()).Times(1).Return(nil, nil)
	// no error expected
	return nil
}

func (listTc *listCommandTest) mockExecListErrors(args []string) error {
	// setup expected calls
	err := log.NewError("failed to get containers list")
	listTc.mockClient.EXPECT().List(context.Background()).Times(1).Return(nil, err)
	return err
}

func (listTc *listCommandTest) mockExecListByNameNoErrors(args []string) error {
	// setup expected calls
	ctrs := []*types.Container{{
		ID:    listContainerID,
		Name:  listFlagName,
		State: &types.State{},
	}}
	listTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(listFlagName))).Times(1).Return(ctrs, nil)
	// no error expected
	return nil
}

func (listTc *listCommandTest) mockExecListByNameNoCtrs(args []string) error {
	// setup expected calls
	listTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(listFlagName))).Times(1).Return(nil, nil)
	// no error expected
	return nil
}

func (listTc *listCommandTest) mockExecListByNameErrors(args []string) error {
	// setup expected calls
	err := log.NewError("failed to get containers list")
	listTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(listFlagName))).Times(1).Return(nil, err)
	return err
}
