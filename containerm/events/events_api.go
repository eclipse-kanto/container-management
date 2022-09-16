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

package events

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

// ContainerEventsManager provies a simple way of publishing and subscribing to container related events.
type ContainerEventsManager interface {
	// Publish adds a new event to be dispatched based on the provided EventType and EventAction
	Publish(ctx context.Context, eventType types.EventType, eventAction types.EventAction, source *types.Container) error
	// Subscribe provides two channels where the according events and errors can be received via the subscriber context provided
	Subscribe(ctx context.Context) (<-chan *types.Event, <-chan error)
}
