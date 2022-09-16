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

// EventType represents the event's type
type EventType string

const (
	// EventTypeContainers is an event type for the containers
	EventTypeContainers EventType = "containers"
	// in the future more types will be added - e.g. for image changes, etc.
)

// EventAction represents the event's action
type EventAction string

const (
	// EventActionContainersCreated is used when a container is created
	EventActionContainersCreated EventAction = "created"
	// EventActionContainersRunning is used when a container is running
	EventActionContainersRunning EventAction = "running"
	// EventActionContainersPaused is used when a container is paused
	EventActionContainersPaused EventAction = "paused"
	// EventActionContainersResumed is used when a container is resumed
	EventActionContainersResumed EventAction = "resumed"
	// EventActionContainersStopped is used when a container is stopped
	EventActionContainersStopped EventAction = "stopped"
	// EventActionContainersExited is used when a container is exited
	EventActionContainersExited EventAction = "exited"
	// EventActionContainersRemoved is used when a container is removed
	EventActionContainersRemoved EventAction = "removed"
	// EventActionContainersRenamed is used when a container is renamed
	EventActionContainersRenamed EventAction = "renamed"
	// EventActionContainersUpdated is used when a container is updated
	EventActionContainersUpdated EventAction = "updated"
	// EventActionContainersUnknown is used when an unknown action has been performed
	EventActionContainersUnknown EventAction = "unknown"
)

// Event represents an emitted event
type Event struct {
	// the EventType
	Type EventType `json:"type"`
	// the EventAction
	Action EventAction `json:"action"`
	// the container instance that changed
	Source Container `json:"source,omitempty"`
	// time
	Time int64 `json:"time,omitempty"`
}
