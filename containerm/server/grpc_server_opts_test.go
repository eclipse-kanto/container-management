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

package server

import (
	"errors"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

var (
	emptyOpts    = grpcServerOpts{}
	existingOpts = grpcServerOpts{
		network:     "tcp",
		addressPath: "/example/address/path",
	}
	newNetworkFunc     = WithGrpcServerNetwork("unix")
	newAddressPathFunc = WithGrpcServerAddressPath("/new/address/path")
)

// MockWithGrpcServerNetwork is a mock of WithGrpcServerNetwork
func MockWithGrpcServerNetwork() GrpcServerOpt {
	return func(grpcServerOpts *grpcServerOpts) error {
		return errors.New("Example error")
	}
}

func TestWithGrpcServerNetworkAndAddressPath(t *testing.T) {
	testCases := map[string]struct {
		argument    grpcServerOpts
		optsToApply []GrpcServerOpt
		expected    *grpcServerOpts
	}{
		"test_network_from_empty_opts": {
			argument:    emptyOpts,
			optsToApply: []GrpcServerOpt{newNetworkFunc},
			expected: &grpcServerOpts{
				network: "unix",
			},
		},
		"test_network_from_existing_opts": {
			argument:    existingOpts,
			optsToApply: []GrpcServerOpt{newNetworkFunc},
			expected: &grpcServerOpts{
				network:     "unix",
				addressPath: "/example/address/path",
			},
		},
		"test_address_path_from_empty_opts": {
			argument:    emptyOpts,
			optsToApply: []GrpcServerOpt{newAddressPathFunc},
			expected: &grpcServerOpts{
				addressPath: "/new/address/path",
			},
		},
		"test_address_path_from_existing_opts": {
			argument:    existingOpts,
			optsToApply: []GrpcServerOpt{newAddressPathFunc},
			expected: &grpcServerOpts{
				network:     "tcp",
				addressPath: "/new/address/path",
			},
		},
		"test_network_and_address_path_from_empty_opts": {
			argument:    emptyOpts,
			optsToApply: []GrpcServerOpt{newNetworkFunc, newAddressPathFunc},
			expected: &grpcServerOpts{
				network:     "unix",
				addressPath: "/new/address/path",
			},
		},
		"test_network_and_address_path_from_exisitng_opts": {
			argument:    existingOpts,
			optsToApply: []GrpcServerOpt{newNetworkFunc, newAddressPathFunc},
			expected: &grpcServerOpts{
				network:     "unix",
				addressPath: "/new/address/path",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			applyOptsGrpcServer(&testCase.argument, testCase.optsToApply...)
			testutil.AssertEqual(t, testCase.expected, &testCase.argument)
		})
	}
}

func TestApplyGrpcServerOpts(t *testing.T) {
	testCases := map[string]struct {
		argument        grpcServerOpts
		optsToApply     []GrpcServerOpt
		expectedSuccess bool
	}{
		"apply_with_no_error": {
			argument:        existingOpts,
			optsToApply:     []GrpcServerOpt{WithGrpcServerNetwork("unix")},
			expectedSuccess: true,
		},
		"apply_with_error": {
			argument:        existingOpts,
			optsToApply:     []GrpcServerOpt{MockWithGrpcServerNetwork()},
			expectedSuccess: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := applyOptsGrpcServer(&testCase.argument, testCase.optsToApply...)
			testutil.AssertEqual(t, testCase.expectedSuccess, err == nil)
		})
	}
}
