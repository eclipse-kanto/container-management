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

package updateagent

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"

	"github.com/eclipse-kanto/update-manager/api/types"
)

const (
	testContainerName    = "test-container"
	testContainerVersion = "1.2.3"

	testContainerName2    = "test-container2"
	testContainerVersion2 = "11.22.33"
)

func TestFindComponent(t *testing.T) {
	desiredState := &internalDesiredState{
		desiredState: &types.DesiredState{
			Domains: []*types.Domain{{
				Components: []*types.ComponentWithConfig{
					{Component: types.Component{ID: testContainerName, Version: testContainerVersion}},
					{Component: types.Component{ID: testContainerName2, Version: testContainerVersion2}},
				},
			}},
		},
	}
	testCases := map[string]struct {
		target   string
		expected types.Component
	}{
		"test_present_1": {target: testContainerName, expected: types.Component{ID: testContainerName, Version: testContainerVersion}},
		"test_present_2": {target: testContainerName2, expected: types.Component{ID: testContainerName2, Version: testContainerVersion2}},
		"test_missing":   {target: "missing", expected: types.Component{}},
		"test_empty":     {target: "", expected: types.Component{}},
	}
	for testName, testCase := range testCases {
		t.Log("TestName: ", testName)
		testutil.AssertEqual(t, testCase.expected, desiredState.findComponent(testCase.target))
	}
}

func TestToInternalDesiredStateDomainError(t *testing.T) {
	testCases := map[string]([]string){
		"test_no_domains":           nil,
		"test_empty_domains":        {},
		"test_no_containers_domain": {"containerized-apps"},
		"test_too_many_domains":     {"containers", "apps"},
	}
	for testName, testCaseDomains := range testCases {
		t.Log("TestName: ", testName)
		testDesiredState := &types.DesiredState{}
		if testCaseDomains != nil {
			domains := []*types.Domain{}
			for _, domain := range testCaseDomains {
				domains = append(domains, &types.Domain{ID: domain})
			}
			testDesiredState.Domains = domains
		}
		internalDesiredState, err := toInternalDesiredState(testDesiredState, "containers")
		testutil.AssertNil(t, internalDesiredState)
		testutil.AssertNotNil(t, err)
	}
}

func TestToInternalDesiredContainersError(t *testing.T) {
	testDesiredState := &types.DesiredState{Domains: []*types.Domain{{ID: "containers",
		Components: []*types.ComponentWithConfig{{
			Component: types.Component{ID: testContainerName, Version: testContainerVersion},
			Config:    []*types.KeyValuePair{{Key: "image", Value: ""}},
		}},
	}}}
	internalDesiredState, err := toInternalDesiredState(testDesiredState, "containers")
	testutil.AssertNil(t, internalDesiredState)
	testutil.AssertNotNil(t, err)
}

func TestToInternalDesiredBaselinesError(t *testing.T) {
	testDesiredState := &types.DesiredState{
		Domains: []*types.Domain{{ID: "containers",
			Components: []*types.ComponentWithConfig{{
				Component: types.Component{ID: testContainerName, Version: testContainerVersion},
			}},
		}},
		Baselines: []*types.Baseline{
			{Title: "test-baseline", Components: []string{"containers:" + testContainerName, "containers:" + testContainerName2}},
		},
	}
	internalDesiredState, err := toInternalDesiredState(testDesiredState, "containers")
	testutil.AssertNil(t, internalDesiredState)
	testutil.AssertNotNil(t, err)
}

func TestToInternalDesiredStateSystemContainers(t *testing.T) {
	testCases := map[string]struct {
		key      string
		value    string
		expected []string
	}{
		"test_valid_system_containers": {key: "systemContainers", value: "sys-container-1, corelib", expected: []string{"sys-container-1", "corelib"}},
		// "test_no_system_containers":    {key: "coreContainers", value: "some-container"},
		// "test_no_config":               {},
	}
	for testName, testCase := range testCases {
		t.Log("TestName: ", testName)
		testDesiredState := &types.DesiredState{
			Domains: []*types.Domain{{ID: "containers",
				Config: []*types.KeyValuePair{{Key: testCase.key, Value: testCase.value}},
				Components: []*types.ComponentWithConfig{
					{Component: types.Component{ID: testContainerName, Version: testContainerVersion}},
					{Component: types.Component{ID: testContainerName2, Version: testContainerVersion2}},
				},
			}},
			Baselines: []*types.Baseline{
				{Title: "test-baseline", Components: []string{"containers:" + testContainerName}},
			},
		}
		internalDesiredState, err := toInternalDesiredState(testDesiredState, "containers")
		testutil.AssertNil(t, err)
		testutil.AssertEqual(t, testDesiredState, internalDesiredState.desiredState)
		testutil.AssertEqual(t, 2, len(internalDesiredState.containers))
		testutil.AssertEqual(t, 2, len(internalDesiredState.baselines))
		testutil.AssertEqual(t, testCase.expected, internalDesiredState.systemContainers)
	}
}
