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

package model

// Registry is the things management entry point
type Registry interface {
	// Create a thing with the provided id and namespace
	Create(id NamespacedID, thing Thing) error
	// Get a thing with the provided id and namespace
	Get(id NamespacedID) Thing
	// Modify a thing with the provided id and namespace
	Update(id NamespacedID) error
	// Remove a thing with the provided id and namespace
	Remove(id NamespacedID) error
	//add commands handler
}
