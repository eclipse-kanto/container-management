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

package matchers

import (
	"fmt"
	"github.com/containerd/containerd/cio"
	"reflect"

	"github.com/golang/mock/gomock"
)

type cioAttachMatcher struct {
	cioAttach cio.Attach
	msg       string
}

// MatchesCioAttach returns a Matcher interface for cio attach function
func MatchesCioAttach(cioCtr cio.Attach) gomock.Matcher {
	return &cioAttachMatcher{cioCtr, ""}
}

func (matcher *cioAttachMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case cio.Attach:
		cioCtr := x.(cio.Attach)

		expected := reflect.ValueOf(cioCtr).Pointer()
		actual := reflect.ValueOf(matcher.cioAttach).Pointer()

		if !reflect.DeepEqual(expected, actual) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
			return false
		}

		return true
	default:
		return false
	}
}

func (matcher *cioAttachMatcher) String() string {
	return matcher.msg
}
