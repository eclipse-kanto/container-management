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

package util

import (
	"strconv"
	"syscall"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"golang.org/x/sys/unix"
)

// ValidateStopOpts validates stop options.
// Returns error if timeout is negative or signal is invalid.
func ValidateStopOpts(opts *types.StopOpts) error {
	if opts.Timeout < 0 {
		return log.NewErrorf("the timeout = %d shouldn't be negative", opts.Timeout)
	}
	signal := ToSignal(opts.Signal)
	if signal < 1 || signal > 255 {
		return log.NewErrorf("invalid signal = %s", opts.Signal)
	}
	return nil
}

// ToSignal parses a string to syscall.Signal.
// Signals are accepted as both number and name: SIGKILL or 9.
// Returns the syscall.Signal for the provided number or name. Returns 0 if a signal with the provided name is not found.
func ToSignal(signal string) syscall.Signal {
	if signalNum, err := strconv.Atoi(signal); err == nil {
		return syscall.Signal(signalNum)
	}
	return unix.SignalNum(signal)
}
