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

package client

import (
	"time"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
)

// InitializedHook is used for initialized notification
type InitializedHook func(client *Client, configuration *Configuration, err error)

// Configuration provides the Client's configuration
type Configuration struct {
	broker                       string
	keepAlive                    time.Duration
	clientUsername               string
	clientPassword               string
	disconnectTimeout            time.Duration
	gatewayDeviceID              string
	deviceName                   string
	deviceAuthID                 string
	devicePassword               string
	deviceTenantID               string
	connectTimeout               time.Duration
	acknowledgeTimeout           time.Duration
	subscribeTimeout             time.Duration
	unsubscribeTimeout           time.Duration
	initHook                     InitializedHook
	thingsRegistryChangedHandler handlers.ThingsRegistryChangedHandler
	tlsConfig                    *tlsConfig
}

// tlsConfig represents the TLS configuration data
type tlsConfig struct {
	RootCA     string
	ClientCert string
	ClientKey  string
}

// NewConfiguration creates a new Configuration instance
func NewConfiguration() *Configuration {
	return &Configuration{
		broker:             defaultBroker,
		keepAlive:          defaultKeepAlive,
		disconnectTimeout:  defaultDisconnectTimeout,
		connectTimeout:     defaultConnectTimeout,
		acknowledgeTimeout: defaultAcknowledgeTimeout,
		subscribeTimeout:   defaultSubscribeTimeout,
		unsubscribeTimeout: defaultUnsubscribeTimeout,
	}
}

// Broker provides the current MQTT broker the client is to connect to
func (cfg *Configuration) Broker() string {
	return cfg.broker
}

// KeepAlive provides the keep alive connection's period
func (cfg *Configuration) KeepAlive() time.Duration {
	return cfg.keepAlive
}

// DisconnectTimeout provides the timeout for disconnecting the client
func (cfg *Configuration) DisconnectTimeout() time.Duration {
	return cfg.disconnectTimeout
}

// ClientUsername provides the currently configured username authentication used for the underlying connection
func (cfg *Configuration) ClientUsername() string {
	return cfg.clientUsername
}

// ClientPassword provides the currently configured password authentication used for the underlying connection
func (cfg *Configuration) ClientPassword() string {
	return cfg.clientPassword
}

// GatewayDeviceID provides the currently configured gateway device ID
func (cfg *Configuration) GatewayDeviceID() string {
	return cfg.gatewayDeviceID
}

// DeviceName provides the currently configured device name
func (cfg *Configuration) DeviceName() string {
	return cfg.deviceName
}

// DeviceAuthID provides the currently configured device authentication ID
func (cfg *Configuration) DeviceAuthID() string {
	return cfg.deviceAuthID
}

// DevicePassword provides the currently configured device password
func (cfg *Configuration) DevicePassword() string {
	return cfg.devicePassword
}

// DeviceTenantID provides the currently configured device tenant ID
func (cfg *Configuration) DeviceTenantID() string {
	return cfg.deviceTenantID
}

// ConnectTimeout provides the currently configured connection timeout
func (cfg *Configuration) ConnectTimeout() time.Duration {
	return cfg.connectTimeout
}

// AcknowledgeTimeout provides the currently configured acknowledge timeout
func (cfg *Configuration) AcknowledgeTimeout() time.Duration {
	return cfg.acknowledgeTimeout
}

// SubscribeTimeout provides the currently configured subscribe timeout
func (cfg *Configuration) SubscribeTimeout() time.Duration {
	return cfg.subscribeTimeout
}

// UnsubscribeTimeout provides the currently configured unsubscribe timeout
func (cfg *Configuration) UnsubscribeTimeout() time.Duration {
	return cfg.unsubscribeTimeout
}

// InitHook provides the currently configured initialized notification
func (cfg *Configuration) InitHook() InitializedHook {
	return cfg.initHook
}

// RegistryChangedHandler provides the currently configured things registry changed handler
func (cfg *Configuration) RegistryChangedHandler() handlers.ThingsRegistryChangedHandler {
	return cfg.thingsRegistryChangedHandler
}

// TLSConfig provides the current TLS configuration
func (cfg *Configuration) TLSConfig() (rootCA, clientCert, clientKey string) {
	if cfg.tlsConfig == nil {
		return "", "", ""
	}
	return cfg.tlsConfig.RootCA, cfg.tlsConfig.ClientCert, cfg.tlsConfig.ClientKey
}

// WithBroker configures the MQTT's broker the Client to connect to
func (cfg *Configuration) WithBroker(broker string) *Configuration {
	cfg.broker = broker
	return cfg
}

// WithKeepAlive configures the keep alive time period for the underlying Client's connection
func (cfg *Configuration) WithKeepAlive(keepAlive time.Duration) *Configuration {
	cfg.keepAlive = keepAlive
	return cfg
}

// WithClientUsername configures the client username
func (cfg *Configuration) WithClientUsername(username string) *Configuration {
	cfg.clientUsername = username
	return cfg
}

// WithClientPassword configures the client password
func (cfg *Configuration) WithClientPassword(password string) *Configuration {
	cfg.clientPassword = password
	return cfg
}

// WithDisconnectTimeout configures the timeout for disconnection of the Client
func (cfg *Configuration) WithDisconnectTimeout(disconnectTimeout time.Duration) *Configuration {
	cfg.disconnectTimeout = disconnectTimeout
	return cfg
}

// WithGatewayDeviceID configures the gateway device ID
func (cfg *Configuration) WithGatewayDeviceID(deviceID string) *Configuration {
	cfg.gatewayDeviceID = deviceID
	return cfg
}

// WithDeviceName configures the device name
func (cfg *Configuration) WithDeviceName(name string) *Configuration {
	cfg.deviceName = name
	return cfg
}

// WithDeviceAuthID configures the device authentication ID
func (cfg *Configuration) WithDeviceAuthID(deviceAuthID string) *Configuration {
	cfg.deviceAuthID = deviceAuthID
	return cfg
}

// WithDevicePassword configures the device password
func (cfg *Configuration) WithDevicePassword(devicePassword string) *Configuration {
	cfg.devicePassword = devicePassword
	return cfg
}

// WithDeviceTenantID configures the device tenant ID
func (cfg *Configuration) WithDeviceTenantID(deviceTenantID string) *Configuration {
	cfg.deviceTenantID = deviceTenantID
	return cfg
}

// WithInitHook configures the initialized notification
func (cfg *Configuration) WithInitHook(hook InitializedHook) *Configuration {
	cfg.initHook = hook
	return cfg
}

// WithThingsRegistryChangedHandler configures the things registry changed handler
func (cfg *Configuration) WithThingsRegistryChangedHandler(handler handlers.ThingsRegistryChangedHandler) *Configuration {
	cfg.thingsRegistryChangedHandler = handler
	return cfg
}

// WithConnectTimeout configures the connect timeout
func (cfg *Configuration) WithConnectTimeout(connectTimeout time.Duration) *Configuration {
	cfg.connectTimeout = connectTimeout
	return cfg
}

// WithAcknowledgeTimeout configures acknowledged timeout
func (cfg *Configuration) WithAcknowledgeTimeout(acknowledgeTimeout time.Duration) *Configuration {
	cfg.acknowledgeTimeout = acknowledgeTimeout
	return cfg
}

// WithSubscribeTimeout configures subscribe timeout
func (cfg *Configuration) WithSubscribeTimeout(subscribeTimeout time.Duration) *Configuration {
	cfg.subscribeTimeout = subscribeTimeout
	return cfg
}

// WithUnsubscribeTimeout configures unsubscribe timeout
func (cfg *Configuration) WithUnsubscribeTimeout(unsubscribeTimeout time.Duration) *Configuration {
	cfg.unsubscribeTimeout = unsubscribeTimeout
	return cfg
}

// WithTLSConfig configures the TLS options to the MQTT server/broker
func (cfg *Configuration) WithTLSConfig(rootCA, clientCert, clientKey string) *Configuration {
	cfg.tlsConfig = &tlsConfig{
		RootCA:     rootCA,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}
	return cfg
}
