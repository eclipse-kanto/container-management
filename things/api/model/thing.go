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

// Thing represents Ditto's Thing model
type Thing interface {
	// Get the thing's namespace
	GetNamespace() string
	// GetID the thing's id
	GetID() NamespacedID

	// Retrieves the thing's policy id
	GetPolicy() NamespacedID

	// Get the thing's definition
	GetDefinition() DefinitionID
	// Updatest he thing;s definition
	SetDefinition(DefinitionID DefinitionID) error
	// Deletes the thing's definition
	RemoveDefinition() error

	// Retrieve the thing's attributes
	GetAttributes() map[string]interface{}
	// Retrieve a thing's attribute
	GetAttribute(id string) interface{}
	// Updates the thing's attributes
	SetAttributes(attributes map[string]interface{}) error
	// Updates a thing's attribute
	SetAttribute(id string, value interface{}) error
	// Delete the thing's attributes
	RemoveAttributes() error
	// Delete a thing's attribute
	RemoveAttribute(id string) error

	// Retrieve the thing's features
	GetFeatures() map[string]Feature
	// Retrieve a thing's features
	GetFeature(id string) Feature
	// Update the thing's features
	SetFeatures(features map[string]Feature) error
	// Update a thing's feature
	SetFeature(id string, feature Feature) error
	// Remove the thing's features
	RemoveFeatures() error
	// Remove a thing's feature
	RemoveFeature(id string) error

	// Set the definition
	SetFeatureDefinition(featureID string, DefinitionID []DefinitionID) error
	// Remove the definition
	RemoveFeatureDefinition(featureID string) error
	// Update the properties
	SetFeatureProperties(featureID string, properties map[string]interface{}) error
	// Update a single property
	SetFeatureProperty(featureID string, propertyID string, value interface{}) error
	// Remove all properties
	RemoveFeatureProperties(featureID string) error
	// Remove a single property
	RemoveFeatureProperty(featureID string, propertyID string) error

	// SendMessage initiates the sending of a live outbox message from the Thing
	SendMessage(action string, value interface{}) error

	// SendFeatureMessage initiates the sending of a live outbox message from the provided feature of the Thing
	SendFeatureMessage(featureID string, action string, value interface{}) error

	// Get the thing's revision
	GetRevision() int64

	// Get the thing's last modification's timestamp
	GetLastModified() string
}
