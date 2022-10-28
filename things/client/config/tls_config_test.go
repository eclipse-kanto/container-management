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
	"errors"
	"fmt"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/client/config"
)

var (
	certFile    = "testdata/certificate.pem"
	keyFile     = "testdata/key.pem"
	nonExisting = "nonexisting.test"
)

func TestUseCertificateSettingsOK(t *testing.T) {
	use, err := config.NewFSTLSConfig(nil, "", "")

	testutil.AssertError(t, errors.New("failed to load X509 key pair: open : no such file or directory"), err)
	testutil.AssertNil(t, use)

	use, err = config.NewFSTLSConfig(nil, certFile, keyFile)
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, len(use.Certificates) > 0)
	testutil.AssertTrue(t, len(use.CipherSuites) > 0)
	// assert that cipher suites identifiers are contained in tls.CipherSuites
	for _, csID := range use.CipherSuites {
		testutil.AssertTrue(t, func() bool {
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
	expectedErrorStr := "failed to load X509 key pair: open %s: no such file or directory"

	assertCertError(t, "", "", fmt.Errorf(expectedErrorStr, ""))

	assertCertError(t, certFile, "", fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, certFile, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))
	assertCertError(t, nonExisting, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))

	assertCertError(t, certFile, "", fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, certFile, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))
	assertCertError(t, nonExisting, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))

	assertCertError(t, "", keyFile, fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, nonExisting, keyFile, fmt.Errorf(expectedErrorStr, nonExisting))

	assertCertError(t, "", keyFile, fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, nonExisting, keyFile, fmt.Errorf(expectedErrorStr, nonExisting))

	_, err := config.NewCAPool("tls_config.go")
	testutil.AssertError(t, errors.New("failed to parse CA tls_config.go"), err)
}

func assertCertError(t *testing.T, certFile, keyFile string, expectedErr error) {
	use, err := config.NewFSTLSConfig(nil, certFile, keyFile)
	testutil.AssertError(t, expectedErr, err)
	testutil.AssertNil(t, use)
}
