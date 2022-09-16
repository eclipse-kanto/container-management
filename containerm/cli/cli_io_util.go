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

import (
	"github.com/eclipse-kanto/container-management/containerm/log"
	"golang.org/x/crypto/ssh/terminal"
)

type terminalManager interface {
	CheckTty(attachStdin, ttyMode bool, fd uintptr) error
	SetRawMode(stdin, stdout bool) (*terminal.State, *terminal.State, error)
	RestoreMode(in, out *terminal.State) error
}
type termMgr struct{}

// CheckTty checks if we are trying to attach to a container tty
// from a non-tty client input stream, and if so, returns an error.
func (tm *termMgr) CheckTty(attachStdin, ttyMode bool, fd uintptr) error {
	if ttyMode && attachStdin && !terminal.IsTerminal(int(fd)) {
		return log.NewError("the input device is not a TTY")
	}
	return nil
}

func (tm *termMgr) SetRawMode(stdin, stdout bool) (*terminal.State, *terminal.State, error) {
	var (
		in  *terminal.State
		out *terminal.State
		err error
	)

	if stdin {
		if in, err = terminal.MakeRaw(0); err != nil {
			return nil, nil, err
		}
	}
	if stdout {
		if out, err = terminal.MakeRaw(1); err != nil {
			return nil, nil, err
		}
	}

	return in, out, nil
}

func (tm *termMgr) RestoreMode(in, out *terminal.State) error {
	if in != nil {
		if err := terminal.Restore(0, in); err != nil {
			return err
		}
	}
	if out != nil {
		if err := terminal.Restore(1, out); err != nil {
			return err
		}
	}
	return nil
}
