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

package datatypes

// RemoveAction represents the SoftwareRemoveAction Vorto SUv2 datatype
type RemoveAction struct {
	CorrelationID string                   `json:"correlationId"`
	Software      []*DependencyDescription `json:"software"`
	Weight        int                      `json:"weight,omitempty"`
	Metadata      map[string]string        `json:"metadata,omitempty"`
	Forced        bool                     `json:"forced,omitempty"`
}
