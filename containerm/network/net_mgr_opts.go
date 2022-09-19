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

package network

// NetOpt provides Network Manager Config Options
type NetOpt func(netOpts *netOpts) error

type netOpts struct {
	netType  string
	metaPath string
	execRoot string

	// default bridge config
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

func applyOptsNet(netOpts *netOpts, opts ...NetOpt) error {
	for _, o := range opts {
		if err := o(netOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithLibNetType sets the default network interface type for the network manager service to use per container.
func WithLibNetType(netType string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.netType = netType
		return nil
	}
}

// WithLibNetMetaPath sets network manager service's meta path.
func WithLibNetMetaPath(metaPath string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.metaPath = metaPath
		return nil
	}
}

// WithLibNetExecRoot sets network manager service's exec root.
func WithLibNetExecRoot(execRoot string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.execRoot = execRoot
		return nil
	}
}

// WithLibNetDisableBridge disables the default network bridge interface creation and usage.
func WithLibNetDisableBridge(disableBridge bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.disableBridge = disableBridge
		return nil
	}
}

// WithLibNetName sets the default bridge network interface name.
func WithLibNetName(name string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.name = name
		return nil
	}
}

// WithLibNetIPV4 sets network IPv4.
func WithLibNetIPV4(ipv4 string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.ipV4 = ipv4
		return nil
	}
}

// WithLibNetFixedCIDRv4 sets network fixed CIDRv4.
func WithLibNetFixedCIDRv4(fixedCIDRv4 string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.fixedCIDRv4 = fixedCIDRv4
		return nil
	}
}

// WithLibNetGatewayIPv4 sets network gateway IPv4.
func WithLibNetGatewayIPv4(gatewayIPv4 string) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.gatewayIPv4 = gatewayIPv4
		return nil
	}
}

// WithLibNetEnableIPv6 enables IPv6
func WithLibNetEnableIPv6(enableIPv6 bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.enableIPv6 = enableIPv6
		return nil
	}
}

// WithLibNetMtu sets network MTU
func WithLibNetMtu(mtu int) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.mtu = mtu
		return nil
	}
}

// WithLibNetIcc enables network ICC
func WithLibNetIcc(icc bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.icc = icc
		return nil
	}
}

// WithLibNetIPTables enables network IP tables
func WithLibNetIPTables(ipTables bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.ipTables = ipTables
		return nil
	}
}

// WithLibNetIPForward enables network IP forwarding
func WithLibNetIPForward(ipForward bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.ipForward = ipForward
		return nil
	}
}

// WithLibNetIPMasq sets network IP masquerade
func WithLibNetIPMasq(ipMasq bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.ipMasq = ipMasq
		return nil
	}
}

// WithLibNetUserlandProxy enables usage of a network userland proxy
func WithLibNetUserlandProxy(userlandProxy bool) NetOpt {
	return func(netOpts *netOpts) error {
		netOpts.userlandProxy = userlandProxy
		return nil
	}
}
