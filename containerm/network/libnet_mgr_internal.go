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

package network

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/libnetwork"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func (netMgr *libnetworkMgr) setupContainerNetworkSandbox(container *types.Container) (libnetwork.Sandbox, error) {
	if netMgr.netController == nil {
		return nil, nil
	}
	var (
		sb  libnetwork.Sandbox
		err error
	)
	sb = getNetworkSandbox(netMgr.netController, container.ID)
	if sb != nil {
		if err := netMgr.netController.SandboxDestroy(container.ID); err != nil {
			return nil, err
		}
	}
	options, err := buildSandboxOptions(container, netMgr.config)
	if err != nil {
		return nil, err
	}

	sb, err = netMgr.netController.NewSandbox(container.ID, options...)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (netMgr *libnetworkMgr) setupContainerNetworkEndpoint(network libnetwork.Network, container *types.Container) (libnetwork.Endpoint, error) {
	var (
		ep              libnetwork.Endpoint
		epCreateOptions []libnetwork.EndpointOption
		err             error
	)
	ep, err = getNetworkEndPoint(container, network)
	if err != nil {
		log.ErrorErr(err, "could not get network endpoint for container ID = %s", container.ID)
		return nil, err
	}
	if ep != nil {
		if err := ep.Delete(true); err != nil {
			log.ErrorErr(err, "could not delete existing network endpoint for container id = %s ", container.ID)
			return nil, err
		}
		ep = nil
	}

	if ep == nil {
		epCreateOptions, err = buildEndpointOptions()
		if err != nil {
			return nil, err
		}
		ep, err = network.CreateEndpoint(container.ID+"-ep", epCreateOptions...)
		if err != nil {
			return nil, err
		}
	}
	return ep, nil
}

// BuildHostnameFile writes the container's hostname file.
func (netMgr *libnetworkMgr) setupNetworkingRelatedPaths(container *types.Container) error {
	containerMetaPath := getContainerNetMetaPath(netMgr.config, container.ID)
	container.ResolvConfPath = filepath.Join(containerMetaPath, "resolv.conf")
	container.HostsPath = filepath.Join(containerMetaPath, "hosts")
	container.HostnamePath = filepath.Join(containerMetaPath, "hostname")
	if err := util.MkDir(containerMetaPath); err != nil {
		return log.NewErrorf("could not initialize meta path for container %s", container.ID)
	}
	return ioutil.WriteFile(container.HostnamePath, []byte(container.HostName+"\n"), 0644)
}

func (netMgr *libnetworkMgr) removeEndpoint(ctx context.Context, sandboxID string, endpoint *types.EndpointSettings) error {
	var (
		ep libnetwork.Endpoint
	)

	// find endpoint in network and delete it
	sb, err := netMgr.netController.SandboxByID(sandboxID)
	if err != nil {
		return err
	}
	if sb == nil {
		return log.NewErrorf("failed to get sandbox with id = %s", sandboxID)
	}

	sbEndpoints := sb.Endpoints()
	if len(sbEndpoints) == 0 {
		return log.NewErrorf("no endpoints in sandbox with id = %s", sandboxID)
	}

	for _, e := range sbEndpoints {
		if e.ID() == endpoint.ID {
			ep = e
			break
		}
	}

	if ep == nil {
		return log.NewErrorf("the endpoint %s is not connected", endpoint.ID)
	}

	if err := ep.Leave(sb); err != nil {
		return log.NewErrorf("error while leaving network %s", ep.Name())
	}

	if err := ep.Delete(false); err != nil {
		return log.NewErrorf("error while deleting endpoint with id = %s", ep.ID())
	}

	// delete the container's sandbox if there are not other endpoints connected
	sbEndpoints = sb.Endpoints()
	if len(sbEndpoints) == 0 {
		if err := sb.Delete(); err != nil {
			log.ErrorErr(err, "error while deleting sandbox with id = %s", sandboxID)
			return err
		}
	}
	return nil
}
