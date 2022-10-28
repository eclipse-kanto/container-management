// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

package things

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type decryptionConfiguration struct {
	Keys       []string `json:"keys,omitempty"`
	Recipients []string `json:"recipients,omitempty"`
}

func fromAPIDecryptConfig(apiDev *types.DecryptConfig) *decryptionConfiguration {
	return &decryptionConfiguration{
		Keys:       apiDev.Keys,
		Recipients: apiDev.Recipients,
	}
}

func toAPIDecryptConfig(internalDev *decryptionConfiguration) *types.DecryptConfig {
	return &types.DecryptConfig{
		Keys:       internalDev.Keys,
		Recipients: internalDev.Recipients,
	}
}
