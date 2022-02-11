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

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/golang/mock/gomock"
)

type containerUpdateMatcher struct {
	rp  *types.RestartPolicy
	r   *types.Resources
	msg string
}

// MatchesContainerUpdate returns a Matcher interface for the Container after update
// Nil field means: do not match.
func MatchesContainerUpdate(rp *types.RestartPolicy, r *types.Resources) gomock.Matcher {
	return &containerUpdateMatcher{rp, r, ""}
}

func (o *containerUpdateMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case *types.Container:
		config := x.(*types.Container).HostConfig
		if o.rp != nil && config.RestartPolicy != o.rp {
			o.msg = fmt.Sprintf("expected restart policy - %v, actual restart policy - %v", o.rp, config.RestartPolicy)
			return false
		}
		if o.r != nil && config.Resources != o.r {
			o.msg = fmt.Sprintf("expected resources - %v, actual resources - %v", o.r, config.Resources)
			return false
		}
		return true
	default:
		return false
	}
}

func (o *containerUpdateMatcher) String() string {
	return o.msg
}
