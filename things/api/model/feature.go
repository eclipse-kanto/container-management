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

package model

// Feature represents Ditto's Feature model and all its properties
type Feature interface {
	GetID() string
	// Get the definition
	GetDefinition() []DefinitionID
	// Set the definition
	SetDefinition(definition []DefinitionID)
	// Remove the definition
	RemoveDefinition()

	// Get all properties
	GetProperties() map[string]interface{}
	// Get a specific property
	GetProperty(id string) interface{}
	// Update the properties
	SetProperties(properties map[string]interface{})
	// Update a single property
	SetProperty(id string, value interface{})
	// Remove all properties
	RemoveProperties()
	// Remove a single property
	RemoveProperty(id string)
}
