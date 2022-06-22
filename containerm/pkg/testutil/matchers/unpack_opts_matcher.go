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
	"context"
	"fmt"
	"reflect"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
)

type unpackOptsMatcher struct {
	opts []containerd.UnpackOpt
	msg  string
}

// MatchesUnpackOpts returns a Matcher interface for the containerd.UnpackOpt
func MatchesUnpackOpts(opts ...containerd.UnpackOpt) gomock.Matcher {
	return &unpackOptsMatcher{opts, ""}
}

func (matcher *unpackOptsMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []containerd.UnpackOpt:
		opts := x.([]containerd.UnpackOpt)
		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		ctx := context.TODO()
		expected := &containerd.UnpackConfig{}
		actual := &containerd.UnpackConfig{}
		for i := range opts {
			_ = opts[i](ctx, actual)
			_ = matcher.opts[i](ctx, expected)
		}
		if len(expected.ApplyOpts) != len(actual.ApplyOpts) {
			matcher.msg = fmt.Sprintf("expected number of ApplyOpts %d , got %d", len(expected.ApplyOpts), len(actual.ApplyOpts))
			return false
		}
		for i := range expected.ApplyOpts {
			expectedA := reflect.ValueOf(expected.ApplyOpts[i]).Pointer()
			actualA := reflect.ValueOf(actual.ApplyOpts[i]).Pointer()
			if !reflect.DeepEqual(expectedA, actualA) {
				matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
				return false
			}
		}
		if len(expected.SnapshotOpts) != len(actual.SnapshotOpts) {
			matcher.msg = fmt.Sprintf("expected number of SnapshotOpts %d , got %d", len(expected.SnapshotOpts), len(actual.SnapshotOpts))
			return false
		}
		for i := range expected.SnapshotOpts {
			expectedS := reflect.ValueOf(expected.SnapshotOpts[i]).Pointer()
			actualS := reflect.ValueOf(actual.SnapshotOpts[i]).Pointer()
			if !reflect.DeepEqual(expectedS, actualS) {
				matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
				return false
			}
		}
		if !reflect.DeepEqual(expected.CheckPlatformSupported, actual.CheckPlatformSupported) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expected, actual)
			return false
		}
		return true
	default:
		return false
	}
}

func (matcher *unpackOptsMatcher) String() string {
	return matcher.msg
}
