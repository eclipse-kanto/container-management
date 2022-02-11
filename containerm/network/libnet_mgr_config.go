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

// config defines the network configuration.
type config struct {
	netType string //bridge

	metaPath string // meta store >> /var/lib/gw-man
	execRoot string // exec root >> /var/run/gw-man

	// bridge config
	bridgeConfig bridgeConfig

	activeSandboxes map[string]interface{}
}

// bridgeConfig defines the bridge network configuration.
type bridgeConfig struct {
	disableBridge bool
	name          string
	ipV4          string
	fixedCIDRv4   string
	gatewayIPv4   string
	enableIPv6    bool

	mtu           int
	icc           bool
	ipTables      bool
	ipForward     bool
	ipMasq        bool
	userlandProxy bool
}
