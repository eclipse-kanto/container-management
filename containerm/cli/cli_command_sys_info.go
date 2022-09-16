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
	"fmt"

	"github.com/eclipse-kanto/container-management/containerm/version"
	"github.com/spf13/cobra"
)

type sysInfoCmd struct {
	baseCommand
}

const (
	containermInfoFormat = "Engine v%s, API v%s, (build %s %s) \n"
	cliInfoFormat        = "CLI v%s, API v%s, (build %s %s) \n"
)

func (cc *sysInfoCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "sysinfo",
		Short: "Show information about the container management runtime and its environment.",
		Long:  "Show information about the container management runtime and its environment.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "sysinfo",
	}
	cc.setupFlags()
}

func (cc *sysInfoCmd) run(args []string) error {
	cmVersion, err := cc.cli.gwManClient.ProjectInfo(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf(containermInfoFormat,
		cmVersion.ProjectVersion, cmVersion.APIVersion, cmVersion.GitCommit, cmVersion.BuildTime)
	fmt.Printf(cliInfoFormat,
		version.ProjectVersion, version.APIVersion, version.GitCommit, version.BuildTime)
	return nil
}
