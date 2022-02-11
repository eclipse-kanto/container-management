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
	"io"
	"os"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	"github.com/spf13/cobra"
)

type startCmd struct {
	baseCommand
	config  startConfig
	termMgr terminalManager
}

type startConfig struct {
	attached    bool
	interactive bool
	name        string
}

func (cc *startCmd) init(cli *cli) {
	cc.cli = cli
	cc.termMgr = &termMgr{}
	cc.cmd = &cobra.Command{
		Use:   "start <container-id>",
		Short: "Start a container.",
		Long:  "Start a container.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "start <container-id>\n start --name <container-name>\n start -n <container-name>",
	}
	cc.setupFlags()
}

func (cc *startCmd) run(args []string) error {
	var (
		container *types.Container
		err       error
		ctx       = context.Background()
	)
	// parse parameters
	if container, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}

	// protect streams attachment prior to starting the container
	if err = validateContainerState(container); err != nil {
		return err
	}

	if cc.config.interactive || cc.config.attached {
		var wait chan struct{}

		if err := cc.termMgr.CheckTty(cc.config.attached, container.IOConfig.Tty, os.Stdout.Fd()); err != nil {
			return err
		}

		if container.IOConfig.Tty {
			in, out, err := cc.termMgr.SetRawMode(cc.config.interactive, false)
			if err != nil {
				return log.NewError("failed to set raw mode")
			}
			defer func() {
				if err := cc.termMgr.RestoreMode(in, out); err != nil {
					log.ErrorErr(err, "failed to restore term mode")
				}
			}()
		}

		writer, reader, err := cc.cli.gwManClient.Attach(ctx, container.ID, cc.config.interactive)
		if err != nil {
			return err
		}

		defer reader.Close()
		defer writer.(io.WriteCloser).Close()

		wait = make(chan struct{})
		go func() {
			io.Copy(os.Stdout, reader)
			close(wait)
		}()

		go func() {
			io.Copy(writer, os.Stdin)
		}()

		// start container
		if err := cc.cli.gwManClient.Start(ctx, container.ID); err != nil {
			return err
		}

		// wait the io to finish.
		<-wait

		reader.Close()

	} else {
		// start container
		if err := cc.cli.gwManClient.Start(ctx, container.ID); err != nil {
			return err
		}
	}
	return nil

}

func (cc *startCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init attach flags
	flagSet.BoolVar(&cc.config.attached, "a", false, "Attach to the container's process STDOUT/STDERR and forward signals")
	// init interactive flags
	flagSet.BoolVar(&cc.config.interactive, "i", false, "Enable interaction with the current container")
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Start a container with a specific name.")
}

func validateContainerState(container *types.Container) error {
	if container.State.Dead {
		return log.NewErrorf("the container with ID = %s is dead and to be removed - cannot start it", container.ID)
	}
	if container.State.Running {
		return log.NewErrorf("the container with ID = %s is already running - cannot start it again", container.ID)
	}
	if container.State.Paused {
		return log.NewErrorf("the container with ID = %s is paused - cannot start it - use unpause instead", container.ID)
	}
	return nil
}
