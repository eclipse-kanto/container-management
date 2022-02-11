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
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
	"github.com/golang/mock/gomock"
)

// Tests ------------------------------
func TestSysInfoCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	sysInfoCliTest := &sysInfoCommandTest{}
	sysInfoCliTest.initWithCtrl(controller)

	execTestsRun(t, sysInfoCliTest)
}

// EOF Tests --------------------------

type sysInfoCommandTest struct {
	cliCommandTestBase
	sysInfoCmd *sysInfoCmd
}

func (sysInfoTc *sysInfoCommandTest) prepareCommand(flagsCfg map[string]string) {
	// setup command to test
	cmd := &sysInfoCmd{}
	sysInfoTc.sysInfoCmd, sysInfoTc.baseCmd = cmd, cmd

	sysInfoTc.sysInfoCmd.init(sysInfoTc.mockRootCommand)
	// setup command flags
	setCmdFlags(flagsCfg, sysInfoTc.sysInfoCmd.cmd)
}

func (sysInfoTc *sysInfoCommandTest) runCommand(args []string) error {
	return sysInfoTc.sysInfoCmd.run(args)
}

func (sysInfoTc *sysInfoCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		"test_sys_info_default": {
			mockExecution: sysInfoTc.mockExecSysInfoNoErrors,
		},
		"test_sys_info_err": {
			mockExecution: sysInfoTc.mockExecSysInfoErrors,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (sysInfoTc *sysInfoCommandTest) mockExecSysInfoNoErrors(args []string) error {
	// setup expected calls
	info := types.ProjectInfo{
		ProjectVersion: "test-project-version",
		BuildTime:      "test-build-time",
		APIVersion:     "test-api-version",
		GitCommit:      "test-git-commit",
	}
	sysInfoTc.mockClient.EXPECT().ProjectInfo(gomock.AssignableToTypeOf(context.Background())).Times(1).Return(info, nil)
	// no error expected
	return nil
}

func (sysInfoTc *sysInfoCommandTest) mockExecSysInfoErrors(args []string) error {
	// setup expected calls
	err := errors.New("failed to get sys info")
	sysInfoTc.mockClient.EXPECT().ProjectInfo(gomock.AssignableToTypeOf(context.Background())).Times(1).Return(types.ProjectInfo{}, err)
	// no error expected
	return err
}
