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

package ctr

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestCioCache(t *testing.T) {
	t.Run("test_cio_cache", func(t *testing.T) {
		io := &containerIO{id: "test-io-id"}
		testID := "test-id"
		testID2 := "test-id-2"
		cioCache := newCache()

		// test put
		cioCache.Put(testID, io)

		// test get
		expIo := cioCache.Get(testID)
		testutil.AssertEqual(t, expIo, io)
		// test get non existant
		expIo = cioCache.Get(testID2)
		testutil.AssertNil(t, expIo)

		// test remove
		cioCache.Remove(testID)
		testutil.AssertNil(t, cioCache.Get(testID))

	})
}
