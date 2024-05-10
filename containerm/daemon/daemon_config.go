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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// config refers to daemon's whole configurations.
type config struct {
	Log *log.Config `json:"log,omitempty"`

	DeploymentManagerConfig *deploymentManagerConfig `json:"deployment,omitempty"`

	ManagerConfig *managerConfig `json:"manager,omitempty"`

	ContainerClientConfig *containerRuntimeConfig `json:"containers,omitempty"`

	NetworkConfig *networkConfig `json:"network,omitempty"`

	GrpcServerConfig *grpcServerConfig `json:"grpc_server,omitempty"`

	ThingsConfig *thingsConfig `json:"things,omitempty"`

	UpdateAgentConfig *updateAgentConfig `json:"update_agent,omitempty"`

	LocalConnection *localConnectionConfig `json:"connection,omitempty"`
}

// container mgr config
type managerConfig struct {
	MgrMetaPath               string `json:"home_dir,omitempty"`
	MgrExecPath               string `json:"exec_root_dir,omitempty"`
	MgrCtrClientServiceID     string `json:"container_client_sid,omitempty"`
	MgrNetMgrServiceID        string `json:"network_manager_sid,omitempty"`
	MgrDefaultCtrsStopTimeout string `json:"default_ctrs_stop_timeout,omitempty"`
}

func (mc *managerConfig) UnmarshalJSON(data []byte) error {
	type managerConfigPlain managerConfig

	plain := (*managerConfigPlain)(mc)
	err := json.Unmarshal(data, &plain)
	if err == nil {
		return nil
	}

	tmp := struct {
		MgrDefaultCtrsStopTimeout int `json:"default_ctrs_stop_timeout,omitempty"`
		*managerConfigPlain
	}{
		managerConfigPlain: plain,
	}

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.MgrDefaultCtrsStopTimeout != 0 {
		mc.MgrDefaultCtrsStopTimeout = strconv.Itoa(tmp.MgrDefaultCtrsStopTimeout)
	}
	return nil
}

// container client config- e.g. containerd
type containerRuntimeConfig struct {
	CtrNamespace           string                     `json:"default_ns,omitempty"`
	CtrAddressPath         string                     `json:"address_path,omitempty"`
	CtrRegistryConfigs     map[string]*registryConfig `json:"registry_configurations,omitempty"`
	CtrInsecureRegistries  []string                   `json:"insecure_registries,omitempty"`
	CtrRootExec            string                     `json:"exec_root_dir,omitempty"`
	CtrMetaPath            string                     `json:"home_dir,omitempty"`
	CtrImageDecKeys        []string                   `json:"image_dec_keys,omitempty"`
	CtrImageDecRecipients  []string                   `json:"image_dec_recipients,omitempty"`
	CtrRuncRuntime         string                     `json:"runc_runtime,omitempty"`
	CtrImageExpiry         time.Duration              `json:"image_expiry,omitempty"`
	CtrImageExpiryDisable  bool                       `json:"image_expiry_disable,omitempty"`
	CtrLeaseID             string                     `json:"lease_id,omitempty"`
	CtrImageVerifierType   string                     `json:"image_verifier_type,omitempty"`
	CtrImageVerifierConfig verifierConfig             `json:"image_verifier_config,omitempty"`
}

// deployment manager config
type deploymentManagerConfig struct {
	DeploymentEnable   bool   `json:"enable,omitempty"`
	DeploymentMode     string `json:"mode,omitempty"`
	DeploymentMetaPath string `json:"home_dir,omitempty"`
	DeploymentCtrPath  string `json:"ctr_dir,omitempty"`
}

func (cfg *containerRuntimeConfig) UnmarshalJSON(data []byte) error {
	type containerRuntimeConfigPlain containerRuntimeConfig

	tmp := struct {
		CtrImageExpiry string `json:"image_expiry,omitempty"`
		*containerRuntimeConfigPlain
	}{
		containerRuntimeConfigPlain: (*containerRuntimeConfigPlain)(cfg),
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	if tmp.CtrImageExpiry != "" {
		cfg.CtrImageExpiry, err = time.ParseDuration(tmp.CtrImageExpiry)
		if err != nil {
			return err
		}
	}
	return nil
}

// registry config
type registryConfig struct {
	Credentials *authCredentials `json:"credentials,omitempty"`
	Transport   *tlsConfig       `json:"transport"`
}

// basic authentication config
type authCredentials struct {
	UserID   string `json:"user_id,omitempty"`
	Password string `json:"password,omitempty"`
}

// tls-secured communication config
type tlsConfig struct {
	RootCA     string `json:"root_ca"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
}

// network manager config - e.g. for the Libnetwork client
type networkConfig struct {
	NetType                    string               `json:"type,omitempty"`
	NetMetaPath                string               `json:"home_dir,omitempty"`
	NetExecRoot                string               `json:"exec_root_dir,omitempty"`
	DefaultBridgeNetworkConfig *bridgeNetworkConfig `json:"default_bridge,omitempty"`
}

// network default bridge network config - kanto-cm0
type bridgeNetworkConfig struct {
	NetBridgeDisableBridge bool   `json:"disable,omitempty"`
	NetBridgeName          string `json:"name,omitempty"`
	NetBridgeIPV4          string `json:"ip4,omitempty"`
	NetBridgeFixedCIDRv4   string `json:"fcidr4,omitempty"`
	NetBridgeGatewayIPv4   string `json:"gwip4,omitempty"`
	NetBridgeEnableIPv6    bool   `json:"enable_ip6,omitempty"`

	NetBridgeMtu           int  `json:"mtu,omitempty"`
	NetBridgeIcc           bool `json:"icc,omitempty"`
	NetBridgeIPTables      bool `json:"ip_tables,omitempty"`
	NetBridgeIPForward     bool `json:"ip_forward,omitempty"`
	NetBridgeIPMasq        bool `json:"ip_masq,omitempty"`
	NetBridgeUserlandProxy bool `json:"userland_proxy,omitempty"`
}

// grpc server config
type grpcServerConfig struct {
	GrpcServerNetworkProtocol string `json:"protocol,omitempty"`
	GrpcServerAddressPath     string `json:"address_path,omitempty"`
}

// things client configuration
type thingsConfig struct {
	ThingsEnable           bool                    `json:"enable,omitempty"`
	ThingsMetaPath         string                  `json:"home_dir,omitempty"`
	Features               []string                `json:"features,omitempty"`
	ThingsConnectionConfig *thingsConnectionConfig `json:"connection,omitempty"`
}

// things client configuration
type updateAgentConfig struct {
	UpdateAgentEnable      bool     `json:"enable,omitempty"`
	DomainName             string   `json:"domain,omitempty"`
	SystemContainers       []string `json:"system_containers,omitempty"`
	VerboseInventoryReport bool     `json:"verbose_inventory_report,omitempty"`
}

// local connection config
type localConnectionConfig struct {
	BrokerURL          string     `json:"broker_url,omitempty"`
	KeepAlive          string     `json:"keep_alive,omitempty"`
	DisconnectTimeout  string     `json:"disconnect_timeout,omitempty"`
	ClientUsername     string     `json:"client_username,omitempty"`
	ClientPassword     string     `json:"client_password,omitempty"`
	ConnectTimeout     string     `json:"connect_timeout,omitempty"`
	AcknowledgeTimeout string     `json:"acknowledge_timeout,omitempty"`
	SubscribeTimeout   string     `json:"subscribe_timeout,omitempty"`
	UnsubscribeTimeout string     `json:"unsubscribe_timeout,omitempty"`
	Transport          *tlsConfig `json:"transport,omitempty"`
}

// TODO Remove in M5
// things service connection config
type thingsConnectionConfig struct {
	BrokerURL          string     `json:"broker_url,omitempty"`
	KeepAlive          int64      `json:"keep_alive,omitempty"`
	DisconnectTimeout  int64      `json:"disconnect_timeout,omitempty"`
	ClientUsername     string     `json:"client_username,omitempty"`
	ClientPassword     string     `json:"client_password,omitempty"`
	ConnectTimeout     int64      `json:"connect_timeout,omitempty"`
	AcknowledgeTimeout int64      `json:"acknowledge_timeout,omitempty"`
	SubscribeTimeout   int64      `json:"subscribe_timeout,omitempty"`
	UnsubscribeTimeout int64      `json:"unsubscribe_timeout,omitempty"`
	Transport          *tlsConfig `json:"transport,omitempty"`
}

// verifierConfig is alias used as flag value
type verifierConfig map[string]string

// String is representation of verifierConfig as a comma separated {key}={value} pairs
func (vc *verifierConfig) String() string {
	if len(*vc) == 0 {
		return ""
	}
	var pairs []string
	for key, value := range *vc {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(pairs, ",")
}

// Set verifierConfig from string, used for flag set
func (vc *verifierConfig) Set(value string) error {
	if value == "" {
		return log.NewError("the image verifier config could not be empty")
	}
	if *vc == nil {
		*vc = make(map[string]string)
	}

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" || value == "" {
			return log.NewErrorf("could not parse image verification config, invalid key-value pair - %s", pair)
		}
		(*vc)[key] = value
	}
	return nil
}

// Type returns the verifierConfig flag type
func (vc verifierConfig) Type() string {
	return "stringSlice"
}
