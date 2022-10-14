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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestToAPINetworkMode(t *testing.T) {
	t.Run("test_to_api_network_mode_bridge", func(t *testing.T) {
		testutil.AssertEqual(t, bridge.toAPINetworkMode(), types.NetworkModeBridge)
	})

	t.Run("test_to_api_network_mode_host", func(t *testing.T) {
		testutil.AssertEqual(t, host.toAPINetworkMode(), types.NetworkModeHost)
	})
}

func TestFromAPINetworkMode(t *testing.T) {
	t.Run("test_from_api_network_mode_bridge", func(t *testing.T) {
		testutil.AssertEqual(t, bridge, fromAPINetworkMode(types.NetworkModeBridge))
	})

	t.Run("test_from_api_network_mode_host", func(t *testing.T) {
		testutil.AssertEqual(t, host, fromAPINetworkMode(types.NetworkModeHost))
	})
}
