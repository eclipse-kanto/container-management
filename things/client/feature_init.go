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
	"strings"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
)

// FeatureOpt represents feature options
type FeatureOpt func(feature *feature) error

func applyOptsFeature(feature *feature, opts ...FeatureOpt) error {
	for _, o := range opts {
		if err := o(feature); err != nil {
			return err
		}
	}
	return nil
}

// NewFeature creates a new feture from the provided name and feture options
func NewFeature(name string, opts ...FeatureOpt) model.Feature {
	feature := &feature{}
	feature.id = name
	if err := applyOptsFeature(feature, opts...); err != nil {
		return nil
	}
	return feature
}

// WithFeatureDefinitionFromString sets the feture definition from the provided string
func WithFeatureDefinitionFromString(definitions ...string) FeatureOpt {
	return func(feature *feature) error {

		definition := make([]model.DefinitionID, len(definitions))
		for i, def := range definitions {
			elements := strings.Split(def, ":")
			definition[i] = definitionID{namespace: elements[0], name: elements[1], version: elements[2]}
		}

		feature.definition = definition

		return nil
	}
}

// WithFeatureDefinition sets the feature definition from the provided definition id
func WithFeatureDefinition(definitions ...model.DefinitionID) FeatureOpt {
	return func(feature *feature) error {
		feature.definition = append([]model.DefinitionID(nil), definitions...)

		return nil
	}
}

// WithFeatureProperties sets the feature properties from the provided properties
func WithFeatureProperties(properties map[string]interface{}) FeatureOpt {
	return func(feature *feature) error {
		feature.properties = copyMap(properties)
		return nil
	}
}

// WithFeatureProperty sets the feture property from the provided property id and property
func WithFeatureProperty(id string, value interface{}) FeatureOpt {
	return func(feature *feature) error {
		feature.setProperty(id, value)
		return nil
	}
}

// WithFeatureDefinitionChangedHandler sets the feature definition changed handler
func WithFeatureDefinitionChangedHandler(handler handlers.FeatureDefinitionChangedHandler) FeatureOpt {
	return func(feature *feature) error {
		feature.definitionChangedHandler = handler
		return nil
	}
}

// WithFeaturePropertyChangedHandler sets the feature property changed handler
func WithFeaturePropertyChangedHandler(handler handlers.FeaturePropertyChangedHandler) FeatureOpt {
	return func(feature *feature) error {
		feature.propertyChangedHandler = handler
		return nil
	}
}

// WithFeatureOperationsHandler sets the feature operations handler
func WithFeatureOperationsHandler(handler handlers.FeatureOperationsHandler) FeatureOpt {
	return func(feature *feature) error {
		feature.operationsHandler = handler
		return nil
	}
}
