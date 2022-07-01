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

import (
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func newContainerdClient(namespace, socket, rootExec, metaPath string, registryConfigs map[string]*RegistryConfig, imageDecKeys, imageDecRecipients []string, runcRuntime types.Runtime) (ContainerAPIClient, error) {

	//ensure storage
	err := util.MkDir(rootExec)
	if err != nil {
		return nil, err
	}
	err = util.MkDir(metaPath)
	if err != nil {
		return nil, err
	}

	log.Debug("starting container client with default namespace = %s", namespace)
	ctrdClientSpi, err := newContainerdSpi(socket, namespace, containerd.DefaultSnapshotter /*overlayfs for now - TODO add client config*/, metaPath)
	if err != nil {
		return nil, err
	}
	decryptMgr, decrErr := newContainerDecryptManager(imageDecKeys, imageDecRecipients)
	if decrErr != nil {
		return nil, decrErr
	}

	ctrdClient := &containerdClient{
		rootExec:           rootExec,
		metaPath:           metaPath,
		ctrdCache:          newContainerInfoCache(),
		registriesResolver: newContainerImageRegistriesResolver(registryConfigs),
		spi:                ctrdClientSpi,
		ioMgr:              newContainerIOManager(filepath.Join(rootExec, "fifo"), newCache()),
		logsMgr:            newContainerLogsManager(filepath.Join(metaPath, "containers")),
		decMgr:             decryptMgr,
		runcRuntime:        runcRuntime,
	}
	go ctrdClient.processEvents(namespace)
	return ctrdClient, nil
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	createOpts := registryCtx.Config.([]ContainerOpts)
	var opts = &ctrOpts{}
	if err := applyOptsCtr(opts, createOpts...); err != nil {
		return nil, err
	}

	return newContainerdClient(opts.namespace, opts.connectionPath, opts.rootExec, opts.metaPath, opts.registryConfigs, opts.imageDecKeys, opts.imageDecRecipients, opts.runcRuntime)

}
