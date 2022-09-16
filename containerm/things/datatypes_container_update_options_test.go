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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestToAPIUpdateOpts(t *testing.T) {
	t.Run("test_to_api_update_opts_rp", func(t *testing.T) {
		opts := &updateOptions{
			RestartPolicy: &restartPolicy{
				RpType:        onFailure,
				RetryTimeout:  10,
				MaxRetryCount: 3,
			},
			Resources: &resources{
				Memory:            testMemory,
				MemoryReservation: testMemoryReservation,
				MemorySwap:        testMemorySwap,
			},
		}
		apiOpts := toAPIUpdateOptions(opts)
		testutil.AssertEqual(t, apiOpts.RestartPolicy, toAPIRestartPolicy(opts.RestartPolicy))
		testutil.AssertEqual(t, apiOpts.Resources, toAPIResources(opts.Resources))
	})
	t.Run("test_to_api_update_opts_is_nil", func(t *testing.T) {
		testutil.AssertEqual(t, toAPIUpdateOptions(nil), &types.UpdateOpts{})
	})
	t.Run("test_to_api_update_opts_rp_is_nil", func(t *testing.T) {
		testutil.AssertNil(t, toAPIUpdateOptions(&updateOptions{}).RestartPolicy)
	})
	t.Run("test_to_api_update_opts_resources_is_nil", func(t *testing.T) {
		testutil.AssertNil(t, toAPIUpdateOptions(&updateOptions{}).Resources)
	})
}
