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

package main

import (
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/deployment"
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
	managerNetworkManagerServiceIDDefault  = network.LibnetworkManagerServiceLocalID
	managerContainerStopTimeoutDefault     = "30s"

	// default container client config
	containerClientNamespaceDefault   = "kanto-cm"
	containerClientAddressPathDefault = "/run/containerd/containerd.sock"
	containerClientExecRootDefault    = managerExecRootPathDefault
	containerClientMetaPathDefault    = managerMetaPathDefault
	containerClientRuncRuntimeDefault = string(types.RuntimeTypeV2runcV2)
	containerClientImageExpiry        = 31 * 24 * time.Hour // 31 days
	containerClientImageExpiryDisable = false
	containerClientLeaseIDDefault     = "kanto-cm.lease"
	containerClientImageVerifierType  = string(ctr.VerifierNone)

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

	// default things config
	thingsEnableDefault   = true
	thingsMetaPathDefault = managerMetaPathDefault

	// default local connection config
	connectionBrokerURLDefault         = "tcp://localhost:1883"
	connectionKeepAliveDefault         = "20s"
	connectionDisconnectTimeoutDefault = "250ms"
	connectionClientUsername           = ""
	connectionClientPassword           = ""
	connectTimeoutTimeoutDefault       = "30s"
	acknowledgeTimeoutDefault          = "15s"
	subscribeTimeoutDefault            = "15s"
	unsubscribeTimeoutDefault          = "5s"

	// default deployment config
	deploymentEnableDefault   = true
	deploymentModeDefault     = string(deployment.UpdateMode)
	deploymentMetaPathDefault = managerMetaPathDefault
	deploymentCtrPathDefault  = "/etc/container-management/containers"

	// default update agent config
	updateAgentEnableDefault                 = false
	updateAgentDomainDefault                 = "containers"
	updateAgentVerboseInventoryReportDefault = false
)

var (
	// default container client config
	containerClientInsecureRegistriesDefault = []string{"localhost"}

	// default things service features config
	thingsServiceFeaturesDefault = []string{things.ContainerFactoryFeatureID, things.SoftwareUpdatableFeatureID, things.MetricsFeatureID}
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
			MgrDefaultCtrsStopTimeout: managerContainerStopTimeoutDefault,
		},
		ContainerClientConfig: &containerRuntimeConfig{
			CtrNamespace:          containerClientNamespaceDefault,
			CtrAddressPath:        containerClientAddressPathDefault,
			CtrInsecureRegistries: containerClientInsecureRegistriesDefault,
			CtrRootExec:           containerClientExecRootDefault,
			CtrMetaPath:           containerClientMetaPathDefault,
			CtrRuncRuntime:        containerClientRuncRuntimeDefault,
			CtrImageExpiry:        containerClientImageExpiry,
			CtrImageExpiryDisable: containerClientImageExpiryDisable,
			CtrLeaseID:            containerClientLeaseIDDefault,
			CtrImageVerifierType:  containerClientImageVerifierType,
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
				BrokerURL:          connectionBrokerURLDefault,
				KeepAlive:          20000,
				DisconnectTimeout:  250,
				ClientUsername:     connectionClientUsername,
				ClientPassword:     connectionClientPassword,
				ConnectTimeout:     30000,
				AcknowledgeTimeout: 15000,
				SubscribeTimeout:   15000,
				UnsubscribeTimeout: 5000,
			},
		},
		DeploymentManagerConfig: &deploymentManagerConfig{
			DeploymentEnable:   deploymentEnableDefault,
			DeploymentMode:     deploymentModeDefault,
			DeploymentMetaPath: deploymentMetaPathDefault,
			DeploymentCtrPath:  deploymentCtrPathDefault,
		},
		UpdateAgentConfig: &updateAgentConfig{
			UpdateAgentEnable:      updateAgentEnableDefault,
			DomainName:             updateAgentDomainDefault,
			SystemContainers:       []string{}, // no system containers by defaults
			VerboseInventoryReport: updateAgentVerboseInventoryReportDefault,
		},
		LocalConnection: &localConnectionConfig{
			BrokerURL:          connectionBrokerURLDefault,
			KeepAlive:          connectionKeepAliveDefault,
			DisconnectTimeout:  connectionDisconnectTimeoutDefault,
			ClientUsername:     connectionClientUsername,
			ClientPassword:     connectionClientPassword,
			ConnectTimeout:     connectTimeoutTimeoutDefault,
			AcknowledgeTimeout: acknowledgeTimeoutDefault,
			SubscribeTimeout:   subscribeTimeoutDefault,
			UnsubscribeTimeout: unsubscribeTimeoutDefault,
		},
	}
}
