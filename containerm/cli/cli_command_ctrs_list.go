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
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/spf13/cobra"
)

type listCmd struct {
	baseCommand
	config listConfig
}

type listConfig struct {
	name   string
	quiet  bool
	filter []string
}

func (cc *listCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "list",
		Short: "List all containers.",
		Long:  "List all containers.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: " list\n list --name <container-name>\n list --quiet\n list --filter status=created",
	}
	cc.setupFlags()
}

func (cc *listCmd) run(args []string) error {
	var filters []client.Filter
	if cc.config.name != "" {
		filters = append(filters, client.WithName(cc.config.name))
	}
	ctrs, err := cc.cli.gwManClient.List(context.Background(), filters...)
	if err != nil {
		return err
	}
	if len(cc.config.filter) > 0 {
		filtered, err := filterBy(cc.config.filter, ctrs)
		if err != nil {
			return err
		}
		ctrs = filtered
	}
	if cc.config.quiet {
		for i, ctr := range ctrs {
			if i != len(ctrs)-1 {
				fmt.Printf("%s ", ctr.ID)
			} else {
				fmt.Println(ctr.ID)
			}
		}
		return nil
	}
	if len(ctrs) == 0 {
		fmt.Println("No containers found.")
	} else {
		prettyPrint(ctrs)
	}
	return nil
}

func (cc *listCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "List all containers with a specific name.")
	flagSet.BoolVarP(&cc.config.quiet, "quiet", "q", false, "List only container IDs.")
	flagSet.StringSliceVar(&cc.config.filter, "filter", nil, "Lists only containers with a specified filter. The containers can be filtered by their status, image and exitcode.")
}

func filterBy(input []string, ctrs []*types.Container) ([]*types.Container, error) {
	var (
		holderStatus      string
		holderImage       string
		convertedExitCode int = -1
		err               error
	)
	filteredCtrs := []*types.Container{}
	for _, inp := range input {
		if strings.HasPrefix(inp, "status=") {
			holderStatus = strings.TrimPrefix(inp, "status=")
		} else if strings.HasPrefix(inp, "image=") {
			holderImage = strings.TrimPrefix(inp, "image=")
		} else if strings.HasPrefix(inp, "exitcode=") {
			convertedExitCode, err = strconv.Atoi(strings.TrimPrefix(inp, "exitcode="))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("no filter: %s", strings.Split(inp, "=")[0])
		}
	}
	for _, ctr := range ctrs {
		if holderStatus != "" && !strings.EqualFold(ctr.State.Status.String(), holderStatus) {
			continue
		}
		if holderImage != "" && !strings.EqualFold(ctr.Image.Name, holderImage) {
			continue
		}
		if int64(convertedExitCode) != -1 && ctr.State.ExitCode != int64(convertedExitCode) {
			continue
		}
		filteredCtrs = append(filteredCtrs, ctr)
	}
	return filteredCtrs, nil
}

/*
Eventually a pretty print util could be created for the table-formatted
container data printing or a respective 3-rd party go package could be used.
*/
const tableRowTemplate = "%-37s\t%-37s\t%-60s\t%-10s\t%-32s\t%-10s\t\n"

func prettyPrint(ctrs []*types.Container) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', tabwriter.Debug)
	defer w.Flush()
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, tableRowTemplate, "ID", "Name", "Image", "Status", "Finished At", "Exit Code")
	fmt.Fprintf(w, tableRowTemplate, "-------------------------------------", "-------------------------------------", "------------------------------------------------------------", "----------", "------------------------------", "----------")
	for _, ctr := range ctrs {
		fmt.Fprintf(w, tableRowTemplate, ctr.ID, ctr.Name, ctr.Image.Name, ctr.State.Status.String(), ctr.State.FinishedAt, strconv.FormatInt(ctr.State.ExitCode, 10))
	}
	fmt.Fprintln(w, "")
}
