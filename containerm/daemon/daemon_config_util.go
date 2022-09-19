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
	"encoding/json"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"io/ioutil"
	"os"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/network"
	"github.com/eclipse-kanto/container-management/containerm/server"
	"github.com/eclipse-kanto/container-management/containerm/things"
	"github.com/spf13/pflag"
)

const daemonConfigFileFlagID = "cfg-file"

func extractCtrClientConfigOptions(daemonConfig *config) []ctr.ContainerOpts {
	ctrOpts := []ctr.ContainerOpts{}
	ctrOpts = append(ctrOpts,
		ctr.WithCtrdConnectionPath(daemonConfig.ContainerClientConfig.CtrAddressPath),
		ctr.WithCtrdNamespace(daemonConfig.ContainerClientConfig.CtrNamespace),
		ctr.WithCtrdRootExec(daemonConfig.ContainerClientConfig.CtrRootExec),
		ctr.WithCtrdMetaPath(daemonConfig.ContainerClientConfig.CtrMetaPath),
		ctr.WithCtrdRegistryConfigs(parseRegistryConfigs(daemonConfig.ContainerClientConfig.CtrRegistryConfigs, daemonConfig.ContainerClientConfig.CtrInsecureRegistries)),
		ctr.WithCtrdImageDecryptKeys(daemonConfig.ContainerClientConfig.CtrImageDecKeys...),
		ctr.WithCtrdImageDecryptRecipients(daemonConfig.ContainerClientConfig.CtrImageDecRecipients...),
		ctr.WithCtrdRuncRuntime(daemonConfig.ContainerClientConfig.CtrRuncRuntime),
		ctr.WithCtrdImageExpiry(daemonConfig.ContainerClientConfig.CtrImageExpiry),
		ctr.WithCtrdImageExpiryDisable(daemonConfig.ContainerClientConfig.CtrImageExpiryDisable),
		ctr.WithCtrdLeaseID(daemonConfig.ContainerClientConfig.CtrLeaseID),
	)
	return ctrOpts
}

func extractNetManagerConfigOptions(daemonConfig *config) []network.NetOpt {
	netOpts := []network.NetOpt{}
	netOpts = append(netOpts,
		network.WithLibNetType(daemonConfig.NetworkConfig.NetType),
		network.WithLibNetMetaPath(daemonConfig.NetworkConfig.NetMetaPath),
		network.WithLibNetExecRoot(daemonConfig.NetworkConfig.NetExecRoot),
		network.WithLibNetDisableBridge(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeDisableBridge),
		network.WithLibNetName(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeName),
		network.WithLibNetIPV4(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPV4),
		network.WithLibNetFixedCIDRv4(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeFixedCIDRv4),
		network.WithLibNetGatewayIPv4(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeGatewayIPv4),
		network.WithLibNetEnableIPv6(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeEnableIPv6),
		network.WithLibNetMtu(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeMtu),
		network.WithLibNetIcc(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIcc),
		network.WithLibNetIPTables(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPTables),
		network.WithLibNetIPForward(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPForward),
		network.WithLibNetIPMasq(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPMasq),
		network.WithLibNetUserlandProxy(daemonConfig.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeUserlandProxy),
	)
	return netOpts
}

func extractContainerManagerOptions(daemonConfig *config) []mgr.ContainerManagerOpt {
	mgrOpts := []mgr.ContainerManagerOpt{}
	mgrOpts = append(mgrOpts,
		mgr.WithMgrMetaPath(daemonConfig.ManagerConfig.MgrMetaPath),
		mgr.WithMgrRootExec(daemonConfig.ManagerConfig.MgrExecPath),
		mgr.WithMgrContainerClientServiceID(daemonConfig.ManagerConfig.MgrCtrClientServiceID),
		mgr.WithMgrNetworkManagerServiceID(daemonConfig.ManagerConfig.MgrNetMgrServiceID),
		mgr.WithMgrDefaultContainerStopTimeout(daemonConfig.ManagerConfig.MgrDefaultCtrsStopTimeout),
	)
	return mgrOpts
}

func extractGrpcOptions(daemonConfig *config) []server.GrpcServerOpt {
	grpcServerOpts := []server.GrpcServerOpt{}
	grpcServerOpts = append(grpcServerOpts,
		server.WithGrpcServerAddressPath(daemonConfig.GrpcServerConfig.GrpcServerAddressPath),
		server.WithGrpcServerNetwork(daemonConfig.GrpcServerConfig.GrpcServerNetworkProtocol),
	)
	return grpcServerOpts
}

func extractThingsOptions(daemonConfig *config) []things.ContainerThingsManagerOpt {
	thingsOpts := []things.ContainerThingsManagerOpt{}
	thingsOpts = append(thingsOpts,
		things.WithMetaPath(daemonConfig.ThingsConfig.ThingsMetaPath),
		things.WithFeatures(daemonConfig.ThingsConfig.Features),
		things.WithConnectionBroker(daemonConfig.ThingsConfig.ThingsConnectionConfig.BrokerURL),
		things.WithConnectionKeepAlive(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.KeepAlive)*time.Millisecond),
		things.WithConnectionDisconnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)*time.Millisecond),
		things.WithConnectionClientUsername(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientUsername),
		things.WithConnectionClientPassword(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientPassword),
		things.WithConnectionConnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)*time.Millisecond),
		things.WithConnectionAcknowledgeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)*time.Millisecond),
		things.WithConnectionSubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)*time.Millisecond),
		things.WithConnectionUnsubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)*time.Millisecond),
	)
	return thingsOpts
}

func initLogger(daemonConfig *config) {
	log.Configure(daemonConfig.Log)
}

func loadLocalConfig(filePath string, localConfig *config) error {

	fi, fierr := os.Stat(filePath)
	if fierr != nil {
		if os.IsNotExist(fierr) {
			return nil
		}
		return fierr
	} else if fi.IsDir() {
		return log.NewErrorf("provided configuration path %s is a directory", filePath)
	} else if fi.Size() == 0 {
		log.Warn("the file %s is empty", filePath)
		return nil
	} else {
		log.Debug("successfully retrieved file %s stats", filePath)
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, localConfig)
	if err != nil {
		return err
	}
	return nil
}

func parseConfigFilePath() string {
	var cfgFilePath string
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagSet.StringVar(&cfgFilePath, daemonConfigFileFlagID, daemonConfigFileDefault, "Specify the configuration file of container-management")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Info("there are flags set for starting the container management instance with a configuration from the command line - file and default configurations will be overridden")
	}
	log.Info("local daemon configuration is set to [%s]", cfgFilePath)
	return cfgFilePath
}

func dumpConfiguration(configInstance *config) {
	if configInstance == nil {
		return
	}
	// dump debug config
	dumpDebug(configInstance)

	// dump container manager config
	dumpContManager(configInstance)

	// dump container client config
	dumpContClient(configInstance)

	// dump network manager config
	dumpNetworkManager(configInstance)

	// dump grpc server config
	dumpGRPCServer(configInstance)

	// dump things client config
	dumpThingsClient(configInstance)
}

func dumpDebug(configInstance *config) {
	if configInstance.Log != nil {
		log.Debug("[daemon_cfg][log-level] : %v", configInstance.Log.LogLevel)
		log.Debug("[daemon_cfg][log-enable-syslog] : %v", configInstance.Log.Syslog)
		if configInstance.Log.LogFile != "" {
			log.Debug("[daemon_cfg][log-file] : %s", configInstance.Log.LogFile)
			log.Debug("[daemon_cfg][log-file-size] : %d", configInstance.Log.LogFileSize)
			log.Debug("[daemon_cfg][log-file-count] : %d", configInstance.Log.LogFileCount)
			log.Debug("[daemon_cfg][log-file-max-age] : %d", configInstance.Log.LogFileMaxAge)
		}
	}
}

func dumpContManager(configInstance *config) {
	if configInstance.ManagerConfig != nil {
		log.Debug("[daemon_cfg][cm-home-dir] : %s", configInstance.ManagerConfig.MgrMetaPath)
		log.Debug("[daemon_cfg][cm-exec-root-dir] : %s", configInstance.ManagerConfig.MgrExecPath)
		log.Debug("[daemon_cfg][cm-cc-sid] : %s", configInstance.ManagerConfig.MgrCtrClientServiceID)
		log.Debug("[daemon_cfg][cm-net-sid] : %s", configInstance.ManagerConfig.MgrNetMgrServiceID)
		log.Debug("[daemon_cfg][cm-deflt-ctrs-stop-timeout] : %d", configInstance.ManagerConfig.MgrDefaultCtrsStopTimeout)
	}
}

func dumpContClient(configInstance *config) {
	if configInstance.ContainerClientConfig != nil {
		log.Debug("[daemon_cfg][ccl-default-ns] : %s", configInstance.ContainerClientConfig.CtrNamespace)
		log.Debug("[daemon_cfg][ccl-ap] : %s", configInstance.ContainerClientConfig.CtrAddressPath)
		log.Debug("[daemon_cfg][ccl-insecure-registries] : %s", configInstance.ContainerClientConfig.CtrInsecureRegistries)
		registryConfigHosts := dumpRegistryConfigHosts(configInstance.ContainerClientConfig.CtrRegistryConfigs)
		if registryConfigHosts != nil {
			log.Debug("[daemon_cfg][ccl-registry_configurations] : %s", registryConfigHosts)
		}
		log.Debug("[daemon_cfg][ccl-exec-root-dir] : %s", configInstance.ContainerClientConfig.CtrRootExec)
		log.Debug("[daemon_cfg][ccl-home-dir] : %s", configInstance.ContainerClientConfig.CtrMetaPath)
		log.Debug("[daemon_cfg][ccl-image-dec-keys] : %s", configInstance.ContainerClientConfig.CtrImageDecKeys)
		log.Debug("[daemon_cfg][ccl-image-dec-recipients] : %s", configInstance.ContainerClientConfig.CtrImageDecRecipients)
		r := types.Runtime(configInstance.ContainerClientConfig.CtrRuncRuntime)
		log.Debug("[daemon_cfg][ccl-runc-runtime] : %s", r)
		if r == types.RuntimeTypeV1 || r == types.RuntimeTypeV2runcV1 {
			log.Warn("runtime %s is deprecated since containerd v1.4, consider using %s", r, types.RuntimeTypeV2runcV2)
		}
		log.Debug("[daemon_cfg][ccl-image-expiry] : %s", configInstance.ContainerClientConfig.CtrImageExpiry)
		log.Debug("[daemon_cfg][ccl-image-expiry-disable] : %v", configInstance.ContainerClientConfig.CtrImageExpiryDisable)
		log.Debug("[daemon_cfg][ccl-lease-id] : %s", configInstance.ContainerClientConfig.CtrLeaseID)
	}
}

func dumpNetworkManager(configInstance *config) {
	if configInstance.NetworkConfig != nil {
		log.Debug("[daemon_cfg][net-type] : %s", configInstance.NetworkConfig.NetType)
		log.Debug("[daemon_cfg][net-home-dir] : %s", configInstance.NetworkConfig.NetMetaPath)
		log.Debug("[daemon_cfg][net-exec-root-dir] : %s", configInstance.NetworkConfig.NetExecRoot)

		// dump default bridge network config
		if configInstance.NetworkConfig.DefaultBridgeNetworkConfig != nil {
			log.Debug("[daemon_cfg][net-tbr-disable] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeDisableBridge)
			if !configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeDisableBridge {
				log.Debug("[daemon_cfg][net-br-name] : %s", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeName)
				if configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPV4 != "" {
					log.Debug("[daemon_cfg][net-br-ip4] : %s", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPV4)
				}
				if configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeFixedCIDRv4 != "" {
					log.Debug("[daemon_cfg][net-br-fcidr4] : %s", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeFixedCIDRv4)
				}
				if configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeGatewayIPv4 != "" {
					log.Debug("[daemon_cfg][net-br-gwip4] : %s", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeGatewayIPv4)
				}
				log.Debug("[daemon_cfg][net-br-enable-ip6] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeEnableIPv6)
				log.Debug("[daemon_cfg][net-br-mtu] : %d", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeMtu)
				log.Debug("[daemon_cfg][net-br-icc] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIcc)
				log.Debug("[daemon_cfg][et-br-ipt] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPTables)
				log.Debug("[daemon_cfg][net-br-ipfw] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPForward)
				log.Debug("[daemon_cfg][net-br-ipmasq] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeIPMasq)
				log.Debug("[daemon_cfg][net-br-ulp] : %v", configInstance.NetworkConfig.DefaultBridgeNetworkConfig.NetBridgeUserlandProxy)
			}
		}
	}
}

func dumpGRPCServer(configInstance *config) {
	if configInstance.GrpcServerConfig != nil {
		log.Debug("[daemon_cfg][grpc-serv-netp] : %s", configInstance.GrpcServerConfig.GrpcServerNetworkProtocol)
		log.Debug("[daemon_cfg][grpc-serv-ap] : %s", configInstance.GrpcServerConfig.GrpcServerAddressPath)
	}
}

func dumpThingsClient(configInstance *config) {
	if configInstance.ThingsConfig != nil {
		log.Debug("[daemon_cfg][things-enable] : %v", configInstance.ThingsConfig.ThingsEnable)
		if configInstance.ThingsConfig.ThingsEnable {
			log.Debug("[daemon_cfg][things-home-dir] : %s", configInstance.ThingsConfig.ThingsMetaPath)
			log.Debug("[daemon_cfg][things-features] : %s", configInstance.ThingsConfig.Features)
			if configInstance.ThingsConfig.ThingsConnectionConfig != nil {
				log.Debug("[daemon_cfg][things-conn-broker] : %s", configInstance.ThingsConfig.ThingsConnectionConfig.BrokerURL)
				log.Debug("[daemon_cfg][things-conn-keep-alive] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.KeepAlive)
				log.Debug("[daemon_cfg][things-conn-disconnect-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)
				log.Debug("[daemon_cfg][things-conn-connect-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)
				log.Debug("[daemon_cfg][things-conn-ack-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)
				log.Debug("[daemon_cfg][things-conn-sub-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)
				log.Debug("[daemon_cfg][things-conn-unsub-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)
			}
		}
	}
}

func dumpRegistryConfigHosts(configs map[string]*registryConfig) []string {
	if configs != nil {
		secureRegistryHosts := make([]string, len(configs))
		i := 0
		for host := range configs {
			secureRegistryHosts[i] = host
			i++
		}
		return secureRegistryHosts
	}
	return nil
}

func parseRegistryConfigs(configs map[string]*registryConfig, insecureRegs []string) map[string]*ctr.RegistryConfig {
	var ctrRegConfigs map[string]*ctr.RegistryConfig
	if len(configs) != 0 {
		ctrRegConfigs = make(map[string]*ctr.RegistryConfig)
		for host, conf := range configs {
			if host == "" {
				log.Warn("[daemon_cfg] registry configuration parse failed for configuration %+v and it will not be added to the container-management configuration. Host is not provided", conf)
				continue
			}
			regConf := &ctr.RegistryConfig{
				IsInsecure: false,
			}
			if conf.Credentials != nil {
				regConf.Credentials = &ctr.AuthCredentials{
					UserID:   conf.Credentials.UserID,
					Password: conf.Credentials.Password,
				}
			}
			if conf.Transport != nil {
				regConf.Transport = &ctr.TLSConfig{
					RootCA:     conf.Transport.RootCA,
					ClientCert: conf.Transport.ClientCert,
					ClientKey:  conf.Transport.ClientKey,
				}
			}
			ctrRegConfigs[host] = regConf
			log.Debug("[daemon_cfg] successfully parsed configuration for secure registry with host %s", host)
		}
	}
	return applyInsecureRegistryConfig(ctrRegConfigs, insecureRegs)
}

func applyInsecureRegistryConfig(registriesConfig map[string]*ctr.RegistryConfig, insecureRegs []string) map[string]*ctr.RegistryConfig {
	if insecureRegs == nil || len(insecureRegs) == 0 {
		log.Debug("no insecure registries provided")
		return registriesConfig
	}
	res := registriesConfig
	addAll := res == nil
	if addAll {
		res = make(map[string]*ctr.RegistryConfig)
	}
	for _, insecReg := range insecureRegs {
		if addAll || res[insecReg] == nil {
			res[insecReg] = &ctr.RegistryConfig{
				IsInsecure: true,
			}
		} else {
			res[insecReg].IsInsecure = true
		}
		log.Debug("[daemon_cfg] successfully parsed configuration for insecure registry with host %s", insecReg)
	}
	return res
}
