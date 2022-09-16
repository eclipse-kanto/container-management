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

package main

const (
	cmdAddressPathDefault = "/run/container-management/container-management.sock"
	cmdDebugDefault       = false
)

type config struct {
	addressPath string
	debug       bool
}

func (c *cli) setupCommandFlags() {
	flagSet := c.rootCmd.Flags()

	// init connection address to the GW CM daemon flag
	flagSet.StringVar(&c.config.addressPath, "host", cmdAddressPathDefault, "Specify the address path to the Eclipse Kanto container management")

	// init debug flags
	flagSet.BoolVar(&c.config.debug, "debug", cmdDebugDefault, "Switch commands log level to DEBUG mode")
}
