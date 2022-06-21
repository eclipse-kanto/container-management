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

package ctr

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testFilePathNotAbsolute = "not-absolute-path"
	testEmptyCertFile       = "../pkg/testutil/certs/emptyTestCertFile.crt"
	testRootCertPath        = "../pkg/testutil/certs/testRootCert.crt"
	testClientCertPath      = "../pkg/testutil/certs/testClientCert.cert"
	testClientKeyPath       = "../pkg/testutil/certs/testClientKey.key"
)

func TestCreateDefaultTLSConfig(t *testing.T) {
	t.Run("test_create_default_tls_config", func(t *testing.T) {
		var testDefaultConfig = &tls.Config{}
		testDefaultConfig = createDefaultTLSConfig(false)
		testutil.AssertEqual(t, uint16(tls.VersionTLS12), testDefaultConfig.MinVersion)
		testutil.AssertEqual(t, uint16(tls.VersionTLS13), testDefaultConfig.MaxVersion)
		testutil.AssertFalse(t, testDefaultConfig.InsecureSkipVerify)
		testutil.AssertTrue(t, len(testDefaultConfig.CipherSuites) > 0)
		// assert that cipher suites identifiers are contained in tls.CipherSuites
		for _, csID := range testDefaultConfig.CipherSuites {
			testutil.AssertTrue(t, func() bool {
				for _, cs := range tls.CipherSuites() {
					if cs.ID == csID {
						return true
					}
				}
				return false
			}())
		}
	})
}

type testValidateTLSConfigFileArgs struct {
	file     string
	fileExt  string
	expError error
}

func TestValidateTLSConfigFile(t *testing.T) {
	testEmptyCertFileAbsPath, _ := filepath.Abs(testEmptyCertFile)
	testRootCertAbsPath, _ := filepath.Abs(testRootCertPath)
	ext := filepath.Ext(testRootCertAbsPath)
	tests := map[string]struct {
		args testValidateTLSConfigFileArgs
	}{
		"test_validate_tls_config_file_empty_file_name": {
			args: testValidateTLSConfigFileArgs{
				file:     "",
				fileExt:  fileExtRootCA,
				expError: log.NewErrorf("TLS configuration data is missing"),
			},
		},
		"test_validate_tls_config_non_absolute_filepath": {
			args: testValidateTLSConfigFileArgs{
				file:     testFilePathNotAbsolute,
				fileExt:  fileExtRootCA,
				expError: log.NewErrorf("provided path must be absolute - " + testFilePathNotAbsolute),
			},
		},
		"test_validate_tls_config_empty_file": {
			args: testValidateTLSConfigFileArgs{
				file:     testEmptyCertFileAbsPath,
				fileExt:  fileExtRootCA,
				expError: log.NewErrorf("file " + testEmptyCertFileAbsPath + " is empty"),
			},
		},
		"test_validate_tls_wrong_file_ext": {
			args: testValidateTLSConfigFileArgs{
				file:     testRootCertAbsPath,
				fileExt:  fileExtClientCert,
				expError: log.NewErrorf("unsupported file format " + ext + " - must be " + fileExtClientCert),
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			err := validateTLSConfigFile(testCase.args.file, testCase.args.fileExt)
			testutil.AssertError(t, testCase.args.expError, err)
		})
	}
}

type testTLSConfigArgs struct {
	config   *TLSConfig
	expError error
}

func TestValidateTLSConfig(t *testing.T) {
	testRootCertAbsPath, _ := filepath.Abs(testRootCertPath)
	testClientCertAbsPath, _ := filepath.Abs(testClientCertPath)
	testClientKeyAbsPath, _ := filepath.Abs(testClientKeyPath)

	tests := map[string]struct {
		args testTLSConfigArgs
	}{
		"test_validate_tls_config": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testRootCertAbsPath,
					ClientCert: testClientCertAbsPath,
					ClientKey:  testClientKeyAbsPath,
				},
				expError: nil,
			},
		},
		"test_validate_tls_config_root_cert_wrong_format": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testClientCertAbsPath,
					ClientCert: testClientCertAbsPath,
					ClientKey:  testClientKeyAbsPath,
				},
				expError: log.NewErrorf("unsupported file format " + fileExtClientCert + " - must be " + fileExtRootCA),
			},
		},
		"test_validate_tls_config_client_cert_wrong_format": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testRootCertAbsPath,
					ClientCert: testClientKeyAbsPath,
					ClientKey:  testClientKeyAbsPath,
				},
				expError: log.NewErrorf("unsupported file format " + fileExtClientCertKey + " - must be " + fileExtClientCert),
			},
		},
		"test_validate_tls_config_client_cert_key_wrong_format": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testRootCertAbsPath,
					ClientCert: testClientCertAbsPath,
					ClientKey:  testRootCertAbsPath,
				},
				expError: log.NewErrorf("unsupported file format " + fileExtRootCA + " - must be " + fileExtClientCertKey),
			},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			err := validateTLSConfig(testCase.args.config)
			testutil.AssertError(t, testCase.args.expError, err)
		})
	}
}

func TestApplyTLSConfig(t *testing.T) {
	defaultTLSConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       supportedCipherSuites(),
		InsecureSkipVerify: false,
	}
	testRootCertAbsPath, _ := filepath.Abs(testRootCertPath)
	testClientCertAbsPath, _ := filepath.Abs(testClientCertPath)
	testClientKeyAbsPath, _ := filepath.Abs(testClientKeyPath)
	tests := map[string]struct {
		args testTLSConfigArgs
	}{
		"test_apply_tls_config_no_err": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testRootCertAbsPath,
					ClientCert: testClientCertAbsPath,
					ClientKey:  testClientKeyAbsPath,
				},
				expError: nil,
			},
		},
		"test_apply_tls_config_err": {
			args: testTLSConfigArgs{
				config: &TLSConfig{
					RootCA:     testClientCertAbsPath,
					ClientCert: testClientCertAbsPath,
					ClientKey:  testClientKeyAbsPath,
				},
				expError: log.NewErrorf("unsupported file format " + fileExtClientCert + " - must be " + fileExtRootCA),
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			err := applyLocalTLSConfig(testCase.args.config, defaultTLSConfig)
			testutil.AssertError(t, testCase.args.expError, err)
		})
	}
}
