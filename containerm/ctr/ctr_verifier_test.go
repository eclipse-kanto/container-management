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

package ctr

import (
	"context"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/notaryproject/notation-go/dir"
)

func TestContainerVerifier(t *testing.T) {
	t.Run("unknown_verifier", func(t *testing.T) {
		v, err := newContainerVerifier("unknown", nil, nil)
		testutil.AssertNotNil(t, err)
		testutil.AssertNil(t, v)
	})
	t.Run("skip_verifier", func(t *testing.T) {
		v, err := newContainerVerifier(VerifierNone, nil, nil)
		testutil.AssertNil(t, err)
		testutil.AssertNotNil(t, v)
		testutil.AssertNil(t, v.Verify(context.Background(), types.Image{}))
	})
	t.Run("notation_verifier", func(t *testing.T) {
		config := map[string]string{
			notationKeyConfigDir:  "testConfigDir",
			notationKeyLibexecDir: "testLibexecDir",
		}
		registryConfig := map[string]*RegistryConfig{
			testHost: testRegConfig,
		}

		v, err := newContainerVerifier(VerifierNotation, config, registryConfig)
		testutil.AssertNil(t, err)
		testutil.AssertNotNil(t, v)
		testutil.AssertNotNil(t, v.Verify(context.Background(), types.Image{})) // expected fail due to invalid config dir

		nv := v.(*notationVerifier)
		testutil.AssertEqual(t, registryConfig, nv.registryConfig)
		testutil.AssertEqual(t, config[notationKeyConfigDir], dir.UserConfigDir)
		testutil.AssertEqual(t, config[notationKeyLibexecDir], dir.UserLibexecDir)

	})

}
