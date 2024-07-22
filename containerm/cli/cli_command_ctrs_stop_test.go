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
	"strconv"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	stopCmdFlagTimeout = "time"
	stopCmdFlagName    = "name"
	stopCmdFlagForce   = "force"
	stopCmdFlagSignal  = "signal"

	// test input constants
	stopContainerID   = "test-ctr"
	stopContainerName = "test-ctr-name"
	sigterm           = "SIGTERM"
)

var (
	// Stop command args ---------------
	stopCmdArgs     = []string{stopContainerID}
	defaultStopOpts = &types.StopOpts{
		Timeout: 0,
		Force:   false,
		Signal:  sigterm,
	}
	forceStopOpts = &types.StopOpts{
		Timeout: 0,
		Force:   true,
		Signal:  sigterm,
	}
)

// Tests ------------------------------
func TestStopCmdInit(t *testing.T) {
	stopCliTest := &stopCommandTest{}
	stopCliTest.init()

	execTestInit(t, stopCliTest)
}

func TestStopCmdFlags(t *testing.T) {
	stopCliTest := &stopCommandTest{}
	stopCliTest.init()

	expectedCfg := stopConfig{
		timeout: "50",
		name:    stopContainerName,
		force:   true,
		signal:  sigterm,
	}

	conv, _ := strconv.ParseInt(expectedCfg.timeout, 10, 64)

	flagsToApply := map[string]string{
		stopCmdFlagTimeout: strconv.FormatInt(conv, 10),
		stopCmdFlagName:    expectedCfg.name,
		stopCmdFlagForce:   strconv.FormatBool(expectedCfg.force),
	}

	execTestSetupFlags(t, stopCliTest, flagsToApply, expectedCfg)
}

func TestStopCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	stopCliTest := &stopCommandTest{}
	stopCliTest.initWithCtrl(controller)

	execTestsRun(t, stopCliTest)
}

// EOF Tests --------------------------

type stopCommandTest struct {
	cliCommandTestBase
	cmdStop *stopCmd
}

func (stopTc *stopCommandTest) commandConfig() interface{} {
	return stopTc.cmdStop.config
}

func (stopTc *stopCommandTest) commandConfigDefault() interface{} {
	return stopConfig{
		timeout: "",
		name:    "",
		force:   false,
		signal:  sigterm,
	}
}

func (stopTc *stopCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &stopCmd{}
	stopTc.cmdStop, stopTc.baseCmd = cmd, cmd

	stopTc.cmdStop.init(stopTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, stopTc.cmdStop.cmd)
}

func (stopTc *stopCommandTest) runCommand(args []string) error {
	return stopTc.cmdStop.run(args)
}

func (stopTc *stopCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_stop_id_and_name_provided": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopIDAndName,
		},
		"test_stop_no_id_or_name_provided": {
			mockExecution: stopTc.mockExecStopNoIDorName,
		},
		"test_stop_by_id_get_err": {
			args:          stopCmdArgs,
			mockExecution: stopTc.mockExecStopByIDGetErr,
		},
		"test_stop_by_id_nil_ctr": {
			args:          stopCmdArgs,
			mockExecution: stopTc.mockExecStopByIDGetNilCtr,
		},
		"test_stop_by_name_list_err": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameListErr,
		},
		"test_stop_by_name_list_nil_ctrs": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameListNilCtrs,
		},
		"test_stop_by_name_list_zero_ctrs": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameListZeroCtrs,
		},
		"test_stop_by_name_list_more_than_one_ctrs": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameListMoreThanOneCtrs,
		},
		"test_stop_no_errs": {
			args:          stopCmdArgs,
			mockExecution: stopTc.mockExecStopNoErrors,
		},
		"test_stop_with_timeout": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagTimeout: "20s",
			},
			mockExecution: stopTc.mockExecStopWithTimeout,
		},
		"test_stop_error": {
			args:          stopCmdArgs,
			mockExecution: stopTc.mockExecStopError,
		},
		"test_stop_with_negative_timeout_error": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagTimeout: "-10s",
			},
			mockExecution: stopTc.mockExecStopWithNegativeTimeoutError,
		},
		"test_stop_by_name_no_errs": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameNoErrors,
		},
		"test_stop_by_name_with_timeout": {
			flags: map[string]string{
				stopCmdFlagTimeout: "20s",
				stopCmdFlagName:    stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameWithTimeout,
		},
		"test_stop_by_name_error": {
			flags: map[string]string{
				stopCmdFlagName: stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameError,
		},
		"test_stop_by_name_with_negative_timeout_error": {
			flags: map[string]string{
				stopCmdFlagTimeout: "-10s",
				stopCmdFlagName:    stopContainerName,
			},
			mockExecution: stopTc.mockExecStopByNameWithNegativeTimeoutError,
		},
		"test_stop_with_force": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagForce: "true",
			},
			mockExecution: stopTc.mockExecStopWithForce,
		},
		"test_stop_by_name_with_force": {
			flags: map[string]string{
				stopCmdFlagName:  stopContainerName,
				stopCmdFlagForce: "true",
			},
			mockExecution: stopTc.mockExecStopByNameWithForce,
		},
		"test_stop_with_signal_1": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagSignal: "1",
			},
			mockExecution: stopTc.mockExecStopWithHupSignal,
		},
		"test_stop_with_SIGKILL": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagSignal: "SIGKILL",
			},
			mockExecution: stopTc.mockExecStopWithKillSignal,
		},
		"test_stop_with_negative_signal_error": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagSignal: "-10",
			},
			mockExecution: stopTc.mockExecStopWithNegativeSignalError,
		},
		"test_stop_with_invalid_signal_error": {
			args: stopCmdArgs,
			flags: map[string]string{
				stopCmdFlagSignal: "SIGINVALID",
			},
			mockExecution: stopTc.mockExecStopWithInvalidSignalError,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (stopTc *stopCommandTest) mockExecStopIDAndName(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().Stop(context.Background(), args[0], gomock.Any()).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}

func (stopTc *stopCommandTest) mockExecStopNoIDorName(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
}

func (stopTc *stopCommandTest) mockExecStopByIDGetErr(args []string) error {
	err := log.NewError("could not get container")
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, err)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return err
}

func (stopTc *stopCommandTest) mockExecStopByIDGetNilCtr(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, nil)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with ID = %s was not found.", args[0])
}

func (stopTc *stopCommandTest) mockExecStopByNameListErr(args []string) error {
	err := log.NewError("could not get containers")
	stopTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(nil, err)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return err
}

func (stopTc *stopCommandTest) mockExecStopByNameListNilCtrs(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(nil, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", startContainerName)
}

func (stopTc *stopCommandTest) mockExecStopByNameListZeroCtrs(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return([]*types.Container{}, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", startContainerName)
}

func (stopTc *stopCommandTest) mockExecStopByNameListMoreThanOneCtrs(args []string) error {
	stopTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return([]*types.Container{{}, {}}, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", startContainerName)
}

func (stopTc *stopCommandTest) mockExecStopNoErrors(args []string) error {
	// setup expected calls
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, defaultStopOpts).Times(1).Return(nil)
	// no error expected
	return nil
}

func (stopTc *stopCommandTest) mockExecStopWithTimeout(args []string) error {
	ctr := &types.Container{
		ID: args[0],
	}
	opts := *defaultStopOpts
	opts.Timeout = 20
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, &opts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopError(args []string) error {
	err := log.NewError("failed to stop container")
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, defaultStopOpts).Times(1).Return(err)
	return err
}

func (stopTc *stopCommandTest) mockExecStopWithNegativeTimeoutError(args []string) error {
	err := log.NewError("the timeout = -10 shouldn't be negative")
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, gomock.Any()).Times(0)
	return err
}

func (stopTc *stopCommandTest) mockExecStopByNameWithNegativeTimeoutError(args []string) error {
	err := log.NewError("the timeout = -10 shouldn't be negative")
	ctrs := []*types.Container{{
		ID:   startContainerID,
		Name: stopContainerName,
	}}
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(stopContainerName))).Times(1).Return(ctrs, nil)
	stopTc.mockClient.EXPECT().Stop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	return err
}

func (stopTc *stopCommandTest) mockExecStopByNameNoErrors(args []string) error {
	// setup expected calls
	ctrs := []*types.Container{{
		ID:   startContainerID,
		Name: stopContainerName,
	}}
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(stopContainerName))).Times(1).Return(ctrs, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctrs[0].ID, defaultStopOpts).Times(1).Return(nil)
	// no error expected
	return nil
}

func (stopTc *stopCommandTest) mockExecStopByNameWithTimeout(args []string) error {
	ctrs := []*types.Container{{
		ID:   startContainerID,
		Name: stopContainerName,
	}}
	opts := *defaultStopOpts
	opts.Timeout = 20
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(stopContainerName))).Times(1).Return(ctrs, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctrs[0].ID, &opts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopByNameError(args []string) error {
	err := log.NewError("failed to stop container")
	ctrs := []*types.Container{{
		ID:   startContainerID,
		Name: stopContainerName,
	}}
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(stopContainerName))).Times(1).Return(ctrs, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctrs[0].ID, defaultStopOpts).Times(1).Return(err)
	return err
}

func (stopTc *stopCommandTest) mockExecStopWithForce(args []string) error {
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, forceStopOpts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopByNameWithForce(args []string) error {
	ctrs := []*types.Container{{
		ID:   startContainerID,
		Name: stopContainerName,
	}}
	stopTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(stopContainerName))).Times(1).Return(ctrs, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctrs[0].ID, forceStopOpts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopWithHupSignal(args []string) error {
	ctr := &types.Container{
		ID: args[0],
	}
	opts := &types.StopOpts{
		Signal: "1",
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, opts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopWithKillSignal(args []string) error {
	ctr := &types.Container{
		ID: args[0],
	}
	opts := &types.StopOpts{
		Signal: "SIGKILL",
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, opts).Times(1).Return(nil)
	return nil
}

func (stopTc *stopCommandTest) mockExecStopWithNegativeSignalError(args []string) error {
	err := log.NewError("invalid signal = -10")
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, gomock.Any()).Times(0)
	return err
}

func (stopTc *stopCommandTest) mockExecStopWithInvalidSignalError(args []string) error {
	err := log.NewError("invalid signal = SIGINVALID")
	ctr := &types.Container{
		ID: args[0],
	}
	stopTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(ctr, nil)
	stopTc.mockClient.EXPECT().Stop(context.Background(), ctr.ID, gomock.Any()).Times(0)
	return err
}
