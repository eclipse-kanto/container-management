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

package handlers

import "github.com/eclipse-kanto/container-management/things/api/model"

// ThingsRegistryChangedHandler is used when a thing has been modified, created or deleted
type ThingsRegistryChangedHandler func(changedType ThingsRegistryChangedType, thing model.Thing)

// ThingRegistryHandler is used to manage the handlers for incoming thing changes and operations in the local things registy
type ThingRegistryHandler interface {
	// Set the registry's handler
	SetThingsRegistryChangedHandler(handler ThingsRegistryChangedHandler)

	// Get the registry's handler
	GetThingsRegistryChangedHandler() ThingsRegistryChangedHandler
}
