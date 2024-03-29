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

package types

// IOConfig represents a container's IO configuration
type IOConfig struct {
	// Whether to attach to `stderr`.
	AttachStderr bool `json:"attach_stderr"`

	// Whether to attach to `stdin`.
	AttachStdin bool `json:"attach_stdin"`

	// Whether to attach to `stdout`.
	AttachStdout bool `json:"attach_stdout"`

	// Open `stdin`
	OpenStdin bool `json:"open_stdin"`

	// Close `stdin` after one attached client disconnects
	StdinOnce bool `json:"stdin_once"`

	// Attach standard streams to a TTY, including `stdin` if it is not closed.
	Tty bool `json:"tty"`
}
