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
	"encoding/json"

	"github.com/eclipse-kanto/container-management/things/api/model"
)

type clientLocalConfig struct {
	id       model.NamespacedID
	tenantID string
	policyID model.NamespacedID
}

func (lc *clientLocalConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonLocalClientConfig{
		DeviceID: lc.id.String(),
		TenantID: lc.tenantID,
		PolicyID: lc.policyID.String(),
	})
}

func (lc *clientLocalConfig) UnmarshalJSON(data []byte) error {
	var v = &jsonLocalClientConfig{}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	lc.id = NewNamespacedIDFromString(v.DeviceID)
	lc.tenantID = v.TenantID
	lc.policyID = NewNamespacedIDFromString(v.PolicyID)
	return nil
}
