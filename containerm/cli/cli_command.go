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

import "github.com/spf13/cobra"

// Command define some interfaces that the command must implement them.
type command interface {
	init(*cli)
	command() *cobra.Command
}

type baseCommand struct {
	cmd *cobra.Command
	cli *cli
}

func (baseCmd *baseCommand) init(cli *cli) {
	// init method implemented by each individual command to initialize its description, arguments, etc.
}
func (baseCmd *baseCommand) setupFlags() {
	// setupFlags method implemented by each individual command to setup its flags
}
func (baseCmd *baseCommand) run(args []string) error { return nil }
func (baseCmd *baseCommand) command() *cobra.Command { return baseCmd.cmd }
