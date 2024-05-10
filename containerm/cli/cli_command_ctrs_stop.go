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
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/util"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	"github.com/spf13/cobra"
)

type stopCmd struct {
	baseCommand
	config stopConfig
}

type stopConfig struct {
	timeout string
	name    string
	force   bool
	signal  string
}

func (cc *stopCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "stop <container-id>",
		Short: "Stop a container.",
		Long:  "Stop a container.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: " stop <container-id>\n stop --name <container-name>\n stop -n <container-name>",
	}
	cc.setupFlags()
}

func (cc *stopCmd) run(args []string) error {
	var (
		container *types.Container
		err       error
		ctx       = context.Background()
	)
	// parse parameters
	if container, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}
	stopOpts := &types.StopOpts{
		Force:  cc.config.force,
		Signal: cc.config.signal,
	}
	if stopOpts.Timeout, err = durationStringToSeconds(cc.config.timeout); err != nil {
		return err
	}
	if err = util.ValidateStopOpts(stopOpts); err != nil {
		return err
	}
	return cc.cli.gwManClient.Stop(ctx, container.ID, stopOpts)
}

func durationStringToSeconds(duration string) (int64, error) {
	if duration == "" {
		return 0, nil
	}
	stopTime, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return int64(stopTime.Round(time.Second).Seconds()), nil
}

func (cc *stopCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init timeout flag
	flagSet.StringVarP(&cc.config.timeout, "time", "t", "", "Sets the timeout period to gracefully stop the container as duration string, e.g. 15s or 1m15s. When timeout expires the container process would be forcibly killed. If not specified the daemon default container stop timeout will be used.")
	// init name flag
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Stop a container with a specific name.")
	// init force flag
	flagSet.BoolVarP(&cc.config.force, "force", "f", false, "Whether to send a SIGKILL signal to the container's process if it does not finish within the timeout specified.")
	// init signal flag
	flagSet.StringVarP(&cc.config.signal, "signal", "s", "SIGTERM", "Stop a container using a specific signal. Signals could be specified by using their names or numbers, e.g. SIGINT or 2.")
}
