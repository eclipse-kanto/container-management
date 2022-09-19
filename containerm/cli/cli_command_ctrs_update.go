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
	"math"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/util"
	utilcli "github.com/eclipse-kanto/container-management/containerm/util/cli"
	"github.com/spf13/cobra"
)

type updateCmd struct {
	baseCommand
	config updateConfig
}

type updateConfig struct {
	// container name
	name string

	restartPolicy
	resources
}

func (cc *updateCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "update <container-id>",
		Short: "Update a container.",
		Long:  "Update a container without recreating it. The provided configurations will be merged with the current one.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "update <container-id>\nupdate -n <container-name>\n",
	}
	cc.setupFlags()
}

func (cc *updateCmd) run(args []string) error {
	var (
		container *types.Container
		err       error
		ctx       = context.Background()
	)

	// parse parameters
	if container, err = utilcli.ValidateContainerByNameArgsSingle(ctx, args, cc.config.name, cc.cli.gwManClient); err != nil {
		return err
	}

	updateOpts := &types.UpdateOpts{}

	updateOpts.RestartPolicy = cc.updatedRestartPolicy(container.HostConfig.RestartPolicy)
	if err = util.ValidateRestartPolicy(updateOpts.RestartPolicy); err != nil {
		return err
	}

	updateOpts.Resources = cc.updatedResources(container.HostConfig.Resources)
	if err = util.ValidateResources(updateOpts.Resources); err != nil {
		return err
	}

	return cc.cli.gwManClient.Update(ctx, container.ID, updateOpts)
}

func (cc *updateCmd) updatedResources(resources *types.Resources) *types.Resources {
	if cc.config.resources.memory == "" && cc.config.resources.memoryReservation == "" && cc.config.resources.memorySwap == "" {
		// nothing to update
		return nil
	}

	if resources == nil {
		return getResourceLimits(cc.config.resources)
	}

	get := func(newValue string, defaultValue string) string {
		if newValue == "" {
			return defaultValue
		}
		if newValue == types.MemoryUnlimited { // update to unlimited
			return ""
		}
		return newValue
	}
	return &types.Resources{
		Memory:            get(cc.config.resources.memory, resources.Memory),
		MemoryReservation: get(cc.config.resources.memoryReservation, resources.MemoryReservation),
		MemorySwap:        get(cc.config.resources.memorySwap, resources.MemorySwap),
	}
}

func (cc *updateCmd) updatedRestartPolicy(restartPolicy *types.RestartPolicy) *types.RestartPolicy {
	if cc.config.restartPolicy.kind == "" &&
		cc.config.restartPolicy.timeout == math.MinInt64 &&
		cc.config.restartPolicy.maxRetryCount == math.MinInt32 {
		// nothing to update
		return nil
	}

	newRestartPolicy := &types.RestartPolicy{}
	if cc.config.restartPolicy.kind == "" {
		newRestartPolicy.Type = restartPolicy.Type
	} else {
		newRestartPolicy.Type = types.PolicyType(cc.config.restartPolicy.kind)
	}

	// load current values
	if newRestartPolicy.Type == types.OnFailure {
		newRestartPolicy.RetryTimeout = restartPolicy.RetryTimeout
		newRestartPolicy.MaximumRetryCount = restartPolicy.MaximumRetryCount
	}

	if cc.config.restartPolicy.timeout != math.MinInt64 {
		newRestartPolicy.RetryTimeout = time.Duration(cc.config.restartPolicy.timeout) * time.Second
	}

	if cc.config.restartPolicy.maxRetryCount != math.MinInt32 {
		newRestartPolicy.MaximumRetryCount = cc.config.restartPolicy.maxRetryCount
	}

	return newRestartPolicy
}

func (cc *updateCmd) setupFlags() {
	flagSet := cc.cmd.Flags()

	// name
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Updates the container with a specific name.")

	// restart policy
	flagSet.StringVar(&cc.config.restartPolicy.kind, "rp", "",
		"Updates the restart policy for the container. The policy will be applied when the container exits. Supported restart policies are - no, always, unless-stopped, on-failure. \n"+
			"no - no attempts to restart the container for any reason will be made \n"+
			"always - an attempt to restart the container will be made each time the container exits regardless of the exit code \n"+
			"unless-stopped - restart attempts will be made only if the container has not been stopped by the user \n"+
			"on-failure - restart attempts will be made if the container exits with an exit code != 0; \n"+
			"the additional flags (--rp-cnt and --rp-to) apply only for this policy; if max retry count is not provided - the system will retry until it succeeds endlessly \n")
	flagSet.IntVar(&cc.config.restartPolicy.maxRetryCount, "rp-cnt", math.MinInt32, "Updates the number of retries that will be made to restart the container on exit if the policy is on-failure")
	flagSet.Int64Var(&cc.config.restartPolicy.timeout, "rp-to", math.MinInt64, "Updates the time out period in seconds for each retry that will be made to restart the container on exit if the policy is set to on-failure")
	flagSet.StringVarP(&cc.config.resources.memory, "memory", "m", "", "Updates the max amount of memory the container can use in the form of 200m, 1.2g.\n"+
		"Use -1, to remove the memory usage limit.")
	flagSet.StringVar(&cc.config.resources.memoryReservation, "memory-reservation", "", "Updates the soft memory limitation in the form of 200m, 1.2g.\n"+
		"Use -1, to remove the reservation memory limit.")
	flagSet.StringVar(&cc.config.resources.memorySwap, "memory-swap", "", "Updates the total amount of memory + swap that the container can use in the form of 200m, 1.2g.\n"+
		"Use -1, to remove the swap memory limit.")
}
