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

package matchers

import (
	"fmt"
	"reflect"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
)

type newContainerOptMatcher struct {
	opts []containerd.NewContainerOpts
	msg  string
}

// MatchesNewContainerOpts returns a Matcher interface for the New Container Opts
func MatchesNewContainerOpts(opts ...containerd.NewContainerOpts) gomock.Matcher {
	return &newContainerOptMatcher{opts, ""}
}

func (matcher *newContainerOptMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []containerd.NewContainerOpts:
		opts := x.([]containerd.NewContainerOpts)
		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		for i := range opts {
			actual := reflect.ValueOf(opts[i]).Pointer()
			expected := reflect.ValueOf(matcher.opts[i]).Pointer()
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

func (matcher *newContainerOptMatcher) String() string {
	return matcher.msg
}
