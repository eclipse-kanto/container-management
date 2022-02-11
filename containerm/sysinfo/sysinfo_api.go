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

package sysinfo

import "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"

// SystemInfoManager provides access to the system information related to the current runtime - both environment and daemon's specifics
type SystemInfoManager interface {
	// GetProjectInfo provides information about the current daemon's implementation
	GetProjectInfo() types.ProjectInfo
	// ... will add mo information in the future - e.g. Go version. Go runtime, OS specifics, etc.
}
