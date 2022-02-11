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

package types

// UpdateOpts represent options for updating a container.
type UpdateOpts struct {

	// RestartPolicy to be used for the container.
	RestartPolicy *RestartPolicy `json:"restart_policy"`

	// Resources of the container.
	Resources *Resources `json:"resources"`
}
