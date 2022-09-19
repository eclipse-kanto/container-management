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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"testing"
)

const (
	testMemory            = "500M"
	testMemoryReservation = "300M"
	testMemorySwap        = "1G"
)

func TestFromAPIResources(t *testing.T) {
	apiResources := &types.Resources{
		Memory:            testMemory,
		MemoryReservation: testMemoryReservation,
		MemorySwap:        testMemorySwap,
	}

	thingsResources := fromAPIResources(apiResources)

	t.Run("test_from_api_resource_memory", func(t *testing.T) {
		testutil.AssertEqual(t, apiResources.Memory, thingsResources.Memory)
	})

	t.Run("test_from_api_resource_memory_reservation", func(t *testing.T) {
		testutil.AssertEqual(t, apiResources.MemoryReservation, thingsResources.MemoryReservation)
	})

	t.Run("test_from_api_resource_memory_swap", func(t *testing.T) {
		testutil.AssertEqual(t, apiResources.MemorySwap, thingsResources.MemorySwap)
	})

}

func TestToAPIResources(t *testing.T) {
	thingsResources := &resources{
		Memory:            testMemory,
		MemoryReservation: testMemoryReservation,
		MemorySwap:        testMemorySwap,
	}

	apiResources := toAPIResources(thingsResources)

	t.Run("test_to_api_resource_memory", func(t *testing.T) {
		testutil.AssertEqual(t, thingsResources.Memory, apiResources.Memory)
	})
	t.Run("test_to_api_resource_memory_reservation", func(t *testing.T) {
		testutil.AssertEqual(t, thingsResources.MemoryReservation, apiResources.MemoryReservation)
	})
	t.Run("test_to_api_resource_memory_swap", func(t *testing.T) {
		testutil.AssertEqual(t, thingsResources.MemorySwap, apiResources.MemorySwap)
	})
}
