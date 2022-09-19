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

package client

import (
	"encoding/json"
	"sync"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
)

type feature struct {
	sync.RWMutex
	id         string
	definition []model.DefinitionID
	properties map[string]interface{}

	//event handlers
	definitionChangedHandler handlers.FeatureDefinitionChangedHandler
	propertyChangedHandler   handlers.FeaturePropertyChangedHandler
	operationsHandler        handlers.FeatureOperationsHandler
}

func (feature *feature) MarshalJSON() ([]byte, error) {
	defToStrings := func(defs []model.DefinitionID) []string {
		var defsAsString []string
		for _, def := range defs {
			defsAsString = append(defsAsString, def.String())
		}
		return defsAsString
	}

	return json.Marshal(&jsonFeature{
		Definition: defToStrings(feature.GetDefinition()),
		Properties: feature.GetProperties(),
	})
}
func (feature *feature) UnmarshalJSON(data []byte) error {
	var jsonFeature = &jsonFeature{}
	if err := json.Unmarshal(data, jsonFeature); err != nil {
		return err
	}

	stringsToDef := func(stringsDef []string) []model.DefinitionID {
		def := make([]model.DefinitionID, len(stringsDef))
		for _, defID := range stringsDef {
			def = append(def, NewDefinitionIDFromString(defID))
		}
		return def
	}
	feature.SetDefinition(stringsToDef(jsonFeature.Definition))
	feature.SetProperties(jsonFeature.Properties)

	return nil
}

func (feature *feature) GetID() string {
	return feature.id
}

// Get the definition
func (feature *feature) GetDefinition() []model.DefinitionID {
	feature.RLock()
	defer feature.RUnlock()

	return append([]model.DefinitionID(nil), feature.definition...)
}

// Set the definition
func (feature *feature) SetDefinition(definition []model.DefinitionID) {
	feature.Lock()
	defer feature.Unlock()

	feature.definition = append([]model.DefinitionID(nil), definition...)
}

// Remove the definition
func (feature *feature) RemoveDefinition() {
	feature.Lock()
	defer feature.Unlock()

	feature.definition = nil
}

// Get all properties
func (feature *feature) GetProperties() map[string]interface{} {
	feature.RLock()
	defer feature.RUnlock()

	propsCopy := make(map[string]interface{}, len(feature.properties))
	for id, prop := range feature.properties {
		propsCopy[id] = prop
	}
	return propsCopy
}

// Get a specific property
func (feature *feature) GetProperty(id string) interface{} {
	feature.RLock()
	defer feature.RUnlock()

	return feature.properties[id]
}

// Update the properties
func (feature *feature) SetProperties(properties map[string]interface{}) {
	feature.Lock()
	defer feature.Unlock()

	feature.properties = copyMap(properties)
}

// Update the properties
func (feature *feature) SetProperty(id string, value interface{}) {
	feature.Lock()
	defer feature.Unlock()

	feature.setProperty(id, value)
}

// Remove all properties
func (feature *feature) RemoveProperties() {
	feature.Lock()
	defer feature.Unlock()

	feature.properties = nil
}

// Remove all properties
func (feature *feature) RemoveProperty(id string) {
	feature.Lock()
	defer feature.Unlock()

	delete(feature.properties, id)
}

// Set the feature's definition changed handler
func (feature *feature) SetDefinitionChangedHandler(handler handlers.FeatureDefinitionChangedHandler) {
	feature.definitionChangedHandler = handler
}

// Get the feature's definition changed handler
func (feature *feature) GetDefinitionChangedHandler() handlers.FeatureDefinitionChangedHandler {
	return feature.definitionChangedHandler
}

// Set the feature's property changed handler
func (feature *feature) SetPropertyChangedHandler(handler handlers.FeaturePropertyChangedHandler) {
	feature.propertyChangedHandler = handler
}

// Get the feature's property changed handler
func (feature *feature) GetPropertyChangedHandler() handlers.FeaturePropertyChangedHandler {
	return feature.propertyChangedHandler
}

// Set the feature's operations handler
func (feature *feature) SetOperationsHandler(handler handlers.FeatureOperationsHandler) {
	feature.operationsHandler = handler
}

// Get the feature's operations handler
func (feature *feature) GetOperationsHandler() handlers.FeatureOperationsHandler {
	return feature.operationsHandler
}
