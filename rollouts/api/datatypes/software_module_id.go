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

package datatypes

// SoftwareModuleID represents the SoftwareModuleID Vorto SUv2 datatype
type SoftwareModuleID struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
