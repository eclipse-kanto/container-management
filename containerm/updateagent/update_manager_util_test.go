// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package updateagent

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestFindContainerVersion(t *testing.T) {
	testCases := []struct {
		imageName string
		version   string
	}{
		{imageName: "mycontainerregistry.com:8080/my-container:v1", version: "v1"},
		{imageName: "my-container:my-branch:123456", version: "my-branch:123456"},
		{imageName: "my-container@sha256:1234567890123456789012345678901234567890123456789012345678901234", version: "sha256:1234567890123456789012345678901234567890123456789012345678901234"},
		{imageName: "my-container", version: "n/a"},
		{imageName: "my-container:", version: "n/a"},
		{imageName: "my-container@", version: "n/a"},
		{imageName: "", version: "n/a"},
	}

	for _, testCase := range testCases {
		t.Log(testCase.imageName)
		testutil.AssertEqual(t, testCase.version, findContainerVersion(testCase.imageName))
	}
}
