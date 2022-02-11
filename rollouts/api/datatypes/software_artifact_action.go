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

// SoftwareArtifactAction represents the SoftwareArtifactAction Vorto SUv2 datatype
type SoftwareArtifactAction struct {
	FileName  string              `json:"filename"`
	Download  map[Protocol]*Links `json:"download"`
	Checksums map[Hash]string     `json:"checksums"`
	Size      uint64              `json:"size"`
}
