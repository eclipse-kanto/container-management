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

package things

import (
	"net"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/golang/mock/gomock"
)

const (
	testMQTTUsername  = "test-username"
	testMQTTPassword  = "test-passoword"
	testMQTTBrokerURL = "localhost:9999"
)

// runs a dummy MQTT broker over TCP to assert the basic auth credentials in the MQTT Connect packet headers.
func TestThingsContainerServiceConnectWithCredentials(t *testing.T) {
	controller := gomock.NewController(t)
	defer func() {
		controller.Finish()
	}()
	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingsContainerManager(controller)
	testThingsMgr, err := newThingsContainerManager(mockContainerManager, mockEventsManager,
		testMQTTBrokerURL,
		0,
		0,
		testMQTTUsername,
		testMQTTPassword,
		testThingsStoragePath,
		testThingsFeaturesDefaultSet,
		0,
		0,
		0,
		0,
		&tlsConfig{},
	)
	if err != nil {
		t.Errorf("unable to create things container manager: %s", err)
	}

	setupThingMock(controller)

	listener, err := net.Listen("tcp4", testMQTTBrokerURL)
	defer listener.Close()
	go func() {
		// wait the tcp listener to initialize
		time.Sleep(1 * time.Second)
		testThingsMgr.thingsClient.Connect()
	}()
	conn, err := listener.Accept()

	if err != nil {
		t.Errorf("Connection accept failure: %s", err)
	}
	controlPacket, err := packets.ReadPacket(conn)
	if err != nil {
		t.Errorf("reading err: %s", err)
	}

	connectPacket := controlPacket.(*packets.ConnectPacket)
	testutil.AssertEqual(t, testMQTTUsername, connectPacket.Username)
	testutil.AssertEqual(t, testMQTTPassword, string(connectPacket.Password))
}

// runs a dummy MQTT broker over TCP to assert the MQTT Connect packet headers.
func TestThingsContainerServiceConnectNoCredentials(t *testing.T) {
	controller := gomock.NewController(t)
	defer func() {
		controller.Finish()
	}()
	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingsContainerManager(controller)
	testThingsMgr, err := newThingsContainerManager(mockContainerManager, mockEventsManager,
		testMQTTBrokerURL,
		0,
		0,
		"",
		"",
		testThingsStoragePath,
		testThingsFeaturesDefaultSet,
		0,
		0,
		0,
		0,
		&tlsConfig{},
	)
	if err != nil {
		t.Errorf("unable to create things container manager: %s", err)
	}

	setupThingMock(controller)

	listener, err := net.Listen("tcp4", testMQTTBrokerURL)
	defer listener.Close()
	go func() {
		// wait the tcp listener to initialize
		time.Sleep(1 * time.Second)
		testThingsMgr.thingsClient.Connect()
	}()
	conn, err := listener.Accept()

	if err != nil {
		t.Errorf("Connection accept failure: %s", err)
	}

	controlPacket, err := packets.ReadPacket(conn)
	if err != nil {
		t.Errorf("reading err: %s", err)
	}

	connectPacket := controlPacket.(*packets.ConnectPacket)
	testutil.AssertEqual(t, "", connectPacket.Username)
	testutil.AssertEqual(t, "", string(connectPacket.Password))
}
