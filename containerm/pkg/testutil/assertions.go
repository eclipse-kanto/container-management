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

package testutil

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// AssertError asserts that the expected and actual errors are the same
func AssertError(t *testing.T, expected error, actual error) {
	if expected == nil {
		if actual != nil {
			t.Errorf("expected nil , got %v", actual)
			t.Fail()
		}
	} else {
		if actual == nil {
			t.Errorf("expected %v , got nil", expected)
			t.Fail()
		} else {
			if expected.Error() != actual.Error() {
				t.Errorf("expected %v , got %v", expected, actual)
				t.Fail()
			}
		}
	}
}

// AssertEqual asserts that the expected and actual values are deeply equal
func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v , got %v", expected, actual)
		t.Fail()
	}
}

// AssertNotEqual asserts that the expected and actual values are not deeply equal
func AssertNotEqual(t *testing.T, expected interface{}, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v and %v to be different", expected, actual)
		t.Fail()
	}
}

// AssertTrue asserts that the actual value is true
func AssertTrue(t *testing.T, actual bool) {
	AssertEqual(t, true, actual)
}

// AssertFalse sserts that the actual value is false
func AssertFalse(t *testing.T, actual bool) {
	AssertEqual(t, false, actual)
}

// AssertContainsString asserts that a string contains a substring
func AssertContainsString(t *testing.T, str string, substring string) {
	if !strings.Contains(str, substring) {
		t.Fatalf("%s was not contained in: %s", str, substring)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}) {
	if !IsNil(value) {
		t.Fatalf("expected nil, but was %+v", value)
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	if IsNil(value) {
		t.Fatalf("expected value not to be nil")
	}
}

// IsNil checks if a value is nil
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() != reflect.Struct && rv.IsNil()
}

// AssertWithTimeout asserts that an operation is completed within a certain period of time
func AssertWithTimeout(t *testing.T, waitGroup *sync.WaitGroup, testTimeout time.Duration) {
	testWaitChan := make(chan struct{})
	go func() {
		defer close(testWaitChan)
		waitGroup.Wait()
	}()
	select {
	case <-testWaitChan:
		return // completed normally
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for ", testTimeout)
	}
}
