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

type status string

const (
	creating status = "CREATING"
	created  status = "CREATED"
	running  status = "RUNNING"
	stopped  status = "STOPPED"
	paused   status = "PAUSED"
	exited   status = "EXITED"
	dead     status = "DEAD"
	unknown  status = "UNKNOWN"
)

func toAPIStatus(internalStatus status) types.Status {
	switch internalStatus {
	case creating:
		return types.Creating
	case created:
		return types.Created
	case running:
		return types.Running
	case stopped:
		return types.Stopped
	case paused:
		return types.Paused
	case exited:
		return types.Exited
	case dead:
		return types.Dead
	default:
		return types.Unknown
	}
}

func fromAPIStatus(apiStatus types.Status) status {
	switch apiStatus {
	case types.Creating:
		return creating
	case types.Created:
		return created
	case types.Running:
		return running
	case types.Stopped:
		return stopped
	case types.Paused:
		return paused
	case types.Exited:
		return exited
	case types.Dead:
		return dead
	default:
		return unknown
	}
}
