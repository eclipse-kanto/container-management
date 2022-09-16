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

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	"github.com/spf13/cobra"
)

type renameCtrCmd struct {
	baseCommand
	config renameConfig
}
type renameConfig struct {
	name string
}

func (cc *renameCtrCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "rename <container-id> <container-new-name>",
		Short: "Renames a given container.",
		Long:  "Renames a given container.",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "rename <container-id> <container-new-name>\n rename --name <container-name> <container-new-name>\n rename -n <container-name> <container-new-name>",
	}
	cc.setupFlags()
}

func (cc *renameCtrCmd) run(args []string) error {
	var (
		name string
		ctr  *types.Container
		err  error
		ctx  = context.Background()
	)

	if len(args) == 2 {
		name = args[1]
		args = args[:1]
	} else {
		name = args[0]
		args = args[:0]
	}
	if err = util.ValidateName(name); err != nil {
		return err
	}

	if ctr, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}
	if name == ctr.Name {
		return log.NewErrorf("the new name = %s shouldn't be the same", name)
	}
	return cc.cli.gwManClient.Rename(ctx, ctr.ID, name)
}

func (cc *renameCtrCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Renames a container with a specific name.")
}
