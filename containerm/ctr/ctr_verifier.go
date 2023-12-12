// Copyright (c) 2023 Contributors to the Eclipse Foundation
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
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	// VerifierNone is a VerifierType denoting that no verification will be performed
	VerifierNone = VerifierType("none")
	// VerifierNotation is a VerifierType denoting that verification will be performed with notation
	VerifierNotation      = VerifierType("notation")
	notationKeyConfigDir  = "configDir"
	notationKeyLibexecDir = "libexecDir"
)

// VerifierType  image verifier type - possible values are none and notation, when set to none image signatures wil not be verified.
type VerifierType string

type containerVerifier interface {
	Verify(context.Context, types.Image) error
}

func newContainerVerifier(verifierType VerifierType, verifierConfig map[string]string, registryConfig map[string]*RegistryConfig) (containerVerifier, error) {
	switch verifierType {
	case VerifierNone:
		return &skipVerifier{}, nil
	case VerifierNotation:
		return newNotationVerifier(verifierConfig, registryConfig)
	default:
		return nil, log.NewErrorf("unknown verifier type - %s", verifierType)
	}
}

type skipVerifier struct{}

func (*skipVerifier) Verify(_ context.Context, _ types.Image) error {
	return nil
}
