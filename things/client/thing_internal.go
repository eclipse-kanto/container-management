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
	"github.com/eclipse-kanto/container-management/things/api/model"
)

func (thing *thing) sendModifyDefinition(defID model.DefinitionID) error {
	msg := generateCommandModifyThingDefinition(thing, defID, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteDefinition() error {
	msg := generateCommandDeleteThingDefinition(thing, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendModifyAttributes(attributes map[string]interface{}) error {
	msg := generateCommandModifyThingAttributes(thing, attributes, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}
func (thing *thing) sendDeleteAttributes() error {
	msg := generateCommandDeleteThingAttributes(thing, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}
func (thing *thing) sendModifyAttribute(id string, value interface{}) error {
	msg := generateCommandModifyThingAttributeSingle(thing, id, value, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}
func (thing *thing) sendDeleteAttribute(id string) error {
	msg := generateCommandDeleteThingAttributeSingle(thing, id, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendModifyFeatures(features map[string]model.Feature) error {
	msg := generateCommandModifyThingFeatures(thing, features, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}
func (thing *thing) sendModifyFeature(id string, feature model.Feature) error {
	msg := generateCommandModifyThingFeatureSingle(thing, id, feature, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteFeatures() error {
	msg := generateCommandDeleteThingFeatures(thing, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteFeature(id string) error {
	msg := generateCommandDeleteThingFeatureSingle(thing, id, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendModifyFeatureDefinition(featureID string, definition []model.DefinitionID) error {
	msg := generateCommandModifyThingFeatureDefinitionSingle(thing, featureID, definition, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteFeatureDefinition(featureID string) error {
	msg := generateCommandDeleteThingFeatureDefinitionSingle(thing, featureID, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendModifyFeatureProperties(featureID string, props map[string]interface{}) error {
	msg := generateCommandModifyThingFeatureProperties(thing, featureID, props, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendModifyFeatureProperty(featureID string, propertyID string, value interface{}) error {
	msg := generateCommandModifyThingFeaturePropertySingle(thing, featureID, propertyID, value, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteFeatureProperties(featureID string) error {
	msg := generateCommandDeleteThingFeatureProperties(thing, featureID, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendDeleteFeatureProperty(featureID string, propertyID string) error {
	msg := generateCommandDeleteThingFeaturePropertySingle(thing, featureID, propertyID, defaultHeaders...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendOutboxMessageForThing(action string, value interface{}) error {
	headerOpts := append([]HeaderOpt{}, defaultHeaders...)
	if value != nil {
		headerOpts = append(headerOpts, WithContentType(jsonContent))
	}
	msg := generateOutboxMessageFromThing(thing, action, value, headerOpts...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) sendOutboxMessageForFeature(featureID string, action string, value interface{}) error {
	headerOpts := append([]HeaderOpt{}, defaultHeaders...)
	if value != nil {
		headerOpts = append(headerOpts, WithContentType(jsonContent))
	}
	msg := generateOutboxMessageFromThingFeature(thing, featureID, action, value, headerOpts...)
	return thing.hubDevice.GetConnection().SendEvent(msg)
}

func (thing *thing) removeFeaturesSafe() map[string]model.Feature {
	thing.Lock()
	defer func() {
		thing.features = nil
		thing.Unlock()
	}()
	return thing.features
}

func (thing *thing) removeFeatureSafe(id string) model.Feature {
	thing.Lock()
	defer func() {
		delete(thing.features, id)
		thing.Unlock()
	}()
	return thing.features[id]
}

func (thing *thing) setFeaturesSafe(features map[string]model.Feature) {
	thing.Lock()
	defer thing.Unlock()

	thing.features = copyFeaturesMap(features)
}

func (thing *thing) setFeature(id string, feature model.Feature) {
	if thing.features == nil {
		thing.features = map[string]model.Feature{}
	}
	thing.features[id] = feature
}

func (thing *thing) setFeatureSafe(id string, feature model.Feature) {
	thing.Lock()
	defer thing.Unlock()

	thing.setFeature(id, feature)
}

func (thing *thing) removeAttributesSafe() map[string]interface{} {
	thing.Lock()
	defer func() {
		thing.attributes = nil
		thing.Unlock()
	}()
	return thing.attributes
}

func (thing *thing) setAttributesSafe(attributes map[string]interface{}) {
	thing.Lock()
	defer thing.Unlock()

	thing.attributes = copyMap(attributes)
}

func (thing *thing) setAttribute(id string, value interface{}) {
	if thing.attributes == nil {
		thing.attributes = make(map[string]interface{})
	}
	thing.attributes[id] = value
}

func (thing *thing) setAttributeSafe(id string, value interface{}) {
	thing.Lock()
	defer thing.Unlock()

	thing.setAttribute(id, value)
}

func (thing *thing) removeAttributeSafe(id string) interface{} {
	thing.Lock()
	defer func() {
		delete(thing.attributes, id)
		thing.Unlock()
	}()
	return thing.attributes[id]
}
