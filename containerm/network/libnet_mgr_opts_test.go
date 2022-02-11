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

package network

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestApplyMgrOpts(t *testing.T) {
	tests := map[string]struct {
		testOpts      []NetOpt
		expectedError error
	}{
		"test_apply_with_no_error": {
			testOpts: []NetOpt{WithLibNetName("test0")},
		},
		"test_apply_with_error": {
			testOpts: []NetOpt{func() NetOpt {
				return func(netOpts *netOpts) error {
					return log.NewError("test error")
				}
			}()},
			expectedError: log.NewError("test error"),
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			resultOpts := &netOpts{}
			err := applyOptsNet(resultOpts, testCase.testOpts...)
			testutil.AssertError(t, testCase.expectedError, err)
		})
	}
}

func TestMgrOpts(t *testing.T) {
	tests := map[string]struct {
		expectedOpts *netOpts
		testOpts     []NetOpt
	}{
		"netmgr_test_opts_disable_bridge": {
			expectedOpts: &netOpts{disableBridge: true},
			testOpts:     []NetOpt{WithLibNetDisableBridge(true)},
		},
		"netmgr_test_opts_enable_ipv6": {
			expectedOpts: &netOpts{enableIPv6: true},
			testOpts:     []NetOpt{WithLibNetEnableIPv6(true)},
		},
		"netmgr_test_opts_exec_root": {
			expectedOpts: &netOpts{execRoot: testNetExecRoot},
			testOpts:     []NetOpt{WithLibNetExecRoot(testNetExecRoot)},
		},
		"netmgr_test_opts_fixed_cidr": {
			expectedOpts: &netOpts{fixedCIDRv4: "fixed"},
			testOpts:     []NetOpt{WithLibNetFixedCIDRv4("fixed")},
		},
		"netmgr_test_opts_gw_ipv4": {
			expectedOpts: &netOpts{gatewayIPv4: netSettingsIPGW},
			testOpts:     []NetOpt{WithLibNetGatewayIPv4(netSettingsIPGW)},
		},
		"netmgr_test_opts_net_icc": {
			expectedOpts: &netOpts{icc: true},
			testOpts:     []NetOpt{WithLibNetIcc(true)},
		},
		"netmgr_test_opts_net_ip_fwd": {
			expectedOpts: &netOpts{ipForward: true},
			testOpts:     []NetOpt{WithLibNetIPForward(true)},
		},
		"netmgr_test_opts_net_ip_masq": {
			expectedOpts: &netOpts{ipMasq: true},
			testOpts:     []NetOpt{WithLibNetIPMasq(true)},
		},
		"netmgr_test_opts_net_ip_tables": {
			expectedOpts: &netOpts{ipTables: true},
			testOpts:     []NetOpt{WithLibNetIPTables(true)},
		},
		"netmgr_test_opts_net_ipv4": {
			expectedOpts: &netOpts{ipV4: netSettingsIP},
			testOpts:     []NetOpt{WithLibNetIPV4(netSettingsIP)},
		}, "netmgr_test_opts_net_meta": {
			expectedOpts: &netOpts{metaPath: testNetMetaPath},
			testOpts:     []NetOpt{WithLibNetMetaPath(testNetMetaPath)},
		},
		"netmgr_test_opts_net_mtu": {
			expectedOpts: &netOpts{mtu: 1500},
			testOpts:     []NetOpt{WithLibNetMtu(1500)},
		},
		"netmgr_test_opts_net_name": {
			expectedOpts: &netOpts{name: "test0"},
			testOpts:     []NetOpt{WithLibNetName("test0")},
		},
		"netmgr_test_opts_net_type": {
			expectedOpts: &netOpts{netType: bridgeNetworkName},
			testOpts:     []NetOpt{WithLibNetType(bridgeNetworkName)},
		},
		"netmgr_test_opts_net_userland_proxy": {
			expectedOpts: &netOpts{userlandProxy: true},
			testOpts:     []NetOpt{WithLibNetUserlandProxy(true)},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			resultOpts := &netOpts{}
			applyOptsNet(resultOpts, testCase.testOpts...)
			testutil.AssertEqual(t, testCase.expectedOpts, resultOpts)
		})
	}
}
