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

package protocol

// Envelope represents the content of the messages
type Envelope struct {
	Topic     string      `json:"topic"`
	Headers   Headers     `json:"headers"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value,omitempty"`
	Status    int         `json:"status,omitempty"`
	Revision  int         `json:"revision,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}
