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
	"encoding/json"
	"fmt"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	"github.com/spf13/cobra"
)

type getCtrInfoCmd struct {
	baseCommand
	config getInfoConfig
}
type getInfoConfig struct {
	name string
}

func (cc *getCtrInfoCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "get <container-id>",
		Short: "Get detailed information about a given container.",
		Long:  "Get detailed information about a given container.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "get <container-id>\n get --name <container-name>\n get -n <container-name>",
	}
	cc.setupFlags()
}

func (cc *getCtrInfoCmd) run(args []string) error {
	var (
		ctr *types.Container
		err error
		ctx = context.Background()
	)
	// parse parameters
	if ctr, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}
	printInfo(ctr)

	return nil
}
func (cc *getCtrInfoCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Get the information about a container with a specific name.")
}

func printInfo(ctr *types.Container) {

	byteArray, err := json.MarshalIndent(ctr, "", "   ")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(byteArray))
}
