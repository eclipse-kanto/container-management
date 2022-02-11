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

// DeviceMapping represents a device mapping between the host and the container
type DeviceMapping struct {
	PathOnHost        string `json:"path_on_host"`
	PathInContainer   string `json:"path_in_container"`
	CgroupPermissions string `json:"cgroup_permissions"` //rwm
}
