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

// returns shallow copy
func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(original))
	for id, prop := range original {
		copy[id] = prop
	}
	return copy
}

// returns shallow copy
func copyFeaturesMap(original map[string]model.Feature) map[string]model.Feature {
	copy := make(map[string]model.Feature, len(original))
	for id, feature := range original {
		copy[id] = feature
	}
	return copy
}
