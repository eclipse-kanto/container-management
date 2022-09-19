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

package types

// Status represents a container's status
type Status int

// constants for the supported statuses
const (
	Creating Status = iota
	Created
	Running
	Stopped
	Paused
	Exited
	Dead
	Unknown
)

func (status Status) String() string {
	return [...]string{"Creating", "Created", "Running", "Stopped", "Paused", "Exited", "Dead", "Unknown"}[status]
}

// State represents a container's state
type State struct {
	// Pid represents the container's process's PID
	Pid int64 `json:"pid"`

	// StartedAt defines the time when this container was last started
	StartedAt string `json:"started_at"`

	// Error indicates whether there was a problem that has occurred while changing the state of a container
	Error string `json:"error"`

	// ExitCode represents the last exit code of the container's internal root process
	ExitCode int64 `json:"exit_code"`

	// FinishedAt defines a timestamp of the last container's exit
	FinishedAt string `json:"finished_at"`

	// Exited defines whether the container has exited on its own for some reason - daemon reboot or internal error - distinguishes between manual stop and internal exit
	Exited bool `json:"exited"`

	// Dead identifies whether the container is dead
	Dead bool `json:"dead"`

	// Restarting identifies whether the container is currently restarting
	Restarting bool `json:"restarting"`

	// Paused indicates whether this container is paused
	Paused bool `json:"paused"`

	// Running indicates whether this container is running
	// Note: Paused and running are not mutually exclusive as pausing actually requires the process to be running - it's only 'freezed' but still running
	Running bool `json:"running"`

	// OOMKilled indicates whether the container is killed due to out of memory
	OOMKilled bool `json:"oomKilled"`

	// Status represents the status of this container
	Status Status `json:"status"`
}
