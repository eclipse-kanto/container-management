// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package deployment

import (
	"context"
)

// Mode indicates available deployment modes
type Mode string

const (
	// InitialDeployMode means that the deployment service will deploy new containers only on initial start of container management
	InitialDeployMode Mode = "init"
	// UpdateMode means that the deployment service will deploy new containers and/or update existing containers on each start of container management
	UpdateMode Mode = "update"
)

// Manager represents the container deployment manager abstraction
type Manager interface {

	// Deploy initially deploys or updates containers described in configured local path
	Deploy(ctx context.Context) error

	// Dispose stops running deployments
	Dispose(ctx context.Context) error
}
