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
	"os"

	"github.com/docker/libnetwork"
	libnetcfg "github.com/docker/libnetwork/config"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const (
	// LibetworkManagerServiceLocalID sets service local ID for libnetwork manager
	LibetworkManagerServiceLocalID = "container-management.service.local.v1.service-libnetwork-manager"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       LibetworkManagerServiceLocalID,
		Type:     registry.NetworkManagerService,
		InitFunc: registryInit,
	})
}

type libnetworkMgr struct {
	config        *config
	netController libnetwork.NetworkController //internal libnetwork controller fields
}

func (netMgr *libnetworkMgr) Manage(ctx context.Context, container *types.Container) error {

	if netMgr.netController == nil {
		return log.NewErrorf("no network controller to connect to default network")
	}

	var (
		sb  libnetwork.Sandbox
		err error
	)
	defer func() {
		if err != nil {
			if sb != nil {
				sb.Delete()
			}
		}
	}()

	if util.IsContainerNetworkHost(container) {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		container.HostName = hostname
		log.Debug("container ID = %s hostname set to the engine's hostname [%s] as the network configuration is set to [host]", container.ID, hostname)
	}
	//build hostname file
	err = netMgr.setupNetworkingRelatedPaths(container)
	if err != nil {
		return err
	}

	//init sandbox
	sb, err = netMgr.setupContainerNetworkSandbox(container)
	if err != nil {
		return err
	}

	return nil
}

func (netMgr *libnetworkMgr) Connect(ctx context.Context, container *types.Container) error {

	var (
		sb          libnetwork.Sandbox
		ep          libnetwork.Endpoint
		network     libnetwork.Network
		cEpSettings *types.EndpointSettings
		err         error
	)

	ctrNetworkName := string(container.HostConfig.NetworkMode)
	defer func() {
		if err != nil {
			if ep != nil {
				ep.Delete(true)
			}
			if container.NetworkSettings != nil && container.NetworkSettings.Networks != nil {
				delete(container.NetworkSettings.Networks, ctrNetworkName)
			}
		}
	}()
	// get network
	network, err = netMgr.netController.NetworkByName(ctrNetworkName)
	if err != nil {
		err = log.NewErrorf("no network [%s] found while connecting container %s ", ctrNetworkName, container.ID)
		return err
	}

	// get sandbox
	sb = getNetworkSandbox(netMgr.netController, container.ID)
	if sb == nil {
		err = log.NewErrorf("no network sandbox for container %s ", container.ID)
		return err
	}

	//init endpoint
	ep, err = netMgr.setupContainerNetworkEndpoint(network, container)
	if err != nil {
		return err
	}

	//joint the network sandbox
	if err = ep.Join(sb); err != nil {
		return err
	}

	//update container network config
	cEpSettings = mapToContainerEndpointSettings(network, ep)

	if container.NetworkSettings == nil {
		container.NetworkSettings = &types.NetworkSettings{}
	}
	container.NetworkSettings.SandboxID = sb.ID()
	container.NetworkSettings.SandboxKey = sb.Key()
	container.NetworkSettings.NetworkControllerID = netMgr.netController.ID()

	if container.NetworkSettings.Networks == nil {
		container.NetworkSettings.Networks = make(map[string]*types.EndpointSettings)
	}

	container.NetworkSettings.Networks[ctrNetworkName] = cEpSettings

	return nil

}

func (netMgr *libnetworkMgr) Restore(ctx context.Context, containers []*types.Container) error {
	var (
		err           error
		netController libnetwork.NetworkController
		netOptions    []libnetcfg.Option
	)
	if containers != nil {
		//restore active containers sandboxes
		if netMgr.config.activeSandboxes == nil {
			netMgr.config.activeSandboxes = make(map[string]interface{})
		}
		for _, ctr := range containers {
			if ctr.NetworkSettings == nil || ctr.NetworkSettings.SandboxID == "" {
				log.Warn("no network settings are restored for container id = %s", ctr.ID)
				continue
			}
			sbOpts, err := buildSandboxOptions(ctr, netMgr.config)
			if err != nil {
				log.ErrorErr(err, "error building sandbox options for restored container id = %s ", ctr.ID)
				continue
			}

			netMgr.config.activeSandboxes[ctr.NetworkSettings.SandboxID] = sbOpts
			log.Debug("added network sandbox config for container id = %s with sandbox id = %s", ctr.ID, ctr.NetworkSettings.SandboxID)
		}
	}
	//ensure storage
	err = util.MkDirs(netMgr.config.metaPath, netMgr.config.execRoot)
	if err != nil {
		return err
	}

	//init controller
	netOptions, err = buildNetworkControllerOptions(netMgr.config)
	if err != nil {
		return err
	}

	netController, err = libnetwork.New(netOptions...)
	if err != nil {
		return err
	}
	netMgr.netController = netController

	return nil
}

func (netMgr *libnetworkMgr) Disconnect(ctx context.Context, container *types.Container, force bool) error {
	return nil
}

func (netMgr *libnetworkMgr) ReleaseNetworkResources(ctx context.Context, container *types.Container) error {
	if container.NetworkSettings == nil {
		return nil
	}

	defer func() {
		log.Debug("clearing internal network settings cache for container ID = %s", container.ID)
		container.NetworkSettings = nil
	}()

	for netName, endpointSettings := range container.NetworkSettings.Networks {
		if err := netMgr.removeEndpoint(context.Background(), container.NetworkSettings.SandboxID, endpointSettings); err != nil {
			log.ErrorErr(err, "error removing endpoint for network %s for container ID = %s", netName, container.ID)
			return err
		}
	}
	return nil
}

func (netMgr *libnetworkMgr) Dispose(ctx context.Context) error {
	netMgr.netController.Stop()
	return nil
}

func (netMgr *libnetworkMgr) Initialize(ctx context.Context) error {

	//init default host network
	hostNet, err := initializeDefaultHostNetwork(netMgr.netController)
	if err != nil {
		return err
	}
	log.Debug("successfully created and initialized the new default network [%s] from scratch ", hostNet.Name())

	if netMgr.config.activeSandboxes != nil && len(netMgr.config.activeSandboxes) > 0 {
		log.Debug("there are active sandboxes - a new default bridge network will not be initialized")

		defaultBridgeNetwork, err := netMgr.netController.NetworkByName(bridgeNetworkName)
		if err != nil {
			log.ErrorErr(err, "could not load default bridge network")
			return err
		}
		log.Debug("successfully initialized existing bridge network %s", defaultBridgeNetwork.Name())
		return nil
	}
	log.Debug("there are no active sandboxes - the default bridge network will be initialized")
	//init default bridge network
	brNet, err := initializeDefaultBridgeNetwork(netMgr.netController, netMgr.config)
	if err != nil {
		return err
	}
	log.Debug("successfully created and initialized the new default bridge network [%s] from scratch ", brNet.Name())

	return nil
}

func (netMgr *libnetworkMgr) Metrics(ctx context.Context, container *types.Container) (*types.IOStats, error) {
	sb := getNetworkSandbox(netMgr.netController, container.ID)
	if sb == nil {
		return nil, log.NewErrorf("no network sandbox for container %s ", container.ID)
	}
	interfaceStats, err := sb.Statistics()
	if err != nil {
		return nil, err
	}
	var rx, tx uint64
	for _, is := range interfaceStats {
		rx += is.RxBytes
		tx += is.TxBytes
	}
	return &types.IOStats{Read: rx, Write: tx}, nil
}
