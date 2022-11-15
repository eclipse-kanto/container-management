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
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	sig "github.com/sigstore/sigstore/pkg/signature"
	"strings"
)

type verificationKey struct {
	publicKey    crypto.PublicKey
	hashFunction crypto.Hash
}

const (
	cosignSignatureAnnotationKey = "dev.cosignproject.cosign/signature"
	cosignSignatureSuffix        = ".sig"
)

var errNoSignatureFound = log.NewError("no signature is found")

type containerVerifyMgr interface {
	GetVerificationKeys(config *types.VerifyConfig) ([]*verificationKey, error)
	GetSignature(ctx context.Context, image containerd.Image) (containerd.Image, error)
	PullSignature(ctx context.Context, image containerd.Image, opts ...containerd.RemoteOpt) (containerd.Image, error)
	DeleteSignature(ctx context.Context, image containerd.Image) error
	VerifySignature(ctx context.Context, image, signatureImage containerd.Image, keys []*verificationKey) error
}

type ctrVerifyMgr struct {
	spi  containerdSpi
	keys []*verificationKey
}

type signature struct {
	payload []byte
	base64  string
}

func newContainerVerifyManager(spi containerdSpi, imageVerKeys []string) (containerVerifyMgr, error) {
	keys, err := parseVerificationKeys(imageVerKeys)
	if err != nil {
		return nil, err
	}

	return &ctrVerifyMgr{spi: spi, keys: keys}, nil
}

func (mgr *ctrVerifyMgr) GetVerificationKeys(config *types.VerifyConfig) ([]*verificationKey, error) {
	if config != nil {
		return parseVerificationKeys(config.Keys)
	}
	return mgr.keys, nil
}

func getSignatureReference(image containerd.Image) string {
	d := image.Target().Digest
	return fmt.Sprint(strings.Split(image.Name(), ":")[0], ":", d.Algorithm().String(), "-", d.Hex(), cosignSignatureSuffix)
}

func (mgr *ctrVerifyMgr) GetSignature(ctx context.Context, image containerd.Image) (containerd.Image, error) {
	signatureRef := getSignatureReference(image)
	return mgr.spi.GetImage(ctx, signatureRef)
}

func (mgr *ctrVerifyMgr) PullSignature(ctx context.Context, image containerd.Image, opts ...containerd.RemoteOpt) (containerd.Image, error) {
	signatureRef := getSignatureReference(image)
	signatureImage, err := mgr.spi.GetImage(ctx, signatureRef)
	if err != nil {
		// if the image is not present locally - pull it
		if errdefs.IsNotFound(err) {
			signatureImage, err = mgr.spi.PullImage(ctx, signatureRef, opts...)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return nil, errNoSignatureFound
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return signatureImage, nil
}

func (mgr *ctrVerifyMgr) DeleteSignature(ctx context.Context, image containerd.Image) error {
	signatureRef := getSignatureReference(image)
	return mgr.spi.DeleteImage(ctx, signatureRef)
}

func (mgr *ctrVerifyMgr) VerifySignature(ctx context.Context, image, signatureImage containerd.Image, keys []*verificationKey) error {
	if len(keys) == 0 {
		// no public keys, skip verification
		return nil
	}

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
	log.ErrorErr(verifyErrs, "signature verification failed for image = %s", image.Name())
	return verifyErrs

}
