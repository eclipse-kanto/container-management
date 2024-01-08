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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setupCommandFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	// init daemon general config flag
	flagSet.String(daemonConfigFileFlagID, "", "Specify the configuration file of container-management")

	// init daemon log flags
	flagSet.StringVar(&cfg.Log.LogLevel, "log-level", cfg.Log.LogLevel, "Set the daemon's log level - possible values are ERROR, WARN, INFO, DEBUG, TRACE")
	flagSet.StringVar(&cfg.Log.LogFile, "log-file", cfg.Log.LogFile, "Set the daemon's log file")
	flagSet.IntVar(&cfg.Log.LogFileSize, "log-file-size", cfg.Log.LogFileSize, "Set the maximum size in megabytes of the log file before it gets rotated")
	flagSet.IntVar(&cfg.Log.LogFileCount, "log-file-count", cfg.Log.LogFileCount, "Set the maximum number of old log files to retain")
	flagSet.IntVar(&cfg.Log.LogFileMaxAge, "log-file-max-age", cfg.Log.LogFileMaxAge, "Set the maximum number of days to retain old log files based on the timestamp encoded in their filename")
	flagSet.BoolVar(&cfg.Log.Syslog, "log-syslog", cfg.Log.Syslog, "Enable logging in the local syslog (e.g. /dev/log, /var/run/syslog, /var/run/log)")

	// init deployment flags
	flagSet.BoolVar(&cfg.DeploymentManagerConfig.DeploymentEnable, "deployment-enable", cfg.DeploymentManagerConfig.DeploymentEnable, "Enable the deployment service providing installation/update of containers via the container descriptor files")
	flagSet.StringVar(&cfg.DeploymentManagerConfig.DeploymentMode, "deployment-mode", cfg.DeploymentManagerConfig.DeploymentMode, "Specify the operation mode of deployment manager service, e.g. if it shall run on its initial run only or on every start of container management")
	flagSet.StringVar(&cfg.DeploymentManagerConfig.DeploymentMetaPath, "deployment-home-dir", cfg.DeploymentManagerConfig.DeploymentMetaPath, "Specify the root directory of the deployment manager service")
	flagSet.StringVar(&cfg.DeploymentManagerConfig.DeploymentCtrPath, "deployment-ctr-dir", cfg.DeploymentManagerConfig.DeploymentCtrPath, "Specify a directory with container descriptor files for automated deployment")

	// init container manager flags
	flagSet.StringVar(&cfg.ManagerConfig.MgrMetaPath, "cm-home-dir", cfg.ManagerConfig.MgrMetaPath, "Specify the root directory of the container manager service")
	flagSet.StringVar(&cfg.ManagerConfig.MgrExecPath, "cm-exec-root-dir", cfg.ManagerConfig.MgrExecPath, "Specify the exec root directory of the container manager service")
	flagSet.StringVar(&cfg.ManagerConfig.MgrCtrClientServiceID, "cm-cc-sid", cfg.ManagerConfig.MgrCtrClientServiceID, "Specify the ID of the container runtime client service to be used by the container manager service")
	flagSet.StringVar(&cfg.ManagerConfig.MgrNetMgrServiceID, "cm-net-sid", cfg.ManagerConfig.MgrNetMgrServiceID, "Specify the ID of the network manager service to be used by container manager service")
	flagSet.StringVar(&cfg.ManagerConfig.MgrDefaultCtrsStopTimeout, "cm-deflt-ctrs-stop-timeout", cfg.ManagerConfig.MgrDefaultCtrsStopTimeout, "Specify the default timeout that the container manager service will wait before killing the container's process")

	// init container client flags
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrNamespace, "ccl-default-ns", cfg.ContainerClientConfig.CtrNamespace, "Specify the default namespace to be used for container management isolation")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrAddressPath, "ccl-ap", cfg.ContainerClientConfig.CtrAddressPath, "Specify the address path to communicate with the desired container runtime")
	flagSet.StringSliceVar(&cfg.ContainerClientConfig.CtrInsecureRegistries, "ccl-insecure-registries", cfg.ContainerClientConfig.CtrInsecureRegistries, "Specify insecure image registries - <ip/hostname>[:<port>]")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrRootExec, "ccl-exec-root-dir", cfg.ContainerClientConfig.CtrRootExec, "Specify the exec root dir to be used for container runtime management data")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrMetaPath, "ccl-home-dir", cfg.ContainerClientConfig.CtrMetaPath, "Specify the home directory to be used for container runtime management data")
	flagSet.StringSliceVar(&cfg.ContainerClientConfig.CtrImageDecKeys, "ccl-image-dec-keys", cfg.ContainerClientConfig.CtrImageDecKeys, "Specify a list of private keys filenames (GPG private key ring, JWE and PKCS7 private key). Each entry can include an optional password separated by a colon after the filename.")
	flagSet.StringSliceVar(&cfg.ContainerClientConfig.CtrImageDecRecipients, "ccl-image-dec-recipients", cfg.ContainerClientConfig.CtrImageDecRecipients, "Specify a recipients certificates list of the image (used only for PKCS7 and must be an x509)")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrRuncRuntime, "ccl-runc-runtime", cfg.ContainerClientConfig.CtrRuncRuntime, "Specify a default global runc runtime - possible values are io.containerd.runtime.v1.linux, io.containerd.runc.v1 and io.containerd.runc.v2. ")
	flagSet.DurationVar(&cfg.ContainerClientConfig.CtrImageExpiry, "ccl-image-expiry", cfg.ContainerClientConfig.CtrImageExpiry, "Specify the time period for the cached images and content to be kept in the form of e.g. 72h3m0.5s")
	flagSet.BoolVar(&cfg.ContainerClientConfig.CtrImageExpiryDisable, "ccl-image-expiry-disable", cfg.ContainerClientConfig.CtrImageExpiryDisable, "Disables expiry management of cached images and content - must be used with caution as it may lead to large memory volumes being persistently allocated")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrLeaseID, "ccl-lease-id", cfg.ContainerClientConfig.CtrLeaseID, "Specify the lease identifier to be used for container resources persistence")
	flagSet.StringVar(&cfg.ContainerClientConfig.CtrImageVerifierType, "ccl-image-verifier-type", cfg.ContainerClientConfig.CtrImageVerifierType, "Specify the image verifier type - possible values are none and notation, when set to none image signatures wil not be verified.")
	flagSet.Var(&cfg.ContainerClientConfig.CtrImageVerifierConfig, "ccl-image-verifier-config", "Specify the configuration of the image verifier, as comma separated {key}={value} pairs - possible keys for notation verifier are configDir and libexecDir, for more info https://notaryproject.dev/docs/user-guides/how-to/directory-structure/#user-level")

	// init network manager flags
	flagSet.StringVar(&cfg.NetworkConfig.NetType, "net-type", cfg.NetworkConfig.NetType, "Specify the default network management type for containers")
	flagSet.StringVar(&cfg.NetworkConfig.NetMetaPath, "net-home-dir", cfg.NetworkConfig.NetMetaPath, "Specify the home directory for containers network management data handling")
	flagSet.StringVar(&cfg.NetworkConfig.NetExecRoot, "net-exec-root-dir", cfg.NetworkConfig.NetExecRoot, "Specify the exec root for the network management operations")

	// init default bridge network flags
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeDisableBridge, "net-tbr-disable", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeDisableBridge, "Disables the default container management bridge network")
	flagSet.StringVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeName, "net-br-name", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeName, "The name of the default bridge network interface")
	flagSet.StringVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPV4, "net-br-ip4", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPV4, "The IP v4 for the default bridge network interface")
	flagSet.StringVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeFixedCIDRv4, "net-br-fcidr4", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeFixedCIDRv4, "The fixed container ids range for the default bridge network interface used with IP v4")
	flagSet.StringVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeGatewayIPv4, "net-br-gwip4", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeGatewayIPv4, "The IP v4 of the gateway to be configured for the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeEnableIPv6, "net-br-enable-ip6", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeEnableIPv6, "Specifies whether IP v6 must be enabled for the default bridge network interface")
	flagSet.IntVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeMtu, "net-br-mtu", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeMtu, "Specifies the MTU for the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIcc, "net-br-icc", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIcc, "Enable inter-container communication on the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPTables, "net-br-ipt", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPTables, "Enable Ip Tables management on the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPForward, "net-br-ipfw", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPForward, "Enable IP forwarding on the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPMasq, "net-br-ipmasq", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPMasq, "Enable IP Masquerade on the default bridge network interface")
	flagSet.BoolVar(&cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeUserlandProxy, "net-br-ulp", cfg.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeUserlandProxy, "Enable userland-proxy on the default bridge network interface")

	// init grpc server flags
	flagSet.StringVar(&cfg.GrpcServerConfig.GrpcServerNetworkProtocol, "grpc-serv-netp", cfg.GrpcServerConfig.GrpcServerNetworkProtocol, "Specify the communication protocol with the container management grpc server")
	flagSet.StringVar(&cfg.GrpcServerConfig.GrpcServerAddressPath, "grpc-serv-ap", cfg.GrpcServerConfig.GrpcServerAddressPath, "Specify the address for communication with the container management grpc server")

	// init things client
	flagSet.BoolVar(&cfg.ThingsConfig.ThingsEnable, "things-enable", cfg.ThingsConfig.ThingsEnable, "Enable the things container management service providing remote containers management and their representation via the Bosch IoT Things service")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsMetaPath, "things-home-dir", cfg.ThingsConfig.ThingsMetaPath, "Specify the home directory for the things container management service persistent storage")
	flagSet.StringSliceVar(&cfg.ThingsConfig.Features, "things-features", cfg.ThingsConfig.Features, "Specify the desired Ditto features that will be registered for the containers Ditto thing")

	// init update agent
	flagSet.BoolVar(&cfg.UpdateAgentConfig.UpdateAgentEnable, "ua-enable", cfg.UpdateAgentConfig.UpdateAgentEnable, "Enable the update agent for containers")
	flagSet.StringVar(&cfg.UpdateAgentConfig.DomainName, "ua-domain", cfg.UpdateAgentConfig.DomainName, "Specify the domain name for the containers update agent")
	flagSet.StringSliceVar(&cfg.UpdateAgentConfig.SystemContainers, "ua-system-containers", cfg.UpdateAgentConfig.SystemContainers, "Specify the list of system containers which shall be skipped during update process by the update agent")
	flagSet.BoolVar(&cfg.UpdateAgentConfig.VerboseInventoryReport, "ua-verbose-inventory-report", cfg.UpdateAgentConfig.VerboseInventoryReport, "Enables verbose reporting of current inventory of containers by the update agent")

	// init local communication flags
	flagSet.StringVar(&cfg.LocalConnection.BrokerURL, "conn-broker-url", cfg.LocalConnection.BrokerURL, "Specify the MQTT broker URL to connect to")
	flagSet.StringVar(&cfg.LocalConnection.KeepAlive, "conn-keep-alive", cfg.LocalConnection.KeepAlive, "Specify the keep alive duration for the MQTT requests as duration string")
	flagSet.StringVar(&cfg.LocalConnection.DisconnectTimeout, "conn-disconnect-timeout", cfg.LocalConnection.DisconnectTimeout, "Specify the disconnection timeout for the MQTT connection as duration string")
	flagSet.StringVar(&cfg.LocalConnection.ClientUsername, "conn-client-username", cfg.LocalConnection.ClientUsername, "Specify the MQTT client username to authenticate with")
	flagSet.StringVar(&cfg.LocalConnection.ClientPassword, "conn-client-password", cfg.LocalConnection.ClientPassword, "Specify the MQTT client password to authenticate with")
	flagSet.StringVar(&cfg.LocalConnection.ConnectTimeout, "conn-connect-timeout", cfg.LocalConnection.ConnectTimeout, "Specify the connect timeout for the MQTT as duration string")
	flagSet.StringVar(&cfg.LocalConnection.AcknowledgeTimeout, "conn-ack-timeout", cfg.LocalConnection.AcknowledgeTimeout, "Specify the acknowledgement timeout for the MQTT requests as duration string")
	flagSet.StringVar(&cfg.LocalConnection.SubscribeTimeout, "conn-sub-timeout", cfg.LocalConnection.SubscribeTimeout, "Specify the subscribe timeout for the MQTT requests as duration string")
	flagSet.StringVar(&cfg.LocalConnection.UnsubscribeTimeout, "conn-unsub-timeout", cfg.LocalConnection.UnsubscribeTimeout, "Specify the unsubscribe timeout for the MQTT requests as duration string")

	// init tls support
	if cfg.LocalConnection.Transport == nil {
		cfg.LocalConnection.Transport = &tlsConfig{}
	}
	flagSet.StringVar(&cfg.LocalConnection.Transport.RootCA, "conn-root-ca", cfg.LocalConnection.Transport.RootCA, "Specify the PEM encoded CA certificates file")
	flagSet.StringVar(&cfg.LocalConnection.Transport.ClientCert, "conn-client-cert", cfg.LocalConnection.Transport.ClientCert, "Specify the PEM encoded certificate file to authenticate to the MQTT server/broker")
	flagSet.StringVar(&cfg.LocalConnection.Transport.ClientKey, "conn-client-key", cfg.LocalConnection.Transport.ClientKey, "Specify the PEM encoded unencrypted private key file to authenticate to the MQTT server/broker")

	//TODO remove in M5
	setupDeprecatedCommandFlags(flagSet)
}

func setupDeprecatedCommandFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.BrokerURL, "things-conn-broker", cfg.ThingsConfig.ThingsConnectionConfig.BrokerURL, "DEPRECATED Specify the MQTT broker URL to connect to")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.KeepAlive, "things-conn-keep-alive", cfg.ThingsConfig.ThingsConnectionConfig.KeepAlive, "DEPRECATED Specify the keep alive duration for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout, "things-conn-disconnect-timeout", cfg.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout, "DEPRECATED Specify the disconnection timeout for the MQTT connection in milliseconds")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.ClientUsername, "things-conn-client-username", cfg.ThingsConfig.ThingsConnectionConfig.ClientUsername, "DEPRECATED Specify the MQTT client username to authenticate with")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.ClientPassword, "things-conn-client-password", cfg.ThingsConfig.ThingsConnectionConfig.ClientPassword, "DEPRECATED Specify the MQTT client password to authenticate with")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.ConnectTimeout, "things-conn-connect-timeout", cfg.ThingsConfig.ThingsConnectionConfig.ConnectTimeout, "DEPRECATED Specify the connect timeout for the MQTT in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout, "things-conn-ack-timeout", cfg.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout, "DEPRECATED Specify the acknowledgement timeout for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout, "things-conn-sub-timeout", cfg.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout, "DEPRECATED Specify the subscribe timeout for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout, "things-conn-unsub-timeout", cfg.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout, "DEPRECATED Specify the unsubscribe timeout for the MQTT requests in milliseconds")

	// init tls support
	if cfg.ThingsConfig.ThingsConnectionConfig.Transport == nil {
		cfg.ThingsConfig.ThingsConnectionConfig.Transport = &tlsConfig{}
	}
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.Transport.RootCA, "things-conn-root-ca", cfg.ThingsConfig.ThingsConnectionConfig.Transport.RootCA, "DEPRECATED Specify the PEM encoded CA certificates file")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.Transport.ClientCert, "things-conn-client-cert", cfg.ThingsConfig.ThingsConnectionConfig.Transport.ClientCert, "DEPRECATED Specify the PEM encoded certificate file to authenticate to the MQTT server/broker")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.Transport.ClientKey, "things-conn-client-key", cfg.ThingsConfig.ThingsConnectionConfig.Transport.ClientKey, "DEPRECATED Specify the PEM encoded unencrypted private key file to authenticate to the MQTT server/broker")
}
