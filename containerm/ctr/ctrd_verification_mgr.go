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
	"context"
	"crypto"
	"fmt"
	"strings"

	"github.com/containerd/containerd"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	sig "github.com/sigstore/sigstore/pkg/signature"
)

type verificationKey struct {
	publicKey    crypto.PublicKey
	hashFunction crypto.Hash
}

type containerVerificationMgr interface {
	GetVerificationKeys(config *types.VerificationConfig) ([]*verificationKey, error)
	GetSignatureReference(image containerd.Image) (string, error)
	VerifySignature(ctx context.Context, image, signatureImage containerd.Image, keys ...*verificationKey) error
}

const (
	cosignSignatureAnnotationKey = "dev.cosignproject.cosign/signature"
	cosignSignatureSuffix        = ".sig"
	tagDelimiter                 = ":"
	digestDelimiter              = "@"
)

type ctrVerificationMgr struct {
	keys []*verificationKey
}

type signature struct {
	payload []byte
	base64  string
}

func newContainerVerificationManager(imageVerificationKeys []string) (containerVerificationMgr, error) {
	keys, err := parseVerificationKeys(imageVerificationKeys)
	if err != nil {
		return nil, err
	}

	return &ctrVerificationMgr{keys: keys}, nil
}

func (mgr *ctrVerificationMgr) GetVerificationKeys(config *types.VerificationConfig) ([]*verificationKey, error) {
	if config != nil {
		return parseVerificationKeys(config.Keys)
	}
	return mgr.keys, nil
}

func (mgr *ctrVerificationMgr) GetSignatureReference(image containerd.Image) (string, error) {
	imageRef := image.Name()
	// <image-name> + :<tag-name> or @<digest> )
	index := strings.Index(imageRef, digestDelimiter)
	if index == -1 {
		// last as there might be a port specified
		index = strings.LastIndex(imageRef, tagDelimiter)
	}
	if index == -1 {
		return "", log.NewErrorf("could not get signature reference for image = %s", imageRef)
	}
	d := image.Target().Digest
	return fmt.Sprint(imageRef[:index], tagDelimiter, d.Algorithm().String(), "-", d.Hex(), cosignSignatureSuffix), nil
}

func (mgr *ctrVerificationMgr) VerifySignature(ctx context.Context, image, signatureImage containerd.Image, keys ...*verificationKey) error {
	signatures, err := getSignatures(ctx, signatureImage.ContentStore(), signatureImage.Target())
	if err != nil {
		return err
	}

	// no signatures are found, skip verification
	if len(signatures) == 0 {
		log.Warn("no signatures are found in %s, verification will be skipped for image = %s", signatureImage.Name(), image.Name())
		return nil
	}

	var (
		digest     = image.Target().Digest.String()
		verifyErrs = &errorUtil.CompoundError{}
	)
	for _, key := range keys {
		var verifier sig.Verifier
		verifier, err = sig.LoadVerifier(key.publicKey, key.hashFunction)
		if err != nil {
			verifyErrs.Append(err)
			continue
		}
		err = verifySignatures(signatures, digest, verifier)
		if err != nil {
			verifyErrs.Append(err)
			continue
		}
		// signature is verified
		log.Debug("signature verification is successful for image = %s", image.Name())
		return nil
	}
	if verifyErrs.Size() > 0 {
		log.ErrorErr(verifyErrs, "signature verification failed for image = %s", image.Name())
		return verifyErrs
	}
	return log.NewErrorf("could not verify image = %s", image.Name())
}
