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

// Manager represents the container deployment manager abstraction
type Manager interface {

	// InitialDeploy initially deploys containers described in configured local path
	InitialDeploy(ctx context.Context) error

	// Dispose stops running deployments
	Dispose(ctx context.Context) error
}
