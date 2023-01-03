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
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"

	"github.com/spf13/cobra"
)

type logsCmd struct {
	baseCommand
	config logsConfig
}

type logsConfig struct {
	name string
	tail int32
}

func (cc *logsCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "logs",
		Short: "Print the logs for a container.",
		Long:  "Print the logs for a container.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "logs <container-id>\nlogs --name <container-name>\nlogs -n <container-name>",
	}
	cc.setupFlags()
}

func (cc *logsCmd) run(args []string) error {
	var (
		ctr *types.Container
		err error
		ctx = context.Background()
	)

	// parse parameters
	if ctr, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}

	err = cc.cli.gwManClient.Logs(ctx, ctr.ID, cc.config.tail)
	if err != nil {
		return err
	}

	return nil
}

func (cc *logsCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Print the logs of a container with a specific name.")
	flagSet.Int32VarP(&cc.config.tail, "tail", "t", 100, "Lines of recent log file to display. Setting it to -1 will return all logs.")
}
