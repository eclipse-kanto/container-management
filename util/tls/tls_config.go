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

package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
)

// TLSConfig represents the TLS configuration data
type TLSConfig struct {
	RootCA     string `json:"root_ca"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
}

// NewLocalTLSConfig initializes the Local broker TLS.
func NewLocalTLSConfig(tlsConfig TLSConfig) (*tls.Config, error) {
	caCertPool, err := NewCAPool(tlsConfig.RootCA)
	if err != nil {
		return nil, err
	}

	if len(tlsConfig.ClientCert) > 0 || len(tlsConfig.ClientKey) > 0 {
		return NewFSTLSConfig(caCertPool, tlsConfig.ClientCert, tlsConfig.ClientKey)
	}

	cfg := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       supportedCipherSuites(),
	}

	return cfg, nil
}

// NewCAPool opens a certificates pool.
func NewCAPool(caFile string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load CA")
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.Errorf("failed to parse CA %s", caFile)
	}

	return caCertPool, nil
}

// NewFSTLSConfig initializes a file Hub TLS.
func NewFSTLSConfig(caCertPool *x509.CertPool, certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load X509 key pair")
	}

	cfg := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       supportedCipherSuites(),
	}

	return cfg, nil
}

func supportedCipherSuites() []uint16 {
	cs := tls.CipherSuites()
	cid := make([]uint16, len(cs))
	for i := range cs {
		cid[i] = cs[i].ID
	}
	return cid
}
