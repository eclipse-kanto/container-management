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
	"errors"
	"strconv"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	mockscli "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/cli"
	mocksio "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/io"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	startCmdFlagAttached    = "a"
	startCmdFlagInteractive = "i"
	startCmdFlagName        = "name"

	// test input constants
	startContainerID   = "test-ctr"
	startContainerName = "test-ctr-name"
)

var (
	// command args ---------------
	startCmdArgs = []string{startContainerID}
)

// Tests --------------------
func TestStartCmdInit(t *testing.T) {
	startCliTest := &startCommandTest{}
	startCliTest.init()

	execTestInit(t, startCliTest)
}

func TestStartCmdSetupFlags(t *testing.T) {
	startCliTest := &startCommandTest{}
	startCliTest.init()

	expectedCfg := startConfig{
		attached:    true,
		interactive: true,
		name:        startContainerName,
	}
	flagsToApply := map[string]string{
		startCmdFlagAttached:    strconv.FormatBool(expectedCfg.attached),
		startCmdFlagInteractive: strconv.FormatBool(expectedCfg.interactive),
		startCmdFlagName:        expectedCfg.name,
	}
	execTestSetupFlags(t, startCliTest, flagsToApply, expectedCfg)
}

func TestStartCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	startCliTest := &startCommandTest{}
	startCliTest.initWithCtrl(controller)

	execTestsRun(t, startCliTest)
}

// EOF Tests --------------------------

type startCommandTest struct {
	cliCommandTestBase
	cmdStart *startCmd
}

func (startTc *startCommandTest) commandConfig() interface{} {
	return startTc.cmdStart.config
}

func (startTc *startCommandTest) commandConfigDefault() interface{} {
	return startConfig{
		attached:    false,
		interactive: false,
		name:        "",
	}
}

func (startTc *startCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &startCmd{}
	startTc.cmdStart, startTc.baseCmd = cmd, cmd

	startTc.cmdStart.init(startTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, startTc.cmdStart.cmd)
}

func (startTc *startCommandTest) runCommand(args []string) error {
	return startTc.cmdStart.run(args)
}

func (startTc *startCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		// Test full
		"test_start_id_and_name_provided": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartIDAndName,
		},
		"test_start_no_id_or_name_provided": {
			mockExecution: startTc.mockExecStartNoIDorName,
		},
		"test_start_by_id_get_err": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartByIDGetErr,
		},
		"test_start_by_id_nil_ctr": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartByIDGetNilCtr,
		},
		"test_start_by_name_list_err": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameListErr,
		},
		"test_start_by_name_list_nil_ctrs": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameListNilCtrs,
		},
		"test_start_by_name_list_zero_ctrs": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameListZeroCtrs,
		},
		"test_start_by_name_list_more_than_one_ctrs": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameListMoreThanOneCtrs,
		},
		"test_start_no_ios": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartByIDNoIOsNoErrs,
		},
		"test_start_no_ios_errs": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartByIDNoIOsErrs,
		},
		"test_start_by_name_no_ios": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameNoIOsNoErrs,
		},
		"test_start_by_name_no_ios_errs": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameNoIOsErrs,
		},
		"test_start_default": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagAttached:    "true",
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartDefault,
		},
		"test_start_by_name_default": {
			flags: map[string]string{
				startCmdFlagAttached:    "true",
				startCmdFlagInteractive: "true",
				startCmdFlagName:        startContainerName,
			},
			mockExecution: startTc.mockExecStartByNameDefault,
		},
		"test_start_get_error": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagAttached:    "true",
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartGetError,
		},
		"test_start_state_dead": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartCtrDead,
		},
		"test_start_by_name_state_dead": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartCtrByNameDead,
		},
		"test_start_state_running": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartCtrRunning,
		},
		"test_start_by_name_state_running": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartCtrByNameRunning,
		},
		"test_start_state_paused": {
			args:          startCmdArgs,
			mockExecution: startTc.mockExecStartCtrPaused,
		},
		"test_start_by_name_state_paused": {
			flags: map[string]string{
				startCmdFlagName: startContainerName,
			},
			mockExecution: startTc.mockExecStartCtrByNamePaused,
		},
		"test_start_interactive_only": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyNoErrors,
		},
		"test_start_by_name_interactive_only": {
			flags: map[string]string{
				startCmdFlagInteractive: "true",
				startCmdFlagName:        startContainerName,
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyByNameNoErrors,
		},
		"test_start_interactive_only_attach_err": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyAttachErrors,
		},
		"test_start_by_name_interactive_only_attach_err": {
			flags: map[string]string{
				startCmdFlagInteractive: "true",
				startCmdFlagName:        startContainerName,
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyByNameAttachErrors,
		},
		"test_start_interactive_only_start_err": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyStartErrors,
		},
		"test_start_by_name_interactive_only_start_err": {
			flags: map[string]string{
				startCmdFlagInteractive: "true",
				startCmdFlagName:        startContainerName,
			},
			mockExecution: startTc.mockExecStartInteractiveOnlyByNameStartErrors,
		},
		"test_start_interactive_attached_no_tty": {
			args: startCmdArgs,
			flags: map[string]string{
				startCmdFlagAttached:    "true",
				startCmdFlagInteractive: "true",
			},
			mockExecution: startTc.mockExecStartInteractiveAttachedNoTTY,
		},
		"test_start_by_name_interactive_attached_no_tty": {
			flags: map[string]string{
				startCmdFlagAttached:    "true",
				startCmdFlagInteractive: "true",
				startCmdFlagName:        startContainerName,
			},
			mockExecution: startTc.mockExecStartInteractiveByNameAttachedNoTTY,
		},
	}
}

// Mocked executions ---------------------------------------------------------------------------------
func (startTc *startCommandTest) mockExecStartIDAndName(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), args[0]).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), args[0], gomock.Any()).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}

func (startTc *startCommandTest) mockExecStartNoIDorName(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
}
func (startTc *startCommandTest) mockExecStartByIDGetErr(args []string) error {
	err := log.NewError("could not get container")
	startTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, err)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return err
}
func (startTc *startCommandTest) mockExecStartByIDGetNilCtr(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(nil, nil)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with ID = %s was not found.", args[0])
}

func (startTc *startCommandTest) mockExecStartByNameListErr(args []string) error {
	err := log.NewError("could not get containers")
	startTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(nil, err)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return err
}
func (startTc *startCommandTest) mockExecStartByNameListNilCtrs(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(nil, nil)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", startContainerName)
}
func (startTc *startCommandTest) mockExecStartByNameListZeroCtrs(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return([]*types.Container{}, nil)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", startContainerName)
}
func (startTc *startCommandTest) mockExecStartByNameListMoreThanOneCtrs(args []string) error {
	startTc.mockClient.EXPECT().Get(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return([]*types.Container{{}, {}}, nil)
	startTc.mockClient.EXPECT().Start(context.Background(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", startContainerName)
}
func (startTc *startCommandTest) mockExecStartByIDNoIOsNoErrs(args []string) error {
	testCtr := &types.Container{
		ID: args[0],
		State: &types.State{
			Status: types.Created,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), testCtr.ID).Times(1).Return(nil)
	return nil
}
func (startTc *startCommandTest) mockExecStartByNameNoIOsNoErrs(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerName,
		State: &types.State{
			Status: types.Created,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(testCtrs, nil)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), testCtrs[0].ID).Times(1).Return(nil)
	return nil
}
func (startTc *startCommandTest) mockExecStartByIDNoIOsErrs(args []string) error {
	testCtr := &types.Container{
		ID: args[0],
		State: &types.State{
			Status: types.Created,
		},
	}
	err := log.NewError("failed to start container")
	startTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), testCtr.ID).Times(1).Return(err)
	return err
}
func (startTc *startCommandTest) mockExecStartByNameNoIOsErrs(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerName,
		State: &types.State{
			Status: types.Created,
		},
	}}
	err := log.NewError("failed to start container")
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerName))).Times(1).Return(testCtrs, nil)
	startTc.mockClient.EXPECT().Attach(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(context.Background(), testCtrs[0].ID).Times(1).Return(err)
	return err
}

func (startTc *startCommandTest) mockExecStartDefault(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: true,
		},
		State: &types.State{
			Status: types.Created,
		},
	}

	startTc.mockClient.EXPECT().Get(context.Background(), testCtr.ID).Times(1).Return(testCtr, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(startTc.cmdStart.config.attached, testCtr.IOConfig.Tty, gomock.Any()).Times(1).Return(nil)
	tMgr.EXPECT().SetRawMode(startTc.cmdStart.config.interactive, false).Times(1).Return(nil, nil, nil)
	tMgr.EXPECT().RestoreMode(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtr.ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtr.ID)).Times(1).Return(nil)
	return nil
}

func (startTc *startCommandTest) mockExecStartGetError(args []string) error {
	err := errors.New("failed to get container")
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(nil, err)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)
	return err
}

func (startTc *startCommandTest) mockExecStartCtrDead(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		State: &types.State{
			Dead: true,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is dead and to be removed - cannot start it", startContainerID)
}

func (startTc *startCommandTest) mockExecStartCtrRunning(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		State: &types.State{
			Running: true,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is already running - cannot start it again", startContainerID)
}

func (startTc *startCommandTest) mockExecStartCtrPaused(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		State: &types.State{
			Paused: true,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is paused - cannot start it - use unpause instead", startContainerID)
}

func (startTc *startCommandTest) mockExecStartInteractiveOnlyNoErrors(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtr.IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtr.ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtr.ID)).Times(1).Return(nil)

	return nil
}
func (startTc *startCommandTest) mockExecStartInteractiveOnlyAttachErrors(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtr.IOConfig.Tty, gomock.Any()).Times(1).Return(nil)
	err := errors.New("failed to attach")
	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtr.ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, err)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtr.ID)).Times(0)

	return err
}

func (startTc *startCommandTest) mockExecStartInteractiveOnlyStartErrors(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtr.IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF")).MaxTimes(1)
	mockReadCloser.EXPECT().Close().Times(1)
	mockWriter.EXPECT().Close()

	err := errors.New("failed to attach")
	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtr.ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtr.ID)).Times(1).Return(err)

	return err
}

func (startTc *startCommandTest) mockExecStartInteractiveAttachedNoTTY(args []string) error {
	testCtr := &types.Container{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}
	startTc.mockClient.EXPECT().Get(context.Background(), startContainerID).Times(1).Return(testCtr, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(true, testCtr.IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtr.ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtr.ID)).Times(1).Return(nil)

	return nil
}
func (startTc *startCommandTest) mockExecStartByNameDefault(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: true,
		},
		State: &types.State{
			Status: types.Created,
		},
	}}

	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(startTc.cmdStart.config.attached, testCtrs[0].IOConfig.Tty, gomock.Any()).Times(1).Return(nil)
	tMgr.EXPECT().SetRawMode(startTc.cmdStart.config.interactive, false).Times(1).Return(nil, nil, nil)
	tMgr.EXPECT().RestoreMode(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtrs[0].ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtrs[0].ID)).Times(1).Return(nil)
	return nil
}

func (startTc *startCommandTest) mockExecStartCtrByNameDead(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		State: &types.State{
			Dead: true,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is dead and to be removed - cannot start it", startContainerID)
}

func (startTc *startCommandTest) mockExecStartCtrByNameRunning(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		State: &types.State{
			Running: true,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is already running - cannot start it again", startContainerID)
}

func (startTc *startCommandTest) mockExecStartCtrByNamePaused(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		State: &types.State{
			Paused: true,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Any()).Times(0)

	return log.NewErrorf("the container with ID = %s is paused - cannot start it - use unpause instead", startContainerID)
}

func (startTc *startCommandTest) mockExecStartInteractiveOnlyByNameNoErrors(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtrs[0].IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtrs[0].ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtrs[0].ID)).Times(1).Return(nil)

	return nil
}
func (startTc *startCommandTest) mockExecStartInteractiveOnlyByNameAttachErrors(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtrs[0].IOConfig.Tty, gomock.Any()).Times(1).Return(nil)
	err := errors.New("failed to attach")
	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtrs[0].ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, err)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtrs[0].ID)).Times(0)

	return err
}

func (startTc *startCommandTest) mockExecStartInteractiveOnlyByNameStartErrors(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(false, testCtrs[0].IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF")).MaxTimes(1)
	mockReadCloser.EXPECT().Close().Times(1)
	mockWriter.EXPECT().Close()

	err := errors.New("failed to attach")
	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtrs[0].ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtrs[0].ID)).Times(1).Return(err)

	return err
}

func (startTc *startCommandTest) mockExecStartInteractiveByNameAttachedNoTTY(args []string) error {
	testCtrs := []*types.Container{{
		ID: startContainerID,
		IOConfig: &types.IOConfig{
			Tty: false,
		},
		State: &types.State{
			Status: types.Created,
		},
	}}
	startTc.mockClient.EXPECT().List(context.Background(), gomock.AssignableToTypeOf(client.WithName(startContainerID))).Times(1).Return(testCtrs, nil)

	mockWriter := mocksio.NewMockWriteCloser(startTc.gomockCtrl)
	mockReadCloser := mocksio.NewMockReadCloser(startTc.gomockCtrl)
	tMgr := mockscli.NewMockterminalManager(startTc.gomockCtrl)
	startTc.cmdStart.termMgr = tMgr

	tMgr.EXPECT().CheckTty(true, testCtrs[0].IOConfig.Tty, gomock.Any()).Times(1).Return(nil)

	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, errors.New("EOF"))
	mockReadCloser.EXPECT().Close().Times(2)
	mockWriter.EXPECT().Close()

	startTc.mockClient.EXPECT().Attach(gomock.Any(), gomock.Eq(testCtrs[0].ID), gomock.Eq(true)).Times(1).Return(mockWriter, mockReadCloser, nil)
	startTc.mockClient.EXPECT().Start(gomock.Any(), gomock.Eq(testCtrs[0].ID)).Times(1).Return(nil)

	return nil
}
