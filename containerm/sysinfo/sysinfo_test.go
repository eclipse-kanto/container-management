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

package sysinfo

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
)

var (
	projectInfo = types.ProjectInfo{
		ProjectVersion: "test-project-version",
		BuildTime:      "test-build-time",
		APIVersion:     "test-api-version",
		GitCommit:      "test-git-commit",
	}
	emptyProjectInfo         = types.ProjectInfo{}
	emptyFieldsInProjectInfo = types.ProjectInfo{
		ProjectVersion: "",
		BuildTime:      "",
		APIVersion:     "",
		GitCommit:      "",
	}
)

func TestNewSystemInfoMgr(t *testing.T) {
	tests := map[string]struct {
		arg  types.ProjectInfo
		want *systemInfoMgr
	}{
		"test_valid_project_info": {
			arg: projectInfo,
			want: &systemInfoMgr{
				mgrVersionInfo: projectInfo,
			},
		},
		"test_empty_project_info": {
			arg: emptyProjectInfo,
			want: &systemInfoMgr{
				mgrVersionInfo: emptyProjectInfo,
			},
		},
		"test_empty_fields_in_project_info": {
			arg: emptyFieldsInProjectInfo,
			want: &systemInfoMgr{
				mgrVersionInfo: emptyFieldsInProjectInfo,
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertEqual(t, newSystemInfoMgr(testCase.arg), testCase.want)
		})
	}
}

func TestGetProjectInfo(t *testing.T) {
	testSystemInfoMgr := &systemInfoMgr{
		mgrVersionInfo: projectInfo,
	}

	actual := projectInfo

	expected := testSystemInfoMgr.GetProjectInfo()

	testutil.AssertEqual(t, expected.ProjectVersion, actual.ProjectVersion)
	testutil.AssertEqual(t, expected.BuildTime, actual.BuildTime)
	testutil.AssertEqual(t, expected.APIVersion, actual.APIVersion)
	testutil.AssertEqual(t, expected.GitCommit, actual.GitCommit)
}
