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

// ChangedType represents the type of change that took place in the digital twin when emmiting an event
type ChangedType int

const (
	// Created is used to mark the creation of an new thing in an event
	Created ChangedType = iota
	// Modified is used in an event ater an existing thing has been modified
	Modified
	// Deleted is used in an event ater an existing thing has been deleted
	Deleted
)

// ThingsRegistryChangedType represents the type of change that took place in the local things registry
type ThingsRegistryChangedType int

const (
	// Added is used to mark the addition of a new thing to the local things registry
	Added ThingsRegistryChangedType = iota
	// Updated is used when the data for an existing thing has been updated in the local things registry
	Updated
	// Removed is used then a thing has been removed from the local things registry
	Removed
)
