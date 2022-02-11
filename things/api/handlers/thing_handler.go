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

package handlers

import "github.com/eclipse-kanto/container-management/things/api/model"

// ThingDefinitionChangedHandler is used when a things's definition has been modified, created or deleted
type ThingDefinitionChangedHandler func(changedType ChangedType, defId model.DefinitionID)

// ThingAttributeChangedHandler is used when a things's attribute has been modified, created or deleted
type ThingAttributeChangedHandler func(changedType ChangedType, attributeId string, attributeValue interface{})

// ThingFeatureChangedHandler is used when a things's feature has been modified, created or deleted
type ThingFeatureChangedHandler func(changedType ChangedType, featureId string, feature model.Feature)

// ThingOperationsHandler is used to handle incoming thing operations
type ThingOperationsHandler func(operationName string, args interface{}) (interface{}, error)

// ThingHandler is used to manage the handlers for incoming thing changes and operations
type ThingHandler interface {
	// Set the thing's definition changed handler
	SetDefinitionChangedHandler(handler ThingDefinitionChangedHandler)
	// Get the thing's definition changed handler
	GetDefinitionChangedHandler() ThingDefinitionChangedHandler

	// Set the thing's attribute changed handler
	SetAttributeChangedHandler(handler ThingAttributeChangedHandler)
	// Get the thing's attribute changed handler
	GetAttributeChangedHandler() ThingAttributeChangedHandler

	// Set the thing's feature changed handler
	SetFeatureChangedHandler(handler ThingFeatureChangedHandler)
	// Get the thing's feature changed handler
	GetFeatureChangedHandler() ThingFeatureChangedHandler

	// Set the thing's operations handler
	SetOperationsHandler(handler ThingOperationsHandler)
	// Get the thing's operations handler
	GetOperationsHandler() ThingOperationsHandler
}
