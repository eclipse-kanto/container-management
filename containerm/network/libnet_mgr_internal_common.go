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
	"path/filepath"
	"strconv"

	"github.com/docker/libnetwork"
	libnetconfig "github.com/docker/libnetwork/config"
	"github.com/docker/libnetwork/drivers/bridge"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/netutils"
	"github.com/docker/libnetwork/options"
	"github.com/docker/libnetwork/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

const (
	containersRootDir    = "containers"
	hostNetworkName      = "host"
	bridgeNetworkName    = "bridge"
	libnetworkDriverHost = "host"
)

func driverNetworkOptions(netConfig config) []libnetconfig.Option {
	bridgeDriverConfig := options.Generic{
		"EnableIPForwarding":  netConfig.bridgeConfig.ipForward,     //true
		"EnableIPTables":      netConfig.bridgeConfig.ipTables,      //true
		"EnableUserlandProxy": netConfig.bridgeConfig.userlandProxy, //false
	}
	bridgeOption := options.Generic{netlabel.GenericData: bridgeDriverConfig}

	driverOpts := []libnetconfig.Option{}
	driverOpts = append(driverOpts, libnetconfig.OptionDriverConfig(netConfig.netType, bridgeOption))
	return driverOpts
}

func buildNetworkControllerOptions(netConfig *config) ([]libnetconfig.Option, error) {
	opts := []libnetconfig.Option{}
	if netConfig == nil {
		return opts, nil
	}
	//may be set it optional in the future
	opts = append(opts, libnetconfig.OptionExperimental(false))

	//init directories
	opts = append(opts, libnetconfig.OptionDataDir(netConfig.metaPath))
	opts = append(opts, libnetconfig.OptionExecRoot(netConfig.execRoot))

	//restore active sandboxes
	if netConfig.activeSandboxes != nil && len(netConfig.activeSandboxes) != 0 {
		opts = append(opts, libnetconfig.OptionActiveSandboxes(netConfig.activeSandboxes))
	}

	//init driver options
	opts = append(opts, libnetconfig.OptionDefaultDriver(netConfig.netType))
	opts = append(opts, libnetconfig.OptionDefaultNetwork(bridgeNetworkName))
	opts = append(opts, driverNetworkOptions(*netConfig)...)
	opts = append(opts, libnetconfig.OptionNetworkControlPlaneMTU(netConfig.bridgeConfig.mtu))

	return opts, nil
}

// Remove default bridge interface if present (--bridge=none use case)
func removeDefaultBridgeInterface(defaultBridgeNetowrkName string) {
	if lnk, err := netlink.LinkByName(defaultBridgeNetowrkName); err == nil {
		if err := netlink.LinkDel(lnk); err != nil {
			//TODO add logs
		}
	}
}

func buildBridgeNetworkOptions(netConfig *config) ([]libnetwork.NetworkOption, error) {
	bridgeName := netConfig.bridgeConfig.name
	netOption := map[string]string{
		bridge.BridgeName:         bridgeName,
		bridge.DefaultBridge:      strconv.FormatBool(false),
		netlabel.DriverMTU:        strconv.Itoa(netConfig.bridgeConfig.mtu),
		bridge.EnableIPMasquerade: strconv.FormatBool(netConfig.bridgeConfig.ipMasq),
		bridge.EnableICC:          strconv.FormatBool(netConfig.bridgeConfig.icc),
	}

	var ipamV4Conf *libnetwork.IpamConf
	ipamV4Conf = &libnetwork.IpamConf{AuxAddresses: make(map[string]string)}

	nwList, _, err := netutils.ElectInterfaceAddresses(bridgeName)
	if err != nil {
		return nil, errors.Wrap(err, "list bridge addresses failed")
	}

	nw := nwList[0]
	ipamV4Conf.PreferredPool = types.GetIPNetCanonical(nw).String()
	hip, _ := types.GetHostPartIP(nw.IP, nw.Mask)
	if hip.IsGlobalUnicast() {
		ipamV4Conf.Gateway = nw.IP.String()
	}
	v4Conf := []*libnetwork.IpamConf{ipamV4Conf}

	//generate libnetwork options
	bridgeDriverOptions := []libnetwork.NetworkOption{}
	bridgeDriverOptions = append(bridgeDriverOptions, libnetwork.NetworkOptionPersist(true))
	bridgeDriverOptions = append(bridgeDriverOptions, libnetwork.NetworkOptionEnableIPv6(netConfig.bridgeConfig.enableIPv6))
	bridgeDriverOptions = append(bridgeDriverOptions, libnetwork.NetworkOptionDriverOpts(netOption))
	bridgeDriverOptions = append(bridgeDriverOptions, libnetwork.NetworkOptionIpam("default", "", v4Conf, nil, nil))
	bridgeDriverOptions = append(bridgeDriverOptions, libnetwork.NetworkOptionDeferIPv6Alloc(netConfig.bridgeConfig.enableIPv6))

	return bridgeDriverOptions, nil

}

func initializeDefaultBridgeNetwork(netController libnetwork.NetworkController, netConfig *config) (libnetwork.Network, error) {
	// backwards compatibility for clearing any old bridge networks from the libnetwork controller
	if n, err := netController.NetworkByName(netConfig.bridgeConfig.name); err == nil {
		if err := n.Delete(); err != nil {
			return nil, err
		}
		removeDefaultBridgeInterface(netConfig.bridgeConfig.name)
	}
	if n, err := netController.NetworkByName(bridgeNetworkName); err == nil {
		if err := n.Delete(); err != nil {
			return nil, err
		}
		removeDefaultBridgeInterface(netConfig.bridgeConfig.name)
	}
	brOpts, err := buildBridgeNetworkOptions(netConfig)
	if err != nil {
		return nil, err
	}

	network, err := netController.NewNetwork(netConfig.netType, bridgeNetworkName, "", brOpts...)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func initializeDefaultHostNetwork(netController libnetwork.NetworkController) (libnetwork.Network, error) {
	if n, err := netController.NetworkByName(hostNetworkName); err == nil && n != nil {
		log.Warn("host network already exists - will not create a new one")
		return n, nil
	}

	hostNetworkOpts := buildHostNetworkOptions()
	hostNetwork, err := netController.NewNetwork(libnetworkDriverHost, hostNetworkName, "", hostNetworkOpts...)
	if err != nil {
		return nil, log.NewErrorf("could not create host network: %v", err)
	}
	return hostNetwork, nil
}

func buildHostNetworkOptions() []libnetwork.NetworkOption {
	hostNetworkOptions := []libnetwork.NetworkOption{
		libnetwork.NetworkOptionEnableIPv6(false),
		libnetwork.NetworkOptionPersist(true),
	}
	log.Debug("initialized host network options: %+v", hostNetworkOptions)
	return hostNetworkOptions
}
func buildEndpointOptions() ([]libnetwork.EndpointOption, error) {
	var createOptions []libnetwork.EndpointOption

	//attach to the network  == network type
	createOptions = append(createOptions, libnetwork.CreateOptionAnonymous())
	createOptions = append(createOptions, libnetwork.CreateOptionDisableResolution())
	return createOptions, nil
}

func getContainerNetMetaPath(netConfig *config, containerID string) string {
	return filepath.Join(netConfig.metaPath, containersRootDir, containerID)
}

func netMrgOptsToLibnetConfig(netCreateOpts *netOpts) (config, error) {

	return config{
		netType:  netCreateOpts.netType,
		metaPath: netCreateOpts.metaPath,
		execRoot: netCreateOpts.execRoot,
		bridgeConfig: bridgeConfig{
			disableBridge: netCreateOpts.disableBridge,
			name:          netCreateOpts.name,
			ipV4:          netCreateOpts.ipV4,
			fixedCIDRv4:   netCreateOpts.fixedCIDRv4,
			gatewayIPv4:   netCreateOpts.gatewayIPv4,
			enableIPv6:    netCreateOpts.enableIPv6,
			mtu:           netCreateOpts.mtu,
			icc:           netCreateOpts.icc,
			ipTables:      netCreateOpts.ipTables,
			ipForward:     netCreateOpts.ipForward,
			ipMasq:        netCreateOpts.ipMasq,
			userlandProxy: netCreateOpts.userlandProxy,
		},
		activeSandboxes: make(map[string]interface{}),
	}, nil
}
