// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

package types

// VerificationConfig holds the data needed for verification of signed images
type VerificationConfig struct {
	// Keys filenames of public keys to verify an image signature. Each entry can include
	// an optional hash function(e.g. sha512) separated by a colon after the filename. If a hash
	// function is not included, then sha256 will be used by default.
	Keys []string `json:"keys,omitempty"`
}
