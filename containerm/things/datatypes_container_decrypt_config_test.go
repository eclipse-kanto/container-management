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

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestFromAPIDecryptConfig(t *testing.T) {
	apiDecryptConfig := &types.DecryptConfig{
		Keys:       []string{"key:pass"},
		Recipients: []string{"recipient"},
	}

	thingsDecryptConfig := fromAPIDecryptConfig(apiDecryptConfig)

	t.Run("test_from_api_decrypt_config_keys", func(t *testing.T) {
		testutil.AssertEqual(t, apiDecryptConfig.Keys, thingsDecryptConfig.Keys)
	})

	t.Run("test_from_api_decrypt_config_recipients", func(t *testing.T) {
		testutil.AssertEqual(t, apiDecryptConfig.Recipients, thingsDecryptConfig.Recipients)
	})
}

func TestToAPIDecryptConfig(t *testing.T) {
	thingsDecryptConfig := &decryptConfig{
		Keys:       []string{"key:pass"},
		Recipients: []string{"recipient"},
	}

	apiDecryptConfig := toAPIDecryptConfig(thingsDecryptConfig)

	t.Run("test_from_api_decrypt_config_keys", func(t *testing.T) {
		testutil.AssertEqual(t, thingsDecryptConfig.Keys, apiDecryptConfig.Keys)
	})

	t.Run("test_from_api_decrypt_config_recipients", func(t *testing.T) {
		testutil.AssertEqual(t, thingsDecryptConfig.Recipients, apiDecryptConfig.Recipients)
	})
}
