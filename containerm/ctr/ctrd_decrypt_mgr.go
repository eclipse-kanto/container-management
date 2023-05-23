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
	"encoding/json"
	"sort"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/imgcrypt/images/encryption"
	"github.com/containerd/imgcrypt/images/encryption/parsehelpers"
	ocicryptconfig "github.com/containers/ocicrypt/config"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type containerDecryptMgr interface {
	GetDecryptConfig(config *types.DecryptConfig) (*ocicryptconfig.DecryptConfig, error)
	CheckAuthorization(ctx context.Context, image containerd.Image, config *ocicryptconfig.DecryptConfig) error
}

type ctrDecryptMgr struct {
	cryptoConfig ocicryptconfig.CryptoConfig
}

func newContainerDecryptManager(imageDecKeys, imageDecRecipients []string) (containerDecryptMgr, error) {
	encArgs := parsehelpers.EncArgs{
		Key:          imageDecKeys,
		DecRecipient: imageDecRecipients,
	}
	cc, err := parsehelpers.CreateDecryptCryptoConfig(encArgs, nil)
	if err != nil {
		log.ErrorErr(err, "could not process provided image decrypt keys\n"+
			"decrypt keys: %s\n"+
			"decrypt recipients: %s", imageDecKeys, imageDecRecipients)
		return nil, err
	}
	return &ctrDecryptMgr{cryptoConfig: cc}, nil
}

func (mgr *ctrDecryptMgr) GetDecryptConfig(decryptConfig *types.DecryptConfig) (*ocicryptconfig.DecryptConfig, error) {
	var encArgs parsehelpers.EncArgs
	if decryptConfig != nil {
		encArgs = parsehelpers.EncArgs{
			Key:          decryptConfig.Keys,
			DecRecipient: decryptConfig.Recipients,
		}
		decryptCC, err := parsehelpers.CreateDecryptCryptoConfig(encArgs, nil)
		if err != nil {
			return nil, err
		}
		return decryptCC.DecryptConfig, nil
	}
	return mgr.cryptoConfig.DecryptConfig, nil
}

func (mgr *ctrDecryptMgr) CheckAuthorization(ctx context.Context, image containerd.Image, decryptConfig *ocicryptconfig.DecryptConfig) error {
	dc := decryptConfig
	if dc == nil {
		dc = &ocicryptconfig.DecryptConfig{}
	}

	desc, err := getPlatformSpecificManifest(ctx, image)
	if err != nil {
		log.ErrorErr(err, "could not get the platform specific manifest of image ID = %s", image.Name())
		return err
	}

	return encryption.CheckAuthorization(ctx, image.ContentStore(), desc, dc)
}

// imgcrypt library does not handle properly the case when a multi-arch index descriptor is provided in some corner cases
// e.g. 32bit Raspberry Pi OS running on cpu with 64bit architecture
func getPlatformSpecificManifest(ctx context.Context, image containerd.Image) (ocispec.Descriptor, error) {
	desc := image.Target()
	if desc.MediaType != ocispec.MediaTypeImageIndex && desc.MediaType != images.MediaTypeDockerSchema2ManifestList {
		// if not a multi-arch image, proceed as it is
		return desc, nil
	}

	var index ocispec.Index
	if blob, err := content.ReadBlob(ctx, image.ContentStore(), desc); err != nil {
		return ocispec.Descriptor{}, err
	} else if err = json.Unmarshal(blob, &index); err != nil {
		return ocispec.Descriptor{}, err
	}

	manifests := index.Manifests
	if len(manifests) == 0 {
		return ocispec.Descriptor{}, log.NewErrorf("no manifests are found")
	}

	// strict platform match
	platform := platforms.DefaultSpec()
	for _, manifest := range manifests {
		if platforms.NewMatcher(platform).Match(*manifest.Platform) {
			return manifest, nil
		}
	}

	// default to less strict platform match - e.g. fallback to lower arm variants
	matcher := platforms.Default()
	sort.SliceStable(manifests, func(i, j int) bool {
		if manifests[i].Platform == nil {
			return false
		}
		if manifests[j].Platform == nil {
			return true
		}
		return matcher.Less(*manifests[i].Platform, *manifests[j].Platform)
	})

	if matcher.Match(*manifests[0].Platform) {
		return manifests[0], nil
	}

	return ocispec.Descriptor{}, log.NewErrorf("no match for platform = %+v", platform)
}
