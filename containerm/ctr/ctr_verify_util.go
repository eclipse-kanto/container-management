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

package ctr

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	sig "github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/payload"
)

var supportedHashFunctions = map[string]crypto.Hash{
	"sha224": crypto.SHA224,
	"sha256": crypto.SHA256,
	"sha384": crypto.SHA384,
	"sha512": crypto.SHA512,
}

func verifySignatures(signatures []signature, digest string, verifier sig.Verifier) error {
	verifyErrs := &errorUtil.CompoundError{}
	for _, sig := range signatures {
		err := verifySignature(sig, digest, verifier)
		if err != nil {
			verifyErrs.Append(err)
			continue
		}
		// successful verification
		return nil
	}
	return verifyErrs
}

func verifySignature(signature signature, digest string, verifier sig.Verifier) error {
	signatureDecoded, err := base64.StdEncoding.DecodeString(signature.base64)
	if err != nil {
		return err
	}
	if err = verifier.VerifySignature(bytes.NewReader(signatureDecoded), bytes.NewReader(signature.payload)); err != nil {
		return err
	}
	sci := &payload.SimpleContainerImage{}
	if err = json.Unmarshal(signature.payload, sci); err != nil {
		return err
	}

	foundDigest := sci.Critical.Image.DockerManifestDigest
	if foundDigest != digest {
		return log.NewErrorf("unexpected digest = %s", foundDigest)
	}
	return nil
}

func getSignatures(ctx context.Context, store content.Store, desc ocispec.Descriptor) ([]signature, error) {
	switch desc.MediaType {
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		// find manifest by platform
		indexRaw, err := content.ReadBlob(ctx, store, desc)
		if err != nil {
			return nil, err
		}

		var index ocispec.Index
		if err = json.Unmarshal(indexRaw, &index); err != nil {
			return nil, err
		}

		platform := platforms.DefaultSpec()
		matcher := platforms.NewMatcher(platform)
		for _, manifest := range index.Manifests {
			if matcher.Match(*manifest.Platform) {
				return getSignatures(ctx, store, desc)
			}
		}
		return nil, log.NewErrorf("could not get signatures, no match for platform = %v", platform)
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		// expected, keep going
	default:
		return nil, log.NewErrorf("could not get signatures, unexpected media type for signature image = %s, ", desc.MediaType)
	}

	manifestRaw, err := content.ReadBlob(ctx, store, desc)
	if err != nil {
		return nil, err
	}
	var manifest ocispec.Manifest
	if err = json.Unmarshal(manifestRaw, &manifest); err != nil {
		return nil, err
	}

	signatures := make([]signature, 0, len(manifest.Layers))
	for _, d := range manifest.Layers {
		base64, ok := d.Annotations[cosignSignatureAnnotationKey]
		if !ok {
			return nil, log.NewErrorf("no key annotation found in signature layer %s ", d.Digest)
		}

		var layer []byte
		layer, err = content.ReadBlob(ctx, store, d)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature{payload: layer, base64: base64})
	}
	return signatures, nil
}

func parseVerificationKeys(keys []string) ([]*verificationKey, error) {
	verificationKeys := make([]*verificationKey, len(keys))
	for i, key := range keys {
		pk, hf, err := util.ParseVerificationKey(key, supportedHashFunctions)
		if err != nil {
			return nil, err
		}
		verificationKeys[i] = &verificationKey{
			publicKey:    pk,
			hashFunction: hf,
		}
	}
	return verificationKeys, nil
}
