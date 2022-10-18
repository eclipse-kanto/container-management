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

package config_test

import (
	"crypto/tls"
	"testing"

	"github.com/eclipse-kanto/container-management/things/client/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCertificateSettingsOK(t *testing.T) {
	certFile := "testdata/certificate.pem"
	keyFile := "testdata/key.pem"

	use, err := config.NewFSTLSConfig(nil, "", "")

	require.Error(t, err)
	assert.Nil(t, use)

	use, err = config.NewFSTLSConfig(nil, certFile, keyFile)
	require.NoError(t, err)
	assert.True(t, len(use.Certificates) > 0)
	assert.True(t, len(use.CipherSuites) > 0)
	// assert that cipher suites identifiers are contained in tls.CipherSuites
	for _, csID := range use.CipherSuites {
		assert.True(t, func() bool {
			for _, cs := range tls.CipherSuites() {
				if cs.ID == csID {
					return true
				}
			}
			return false
		}())
	}
}

func TestUseCertificateSettingsFail(t *testing.T) {
	certFile := "testdata/certificate.pem"
	keyFile := "testdata/key.pem"
	nonExisting := "nonexisting.test"

	assertCertError(t, true, "", "")

	assertCertError(t, true, certFile, "")
	assertCertError(t, true, certFile, nonExisting)
	assertCertError(t, true, nonExisting, nonExisting)

	assertCertError(t, false, certFile, "")
	assertCertError(t, false, certFile, nonExisting)
	assertCertError(t, false, nonExisting, nonExisting)

	assertCertError(t, true, "", keyFile)
	assertCertError(t, true, nonExisting, keyFile)

	assertCertError(t, false, "", keyFile)
	assertCertError(t, false, nonExisting, keyFile)

	_, err := config.NewCAPool("tls_config.go")
	assert.Error(t, err)
}

func assertCertError(t *testing.T, useCertificate bool, certFile, keyFile string) {
	use, err := config.NewFSTLSConfig(nil, certFile, keyFile)
	assert.Error(t, err, useCertificate, certFile, keyFile)
	assert.Nil(t, use)
}
