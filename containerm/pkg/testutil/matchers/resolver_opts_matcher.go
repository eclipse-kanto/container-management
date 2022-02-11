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

package matchers

import (
	"fmt"
	"reflect"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
)

type resolverOptsMatcher struct {
	opts []containerd.RemoteOpt
	msg  string
}

// MatchesResolverOpts returns a Matcher interface for containerd.RemoteOpt
func MatchesResolverOpts(opts ...containerd.RemoteOpt) gomock.Matcher {
	return &resolverOptsMatcher{opts, ""}
}

func (matcher *resolverOptsMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []containerd.RemoteOpt:
		opts := x.([]containerd.RemoteOpt)
		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		client := &containerd.Client{}
		actualCtx := &containerd.RemoteContext{}
		expectedCtx := &containerd.RemoteContext{}
		for i := range opts {
			_ = opts[i](client, actualCtx)
			_ = matcher.opts[i](client, expectedCtx)
		}
		if !reflect.DeepEqual(expectedCtx, actualCtx) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expectedCtx, actualCtx)
			return false
		}
		return true
	default:
		return false
	}
}

func (matcher *resolverOptsMatcher) String() string {
	return matcher.msg
}
