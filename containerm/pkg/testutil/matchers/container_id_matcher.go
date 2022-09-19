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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/golang/mock/gomock"
)

type containerIDMatcher struct{ containerID string }

// MatchesContainerID returns a Matcher interface for the Container's ID
func MatchesContainerID(t string) gomock.Matcher {
	return &containerIDMatcher{t}
}

func (o *containerIDMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case *types.Container:
		return x.(*types.Container).ID == o.containerID
	case types.Container:
		return x.(types.Container).ID == o.containerID
	default:
		return false
	}
}

func (o *containerIDMatcher) String() string {
	return "container id is not " + o.containerID
}
