// Copyright (c) 2022 Contributors to the Eclipse Foundation
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
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/imgcrypt/images/encryption"
	"github.com/containerd/imgcrypt/images/encryption/parsehelpers"
	ocicryptconfig "github.com/containers/ocicrypt/config"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
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
	return encryption.CheckAuthorization(ctx, image.ContentStore(), image.Target(), dc)
}
