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
	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
)

// ThingOpt represents thing options
type ThingOpt func(thing *thing) error

func applyOptsThing(thing *thing, opts ...ThingOpt) error {
	for _, o := range opts {
		if err := o(thing); err != nil {
			return err
		}
	}
	return nil
}

// NewThing creates a new thing instance from the provided thing options
func NewThing(opts ...ThingOpt) model.Thing {
	thing := &thing{attributes: make(map[string]interface{}), features: make(map[string]model.Feature)}
	if err := applyOptsThing(thing, opts...); err != nil {
		return nil
	}

	return thing
}

// WithThingID sets the provided namespace and name to the thing's ID
func WithThingID(namespace string, name string) ThingOpt {
	return func(thing *thing) error {
		thing.id = namespacedID{namespace: namespace, name: name}
		return nil
	}
}

// WithThingDefinition sets the provided namespace, name and version to the thing's definition
func WithThingDefinition(namespace string, name string, version string) ThingOpt {
	return func(thing *thing) error {
		thing.definitionID = &definitionID{namespace: namespace, name: name, version: version}
		return nil
	}
}

// WithThingPolicy sets the provided namespace and name to the thing's policy ID
func WithThingPolicy(namespace string, name string) ThingOpt {
	return func(thing *thing) error {
		thing.policyID = namespacedID{namespace: namespace, name: name}
		return nil
	}
}

// WithThingAttributes sets all attributes to the thing's attributes
func WithThingAttributes(attrs map[string]interface{}) ThingOpt {
	return func(thing *thing) error {
		thing.attributes = copyMap(attrs)
		return nil
	}
}

// WithThingAttribute sets/adds an attribute to the thing's attributes
func WithThingAttribute(id string, value interface{}) ThingOpt {
	return func(thing *thing) error {
		thing.setAttribute(id, value)
		return nil
	}
}

// WithThingFeatures sets all features to the thing's features
func WithThingFeatures(features map[string]model.Feature) ThingOpt {
	return func(thing *thing) error {
		thing.features = copyFeaturesMap(features)
		return nil
	}
}

// WithThingFeature sets/adds a feature from the provided Feature to the thing's features
func WithThingFeature(id string, value model.Feature) ThingOpt {
	return func(thing *thing) error {
		thing.setFeature(id, value)
		return nil
	}
}

// WithThingFeatureFrom sets/adds a feature from the provided Feature options to the thing's features
func WithThingFeatureFrom(id string, featureOpts ...FeatureOpt) ThingOpt {
	return func(thing *thing) error {
		thing.setFeature(id, NewFeature(id, featureOpts...))
		return nil
	}
}

// WithThingDefinitionChangedHandler sets a definition changed handler to the thing's definition changed handler
func WithThingDefinitionChangedHandler(handler handlers.ThingDefinitionChangedHandler) ThingOpt {
	return func(thing *thing) error {
		thing.definitionChangedHandler = handler
		return nil
	}
}

// WithThingAttributeChangedHandler sets an attribute changed handler to the thing's attribute changed handlers
func WithThingAttributeChangedHandler(handler handlers.ThingAttributeChangedHandler) ThingOpt {
	return func(thing *thing) error {
		thing.attributeChangedHandler = handler
		return nil
	}
}

// WithThingFeatureChangedHandler sets a feature changed handler to the thing's feature changed handler
func WithThingFeatureChangedHandler(handler handlers.ThingFeatureChangedHandler) ThingOpt {
	return func(thing *thing) error {
		thing.featureChangedHandler = handler
		return nil
	}
}

// WithThingOperationsHandler sets an operations handler to the thing's operations handler
func WithThingOperationsHandler(handler handlers.ThingOperationsHandler) ThingOpt {
	return func(thing *thing) error {
		thing.operationsHandler = handler
		return nil
	}
}
