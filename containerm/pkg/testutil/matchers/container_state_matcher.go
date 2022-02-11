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

type containerStateMatcher struct {
	containerID   string
	requiredState types.State
}

// MatchesContainerState returns a Matcher interface for the Container's state
func MatchesContainerState(id string, state types.State) gomock.Matcher {
	return &containerStateMatcher{
		containerID:   id,
		requiredState: state,
	}
}

func (o *containerStateMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case *types.Container:
		return x.(*types.Container).ID == o.containerID &&
			o.CompareStates(&o.requiredState, x.(*types.Container).State)
	case types.Container:
		return x.(types.Container).ID == o.containerID &&
			o.CompareStates(&o.requiredState, x.(*types.Container).State)
	default:
		return false
	}
}

func (o *containerStateMatcher) CompareStates(expected *types.State, actual *types.State) bool {
	stoppedStates := expected.Paused == actual.Paused && expected.Dead == actual.Dead &&
		expected.Exited == actual.Exited
	runningStates := expected.Running == actual.Running && expected.Restarting == actual.Restarting
	return stoppedStates && runningStates && expected.Status == actual.Status
}

func (o *containerStateMatcher) String() string {
	return fmt.Sprintf("container id is not %s or state is not %+v", o.containerID, o.requiredState)
}
