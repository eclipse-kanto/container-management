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
	"errors"
	"strconv"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	removeCmdFlagForce = "force"
	removeCmdFlagName  = "name"

	// test input constants
	removeContainerID   = "test-ctr"
	removeContainerName = "test-ctr-name"
)

var (
	// command args ---------------
	removeCmdArgs = []string{removeContainerID}
)

// Tests --------------------
func TestRemoveCmdInit(t *testing.T) {
	rmCliTest := &removeCommandTest{}
	rmCliTest.init()

	execTestInit(t, rmCliTest)
}

func TestRemoveCmdSetupFlags(t *testing.T) {
	rmCliTest := &removeCommandTest{}
	rmCliTest.init()

	expectedCfg := removeConfig{
		force: true,
		name:  removeContainerName,
	}
	flagsToApply := map[string]string{
		removeCmdFlagForce: strconv.FormatBool(expectedCfg.force),
		removeCmdFlagName:  expectedCfg.name,
	}

	execTestSetupFlags(t, rmCliTest, flagsToApply, expectedCfg)
}

func TestRemoveCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	rmCliTest := &removeCommandTest{}
	rmCliTest.initWithCtrl(controller)

	execTestsRun(t, rmCliTest)
}

// EOF Tests --------------------------

type removeCommandTest struct {
	cliCommandTestBase
	cmdRemove *removeCmd
}

func (rmTc *removeCommandTest) commandConfig() interface{} {
	return rmTc.cmdRemove.config
}

func (rmTc *removeCommandTest) commandConfigDefault() interface{} {
	return removeConfig{
		force: false,
		name:  "",
	}
}

func (rmTc *removeCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &removeCmd{}
	rmTc.cmdRemove, rmTc.baseCmd = cmd, cmd

	rmTc.cmdRemove.init(rmTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, rmTc.cmdRemove.cmd)
}

func (rmTc *removeCommandTest) runCommand(args []string) error {
	return rmTc.cmdRemove.run(args)
}

func (rmTc *removeCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_remove_id_and_name_provided": {
			args: removeCmdArgs,
			flags: map[string]string{
				removeCmdFlagName: removeContainerName,
			},
			mockExecution: rmTc.mockExecRemoveIDAndName,
		},
		"test_remove_no_id_or_name_provided": {
			mockExecution: rmTc.mockExecRemoveNoIDorName,
		},
		"test_remove_default": {
			args:          removeCmdArgs,
			mockExecution: rmTc.mockExecRemoveNoErrors,
		},
		"test_remove_by_id_err": {
			args:          removeCmdArgs,
			mockExecution: rmTc.mockExecRemoveGetError,
		},
		"test_remove_by_id_get_err": {
			args:          removeCmdArgs,
			mockExecution: rmTc.mockExecRemoveGetError,
		},
		"test_remove_by_id_get_nil_err": {
			args:          removeCmdArgs,
			mockExecution: rmTc.mockExecRemoveGetNilError,
		},
		"test_remove_by_name_default": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameNoErrors,
		},
		"test_remove_by_name_err": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameError,
		},
		"test_remove_by_name_list_err": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameListError,
		},
		"test_remove_by_name_list_nil_ctrs": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameListNilCtrs,
		},
		"test_remove_by_name_list_zero_ctrs": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameListZeroCtrs,
		},
		"test_remove_by_name_list_more_than_one_ctrs": {
			flags:         map[string]string{removeCmdFlagName: removeContainerName},
			mockExecution: rmTc.mockExecRemoveByNameListMoreThanOneCtrs,
		},
		"test_remove_force": {
			args:          removeCmdArgs,
			flags:         map[string]string{removeCmdFlagForce: "true"},
			mockExecution: rmTc.mockExecForceRemoveNoErrors,
		},
		"test_remove_force_err": {
			args:          removeCmdArgs,
			flags:         map[string]string{removeCmdFlagForce: "true"},
			mockExecution: rmTc.mockExecForceRemoveError,
		},
	}
}

// Mocked executions ---------------------------------------------------------------------------------
func (rmTc *removeCommandTest) mockExecRemoveIDAndName(args []string) error {
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], false).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}
func (rmTc *removeCommandTest) mockExecRemoveNoIDorName(args []string) error {
	rmTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	rmTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	return log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
}
func (rmTc *removeCommandTest) mockExecRemoveNoErrors(args []string) error {
	// setup expected calls
	ctr := &types.Container{
		ID:   args[0],
		Name: removeContainerName,
	}
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], false).Times(1).Return(nil)
	// no error expected
	return nil
}

func (rmTc *removeCommandTest) mockExecRemoveError(args []string) error {
	// setup expected calls
	err := errors.New("failed to remove container")
	ctr := &types.Container{
		ID:   args[0],
		Name: removeContainerName,
	}
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], false).Times(1).Return(err)
	return err
}
func (rmTc *removeCommandTest) mockExecRemoveByNameNoErrors(args []string) error {
	// setup expected calls
	ctrs := []*types.Container{{ID: removeContainerID, Name: removeContainerName}}
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return(ctrs, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), ctrs[0].ID, false).Times(1).Return(nil)
	// no error expected
	return nil
}

func (rmTc *removeCommandTest) mockExecRemoveByNameError(args []string) error {
	// setup expected calls
	err := errors.New("failed to remove container")
	ctrs := []*types.Container{{ID: removeContainerID, Name: removeContainerName}}
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return(ctrs, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), ctrs[0].ID, false).Times(1).Return(err)
	// no error expected
	return err
}
func (rmTc *removeCommandTest) mockExecRemoveByNameListError(args []string) error {
	// setup expected calls
	err := errors.New("failed to list containers")
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return(nil, err)
	rmTc.mockClient.EXPECT().Remove(context.Background(), gomock.Any(), false).Times(0)
	return err
}
func (rmTc *removeCommandTest) mockExecRemoveByNameListNilCtrs(args []string) error {
	// setup expected calls
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return(nil, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), gomock.Any(), false).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", removeContainerName)
}
func (rmTc *removeCommandTest) mockExecRemoveByNameListZeroCtrs(args []string) error {
	// setup expected calls
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return([]*types.Container{}, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), gomock.Any(), false).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", removeContainerName)
}
func (rmTc *removeCommandTest) mockExecRemoveByNameListMoreThanOneCtrs(args []string) error {
	// setup expected calls
	rmTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(removeContainerName))).Times(1).Return([]*types.Container{{}, {}}, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), gomock.Any(), false).Times(0)
	return log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", removeContainerName)
}
func (rmTc *removeCommandTest) mockExecRemoveGetNilError(args []string) error {
	// setup expected calls
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], false).Times(0)
	return log.NewErrorf("The requested container with ID = %s was not found.", args[0])
}
func (rmTc *removeCommandTest) mockExecRemoveGetError(args []string) error {
	// setup expected calls
	err := errors.New("failed to remove container")
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, err)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], false).Times(0)
	return err
}

func (rmTc *removeCommandTest) mockExecForceRemoveNoErrors(args []string) error {
	// setup expected calls
	ctr := &types.Container{
		ID:   args[0],
		Name: removeContainerName,
	}
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], true).Times(1).Return(nil)
	// no error expected
	return nil
}

func (rmTc *removeCommandTest) mockExecForceRemoveError(args []string) error {
	// setup expected calls
	err := errors.New("failed to remove container")
	ctr := &types.Container{
		ID:   args[0],
		Name: removeContainerName,
	}
	rmTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	rmTc.mockClient.EXPECT().Remove(context.Background(), args[0], true).Times(1).Return(err)
	// no error expected
	return err
}
