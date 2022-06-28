// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package types

// DecryptConfig holds the data needed for image decryption
type DecryptConfig struct {
	Keys       []string `json:"keys,omitempty"`
	Recipients []string `json:"recipients,omitempty"`
}
