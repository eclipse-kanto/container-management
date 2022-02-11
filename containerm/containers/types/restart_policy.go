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

import "time"

// PolicyType represents a container's policy type
type PolicyType string

// constants for the supported policy types
const (
	No            PolicyType = "no"
	Always        PolicyType = "always"
	UnlessStopped PolicyType = "unless-stopped"
	OnFailure     PolicyType = "on-failure"
)

// RestartPolicy represents a container's restart policy
type RestartPolicy struct {
	// maximum retry count
	MaximumRetryCount int `json:"maximum_retry_count"`

	// retry timeout in seconds
	RetryTimeout time.Duration `json:"retry_timeout"`

	// type
	Type PolicyType `json:"type"`
}
