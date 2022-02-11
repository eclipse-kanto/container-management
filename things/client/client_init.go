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

package client

import (
	"github.com/eclipse-kanto/container-management/things/api/model"
)

// NewClient creates a new client from the provided configuration
func NewClient(cfg *Configuration) *Client {

	client := &Client{
		cfg:    cfg,
		things: make(map[model.NamespacedID]*thing),
	}

	return client

}
