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

package util

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"github.com/eclipse-kanto/container-management/containerm/log"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	"os"
	"strings"
)

const separator = ":"

// ParseVerificationKey parses verification key that consist of public key filename and optional hash
// function(e.g. sha512) separated by a colon after the filename. If a hash function is not included,
// then sha256 will be returned.
func ParseVerificationKey(key string, supportedHashFunc map[string]crypto.Hash) (crypto.PublicKey, crypto.Hash, error) {
	filename, hashFuncionStr, err := splitVerificationKey(key)
	if err != nil {
		return nil, crypto.Hash(0), err
	}

	var (
		hashFunction = crypto.SHA256
		publicKey    crypto.PublicKey
	)
	if len(hashFuncionStr) > 0 {
		if hashFunction, err = ParseHashFunc(hashFuncionStr, supportedHashFunc); err != nil {
			return nil, crypto.Hash(0), err
		}
	}
	publicKey, err = ParsePublicKey(filename)
	if err != nil {
		return nil, crypto.Hash(0), err
	}
	return publicKey, hashFunction, nil
}

func splitVerificationKey(key string) (string, string, error) {
	fileHashPair := strings.Split(strings.TrimSpace(key), separator)
	if len(fileHashPair) > 2 {
		return "", "", log.NewErrorf("invalid verification key - %s", key)
	}
	if len(fileHashPair) == 2 {
		return fileHashPair[0], fileHashPair[1], nil
	}
	return fileHashPair[0], "", nil
}

// ParsePublicKey parses PKIX or PKCS #1 public key encoded in PEM.
func ParsePublicKey(filename string) (crypto.PublicKey, error) {
	// PEM file
	pemBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	derBytes, _ := pem.Decode(pemBytes)
	if derBytes == nil {
		return nil, log.NewErrorf("PEM decoding failed for public key = %s", filename)
	}

	var (
		pubKey crypto.PublicKey
		errors = &errorUtil.CompoundError{}
	)
	pubKey, err = x509.ParsePKIXPublicKey(derBytes.Bytes)
	if err != nil {
		errors.Append(err)
		// try to use PKCS1
		pubKey, err = x509.ParsePKCS1PublicKey(derBytes.Bytes)
		if err != nil {
			errors.Append(err)
			return nil, errors
		}
	}
	return pubKey, nil
}

// ParseHashFunc parses a string representation of hash functions to crypto.Hash.
// Returns error if the hash function is not found within the supported or the hash function is not available.
func ParseHashFunc(hashFuncStr string, supportedHashFunc map[string]crypto.Hash) (crypto.Hash, error) {
	normalizedHashFunc := strings.ToLower(strings.TrimSpace(hashFuncStr))
	hashFunc, exists := supportedHashFunc[normalizedHashFunc]
	if !exists {
		return crypto.Hash(0), log.NewErrorf("unsupported hash function - %s", hashFuncStr)
	}
	if !hashFunc.Available() {
		return crypto.Hash(0), log.NewErrorf("hash function is not available - %s", hashFuncStr)
	}
	return hashFunc, nil
}
