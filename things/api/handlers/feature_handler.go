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

// FeatureDefinitionChangedHandler is used when a feature's definition has been modified, created or deleted
type FeatureDefinitionChangedHandler func(changedType ChangedType, defId []model.DefinitionID)

// FeaturePropertyChangedHandler is used when a feature's property has been modified, created or deleted
type FeaturePropertyChangedHandler func(changedType ChangedType, propertyId string, propertyValue interface{})

// FeatureOperationsHandler is used to handle incoming feature operations
type FeatureOperationsHandler func(operationName string, args interface{}) (interface{}, error)

// FeatureHandler is used to manage incoming feature changes and operations
type FeatureHandler interface {
	// Set the feature's definition changed handler
	SetDefinitionChangedHandler(handler FeatureDefinitionChangedHandler)
	// Get the feature's definition changed handler
	GetDefinitionChangedHandler() FeatureDefinitionChangedHandler

	// Set the feature's property changed handler
	SetPropertyChangedHandler(handler FeaturePropertyChangedHandler)
	// Get the feature's property changed handler
	GetPropertyChangedHandler() FeaturePropertyChangedHandler

	// Set the feature's operations handler
	SetOperationsHandler(handler FeatureOperationsHandler)
	// Get the feature's operations handler
	GetOperationsHandler() FeatureOperationsHandler
}
