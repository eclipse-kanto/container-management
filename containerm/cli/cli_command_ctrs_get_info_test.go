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
	getInfoCmdFlagName = "name"

	// test input constants
	getInfoContainerID = "test-ctr"
	getInfoFlagName    = "test-ctr-name"
)

var (
	// command args ---------------
	getInfoCmdArgs = []string{getInfoContainerID}
)

// Tests ------------------------------
func TestGetInfoCmdInit(t *testing.T) {
	getInfoCliTest := &getInfoCommandTest{}
	getInfoCliTest.init()

	execTestInit(t, getInfoCliTest)
}

func TestGetInfoCmdFlags(t *testing.T) {
	getInfoCliTest := &getInfoCommandTest{}
	getInfoCliTest.init()

	expectedCfg := getInfoConfig{
		name: getInfoFlagName,
	}

	flagsToApply := map[string]string{
		getInfoCmdFlagName: expectedCfg.name,
	}

	execTestSetupFlags(t, getInfoCliTest, flagsToApply, expectedCfg)
}

func TestGetInfoCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	getInfoCliTest := &getInfoCommandTest{}
	getInfoCliTest.initWithCtrl(controller)

	execTestsRun(t, getInfoCliTest)
}

// EOF Tests --------------------------

type getInfoCommandTest struct {
	cliCommandTestBase
	getInfoCmd *getCtrInfoCmd
}

func (getInfoTc *getInfoCommandTest) commandConfig() interface{} {
	return getInfoTc.getInfoCmd.config
}

func (getInfoTc *getInfoCommandTest) commandConfigDefault() interface{} {
	return getInfoConfig{
		name: "",
	}
}

func (getInfoTc *getInfoCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &getCtrInfoCmd{}
	getInfoTc.getInfoCmd, getInfoTc.baseCmd = cmd, cmd

	getInfoTc.getInfoCmd.init(getInfoTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, getInfoTc.getInfoCmd.cmd)
}

func (getInfoTc *getInfoCommandTest) runCommand(args []string) error {
	return getInfoTc.getInfoCmd.run(args)
}

func (getInfoTc *getInfoCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_get_info_id_and_name_provided": {
			args: getInfoCmdArgs,
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoIDAndName,
		},
		"test_get_info_no_id_or_name_provided": {
			mockExecution: getInfoTc.mockExecGetInfoNoIDorName,
		},
		"test_get_info_by_id_default": {
			args:          getInfoCmdArgs,
			mockExecution: getInfoTc.mockExecGetInfoNoErrors,
		},
		"test_get_info_by_id_ctr_nil": {
			args:          getInfoCmdArgs,
			mockExecution: getInfoTc.mockExecGetInfoByIDNilCtr,
		},
		"test_get_info_by_id_err": {
			args:          getInfoCmdArgs,
			mockExecution: getInfoTc.mockExecGetInfoErrors,
		},
		"test_get_info_by_name_err": {
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoByNameErr,
		},
		"test_get_info_by_name_nil_ctr": {
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoByNameNilCtr,
		},
		"test_get_info_by_name_zero_ctrs": {
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoByNameZeroCtrs,
		},
		"test_get_info_by_name_more_than_one_ctrs": {
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoByNameMoreThanOneCtrs,
		},
		"test_get_info_by_name_no_errs": {
			flags: map[string]string{
				getInfoCmdFlagName: getInfoFlagName,
			},
			mockExecution: getInfoTc.mockExecGetInfoByNameDefault,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (getInfoTc *getInfoCommandTest) mockExecGetInfoIDAndName(args []string) error {
	getInfoTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoNoIDorName(args []string) error {
	getInfoTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	return log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoNoErrors(args []string) error {
	ctr := &types.Container{ID: args[0]}
	getInfoTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	return nil
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByIDNilCtr(args []string) error {
	getInfoTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, nil)
	return log.NewErrorf("The requested container with ID = %s was not found.", args[0])
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoErrors(args []string) error {
	err := log.NewError("error getting container")
	getInfoTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, err)
	return err
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByNameErr(args []string) error {
	err := log.NewError("error listing containers")
	getInfoTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(getInfoFlagName))).Times(1).Return(nil, err)
	return err
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByNameNilCtr(args []string) error {
	getInfoTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(getInfoFlagName))).Times(1).Return(nil, nil)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", getInfoFlagName)
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByNameZeroCtrs(args []string) error {
	getInfoTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(getInfoFlagName))).Times(1).Return([]*types.Container{}, nil)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", getInfoFlagName)
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByNameMoreThanOneCtrs(args []string) error {
	res := []*types.Container{{Name: getInfoFlagName}, {Name: getInfoFlagName}}
	getInfoTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(getInfoFlagName))).Times(1).Return(res, nil)
	return log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", getInfoFlagName)
}
func (getInfoTc *getInfoCommandTest) mockExecGetInfoByNameDefault(args []string) error {
	res := []*types.Container{{
		ID:   getInfoContainerID,
		Name: getInfoFlagName,
	}}
	getInfoTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(getInfoFlagName))).Times(1).Return(res, nil)
	return nil
}
