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

package ctr

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testNamespace      = "test-namespace"
	testConnectionPath = "test-conn-path"
	testRootExec       = "test-root-exec"
	testMetaPath       = "test-meta-path"
	testHost           = "test-host"
	testUser           = "test-user"
	testPass           = "test-pass"
)

var (
	regConfig = &RegistryConfig{
		IsInsecure: false,
		Credentials: &AuthCredentials{
			UserID:   testUser,
			Password: testPass,
		},
		Transport: nil,
	}

	testOpt = &ctrOpts{
		namespace:       testNamespace,
		connectionPath:  testConnectionPath,
		registryConfigs: map[string]*RegistryConfig{testHost: regConfig},
		rootExec:        testRootExec,
		metaPath:        testMetaPath,
	}
)

func TestCtrOpts(t *testing.T) {
	t.Run("test_ctr_opts", func(t *testing.T) {
		opt := []ContainerOpts{}
		opt = append(opt,
			WithCtrdConnectionPath(testConnectionPath),
			WithCtrdNamespace(testNamespace),
			WithCtrdRootExec(testRootExec),
			WithCtrdMetaPath(testMetaPath),
			WithCtrdRegistryConfigs(map[string]*RegistryConfig{testHost: regConfig}),
		)

		opts := &ctrOpts{}
		applyOptsCtr(opts, opt...)

		testutil.AssertEqual(t, testOpt, opts)
	})
}
