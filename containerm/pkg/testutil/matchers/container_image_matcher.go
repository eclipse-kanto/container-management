// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/golang/mock/gomock"
)

type containerImageMatcher struct {
	containerName  string
	containerImage string
}

// MatchesContainerImage returns a Matcher interface for the Container's name and image name
func MatchesContainerImage(name, image string) gomock.Matcher {
	return &containerImageMatcher{containerName: name, containerImage: image}
}

func (o *containerImageMatcher) Matches(x interface{}) bool {
	switch c := x.(type) {
	case *types.Container:
		return c.Name == o.containerName && c.Image.Name == o.containerImage
	case types.Container:
		return c.Name == o.containerName && c.Image.Name == o.containerImage
	default:
		return false
	}
}

func (o *containerImageMatcher) String() string {
	return fmt.Sprintf("container name is not %s or image name is not %+s", o.containerName, o.containerImage)
}
