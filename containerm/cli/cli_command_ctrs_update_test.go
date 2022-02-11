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
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	updateCmdFlagName                       = "name"
	updateCmdFlagRestartPolicy              = "rp"
	updateCmdFlagRestartPolicyTimeout       = "rp-to"
	updateCmdFlagRestartPolicyMaxRetryCount = "rp-cnt"
	updateCmdFlagMemory                     = "memory"
	updateCmdFlagMemoryReservation          = "memory-reservation"
	updateCmdFlagMemorySwap                 = "memory-swap"

	// test input constants
	updateContainerID   = "test-ctr"
	updateContainerName = "test-ctr-name"

	updatedRestartPolicyTimeout       = 10 * time.Second
	updatedRestartPolicyMaxRetryCount = 10
	invalidPolicyType                 = types.PolicyType("invalid")
	updatedMemory                     = "500M"
	updatedSwapLimit                  = "1G"
	invalidMemory                     = "1X"
)

var (
	// command args ---------------
	updateCmdArgs = []string{updateContainerID}

	onFailureRestartPolicy = &types.RestartPolicy{
		Type:              types.OnFailure,
		RetryTimeout:      30 * time.Second,
		MaximumRetryCount: 1,
	}

	testCtr = &types.Container{
		ID:   updateContainerID,
		Name: updateContainerName,
		HostConfig: &types.HostConfig{
			RestartPolicy: &types.RestartPolicy{
				Type: types.UnlessStopped,
			},
		},
	}

	testCtr2 = &types.Container{
		ID:   updateContainerID,
		Name: updateContainerName,
		HostConfig: &types.HostConfig{
			RestartPolicy: onFailureRestartPolicy,
			Resources: &types.Resources{
				Memory:            "300M",
				MemoryReservation: "200M",
				MemorySwap:        "500M",
			},
		},
	}
)

// Tests ------------------------------
func TestUpdateCmdInit(t *testing.T) {
	updateCliTest := &updateCommandTest{}
	updateCliTest.init()

	execTestInit(t, updateCliTest)
}

func TestUpdateCmdFlags(t *testing.T) {
	updateCliTest := &updateCommandTest{}
	updateCliTest.init()

	expectedCfg := updateConfig{
		name: updateContainerName,
		restartPolicy: restartPolicy{
			kind:          string(types.OnFailure),
			timeout:       10000,
			maxRetryCount: 3,
		},
		resources: resources{
			memory:            "2G",
			memoryReservation: "1.5G",
			memorySwap:        "4G",
		},
	}

	flagsToApply := map[string]string{
		updateCmdFlagName:                       expectedCfg.name,
		updateCmdFlagRestartPolicy:              expectedCfg.restartPolicy.kind,
		updateCmdFlagRestartPolicyTimeout:       strconv.FormatInt(expectedCfg.restartPolicy.timeout, 10),
		updateCmdFlagRestartPolicyMaxRetryCount: strconv.Itoa(expectedCfg.restartPolicy.maxRetryCount),
		updateCmdFlagMemory:                     expectedCfg.resources.memory,
		updateCmdFlagMemoryReservation:          expectedCfg.resources.memoryReservation,
		updateCmdFlagMemorySwap:                 expectedCfg.resources.memorySwap,
	}

	execTestSetupFlags(t, updateCliTest, flagsToApply, expectedCfg)
}

func TestUpdateCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	updateCliTest := &updateCommandTest{}
	updateCliTest.initWithCtrl(controller)

	execTestsRun(t, updateCliTest)
}

// EOF Tests --------------------------

type updateCommandTest struct {
	cliCommandTestBase
	updateCmd *updateCmd
}

func (updateTc *updateCommandTest) commandConfig() interface{} {
	return updateTc.updateCmd.config
}

func (updateTc *updateCommandTest) commandConfigDefault() interface{} {
	return updateConfig{
		restartPolicy: restartPolicy{
			timeout:       math.MinInt64,
			maxRetryCount: math.MinInt32,
		},
	}
}

func (updateTc *updateCommandTest) prepareCommand(flagsCfg map[string]string) {
	// setup command to test
	cmd := &updateCmd{}
	updateTc.updateCmd, updateTc.baseCmd = cmd, cmd

	updateTc.updateCmd.init(updateTc.mockRootCommand)
	// setup command flags
	setCmdFlags(flagsCfg, updateTc.updateCmd.cmd)
}

func (updateTc *updateCommandTest) runCommand(args []string) error {
	return updateTc.updateCmd.run(args)
}

func (updateTc *updateCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		// Test name
		"test_update_id_and_name_provided": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagName: updateContainerName,
			},
			mockExecution: updateTc.mockExecUpdateIDAndName,
		},

		// Test restart policy
		"test_update_restart_policy_to_on_failure": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagRestartPolicy:              string(types.OnFailure),
				updateCmdFlagRestartPolicyTimeout:       fmt.Sprintf("%.0f", onFailureRestartPolicy.RetryTimeout.Seconds()),
				updateCmdFlagRestartPolicyMaxRetryCount: strconv.Itoa(onFailureRestartPolicy.MaximumRetryCount),
			},
			mockExecution: updateTc.mockExecUpdateRestartPolicyToOnFailure,
		},
		"test_update_restart_policy_timeout": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagRestartPolicyTimeout: fmt.Sprintf("%.0f", updatedRestartPolicyTimeout.Seconds()),
			},
			mockExecution: updateTc.mockExecUpdateRestartPolicyTimeout,
		},
		"test_update_restart_policy_max_retry_count": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagRestartPolicyMaxRetryCount: strconv.Itoa(updatedRestartPolicyMaxRetryCount),
			},
			mockExecution: updateTc.mockExecUpdateRestartPolicyMaxRetryCount,
		},
		"test_update_restart_policy_error": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagRestartPolicy: string(invalidPolicyType),
			},
			mockExecution: updateTc.mockExecUpdateRestartPolicyError,
		},
		// Test memory
		"test_update_memory": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagMemory: updatedMemory,
			},
			mockExecution: updateTc.mockExecUpdateMemory,
		},
		"test_update_swap_limit": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagMemorySwap: updatedSwapLimit,
			},
			mockExecution: updateTc.mockExecUpdateSwapLimit,
		},
		"test_update_memory_no_limits": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagMemory:            types.MemoryUnlimited,
				updateCmdFlagMemoryReservation: types.MemoryUnlimited,
				updateCmdFlagMemorySwap:        types.MemoryUnlimited,
			},
			mockExecution: updateTc.mockExecUpdateMemoryNoLimit,
		},
		"test_update_memory_error": {
			args: updateCmdArgs,
			flags: map[string]string{
				updateCmdFlagMemory: invalidMemory,
			},
			mockExecution: updateTc.mockExecUpdateMemoryError,
		},
	}
}

// Mocked executions---------------------------------------------------------------------------------
func (updateTc *updateCommandTest) mockExecUpdateIDAndName(args []string) error {
	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(0)
	return log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
}

func (updateTc *updateCommandTest) mockExecUpdateRestartPolicyToOnFailure(args []string) error {
	opts := &types.UpdateOpts{
		RestartPolicy: onFailureRestartPolicy,
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateRestartPolicyTimeout(args []string) error {
	opts := &types.UpdateOpts{
		RestartPolicy: &types.RestartPolicy{
			Type:              onFailureRestartPolicy.Type,
			RetryTimeout:      updatedRestartPolicyTimeout,
			MaximumRetryCount: onFailureRestartPolicy.MaximumRetryCount,
		},
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr2, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr2.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateRestartPolicyMaxRetryCount(args []string) error {
	opts := &types.UpdateOpts{
		RestartPolicy: &types.RestartPolicy{
			Type:              onFailureRestartPolicy.Type,
			RetryTimeout:      onFailureRestartPolicy.RetryTimeout,
			MaximumRetryCount: updatedRestartPolicyMaxRetryCount,
		},
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr2, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr2.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateRestartPolicyError(args []string) error {
	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("unsupported restart policy type %s", invalidPolicyType)
}

func (updateTc *updateCommandTest) mockExecUpdateMemory(args []string) error {
	opts := &types.UpdateOpts{
		Resources: &types.Resources{
			Memory: updatedMemory,
		},
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateSwapLimit(args []string) error {
	opts := &types.UpdateOpts{
		Resources: &types.Resources{
			Memory:            testCtr2.HostConfig.Resources.Memory,
			MemoryReservation: testCtr2.HostConfig.Resources.MemoryReservation,
			MemorySwap:        updatedSwapLimit,
		},
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr2, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr2.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateMemoryNoLimit(args []string) error {
	opts := &types.UpdateOpts{
		Resources: &types.Resources{},
	}

	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr2, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), testCtr2.ID, opts).Times(1)
	return nil
}

func (updateTc *updateCommandTest) mockExecUpdateMemoryError(args []string) error {
	updateTc.mockClient.EXPECT().Get(context.Background(), args[0]).Times(1).Return(testCtr, nil)
	updateTc.mockClient.EXPECT().Update(context.Background(), gomock.Any(), gomock.Any()).Times(0)
	return log.NewErrorf("invalid format of memory - %s", invalidMemory)
}
