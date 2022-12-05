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
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/util/tls"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	defaultDisconnectTimeout  = 250 * time.Millisecond
	defaultKeepAlive          = 30 * time.Second
	defaultConnectTimeout     = 30 * time.Second
	defaultAcknowledgeTimeout = 15 * time.Second
	defaultSubscribeTimeout   = 15 * time.Second
	defaultUnsubscribeTimeout = 5 * time.Second
	defaultBroker             = "tcp://localhost:1883"
)

// type CommandsHandler func(thingId string, featureId string, commandId string, args interface{}) (interface{}, error)

// Client represetns the MQTT client
type Client struct {
	cfg *Configuration

	pahoClient  MQTT.Client
	device      *device
	deviceMutex sync.Mutex

	things map[model.NamespacedID]*thing
}

// if the provided opts are nil or the credentials of the device are nil then the local device configuration is loaded

// Connect connects the client
func (client *Client) Connect() error {

	pahoOpts := MQTT.NewClientOptions().
		AddBroker(client.cfg.broker).
		SetClientID(uuid.New().String()).
		SetDefaultPublishHandler(client.handleDefault).
		SetKeepAlive(client.cfg.keepAlive).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(client.clientConnectHandler).
		SetConnectTimeout(client.cfg.connectTimeout)

	if client.cfg.clientUsername != "" {
		pahoOpts.SetCredentialsProvider(func() (username string, password string) {
			return client.cfg.clientUsername, client.cfg.clientPassword
		})
	}

	if err := setupTLSConfiguration(pahoOpts, client.cfg); err != nil {
		return err
	}

	//create and start a client using the created ClientOptions
	client.pahoClient = MQTT.NewClient(pahoOpts)

	if token := client.pahoClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func setupTLSConfiguration(pahoOpts *MQTT.ClientOptions, configuration *Configuration) error {
	u, err := url.Parse(configuration.broker)
	if err != nil {
		return err
	}

	if isConnectionSecure(u.Scheme) {
		if configuration.tlsConfig == nil {
			return errors.New("connection is secure, but no TLS configuration is provided")
		}
		tlsConfig, err := tls.NewConfig(configuration.tlsConfig.RootCA, configuration.tlsConfig.ClientCert, configuration.tlsConfig.ClientKey)
		if err != nil {
			return err
		}
		pahoOpts.SetTLSConfig(tlsConfig)
	}

	return nil
}

func isConnectionSecure(schema string) bool {
	switch schema {
	case "wss", "ssl", "tls", "mqtts", "mqtt+ssl", "tcps":
		return true
	default:
	}
	return false
}

// Disconnect unsubscribes and disconects the client
func (client *Client) Disconnect() {
	if client.pahoClient.IsConnectionOpen() {
		client.unsubscribe(mqttTopicEdgeThingRsp, false)
		if client.device != nil {
			fmt.Println("root device has been created for this client - will unsubscribe")
			client.unsubscribe(fmt.Sprintf(mqttTopicSubscribeCommandsBase, client.device.id), false)
		}
	}
	client.pahoClient.Disconnect(uint(client.cfg.disconnectTimeout.Milliseconds()))
}

// Get returns a things by provided namespace ID
func (client *Client) Get(id model.NamespacedID) model.Thing {
	return client.things[id]
}

// Create creates a thing with the provided namespace ID
func (client *Client) Create(id model.NamespacedID, thing model.Thing) error {
	// only the root thing will be created internally for now
	return nil
}

// Update modifies a thing with the provided namespace ID
func (client *Client) Update(id model.NamespacedID) error {
	return nil
}

// Remove removes a thing with the provided namespace ID
func (client *Client) Remove(id model.NamespacedID) error {
	return nil
}

// SetThingsRegistryChangedHandler sets the registry's handler
func (client *Client) SetThingsRegistryChangedHandler(handler handlers.ThingsRegistryChangedHandler) {
	client.cfg.thingsRegistryChangedHandler = handler
}

// GetThingsRegistryChangedHandler gets the registry's handler
func (client *Client) GetThingsRegistryChangedHandler() handlers.ThingsRegistryChangedHandler {
	return client.cfg.thingsRegistryChangedHandler
}
