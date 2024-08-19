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

// StopOpts represent options for stoping a container.
type StopOpts struct {

	// Timeout period in seconds to gracefully stop the container.
	Timeout int64 `json:"timeout,omitempty"`

	// Force determines whether a SIGKILL signal will be send to the container's process if it does not finish within the timeout specified.
	Force bool `json:"force,omitempty"`

	// Signal to be send to the container's process. Signal could be specified by using their names or numbers, e.g. SIGINT or 2.
	Signal string `json:"signal,omitempty"`

	RemoveCache bool `json:"removecache,omitempty"`
}
