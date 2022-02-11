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

package version

// Version package values are overwritten at build time.
var (
	// ProjectVersion represents the version of the current binary.
	ProjectVersion = "1.0.0"

	// BuildTime is the time when the binaries are built
	BuildTime = "unknown"

	// APIVersion means the api version that the container manager and cli are build using
	APIVersion = "1.0.0"

	// GitCommit is the commit id to build the binaries
	GitCommit = "unknown"
)
