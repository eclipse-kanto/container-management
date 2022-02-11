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
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/docker/libnetwork"
	libnettypes "github.com/docker/libnetwork/types"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const (
	defaultHostResolvConfPath = "/etc/resolv.conf"
	defaultHostsPath          = "/etc/hosts"

	reservedHostIP       = "host_ip"
	reservedHostIPPrefix = reservedHostIP + "_"
)

var (
	regexReservedAutoresolveHostIPMapping = regexp.MustCompile(fmt.Sprintf("^%s(.+)$", reservedHostIPPrefix))
)

func getNetworkSandbox(netctrl libnetwork.NetworkController, containerID string) libnetwork.Sandbox {
	var sb libnetwork.Sandbox
	netctrl.WalkSandboxes(func(s libnetwork.Sandbox) bool {
		if s.ContainerID() == containerID {
			sb = s
			return true
		}
		return false
	})
	return sb
}

func buildSandboxOptions(container *types.Container, netConfig *config) ([]libnetwork.SandboxOption, error) {
	var sboxOptions []libnetwork.SandboxOption

	sboxOptions = append(sboxOptions, libnetwork.OptionHostname(container.HostName),
		libnetwork.OptionDomainname(container.DomainName))

	if util.IsContainerNetworkHost(container) {
		sboxOptions = append(sboxOptions, libnetwork.OptionUseDefaultSandbox())
		if len(container.HostConfig.ExtraHosts) == 0 {
			sboxOptions = append(sboxOptions, libnetwork.OptionOriginHostsPath(defaultHostsPath))
		}
		// Copy the host's resolv.conf for the container (/etc/resolv.conf or /run/systemd/resolve/resolv.conf)
		sboxOptions = append(sboxOptions, libnetwork.OptionOriginResolvConfPath(defaultHostResolvConfPath))
	} else {
		// OptionUseExternalKey is mandatory for userns support.
		// But optional for non-userns support
		sboxOptions = append(sboxOptions, libnetwork.OptionUseExternalKey())
	}

	//-----------------Resolve paths -------------------------------
	//config the local hosts and resolv.conf for the container
	sboxOptions = append(sboxOptions, libnetwork.OptionHostsPath(container.HostsPath))
	sboxOptions = append(sboxOptions, libnetwork.OptionResolvConfPath(container.ResolvConfPath))
	//-----------------EOF Resolve paths -------------------------------

	//add extra hosts to /etc/hosts
	extraHosts := container.HostConfig.ExtraHosts
	if extraHosts != nil {
		for _, extraHost := range extraHosts {
			host := strings.Split(strings.TrimSpace(extraHost), ":")
			// check for ip_host[_interface]
			resolved, err := resolveToHostIPOnInterface(container, netConfig, host[1])
			if err != nil {
				log.ErrorErr(err, "could not map the reserved host_ip_[interface] to an IP")
			} else {
				sboxOptions = append(sboxOptions, libnetwork.OptionExtraHost(host[0], resolved))
			}
		}
	}

	portMappings := container.HostConfig.PortMappings
	if portMappings != nil {
		var bindings []libnettypes.PortBinding
		for _, mapping := range portMappings {
			binding := libnettypes.PortBinding{
				Proto:       libnettypes.ParseProtocol(mapping.Proto),
				Port:        mapping.ContainerPort,
				HostIP:      net.ParseIP(mapping.HostIP),
				HostPort:    mapping.HostPort,
				HostPortEnd: mapping.HostPortEnd,
			}
			bindings = append(bindings, binding)
		}
		sboxOptions = append(sboxOptions, libnetwork.OptionPortMapping(bindings))
	}
	return sboxOptions, nil
}

func resolveToHostIPOnInterface(container *types.Container, netConfig *config, ipToCheck string) (string, error) {
	var interfaceName string
	if regexReservedAutoresolveHostIPMapping.MatchString(ipToCheck) {
		interfaceName = regexReservedAutoresolveHostIPMapping.FindStringSubmatch(ipToCheck)[1]
	} else if ipToCheck == reservedHostIPPrefix {
		return "", log.NewError("a network interface name must be provided after the reserved host_ip_ prefix - e.g. host_ip_gw0 or use just host_ip if you want to resolve the host's IP on the default bridge network interface")
	} else if ipToCheck == reservedHostIP {
		if container.HostConfig.NetworkMode == types.NetworkModeBridge {
			interfaceName = netConfig.bridgeConfig.name
		} else {
			return "", log.NewError("will not resolve host_ip as container with id = %s is not configured in bridge network mode, thus, not connected to the default bridge network interface")
		}
	} else {
		return ipToCheck, nil
	}

	netIf, _ := net.InterfaceByName(interfaceName)
	ifAddresses, err := netIf.Addrs()
	if err != nil {
		return "", err
	}
	var ip net.IP
	for _, ifAddress := range ifAddresses {
		switch v := ifAddress.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() && v.IP.To4() != nil {
				if ip = v.IP; ip != nil {
					return ip.String(), nil
				}
			}
		default:
			log.Error("The network is not an IP Network")
		}
	}
	return "", log.NewErrorf("could not retrieve the host's IP on interface %s for container id = %s", interfaceName, container.ID)

}

func getNetworkEndPoint(container *types.Container, network libnetwork.Network) (libnetwork.Endpoint, error) {
	if container.NetworkSettings == nil || container.NetworkSettings.Networks == nil || container.NetworkSettings.Networks[network.Name()] == nil || container.NetworkSettings.Networks[network.Name()].ID == "" {
		return nil, nil
	}
	ep, err := network.EndpointByID(container.NetworkSettings.Networks[network.Name()].ID)
	if err == libnetwork.ErrNoSuchEndpoint(container.NetworkSettings.Networks[network.Name()].ID) {
		return nil, nil
	}
	return ep, err
}

func mapToContainerEndpointSettings(network libnetwork.Network, netEp libnetwork.Endpoint) *types.EndpointSettings {
	epInfo := netEp.Info()
	epSettings := &types.EndpointSettings{}

	epSettings.ID = netEp.ID()
	epSettings.NetworkID = network.ID()
	if gw := epInfo.Gateway(); gw != nil {
		epSettings.Gateway = gw.String()
	}

	iface := epInfo.Iface()
	if iface != nil {
		if addr := iface.Address(); addr != nil {
			epSettings.IPAddress = addr.IP.String()
		}
		if mac := iface.MacAddress(); mac != nil {
			epSettings.MacAddress = mac.String()
		}
	}
	return epSettings

}
