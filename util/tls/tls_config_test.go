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

package tls

import (
	"crypto/tls"
	"errors"
	"fmt"
	"testing"
)

var (
	certFile = "testdata/certificate.pem"
	keyFile  = "testdata/key.pem"
)

func TestNewConfig(t *testing.T) {
	cfg, err := NewConfig("testdata/ca.crt", certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	testTLSConfig(t, cfg)
}

func TestUseCertificateSettingsOK(t *testing.T) {
	cfg, err := newFSConfig(nil, certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	testTLSConfig(t, cfg)
}

func TestUseCertificateSettingsFail(t *testing.T) {
	expectedErrorStr := "failed to load X509 key pair: open %s: no such file or directory"
	nonExisting := "nonexisting.test"

	assertCertError(t, "", "", fmt.Errorf(expectedErrorStr, ""))

	assertCertError(t, certFile, "", fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, certFile, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))
	assertCertError(t, nonExisting, nonExisting, fmt.Errorf(expectedErrorStr, nonExisting))

	assertCertError(t, "", keyFile, fmt.Errorf(expectedErrorStr, ""))
	assertCertError(t, nonExisting, keyFile, fmt.Errorf(expectedErrorStr, nonExisting))

	expectedErr := errors.New("failed to parse CA testdata/invalid.pem")
	_, err := newCAPool("testdata/invalid.pem")
	if expectedErr.Error() != err.Error() {
		t.Fatalf("expected error : %s, got: %s", expectedErr, err)
	}

	expectedErr = errors.New("failed to load CA: open \\\000: invalid argument")
	_, err = newCAPool("\\\000")
	if expectedErr.Error() != err.Error() {
		t.Fatalf("expected error : %s, got: %s", expectedErr, err)
	}
}

func assertCertError(t *testing.T, certFile, keyFile string, expectedErr error) {
	use, err := newFSConfig(nil, certFile, keyFile)
	if expectedErr.Error() != err.Error() {
		t.Fatalf("expected error : %s, got: %s", expectedErr, err)
	}
	if use != nil {
		t.Fatalf("expected nil, got: %v", use)
	}
}

func testTLSConfig(t *testing.T, cfg *tls.Config) {
	if len(cfg.Certificates) == 0 {
		t.Fatal("certificates length must not be 0")
	}
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("cipher suites length must not be 0")
	}
	// assert that cipher suites identifiers are contained in tls.CipherSuites
	for _, csID := range cfg.CipherSuites {
		if !func() bool {
			for _, cs := range tls.CipherSuites() {
				if cs.ID == csID {
					return true
				}
			}
			return false
		}() {
			t.Fatalf("cipher suite %d is not implemented", csID)
		}
	}
	if cfg.InsecureSkipVerify {
		t.Fatal("skip verify is set to true")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("invalid min TLS version %d", cfg.MinVersion)
	}
	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Fatalf("invalid max TLS version %d", cfg.MaxVersion)
	}
}
