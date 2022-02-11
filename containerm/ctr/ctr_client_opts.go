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

// ContainerOpts represents container engine client's configuration options.
type ContainerOpts func(ctrOptions *ctrOpts) error

type ctrOpts struct {
	namespace       string
	connectionPath  string
	registryConfigs map[string]*RegistryConfig
	rootExec        string
	metaPath        string
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
