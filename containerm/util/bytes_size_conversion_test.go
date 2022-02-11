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

package util

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestSizeToBytesNoErrors(t *testing.T) {

	// test Kilo
	bytes, _ := SizeToBytes("1k")
	testutil.AssertEqual(t, int64(kb), bytes)

	bytes, _ = SizeToBytes("1K")
	testutil.AssertEqual(t, int64(kb), bytes)

	// test Mega
	bytes, _ = SizeToBytes("1m")
	testutil.AssertEqual(t, int64(mb), bytes)

	bytes, _ = SizeToBytes("1M")
	testutil.AssertEqual(t, int64(mb), bytes)

	// test Giga
	bytes, _ = SizeToBytes("1g")
	testutil.AssertEqual(t, int64(gb), bytes)

	bytes, _ = SizeToBytes("1G")
	testutil.AssertEqual(t, int64(gb), bytes)

	// test float
	expected := 1.1 * mb
	bytes, _ = SizeToBytes("1.1m")
	testutil.AssertEqual(t, int64(expected), bytes)

	bytes, _ = SizeToBytes("1.1M")
	testutil.AssertEqual(t, int64(expected), bytes)

	// test wrong size error
	sizeStr := "1z"
	_, err := SizeToBytes(sizeStr)
	expectedErr := log.NewErrorf("invalid size provided %s", sizeStr)
	testutil.AssertError(t, expectedErr, err)
}

func TestSizeSizeRecalculate(t *testing.T) {
	addTwo := func(v float64) float64 {
		return v + 2
	}
	divideTwo := func(v float64) float64 {
		return v / 2
	}

	// test Kilo
	newSize, _ := SizeRecalculate("1k", addTwo)
	testutil.AssertEqual(t, "3k", newSize)

	newSize, _ = SizeRecalculate("2K", divideTwo)
	testutil.AssertEqual(t, "1K", newSize)

	// test Mega
	newSize, _ = SizeRecalculate("1.5m", addTwo)
	testutil.AssertEqual(t, "3.5m", newSize)

	newSize, _ = SizeRecalculate("3.2M", divideTwo)
	testutil.AssertEqual(t, "1.6M", newSize)

	// test Giga
	newSize, _ = SizeRecalculate("1.53g", addTwo)
	testutil.AssertEqual(t, "3.53g", newSize)

	newSize, _ = SizeRecalculate("3.5G", divideTwo)
	testutil.AssertEqual(t, "1.75G", newSize)

	// test precision
	newSize, _ = SizeRecalculate("1G", func(f float64) float64 {
		return 2.251000001
	})
	testutil.AssertEqual(t, "2.251G", newSize)

	// test wrong size error
	sizeStr := "1z"
	_, err := SizeRecalculate(sizeStr, addTwo)
	expectedErr := log.NewErrorf("invalid size provided %s", sizeStr)
	testutil.AssertError(t, expectedErr, err)
}
