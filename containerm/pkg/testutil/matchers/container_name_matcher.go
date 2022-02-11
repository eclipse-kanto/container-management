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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/golang/mock/gomock"
)

type containerNameMatcher struct{ containerName string }

// MatchesContainerName returns a Matcher interface for the Container's name
func MatchesContainerName(t string) gomock.Matcher {
	return &containerNameMatcher{t}
}

func (o *containerNameMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case *types.Container:
		return x.(*types.Container).Name == o.containerName
	case types.Container:
		return x.(types.Container).Name == o.containerName
	default:
		return false
	}
}

func (o *containerNameMatcher) String() string {
	return "container name is not " + o.containerName
}
