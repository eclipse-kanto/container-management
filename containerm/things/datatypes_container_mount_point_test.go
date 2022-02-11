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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	mountPointSource      = "/mount/point/src"
	mountPointDestination = "/mount/point/dst"
)

func TestToAPIMountPoint(t *testing.T) {
	mountPoint := &mountPoint{
		Destination:     mountPointDestination,
		Source:          mountPointSource,
		PropagationMode: rprivate,
	}
	result := toAPIMountPoint(mountPoint)

	t.Run("test_to_api_mount_point_destination", func(t *testing.T) {
		testutil.AssertEqual(t, mountPoint.Destination, result.Destination)
	})
	t.Run("test_to_api_mount_point_source", func(t *testing.T) {
		testutil.AssertEqual(t, mountPoint.Source, result.Source)
	})
	t.Run("test_to_api_mount_point_pr_mode", func(t *testing.T) {
		testutil.AssertEqual(t, toAPIPRMode(mountPoint.PropagationMode), result.PropagationMode)
	})
}

func TestFromAPIMountPoint(t *testing.T) {
	mountPoint := &types.MountPoint{
		Destination:     mountPointDestination,
		Source:          mountPointSource,
		PropagationMode: types.RPrivatePropagationMode,
	}
	result := fromAPIMountPoint(*mountPoint)

	t.Run("test_to_api_mount_point_destination", func(t *testing.T) {
		testutil.AssertEqual(t, mountPoint.Destination, result.Destination)
	})
	t.Run("test_to_api_mount_point_source", func(t *testing.T) {
		testutil.AssertEqual(t, mountPoint.Source, result.Source)
	})
	t.Run("test_to_api_mount_point_pr_mode", func(t *testing.T) {
		testutil.AssertEqual(t, fromAPIPRMode(mountPoint.PropagationMode), result.PropagationMode)
	})
}

func TestToAPIPRMode(t *testing.T) {
	tests := map[string]struct {
		prMode   propagationMode
		expected string
	}{
		"test_to_api_pr_mode_rprivate": {
			prMode:   rprivate,
			expected: types.RPrivatePropagationMode,
		},
		"test_to_api_pr_mode_private": {
			prMode:   private,
			expected: types.PrivatePropagationMode,
		},
		"test_to_api_pr_mode_rshared": {
			prMode:   rshared,
			expected: types.RSharedPropagationMode,
		},
		"test_to_api_pr_mode_shared": {
			prMode:   shared,
			expected: types.SharedPropagationMode,
		},
		"test_to_api_pr_mode_rslave": {
			prMode:   rslave,
			expected: types.RSlavePropagationMode,
		},
		"test_to_api_pr_mode_slave": {
			prMode:   slave,
			expected: types.SlavePropagationMode,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := toAPIPRMode(testCase.prMode)
			testutil.AssertEqual(t, testCase.expected, actual)

		})
	}

}

func TestFromAPIPRMode(t *testing.T) {
	tests := map[string]struct {
		prMode   string
		expected propagationMode
	}{
		"test_from_api_pr_mode_rprivate": {
			prMode:   types.RPrivatePropagationMode,
			expected: rprivate,
		},
		"test_from_api_pr_mode_private": {
			prMode:   types.PrivatePropagationMode,
			expected: private,
		},
		"test_from_api_pr_mode_rshared": {
			prMode:   types.RSharedPropagationMode,
			expected: rshared,
		},
		"test_from_api_pr_mode_shared": {
			prMode:   types.SharedPropagationMode,
			expected: shared,
		},
		"test_from_api_pr_mode_rslave": {
			prMode:   types.RSlavePropagationMode,
			expected: rslave,
		},
		"test_from_api_pr_mode_slave": {
			prMode:   types.SlavePropagationMode,
			expected: slave,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := fromAPIPRMode(testCase.prMode)
			testutil.AssertEqual(t, testCase.expected, actual)
		})
	}
}
