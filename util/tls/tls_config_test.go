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

func TestCertificateSettings(t *testing.T) {
	const (
		failedToLoadError = "failed to load X509 key pair: open %s: no such file or directory"
		noSuchCAFileError = "failed to load CA: open %s: no such file or directory"
		nonExisting       = "nonexisting.test"
		invalidFile       = "testdata/invalid.pem"
		certFile          = "testdata/certificate.pem"
		keyFile           = "testdata/key.pem"
		caFile            = "testdata/ca.crt"
	)

	testCases := map[string]struct {
		CAFile        string
		KeyFile       string
		CertFile      string
		ExpectedError error
	}{
		"valid_config_no_credentials":     {CAFile: caFile, KeyFile: "", CertFile: "", ExpectedError: nil},
		"valid_config_with_credentials":   {CAFile: caFile, KeyFile: keyFile, CertFile: certFile, ExpectedError: nil},
		"no_files_provided":               {CAFile: "", KeyFile: "", CertFile: "", ExpectedError: fmt.Errorf(noSuchCAFileError, "")},
		"no_ca_file_provided":             {CAFile: "", KeyFile: keyFile, CertFile: certFile, ExpectedError: fmt.Errorf(noSuchCAFileError, "")},
		"invalid_ca_file_arg":             {CAFile: "\\\000", KeyFile: keyFile, CertFile: certFile, ExpectedError: errors.New("failed to load CA: open \\\000: invalid argument")},
		"invalid_ca_file":                 {CAFile: invalidFile, KeyFile: keyFile, CertFile: certFile, ExpectedError: fmt.Errorf("failed to parse CA %s", invalidFile)},
		"no_key_file_provided":            {CAFile: caFile, KeyFile: "", CertFile: certFile, ExpectedError: fmt.Errorf(failedToLoadError, "")},
		"no_cert_file_provided":           {CAFile: caFile, KeyFile: keyFile, CertFile: "", ExpectedError: fmt.Errorf(failedToLoadError, "")},
		"non_existing_ca_file":            {CAFile: nonExisting, KeyFile: nonExisting, CertFile: certFile, ExpectedError: fmt.Errorf(noSuchCAFileError, nonExisting)},
		"non_existing_key_file":           {CAFile: caFile, KeyFile: nonExisting, CertFile: certFile, ExpectedError: fmt.Errorf(failedToLoadError, nonExisting)},
		"non_existing_cert_file":          {CAFile: caFile, KeyFile: keyFile, CertFile: nonExisting, ExpectedError: fmt.Errorf(failedToLoadError, nonExisting)},
		"non_existing_key_and_cert_files": {CAFile: caFile, KeyFile: nonExisting, CertFile: nonExisting, ExpectedError: fmt.Errorf(failedToLoadError, nonExisting)},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			cfg, err := NewConfig(testCase.CAFile, testCase.CertFile, testCase.KeyFile)
			if testCase.ExpectedError != nil {
				if testCase.ExpectedError.Error() != err.Error() {
					t.Fatalf("expected error : %s, got: %s", testCase.ExpectedError, err)
				}
				if cfg != nil {
					t.Fatalf("expected nil, got: %v", cfg)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if len(cfg.Certificates) == 0 && testCase.CertFile != "" && testCase.KeyFile != "" {
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
		})
	}
}
