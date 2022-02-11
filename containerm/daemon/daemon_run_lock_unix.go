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

//go:build linux || netbsd || openbsd || freebsd || darwin
// +build linux netbsd openbsd freebsd darwin

package main

import (
	"syscall"
)

const lockFileName = "lock"

func openFile(name string) (fd int, err error) {
	fd, err = syscall.Open(name, syscall.O_CREAT|syscall.O_RDONLY, 0600)
	return
}

func closeFile(file int) {
	syscall.Close(file)
}

func lockFile(file int) (err error) {
	err = syscall.Flock(file, syscall.LOCK_EX|syscall.LOCK_NB)
	return
}
