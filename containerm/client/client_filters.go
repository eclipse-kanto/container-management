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

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

// WithName filters the containers that match a given name
func WithName(name string) Filter {
	return func(container *types.Container) bool {
		return container.Name == name
	}
}
