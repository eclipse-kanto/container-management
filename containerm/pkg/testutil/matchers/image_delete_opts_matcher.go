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
	"context"
	"fmt"
	"github.com/containerd/containerd/images"
	"github.com/golang/mock/gomock"
	"reflect"
)

type imageDeleteOptsMatcher struct {
	opts []images.DeleteOpt
	msg  string
}

// MatchesImageDeleteOpts returns a Matcher interface for images.DeleteOpt used in variadic functions
func MatchesImageDeleteOpts(opts ...images.DeleteOpt) gomock.Matcher {
	return &imageDeleteOptsMatcher{opts, ""}
}

func (matcher *imageDeleteOptsMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []images.DeleteOpt:
		opts := x.([]images.DeleteOpt)
		if len(matcher.opts) != len(opts) {
			matcher.msg = fmt.Sprintf("expected %d , got %d", len(matcher.opts), len(opts))
			return false
		}
		actualCtx := context.TODO()
		actualDelOptions := &images.DeleteOptions{}
		expectedCtx := context.TODO()
		expectedDeleteOptions := &images.DeleteOptions{}
		for i := range opts {
			_ = opts[i](actualCtx, actualDelOptions)
			_ = matcher.opts[i](expectedCtx, expectedDeleteOptions)
		}
		if !reflect.DeepEqual(expectedCtx, actualCtx) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expectedCtx, actualCtx)
			return false
		}
		if !reflect.DeepEqual(expectedDeleteOptions, actualDelOptions) {
			matcher.msg = fmt.Sprintf("expected %v , got %v", expectedDeleteOptions, actualDelOptions)
			return false
		}
		return true
	default:
		return false
	}
}

func (matcher *imageDeleteOptsMatcher) String() string {
	return matcher.msg
}
