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

package error

import (
	"fmt"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

func TestCompoundError_Append(t *testing.T) {

	t.Run("test_compound_error_empty_message", func(t *testing.T) {
		compoundErr := CompoundError{}
		err := compoundErr.Error()
		if err != "no error" {
			t.Errorf("unexpected compound error: %s", err)
		}
	})

	t.Run("test_compound_error_empty_size", func(t *testing.T) {
		compoundErr := CompoundError{}
		size := compoundErr.Size()
		if size != 0 {
			t.Errorf("unexpected compound error size: %d", size)
		}
	})

	t.Run("test_compound_error_single", func(t *testing.T) {
		compoundErr := CompoundError{}
		errorMsg := "errorMsg"
		firstErr := log.NewError(errorMsg)
		compoundErr.Append(firstErr)
		compound := compoundErr.Error()
		if compound != errorMsg {
			t.Errorf("expected compound error: %s, but was: %s", errorMsg, compound)
		}
	})

	t.Run("test_compound_error_multiple", func(t *testing.T) {
		compoundErr := CompoundError{}
		firstErrorMsg := "firstErrorMsg"
		secondErrorMsg := "secondErrorMsg"
		firstErr := log.NewError(firstErrorMsg)
		secondErr := log.NewError(secondErrorMsg)
		compoundErr.Append(firstErr)
		compoundErr.Append(secondErr)
		compound := compoundErr.Error()
		expected := fmt.Sprintf("%d errors:\n\n%s", 2, "* "+firstErrorMsg+"\n"+"* "+secondErrorMsg)
		if compound != expected {
			t.Errorf("expected compound error:\n%s\nbut was:\n%s", expected, compound)
		}
	})
}
