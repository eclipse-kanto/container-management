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
	"crypto/x509"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
)

const (
	fileExtRootCA        = ".crt"
	fileExtClientCert    = ".cert"
	fileExtClientCertKey = ".key"
)

func createDefaultTLSConfig(skipVerify bool) *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites:       supportedCipherSuites(),
		InsecureSkipVerify: skipVerify,
	}
}

func applyLocalTLSConfig(config *TLSConfig, tlsConfig *tls.Config) error {
	if err := validateTLSConfig(config); err != nil {
		log.ErrorErr(err, "invalid TLS configuration provided")
		return err
	}
	// load root CA
	systemPool, err := createSystemCertPool()
	if err != nil {
		log.ErrorErr(err, "unable to get system cert pool")
		return err
	}
	tlsConfig.RootCAs = systemPool

	data, err := ioutil.ReadFile(config.RootCA)
	if err != nil {
		return err
	}
	tlsConfig.RootCAs.AppendCertsFromPEM(data)

	//load client certificate-key pair
	cert, err := tls.LoadX509KeyPair(config.ClientCert, config.ClientKey)
	if err != nil {
		return err
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	return nil
}

func validateTLSConfig(config *TLSConfig) error {
	allErrs := &errorUtil.CompoundError{}
	if err := validateTLSConfigFile(config.RootCA, fileExtRootCA); err != nil {
		log.NewErrorf("problem accessing provided Root CA file %s", config.RootCA)
		allErrs.Append(err)
	}
	if err := validateTLSConfigFile(config.ClientCert, fileExtClientCert); err != nil {
		log.NewErrorf("problem accessing provided client certificate file %s", config.ClientCert)
		allErrs.Append(err)
	}
	if err := validateTLSConfigFile(config.ClientKey, fileExtClientCertKey); err != nil {
		log.NewErrorf("problem accessing provided client certificate key file %s", config.ClientKey)
		allErrs.Append(err)
	}
	if allErrs.Size() > 0 {
		return allErrs
	}
	return nil
}

func createSystemCertPool() (*x509.CertPool, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil && runtime.GOOS == "windows" {
		return x509.NewCertPool(), nil
	}
	return certPool, err
}

func validateTLSConfigFile(file, expectedFileExt string) error {
	if file == "" {
		return log.NewErrorf("TLS configuration data is missing")
	}
	if !filepath.IsAbs(file) {
		return log.NewErrorf("provided path must be absolute - %s", file)
	}
	if err := util.FileNotExistEmptyOrDir(file); err != nil {
		return err
	}
	if ext := filepath.Ext(file); ext != expectedFileExt {
		return log.NewErrorf("unsupported file format %s - must be %s", ext, expectedFileExt)
	}
	return nil
}

// excludes cipher suites with security issues
func supportedCipherSuites() []uint16 {
	cs := tls.CipherSuites()
	cid := make([]uint16, len(cs))
	for i := range cs {
		cid[i] = cs[i].ID
	}
	return cid
}
