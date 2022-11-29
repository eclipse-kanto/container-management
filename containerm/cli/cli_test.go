// Copyright (c) 2022 Contributors to the Eclipse Foundation
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
	"fmt"
	"testing"
)

const (
	// command flags
	cmdFlagHost  = "host"
	cmdFlagDebug = "debug"

	// test input constants
	testAddressPath = "test-address-path"
)

type cliCmdTest struct {
	cliCommandTestBase
	cli *cli
}

func TestCmdInit(t *testing.T) {
	ct := &cliCmdTest{}
	ct.init()

	execTestInit(t, ct)
}

func TestCmdFlags(t *testing.T) {
	ct := &cliCmdTest{}
	ct.init()

	expectedCfg := config{
		addressPath: testAddressPath,
		debug:       true,
	}

	flagsToApply := map[string]string{
		cmdFlagHost:  expectedCfg.addressPath,
		cmdFlagDebug: fmt.Sprintf("%v", expectedCfg.debug),
	}

	execTestSetupFlags(t, ct, flagsToApply, expectedCfg)
}

func (c *cliCmdTest) prepareCommand(flagsCfg map[string]string) error {
	cli := newCli()
	c.cli, c.baseCmd = cli, &baseCommand{cmd: cli.rootCmd, cli: cli}

	cli.setupCommandFlags()

	return setCmdFlags(flagsCfg, c.cli.rootCmd)
}

func (c *cliCmdTest) commandConfig() interface{} {
	return c.cli.config
}

func (c *cliCmdTest) commandConfigDefault() interface{} {
	return config{
		addressPath: cmdAddressPathDefault,
		debug:       cmdDebugDefault,
	}
}
