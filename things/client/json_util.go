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

	"github.com/eclipse-kanto/container-management/things/api/model"
)

type jsonFeature struct {
	Definition []string               `json:"definition,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type jsonThing struct {
	ThingID    string                 `json:"thingId"`
	Definition string                 `json:"definition,omitempty"`
	PolicyID   string                 `json:"policyId,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Features   map[string]jsonFeature `json:"features,omitempty"`
}

type jsonLocalClientConfig struct {
	DeviceID string `json:"deviceId"`
	TenantID string `json:"tenantId"`
	PolicyID string `json:"policyId"`
}

type jsonEnvelope struct {
	Topic     string                 `json:"topic"`
	Headers   map[string]interface{} `json:"headers"`
	Path      string                 `json:"path"`
	Value     interface{}            `json:"value,omitempty"`
	Status    int                    `json:"status,omitempty"`
	Revision  int                    `json:"revision,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

func marshalToThing(value interface{}) (*thing, error) {
	th := &jsonThing{}
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, th); err != nil {
		return nil, err
	}
	var features map[string]model.Feature

	if th.Features != nil {
		features = make(map[string]model.Feature)
		for key, val := range th.Features {
			features[key] = NewFeature(key, WithFeatureProperties(val.Properties), WithFeatureDefinitionFromString(val.Definition...))
		}
	}
	var (
		policyID     model.NamespacedID
		definitionID model.DefinitionID
	)
	if th.PolicyID != "" {
		policyID = NewNamespacedIDFromString(th.PolicyID)
	}
	if th.Definition != "" {
		definitionID = NewDefinitionIDFromString(th.Definition)
	}

	return &thing{
		id:           NewNamespacedIDFromString(th.ThingID),
		policyID:     policyID,
		definitionID: definitionID,
		attributes:   th.Attributes,
		features:     features,
	}, nil
}

func marshalToDefinitionID(value interface{}) (model.DefinitionID, error) {
	var res string
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, &res); err != nil {
		return nil, err
	}
	return NewDefinitionIDFromString(res), nil
}
func marshalToAttributes(value interface{}) (map[string]interface{}, error) {
	var res map[string]interface{}
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, &res); err != nil {
		return nil, err
	}
	return res, nil
}
func marshalToFeatures(value interface{}) (map[string]model.Feature, error) {
	f := map[string]jsonFeature{}
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, &f); err != nil {
		return nil, err
	}
	res := map[string]model.Feature{}
	for key, val := range f {
		res[key] = NewFeature(key, WithFeatureProperties(val.Properties), WithFeatureDefinitionFromString(val.Definition...))
	}
	return res, nil
}

func marshalToFeature(featureID string, value interface{}) (model.Feature, error) {
	f := &jsonFeature{}
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, f); err != nil {
		return nil, err
	}
	return NewFeature(featureID, WithFeatureProperties(f.Properties), WithFeatureDefinitionFromString(f.Definition...)), nil
}

func marshalToFeatureDefinition(value interface{}) ([]model.DefinitionID, error) {
	f := []string{}
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, &f); err != nil {
		return nil, err
	}
	res := []model.DefinitionID{}
	for _, def := range f {
		res = append(res, NewDefinitionIDFromString(def))
	}
	return res, nil
}
func marshalToFeatureProperties(value interface{}) (map[string]interface{}, error) {
	f := make(map[string]interface{})
	if jsonValue, err := json.Marshal(value); err != nil {
		return nil, err
	} else if err := json.Unmarshal(jsonValue, &f); err != nil {
		return nil, err
	}
	return f, nil
}

func convertFromFeature(feature model.Feature) jsonFeature {
	defsAsString := []string{}
	for _, def := range feature.GetDefinition() {
		defsAsString = append(defsAsString, def.String())
	}
	return jsonFeature{
		Definition: defsAsString,
		Properties: feature.GetProperties(),
	}
}

func convertFromFeaturesMap(features map[string]model.Feature) map[string]jsonFeature {
	if features == nil {
		return nil
	}
	res := make(map[string]jsonFeature)

	for id, feat := range features {
		res[id] = convertFromFeature(feat)
	}
	return res
}
