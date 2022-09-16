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

// Hash represents the supported hashing algorithms
type Hash string

const (
	// SHA1 algorithm
	SHA1 Hash = "SHA1"
	// SHA256 algorithm
	SHA256 Hash = "SHA256"
	// MD5 algorithm
	MD5 Hash = "MD5"
)
