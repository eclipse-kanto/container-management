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

package things

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type state struct {
	Status     status `json:"status"`
	Pid        int64  `json:"pid,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int64  `json:"exitCode,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
	OOMKilled  bool   `json:"oomKilled,omitempty"`
}

func fromAPIContainerState(ctrState *types.State) *state {
	return &state{
		Status:     fromAPIStatus(ctrState.Status),
		Pid:        ctrState.Pid,
		Error:      ctrState.Error,
		ExitCode:   ctrState.ExitCode,
		StartedAt:  ctrState.StartedAt,
		FinishedAt: ctrState.FinishedAt,
		OOMKilled:  ctrState.OOMKilled,
	}
}

func toAPIContainerState(state *state) *types.State {
	if state == nil {
		return &types.State{
			Status: toAPIStatus(running),
		}
	}
	return &types.State{
		Status: toAPIStatus(state.Status),
	}
}
