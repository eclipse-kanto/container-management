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
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/client"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type mockExecution func(args []string) error

type testRunExecutionConfig struct {
	args          []string
	flags         map[string]string
	mockExecution mockExecution
}

type cliCommandTest interface {
	commandConfig() interface{}
	commandConfigDefault() interface{}
	commandFlags() map[string]string
	commandFlagsDefault() map[string]string
	prepareCommand(flagsCfg map[string]string) error
	generateRunExecutionConfigs() map[string]testRunExecutionConfig
	runCommand(args []string) error
}

type cliCommandTestBase struct {
	gomockCtrl      *gomock.Controller
	mockClient      *mocks.MockClient
	mockRootCommand *cli
	baseCmd         command
}

func (cliBase *cliCommandTestBase) commandConfig() interface{}        { return nil }
func (cliBase *cliCommandTestBase) commandConfigDefault() interface{} { return nil }
func (cliBase *cliCommandTestBase) commandFlags() map[string]string {
	flags := map[string]string{}
	cliBase.baseCmd.command().Flags().VisitAll(func(flag *pflag.Flag) {
		flags[flag.Name] = flag.Value.String()
	})
	return flags
}
func (cliBase *cliCommandTestBase) commandFlagsDefault() map[string]string {
	flags := map[string]string{}
	cliBase.baseCmd.command().Flags().VisitAll(func(flag *pflag.Flag) {
		flags[flag.Name] = flag.DefValue
	})
	return flags
}
func (cliBase *cliCommandTestBase) prepareCommand(flagsCfg map[string]string) {
	// must be overridden in the according CLI command test implementation
}
func (cliBase *cliCommandTestBase) runCommand(args []string) error { return nil }
func (cliBase *cliCommandTestBase) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return nil
}
func (cliBase *cliCommandTestBase) initWithCtrl(gomockCtrl *gomock.Controller) {
	cliBase.gomockCtrl = gomockCtrl
	cliBase.mockClient = mocks.NewMockClient(gomockCtrl)
	cliBase.mockRootCommand = &cli{
		config:      config{},
		gwManClient: cliBase.mockClient,
	}
}
func (cliBase *cliCommandTestBase) init() {
	cliBase.mockRootCommand = &cli{}
}

func execTestInit(t *testing.T, cliTest cliCommandTest) {
	err := cliTest.prepareCommand(nil)
	testutil.AssertNil(t, err)

	defaultCfg := cliTest.commandConfigDefault()
	defaultFlags := cliTest.commandFlagsDefault()

	currentCfg := cliTest.commandConfig()
	currentFlags := cliTest.commandFlags()

	// assert command flags properly initialized
	testutil.AssertEqual(t, defaultCfg, currentCfg)
	// assert command configuration properly initialized
	testutil.AssertEqual(t, defaultFlags, currentFlags)

}

func execTestSetupFlags(t *testing.T, cliTest cliCommandTest, flagsToApply map[string]string, expectedCommandCfg interface{}) {
	err := cliTest.prepareCommand(flagsToApply)
	testutil.AssertNil(t, err)

	currentCfg := cliTest.commandConfig()

	// assert command flags properly initialized
	testutil.AssertEqual(t, expectedCommandCfg, currentCfg)
}

func execTestsRun(t *testing.T, cliTest cliCommandTest) {
	// execute tests
	for testName, testCase := range cliTest.generateRunExecutionConfigs() {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			// config command
			err := cliTest.prepareCommand(testCase.flags)
			testutil.AssertNil(t, err)
			// prepare mocks
			expectedRunErr := testCase.mockExecution(testCase.args)
			// perform the real call
			resultErr := cliTest.runCommand(testCase.args)
			// assert result
			if expectedRunErr == nil {
				testutil.AssertNil(t, resultErr)
			} else {
				testutil.AssertNotNil(t, resultErr)
				testutil.AssertContainsString(t, resultErr.Error(), expectedRunErr.Error())
			}
		})
	}
}

func setCmdFlags(flagValues map[string]string, cmd *cobra.Command) error {
	if flagValues != nil {
		for flagKey, flagValue := range flagValues {
			flag := cmd.Flag(flagKey)
			val := flag.Value
			err := val.Set(flagValue)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
