// Copyright (c) 2021 Contributors to the Eclipse Foundation
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

package ctr

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/pkg/errors"
)

func TestNewContainerImageRegistriesResolver(t *testing.T) {
	const (
		insecureHostName = "test_host_insecure"

		secureHostName = "test_host_secure"
		certDir        = "../pkg/testutil/certs"
	)

	var (
		rootCertFile, _ = filepath.Abs(filepath.Join(certDir, "testRootCert.crt"))
		certFile, _     = filepath.Abs(filepath.Join(certDir, "testClientCert.cert"))
		keyFile, _      = filepath.Abs(filepath.Join(certDir, "testClientKey.key"))
	)
	tests := map[string]struct {
		createConfig func() map[string]*RegistryConfig
		assertHosts  func(map[string][]docker.RegistryHost)
	}{
		"test_null_configs": {
			createConfig: func() map[string]*RegistryConfig {
				return nil
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 0, len(hosts))
			},
		},
		"test_empty_configs": {
			createConfig: func() map[string]*RegistryConfig {
				return make(map[string]*RegistryConfig)
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 0, len(hosts))
			},
		},
		"test_one_insecure_config": {
			createConfig: func() map[string]*RegistryConfig {
				result := make(map[string]*RegistryConfig, 1)
				result[insecureHostName] = &RegistryConfig{
					IsInsecure: true,
				}
				return result
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 1, len(hosts))
				assertInsecureRegistryHosts(t, insecureHostName, hosts[insecureHostName], false)
			},
		},
		"test_one_secure_config": {
			createConfig: func() map[string]*RegistryConfig {
				result := make(map[string]*RegistryConfig, 1)
				result[secureHostName] = &RegistryConfig{
					IsInsecure: false,
					Credentials: &AuthCredentials{
						UserID:   "testID",
						Password: "testPassword",
					},
				}
				return result
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 1, len(hosts))
				assertSecureRegistryHosts(t, secureHostName, hosts[secureHostName], true, true)
			},
		},
		"test_one_secure_config_with_invalid_transport": {
			createConfig: func() map[string]*RegistryConfig {
				result := make(map[string]*RegistryConfig, 1)
				result[secureHostName] = &RegistryConfig{
					IsInsecure: false,
					Transport:  &TLSConfig{},
				}
				return result
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 1, len(hosts))
				assertSecureRegistryHosts(t, secureHostName, hosts[secureHostName], false, true)
			},
		},
		"test_one_secure_config_with_transport": {
			createConfig: func() map[string]*RegistryConfig {
				result := make(map[string]*RegistryConfig, 1)
				result[secureHostName] = &RegistryConfig{
					IsInsecure: false,
					Transport: &TLSConfig{
						RootCA:     rootCertFile,
						ClientCert: certFile,
						ClientKey:  keyFile,
					},
				}
				return result
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 1, len(hosts))
				assertSecureRegistryHosts(t, secureHostName, hosts[secureHostName], false, false)
			},
		},
		"test_many_configs": {
			createConfig: func() map[string]*RegistryConfig {
				result := make(map[string]*RegistryConfig, 2)
				result[secureHostName] = &RegistryConfig{
					IsInsecure: false,
				}
				result[insecureHostName] = &RegistryConfig{
					IsInsecure: true,
					Credentials: &AuthCredentials{
						UserID:   "testID",
						Password: "testPassword",
					},
				}
				return result
			},
			assertHosts: func(hosts map[string][]docker.RegistryHost) {
				testutil.AssertEqual(t, 2, len(hosts))
				assertSecureRegistryHosts(t, secureHostName, hosts[secureHostName], false, true)
				assertInsecureRegistryHosts(t, insecureHostName, hosts[insecureHostName], true)
			},
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			configs := testCase.createConfig()
			imageRegResolver := newContainerImageRegistriesResolver(configs).(*ctrImagesResolver)
			testutil.AssertEqual(t, configs, imageRegResolver.registryConfigurations)
			testCase.assertHosts(imageRegResolver.registryHosts)
		})
	}
}

func TestResolveImageRegistry(t *testing.T) {
	hostName := "test_host"
	configurations := make(map[string]*RegistryConfig)
	configurations[hostName] = &RegistryConfig{}
	hosts := make(map[string][]docker.RegistryHost)
	hosts[hostName] = []docker.RegistryHost{
		{
			Host: hostName,
		},
	}
	imageRegResolver := &ctrImagesResolver{
		registryConfigurations: configurations,
		registryHosts:          hosts,
	}

	tests := map[string]struct {
		hostName string
		isNull   bool
	}{
		"test_not_existing_host": {
			hostName: "not_existing",
			isNull:   true,
		},
		"test_existing_host": {
			hostName: "test_host",
			isNull:   false,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := imageRegResolver.ResolveImageRegistry(testCase.hostName)
			testutil.AssertEqual(t, testCase.isNull, actual == nil)
		})
	}
}

func TestGetRegistryHosts(t *testing.T) {
	hostName := "test_host"
	testRegistryHost := docker.RegistryHost{Host: hostName}
	hosts := make(map[string][]docker.RegistryHost)
	hosts[hostName] = []docker.RegistryHost{testRegistryHost}
	imageRegResolver := &ctrImagesResolver{
		registryHosts: hosts,
	}

	tests := map[string]struct {
		hostName      string
		expectedHosts []docker.RegistryHost
		expectedError error
	}{
		"test_not_existing_host": {
			hostName:      "test_not_existing_host",
			expectedError: errors.New("no registry hosts found for host [test_not_existing_host]"),
		},
		"test_host": {
			hostName:      hostName,
			expectedHosts: []docker.RegistryHost{testRegistryHost},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actualHosts, actualErr := imageRegResolver.getRegistryHosts(testCase.hostName)
			testutil.AssertEqual(t, testCase.expectedHosts, actualHosts)
			testutil.AssertError(t, testCase.expectedError, actualErr)
		})
	}
}

func TestGetRegistryAuthCreds(t *testing.T) {
	testUserID := "testID"
	testPassword := "testPassword"

	configurations := make(map[string]*RegistryConfig)
	configurations["test_host_withCredentials"] = &RegistryConfig{
		Credentials: &AuthCredentials{
			UserID:   testUserID,
			Password: testPassword,
		},
	}
	configurations["test_host_withoutCredentials"] = &RegistryConfig{}
	imageRegResolver := &ctrImagesResolver{
		registryConfigurations: configurations,
	}

	tests := map[string]struct {
		hostName       string
		expectedUserID string
		expectedPass   string
		expectedError  error
	}{
		"test_not_existing_credentials": {
			hostName:      "test_host_withoutCredentials",
			expectedError: errors.New("no credentials could be found for registry host test_host_withoutCredentials"),
		},
		"test_existing_credentials": {
			hostName:       "test_host_withCredentials",
			expectedUserID: testUserID,
			expectedPass:   testPassword,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actualID, actualPass, actualErr := imageRegResolver.getRegistryAuthCreds(testCase.hostName)
			testutil.AssertEqual(t, testCase.expectedUserID, actualID)
			testutil.AssertEqual(t, testCase.expectedPass, actualPass)
			testutil.AssertError(t, testCase.expectedError, actualErr)
		})
	}
}

func assertSecureRegistryHosts(t *testing.T, hostName string, registryHosts []docker.RegistryHost, hasAutorizer bool, defaultTLS bool) {
	testutil.AssertEqual(t, 1, len(registryHosts))
	assertSecureRegistryHost(t, hostName, registryHosts[0], hasAutorizer, false, defaultTLS)
}

func assertSecureRegistryHost(t *testing.T, hostName string, registryHost docker.RegistryHost, hasAutorizer bool, skipVerify bool, defaultTLS bool) {
	assertClientTransport(t, registryHost.Client, skipVerify, defaultTLS)
	testutil.AssertEqual(t, hostName, registryHost.Host)
	testutil.AssertEqual(t, registryHostSchemeHTTPS, registryHost.Scheme)
	testutil.AssertEqual(t, registryHostPathV2, registryHost.Path)
	testutil.AssertEqual(t, registryHostCapabilitiesDefault, registryHost.Capabilities)

	if hasAutorizer {
		testutil.AssertNotNil(t, registryHost.Authorizer)
	} else {
		testutil.AssertNil(t, registryHost.Authorizer)
	}
}
func assertInsecureRegistryHosts(t *testing.T, hostName string, registryHosts []docker.RegistryHost, hasAutorizer bool) {
	testutil.AssertEqual(t, 2, len(registryHosts))
	assertInsecureRegistryHost(t, hostName, registryHosts[0], hasAutorizer)
	assertSecureRegistryHost(t, hostName, registryHosts[1], hasAutorizer, true, true)
}

func assertInsecureRegistryHost(t *testing.T, hostName string, registryHost docker.RegistryHost, hasAutorizer bool) {
	testutil.AssertEqual(t, http.DefaultClient, registryHost.Client)
	testutil.AssertEqual(t, hostName, registryHost.Host)
	testutil.AssertEqual(t, registryHostSchemeHTTP, registryHost.Scheme)
	testutil.AssertEqual(t, registryHostPathV2, registryHost.Path)
	testutil.AssertEqual(t, registryHostCapabilitiesDefault, registryHost.Capabilities)

	if hasAutorizer {
		testutil.AssertNotNil(t, registryHost.Authorizer)
	} else {
		testutil.AssertNil(t, registryHost.Authorizer)
	}
}

func assertClientTransport(t *testing.T, client *http.Client, skipVerify bool, defaultTLS bool) {
	testutil.AssertNotNil(t, client)

	transport := client.Transport.(*http.Transport)
	testutil.AssertNotNil(t, transport)
	if defaultTLS {
		testutil.AssertEqual(t, createDefaultTLSConfig(skipVerify), transport.TLSClientConfig)
	} else {
		testutil.AssertNotEqual(t, createDefaultTLSConfig(skipVerify), transport.TLSClientConfig)
	}

	testutil.AssertEqual(t, registryResolverTransportMaxIdeConns, transport.MaxIdleConns)
	testutil.AssertEqual(t, registryResolverTransportIdleConnTimeout, transport.IdleConnTimeout)
	testutil.AssertEqual(t, registryResolverTransportTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
	testutil.AssertEqual(t, registryResolverTransportExpectContinueTimeout, transport.ExpectContinueTimeout)
}
