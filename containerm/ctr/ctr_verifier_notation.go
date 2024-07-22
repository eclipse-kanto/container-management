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
	"fmt"
	"net/http"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orasregistry "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// do not remove: support for verifying those types is added in init() functions in each corresponding packet, it seems that
// the packets does no go in the binary, through indirect dependencies, so add them explicitly otherwise verification fails.
var supportedMediaTypes = []string{jws.MediaTypeEnvelope, cose.MediaTypeEnvelope}

type notationVerifier struct {
	registryConfig map[string]*RegistryConfig
}

func newNotationVerifier(config map[string]string, registryConfig map[string]*RegistryConfig) (containerVerifier, error) {
	// set up notation configuration and library execution directories
	if value, ok := config[notationKeyConfigDir]; ok {
		dir.UserConfigDir = value
	}
	if value, ok := config[notationKeyLibexecDir]; ok {
		dir.UserLibexecDir = value
	}
	return &notationVerifier{
		registryConfig: registryConfig,
	}, nil
}

func (nv *notationVerifier) Verify(ctx context.Context, imageInfo types.Image) error {
	var (
		verifyOpts  = notation.VerifyOptions{MaxSignatureAttempts: 50}
		sigVerifier notation.Verifier
		repo        registry.Repository
		err         error
	)

	if sigVerifier, err = verifier.NewFromConfig(); err != nil {
		return err
	}
	if repo, err = getRepository(imageInfo.Name, nv.registryConfig); err != nil {
		return err
	}
	if _, verifyOpts.ArtifactReference, err = resolveReference(ctx, imageInfo.Name, repo); err != nil {
		return err
	}

	_, outcomes, err := notation.Verify(ctx, sigVerifier, repo, verifyOpts)
	if err != nil {
		return err
	} else if len(outcomes) == 0 {
		return log.NewErrorf("signature verification failed for all signatures of %s", imageInfo.Name)
	}

	outcome := outcomes[0]
	for _, result := range outcomes[0].VerificationResults {
		if result.Error != nil {
			if result.Action == trustpolicy.ActionLog {
				log.WarnErr(result.Error, "%s verification failed", result.Type)
			}
		}
	}

	if outcome.VerificationLevel.Name == trustpolicy.LevelSkip.Name {
		log.Info("signature verification is skipped for %s", imageInfo.Name)
	} else {
		log.Info("signature verification is successful for %s", imageInfo.Name)
	}
	return nil
}

func resolveReference(ctx context.Context, reference string, repo registry.Repository) (ocispec.Descriptor, string, error) {
	var (
		ref          orasregistry.Reference
		manifestDesc ocispec.Descriptor
		err          error
	)

	if ref, err = orasregistry.ParseReference(reference); err != nil {
		return ocispec.Descriptor{}, "", err
	}
	if manifestDesc, err = repo.Resolve(ctx, reference); err != nil {
		return ocispec.Descriptor{}, "", err
	}

	resolvedRef := fmt.Sprintf("%s/%s@%s", ref.Registry, ref.Repository, manifestDesc.Digest.String())
	if _, err := digest.Parse(ref.Reference); err != nil {
		log.Warn("image %s is provided using a tag, tags are mutable, using a digest is the preferred way when verifying a signature", reference)
		return manifestDesc, resolvedRef, nil
	}

	if ref.Reference != manifestDesc.Digest.String() {
		return ocispec.Descriptor{}, "", log.NewErrorf("provided digest %s does not match the resolved digest %s", ref.Reference, manifestDesc.Digest.String())
	}
	return manifestDesc, resolvedRef, nil
}

func getRepository(reference string, registryConfigs map[string]*RegistryConfig) (registry.Repository, error) {
	var (
		err  error
		repo = &remote.Repository{}
	)

	if repo.Reference, err = orasregistry.ParseReference(reference); err != nil {
		return nil, err
	}
	if repo.Client, repo.PlainHTTP, err = getAuthClient(repo.Reference, registryConfigs); err != nil {
		return nil, err
	}
	return registry.NewRepository(repo), nil
}

func getAuthClient(ref orasregistry.Reference, registryConfigs map[string]*RegistryConfig) (*auth.Client, bool, error) {
	authClient := &auth.Client{
		Cache:  auth.NewCache(),
		Header: auth.DefaultClient.Header.Clone(),
	}

	config, ok := registryConfigs[ref.Registry]
	if !ok {
		return authClient, false, nil
	}
	if config.Credentials != nil {
		authClient.Credential = auth.StaticCredential(ref.Host(), auth.Credential{
			Username: config.Credentials.UserID,
			Password: config.Credentials.Password,
		})
	}
	if !config.IsInsecure {
		authClient.Client = &http.Client{Transport: getTransport(false, config.Transport, ref.Registry)}
	}

	return authClient, config.IsInsecure, nil
}
