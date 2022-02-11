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

package things

import (
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

type restartPolicyType string

const (
	always        restartPolicyType = "ALWAYS"
	onFailure     restartPolicyType = "ON_FAILURE"
	unlessStopped restartPolicyType = "UNLESS_STOPPED"
	no            restartPolicyType = "NO"
)

type restartPolicy struct {
	MaxRetryCount int               `json:"maxRetryCount,omitempty"`
	RetryTimeout  float64           `json:"retryTimeout,omitempty"`
	RpType        restartPolicyType `json:"type,omitempty"`
}

func toAPIRestartPolicy(internalRP *restartPolicy) *types.RestartPolicy {
	return &types.RestartPolicy{
		MaximumRetryCount: internalRP.MaxRetryCount,
		RetryTimeout:      time.Duration(internalRP.RetryTimeout) * time.Second,
		Type:              toAPIRPType(internalRP.RpType),
	}
}

func fromAPIRestartPolicy(apiPolicy *types.RestartPolicy) *restartPolicy {
	return &restartPolicy{
		MaxRetryCount: apiPolicy.MaximumRetryCount,
		RetryTimeout:  apiPolicy.RetryTimeout.Seconds(),
		RpType:        fromAPIRPType(apiPolicy.Type),
	}
}

func toAPIRPType(rpType restartPolicyType) types.PolicyType {
	switch rpType {
	case always:
		return types.Always
	case onFailure:
		return types.OnFailure
	case unlessStopped:
		return types.UnlessStopped
	case no:
		return types.No
	default:
		return types.PolicyType(rpType)
	}
}
func fromAPIRPType(apiRpType types.PolicyType) restartPolicyType {
	switch apiRpType {
	case types.Always:
		return always
	case types.OnFailure:
		return onFailure
	case types.UnlessStopped:
		return unlessStopped
	case types.No:
		return no
	default:
		return restartPolicyType(apiRpType)
	}
}
