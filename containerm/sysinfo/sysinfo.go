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

type systemInfoMgr struct {
	mgrVersionInfo types.ProjectInfo
}

func newSystemInfoMgr(mgrVersionInfo types.ProjectInfo) *systemInfoMgr {
	return &systemInfoMgr{mgrVersionInfo: mgrVersionInfo}
}

func (sysInfoMgr *systemInfoMgr) GetProjectInfo() types.ProjectInfo {
	return sysInfoMgr.mgrVersionInfo
}
