// Copyright (c) 2021 Contributors to the Eclipse Foundation
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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"time"
)

// ContainerOpts represents container engine client's configuration options.
type ContainerOpts func(ctrOptions *ctrOpts) error

type ctrOpts struct {
	namespace          string
	connectionPath     string
	registryConfigs    map[string]*RegistryConfig
	rootExec           string
	metaPath           string
	imageDecKeys       []string
	imageDecRecipients []string
	runcRuntime        types.Runtime
	imageExpiry        time.Duration
	imageExpiryDisable bool
	leaseID            string
}

// RegistryConfig represents a single registry's access configuration.
type RegistryConfig struct {
	IsInsecure  bool
	Credentials *AuthCredentials
	Transport   *TLSConfig
}

// AuthCredentials represents credentials for accessing container registries secured via Basic Auth.
type AuthCredentials struct {
	UserID   string
	Password string
}

// TLSConfig represents TLS configuration.
type TLSConfig struct {
	RootCA     string
	ClientCert string
	ClientKey  string
}

func applyOptsCtr(ctrOpts *ctrOpts, opts ...ContainerOpts) error {
	for _, o := range opts {
		if err := o(ctrOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithCtrdNamespace sets the namespace that the container client instance will use within containerd.
func WithCtrdNamespace(namespace string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.namespace = namespace
		return nil
	}
}

// WithCtrdConnectionPath sets the address path to the containerd service communication endpoint (e.g. a local UNIX socket).
func WithCtrdConnectionPath(conPath string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.connectionPath = conPath
		return nil
	}
}

// WithCtrdRootExec sets root executable directory that the client will use.
func WithCtrdRootExec(rootExecDir string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.rootExec = rootExecDir
		return nil
	}
}

// WithCtrdMetaPath sets meta path for the container client service to use for its storage.
func WithCtrdMetaPath(metaPath string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.metaPath = metaPath
		return nil
	}
}

// WithCtrdRegistryConfigs sets the configurations for accessing the provided container registries.
func WithCtrdRegistryConfigs(configs map[string]*RegistryConfig) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.registryConfigs = configs
		return nil
	}
}

// WithCtrdImageDecryptKeys sets the keys for decrypting encrypted container images.
func WithCtrdImageDecryptKeys(keys ...string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.imageDecKeys = keys
		return nil
	}
}

// WithCtrdImageDecryptRecipients sets the recipients for decrypting encrypted container images.
func WithCtrdImageDecryptRecipients(recipients ...string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.imageDecRecipients = recipients
		return nil
	}
}

// WithCtrdRuncRuntime sets the container runc runtime.
func WithCtrdRuncRuntime(runcRuntime string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		switch types.Runtime(runcRuntime) {
		case types.RuntimeTypeV1, types.RuntimeTypeV2runcV1, types.RuntimeTypeV2runcV2:
			ctrOptions.runcRuntime = types.Runtime(runcRuntime)
		default:
			return log.NewErrorf("unexpected runc runtime = %s", runcRuntime)
		}
		return nil
	}
}

// WithCtrdImageExpiry sets images expiry time.
func WithCtrdImageExpiry(expiry time.Duration) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.imageExpiry = expiry
		return nil
	}
}

// WithCtrdImageExpiryDisable disables the images' expiry management.
func WithCtrdImageExpiryDisable(disable bool) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.imageExpiryDisable = disable
		return nil
	}
}

// WithCtrdLeaseID sets the lease that the container client instance will use within containerd.
func WithCtrdLeaseID(leaseID string) ContainerOpts {
	return func(ctrOptions *ctrOpts) error {
		ctrOptions.leaseID = leaseID
		return nil
	}
}
