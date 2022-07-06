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
	"reflect"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
)

type taskKillOptsMatcher struct {
	opts []containerd.KillOpts
	msg  string
}

// MatchesTaskKillOpts returns a Matcher interface for containerd task kill opts
func MatchesTaskKillOpts(opts ...containerd.KillOpts) gomock.Matcher {
	return &taskKillOptsMatcher{opts, ""}
}

func (matcher *taskKillOptsMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []containerd.KillOpts:
		opts := x.([]containerd.KillOpts)

		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		for i := range opts {
			expected := reflect.ValueOf(opts[i]).Pointer()
			actual := reflect.ValueOf(matcher.opts[i]).Pointer()

			if !reflect.DeepEqual(expected, actual) {
				matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
				return false
			}
		}

		return true
	default:
		return false
	}
}

func (matcher *taskKillOptsMatcher) String() string {
	return matcher.msg
}
