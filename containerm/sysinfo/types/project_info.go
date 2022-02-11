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

package types

// ProjectInfo contains the main information about the project
type ProjectInfo struct {
	ProjectVersion string `json:"project_version"`
	BuildTime      string `json:"build_time"`
	APIVersion     string `json:"api_version"`
	GitCommit      string `json:"git_commit"`
}
