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

package client

import (
	"encoding/json"
	"sync"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
)

type thing struct {
	sync.RWMutex
	id           model.NamespacedID
	policyID     model.NamespacedID
	definitionID model.DefinitionID
	attributes   map[string]interface{}
	features     map[string]model.Feature
	revision     int64
	timestamp    string

	hubDevice Device

	//event handlers
	definitionChangedHandler handlers.ThingDefinitionChangedHandler
	attributeChangedHandler  handlers.ThingAttributeChangedHandler
	featureChangedHandler    handlers.ThingFeatureChangedHandler
	operationsHandler        handlers.ThingOperationsHandler
}

func (thing *thing) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonThing{
		ThingID:    thing.id.String(),
		Definition: thing.definitionID.String(),
		PolicyID:   thing.policyID.String(),
		Attributes: thing.GetAttributes(),
		Features:   convertFromFeaturesMap(thing.GetFeatures()),
	})
}

func (thing *thing) UnmarshalJSON(data []byte) error {
	var jsonThing = &jsonThing{}
	if err := json.Unmarshal(data, jsonThing); err != nil {
		return err
	}
	thing.id = NewNamespacedIDFromString(jsonThing.ThingID)
	if jsonThing.Definition != "" {
		thing.definitionID = NewDefinitionIDFromString(jsonThing.Definition)
	}
	if jsonThing.PolicyID != "" {
		thing.policyID = NewNamespacedIDFromString(jsonThing.PolicyID)
	}
	thing.setAttributesSafe(jsonThing.Attributes)

	if jsonThing.Features != nil {
		for featID, feat := range jsonThing.Features {
			thing.setFeatureSafe(featID, NewFeature(featID, WithFeatureProperties(feat.Properties), WithFeatureDefinitionFromString(feat.Definition...)))
		}
	}
	return nil
}

// GetNamespace retrieves the thing's namespace
func (thing *thing) GetNamespace() string {
	return thing.id.GetNamespace()
}

// GetID retrieves the thing's ID
func (thing *thing) GetID() model.NamespacedID {
	return thing.id
}

// GetPolicy retrieves the thing's policy ID
func (thing *thing) GetPolicy() model.NamespacedID {
	return thing.policyID
}

// GetDefinition retrieves the thing's definition
func (thing *thing) GetDefinition() model.DefinitionID {
	return thing.definitionID
}

// SetDefinition updates the thing's definition
func (thing *thing) SetDefinition(defID model.DefinitionID) error {
	if err := thing.sendModifyDefinition(defID); err != nil {
		return err
	}
	thing.definitionID = defID
	return nil
}

// RemoveDefinition deletes the thing's definition
func (thing *thing) RemoveDefinition() error {
	if err := thing.sendDeleteDefinition(); err != nil {
		return err
	}
	thing.definitionID = nil
	return nil
}

// GetAttributes retrieves the thing's attributes
func (thing *thing) GetAttributes() map[string]interface{} {
	thing.RLock()
	defer thing.RUnlock()

	attrsCopy := make(map[string]interface{}, len(thing.attributes))
	for id, attr := range thing.attributes {
		attrsCopy[id] = attr
	}
	return attrsCopy
}

// GetAttribute retrieves a thing's attribute
func (thing *thing) GetAttribute(id string) interface{} {
	thing.RLock()
	defer thing.RUnlock()

	return thing.attributes[id]
}

// SetAttributes updates the thing's attributes
func (thing *thing) SetAttributes(attributes map[string]interface{}) error {
	if err := thing.sendModifyAttributes(attributes); err != nil {
		return err
	}

	thing.setAttributesSafe(attributes)
	return nil
}

// SetAttribute updates a thing's attribute
func (thing *thing) SetAttribute(id string, value interface{}) error {
	if err := thing.sendModifyAttribute(id, value); err != nil {
		return err
	}
	thing.setAttributeSafe(id, value)
	return nil
}

// RemoveAttributes deletes the thing's attributes
func (thing *thing) RemoveAttributes() error {
	if err := thing.sendDeleteAttributes(); err != nil {
		return err
	}
	thing.removeAttributesSafe()
	return nil
}

// RemoveAttribute deletes a thing's attribute
func (thing *thing) RemoveAttribute(id string) error {
	if err := thing.sendDeleteAttribute(id); err != nil {
		return err
	}
	thing.removeAttributeSafe(id)
	return nil
}

// GetFeatures retrieve the thing's features
func (thing *thing) GetFeatures() map[string]model.Feature {
	thing.RLock()
	defer thing.RUnlock()

	featuresCopy := make(map[string]model.Feature, len(thing.features))
	for id, feature := range thing.features {
		featuresCopy[id] = feature
	}
	return featuresCopy
}

// GetFeature retrieves a thing's features
func (thing *thing) GetFeature(id string) model.Feature {
	thing.RLock()
	defer thing.RUnlock()
	if thing.features == nil {
		return nil
	}
	return thing.features[id]
}

// SetFeatures updates the thing's features
func (thing *thing) SetFeatures(features map[string]model.Feature) error {
	if err := thing.sendModifyFeatures(features); err != nil {
		return err
	}
	thing.setFeaturesSafe(features)
	return nil
}

// SetFeature updates a thing's feature
func (thing *thing) SetFeature(id string, feature model.Feature) error {
	if err := thing.sendModifyFeature(id, feature); err != nil {
		return err
	}
	thing.setFeatureSafe(id, feature)
	return nil
}

// RemoveFeatures deletes the thing's features
func (thing *thing) RemoveFeatures() error {
	if err := thing.sendDeleteFeatures(); err != nil {
		return err
	}
	thing.removeFeaturesSafe()
	return nil
}

// RemoveFeature deletes a thing's feature
func (thing *thing) RemoveFeature(id string) error {
	if err := thing.sendDeleteFeature(id); err != nil {
		return err
	}

	thing.removeFeatureSafe(id)
	return nil
}

// SetFeatureDefinition updates the definition
func (thing *thing) SetFeatureDefinition(featureID string, definition []model.DefinitionID) error {
	if err := thing.sendModifyFeatureDefinition(featureID, definition); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].SetDefinition(definition)
	}
	return nil
}

// RemoveFeatureDefinition deletes the definition
func (thing *thing) RemoveFeatureDefinition(featureID string) error {
	if err := thing.sendDeleteFeatureDefinition(featureID); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].RemoveDefinition()
	}
	return nil
}

// SetFeatureProperties updates all properties
func (thing *thing) SetFeatureProperties(featureID string, properties map[string]interface{}) error {
	if err := thing.sendModifyFeatureProperties(featureID, properties); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].SetProperties(properties)
	}
	return nil
}

// SetFeatureProperty updates a single property
func (thing *thing) SetFeatureProperty(featureID string, propertyID string, value interface{}) error {
	if err := thing.sendModifyFeatureProperty(featureID, propertyID, value); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].SetProperty(propertyID, value)
	}
	return nil
}

// RemoveFeatureProperties removes all properties
func (thing *thing) RemoveFeatureProperties(featureID string) error {
	if err := thing.sendDeleteFeatureProperties(featureID); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].RemoveProperties()
	}
	return nil
}

// RemoveFeatureProperty removes a single property
func (thing *thing) RemoveFeatureProperty(featureID string, propertyID string) error {
	if err := thing.sendDeleteFeatureProperty(featureID, propertyID); err != nil {
		return err
	}

	thing.RLock()
	defer thing.RUnlock()
	if thing.features != nil && thing.features[featureID] != nil {
		thing.features[featureID].RemoveProperty(propertyID)
	}
	return nil
}

// SendMessage initiates the sending of a live outbox message from the Thing
func (thing *thing) SendMessage(action string, value interface{}) error {
	return thing.sendOutboxMessageForThing(action, value)
}

// SendFeatureMessage initiates the sending of a live outbox message from the provided feature of the Thing
func (thing *thing) SendFeatureMessage(featureID string, action string, value interface{}) error {
	return thing.sendOutboxMessageForFeature(featureID, action, value)
}

// GetRevision retrieves the thing's revision
func (thing *thing) GetRevision() int64 {
	return thing.revision
}

// GetLastModified retrieves the thing's last modification's timestamp
func (thing *thing) GetLastModified() string {
	return thing.timestamp
}

// SetDefinitionChangedHandler updates the thing's definition changed handler
func (thing *thing) SetDefinitionChangedHandler(handler handlers.ThingDefinitionChangedHandler) {
	thing.definitionChangedHandler = handler
}

// GetDefinitionChangedHandler retrieves the thing's definition changed handler
func (thing *thing) GetDefinitionChangedHandler() handlers.ThingDefinitionChangedHandler {
	return thing.definitionChangedHandler
}

// SetAttributeChangedHandler updates the thing's attribute changed handler
func (thing *thing) SetAttributeChangedHandler(handler handlers.ThingAttributeChangedHandler) {
	thing.attributeChangedHandler = handler
}

// GetAttributeChangedHandler retrieves the thing's attribute changed handler
func (thing *thing) GetAttributeChangedHandler() handlers.ThingAttributeChangedHandler {
	return thing.attributeChangedHandler
}

// SetFeatureChangedHandler updates the thing's feature changed handler
func (thing *thing) SetFeatureChangedHandler(handler handlers.ThingFeatureChangedHandler) {
	thing.featureChangedHandler = handler
}

// GetFeatureChangedHandler retrieves the thing's feature changed handler
func (thing *thing) GetFeatureChangedHandler() handlers.ThingFeatureChangedHandler {
	return thing.featureChangedHandler
}

// SetOperationsHandler updates the thing's operations handler
func (thing *thing) SetOperationsHandler(handler handlers.ThingOperationsHandler) {
	thing.operationsHandler = handler
}

// GetOperationsHandler retrieves the thing's operations handler
func (thing *thing) GetOperationsHandler() handlers.ThingOperationsHandler {
	return thing.operationsHandler
}
