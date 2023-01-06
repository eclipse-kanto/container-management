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
	"fmt"
	"os"
)

func main() {
	cli := newCli()

	// set global flags for rootCmd in cli.
	cli.setupCommandFlags()

	base := &baseCommand{cmd: cli.rootCmd, cli: cli}
	base.cmd.SilenceErrors = true

	cli.addCommand(base, &createCmd{})
	cli.addCommand(base, &removeCmd{})
	cli.addCommand(base, &startCmd{})
	cli.addCommand(base, &stopCmd{})
	cli.addCommand(base, &listCmd{})
	cli.addCommand(base, &getCtrInfoCmd{})
	cli.addCommand(base, &sysInfoCmd{})
	cli.addCommand(base, &updateCmd{})
	cli.addCommand(base, &renameCtrCmd{})
	cli.addCommand(base, &logsCmd{})

	if err := cli.run(); err != nil {
		// not ExitError, print error to os.Stderr, exit code 1.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
