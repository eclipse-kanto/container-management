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
	"errors"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	errorcli "github.com/eclipse-kanto/container-management/containerm/util/error"
	"github.com/spf13/cobra"
)

type removeCmd struct {
	baseCommand
	config removeConfig
}

type removeConfig struct {
	force bool
	name  string
}

func (cc *removeCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "remove <container-id/s>",
		Short: "Remove a container/s.",
		Long:  "Remove a container and frees the associated resources.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "remove <container-id>\n remove <container-id> <container-id> \n remove --name <container-name>\n remove -n <container-name>",
	}
	cc.setupFlags()
}

func (cc *removeCmd) run(args []string) error {
	var (
		ctr  *types.Container
		err  error
		ctx  = context.Background()
		errs errorcli.CompoundError
	)
	// parse parameters
	for _, arg := range args {
		ctr, err = utilcli.ValidateContainerByNameArgsSingle(ctx, []string{arg}, cc.config.name, cc.cli.gwManClient)
		if err == nil {
			cc.cli.gwManClient.Remove(ctx, ctr.ID, cc.config.force)
		} else {
			errs.Append(err)
		}
	}
	if errs.Size() > 0 {
		return errors.New(errs.Error())
	}
	return nil
}

func (cc *removeCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init terminal flags
	flagSet.BoolVarP(&cc.config.force, "force", "f", false, "Force stopping before removing a container")
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Remove a container with a specific name.")
}
