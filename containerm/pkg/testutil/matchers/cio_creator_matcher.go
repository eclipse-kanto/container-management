// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package matchers

import (
	"fmt"
	"github.com/containerd/containerd/cio"
	"reflect"

	"github.com/golang/mock/gomock"
)

type cioCreatorMatcher struct {
	cioCreator cio.Creator
	msg        string
}

// MatchesCioCreator returns a Matcher interface for cio creator function
func MatchesCioCreator(cioCtr cio.Creator) gomock.Matcher {
	return &cioCreatorMatcher{cioCtr, ""}
}

func (matcher *cioCreatorMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case cio.Creator:
		cioCtr := x.(cio.Creator)

		expected := reflect.ValueOf(cioCtr).Pointer()
		actual := reflect.ValueOf(matcher.cioCreator).Pointer()

		if !reflect.DeepEqual(expected, actual) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
			return false
		}

		return true
	default:
		return false
	}
}

func (matcher *cioCreatorMatcher) String() string {
	return matcher.msg
}
