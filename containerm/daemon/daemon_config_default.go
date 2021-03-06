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

package main

import (
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/network"
	"github.com/eclipse-kanto/container-management/containerm/things"
)

const (
	// default daemon config
	daemonConfigFileDefault = ""

	// default daemon log config
	daemonLogFileDefault       = "log/container-management.log"
	daemonLogLevelDefault      = "INFO"
	daemonLogFileSizeDefault   = 2
	daemonLogFileCountDefault  = 5
	daemonLogFileMaxAgeDefault = 28
	daemonEnableSyslogDefault  = false

	// default container manager config
	managerMetaPathDefault                 = "/var/lib/container-management"
	managerExecRootPathDefault             = "/var/run/container-management"
	managerContainerClientServiceIDDefault = ctr.ContainerdClientServiceLocalID
	managerNetworkManagerServiceIDDefault  = network.LibetworkManagerServiceLocalID
	managerNetworkManagerStopTimeout       = 30

	// default container client config
	containerClientNamespaceDefault   = "kanto-cm"
	containerClientAddressPathDefault = "/run/containerd/containerd.sock"
	containerClientExecRootDefault    = managerExecRootPathDefault
	containerClientMetaPathDefault    = managerMetaPathDefault
	containerClientRuncRuntimeDefault = string(types.RuntimeTypeV2runcV2)

	// default network manager config
	networkManagerNetTypeDefault  = string(types.NetworkModeBridge)
	networkManagerMetaPathDefault = managerMetaPathDefault
	networkManagerExecRootDefault = managerExecRootPathDefault

	// default bridge network config
	networkBridgeDisableDefault       = false
	networkBridgeNameDefault          = "kanto-cm0"
	networkBridgeIPV4Default          = ""
	networkBridgeFixedCIDRv4Default   = ""
	networkBridgeGatewayIPV4Default   = ""
	networkBridgeEnableIPV6Default    = false
	networkBridgeMtuDefault           = 1500
	networkBridgeIccDefault           = true
	networkBridgeIPTablesDefault      = true
	networkBridgeIPForwardDefault     = true
	networkBridgeIPMasqDefault        = true
	networkBridgeUserlandProxyDefault = false

	// default grpc server config
	grpcServerNetworkProtocolDefault = "unix"
	grpcServerAddressPathDefault     = "/run/container-management/container-management.sock"

	// default things connection config
	thingsEnableDefault                      = true
	thingsMetaPathDefault                    = managerMetaPathDefault
	thingsConnectionBrokerURLDefault         = "tcp://localhost:1883"
	thingsConnectionKeepAliveDefault         = 20000
	thingsConnectionDisconnectTimeoutDefault = 250
	thingsConnectionClientUsername           = ""
	thingsConnectionClientPassword           = ""
	thingsConnectTimeoutTimeoutDefault       = 30000
	thingsAcknowledgeTimeoutDefault          = 15000
	thingsSubscribeTimeoutDefault            = 15000
	thingsUnsubscribeTimeoutDefault          = 5000
)

var (
	// default container client config
	containerClientInsecureRegistriesDefault = []string{"localhost"}

	// default things service features config
	thingsServiceFeaturesDefault = []string{things.ContainerFactoryFeatureID, things.SoftwareUpdatableFeatureID}
)

func getDefaultInstance() *config {
	return &config{
		Log: &log.Config{
			LogFile:       daemonLogFileDefault,
			LogLevel:      daemonLogLevelDefault,
			LogFileSize:   daemonLogFileSizeDefault,
			LogFileCount:  daemonLogFileCountDefault,
			LogFileMaxAge: daemonLogFileMaxAgeDefault,
			Syslog:        daemonEnableSyslogDefault,
		},
		ManagerConfig: &managerConfig{
			MgrMetaPath:               managerMetaPathDefault,
			MgrExecPath:               managerExecRootPathDefault,
			MgrCtrClientServiceID:     managerContainerClientServiceIDDefault,
			MgrNetMgrServiceID:        managerNetworkManagerServiceIDDefault,
			MgrDefaultCtrsStopTimeout: managerNetworkManagerStopTimeout,
		},
		ContainerClientConfig: &containerRuntimeConfig{
			CtrNamespace:          containerClientNamespaceDefault,
			CtrAddressPath:        containerClientAddressPathDefault,
			CtrInsecureRegistries: containerClientInsecureRegistriesDefault,
			CtrRootExec:           containerClientExecRootDefault,
			CtrMetaPath:           containerClientMetaPathDefault,
			CtrRuncRuntime:        containerClientRuncRuntimeDefault,
		},
		NetworkConfig: &networkConfig{
			NetType:     networkManagerNetTypeDefault,
			NetMetaPath: networkManagerMetaPathDefault,
			NetExecRoot: networkManagerExecRootDefault,
			DefaultBridgeNetworkConfig: &bridgeNetworkConfig{
				NetBridgeDisableBridge: networkBridgeDisableDefault,
				NetBridgeName:          networkBridgeNameDefault,
				NetBridgeIPV4:          networkBridgeIPV4Default,
				NetBridgeFixedCIDRv4:   networkBridgeFixedCIDRv4Default,
				NetBridgeGatewayIPv4:   networkBridgeGatewayIPV4Default,
				NetBridgeEnableIPv6:    networkBridgeEnableIPV6Default,
				NetBridgeMtu:           networkBridgeMtuDefault,
				NetBridgeIcc:           networkBridgeIccDefault,
				NetBridgeIPTables:      networkBridgeIPTablesDefault,
				NetBridgeIPForward:     networkBridgeIPForwardDefault,
				NetBridgeIPMasq:        networkBridgeIPMasqDefault,
				NetBridgeUserlandProxy: networkBridgeUserlandProxyDefault,
			},
		},
		GrpcServerConfig: &grpcServerConfig{
			GrpcServerNetworkProtocol: grpcServerNetworkProtocolDefault,
			GrpcServerAddressPath:     grpcServerAddressPathDefault,
		},
		ThingsConfig: &thingsConfig{
			ThingsEnable:   thingsEnableDefault,
			ThingsMetaPath: thingsMetaPathDefault,
			Features:       thingsServiceFeaturesDefault,
			ThingsConnectionConfig: &thingsConnectionConfig{
				BrokerURL:          thingsConnectionBrokerURLDefault,
				KeepAlive:          thingsConnectionKeepAliveDefault,
				DisconnectTimeout:  thingsConnectionDisconnectTimeoutDefault,
				ClientUsername:     thingsConnectionClientUsername,
				ClientPassword:     thingsConnectionClientPassword,
				ConnectTimeout:     thingsConnectTimeoutTimeoutDefault,
				AcknowledgeTimeout: thingsAcknowledgeTimeoutDefault,
				SubscribeTimeout:   thingsSubscribeTimeoutDefault,
				UnsubscribeTimeout: thingsUnsubscribeTimeoutDefault,
			},
		},
	}
}
