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
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/containerd/containerd"
	"github.com/golang/mock/gomock"
)

type updateTaskOptsMatcher struct {
	opts []containerd.UpdateTaskOpts
	msg  string
}

// MatchesUpdateTaskOpts returns a Matcher interface for the containerd.UpdateTaskOpts
func MatchesUpdateTaskOpts(opts ...containerd.UpdateTaskOpts) gomock.Matcher {
	return &updateTaskOptsMatcher{opts, ""}
}

func (matcher *updateTaskOptsMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []containerd.UpdateTaskOpts:
		opts := x.([]containerd.UpdateTaskOpts)
		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		ctx := context.Background()
		client := &containerd.Client{}
		actual := &containerd.UpdateTaskInfo{}
		expected := &containerd.UpdateTaskInfo{}
		for i := range opts {
			_ = opts[i](ctx, client, actual)
			_ = matcher.opts[i](ctx, client, expected)
		}
		if !reflect.DeepEqual(expected, actual) {
			toString := func(v interface{}) string {
				bytes, _ := json.Marshal(v)
				return string(bytes)
			}
			matcher.msg = fmt.Sprintf("expected %s , got %s", toString(expected), toString(actual))
			return false
		}
		return true
	default:
		return false
	}
}

func (matcher *updateTaskOptsMatcher) String() string {
	return matcher.msg
}
