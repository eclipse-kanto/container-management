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

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client/protocol"

	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttTopicEdgeThingReq = "edge/thing/request"
	mqttTopicEdgeThingRsp = "edge/thing/response"
)

func (client *Client) clientConnectHandler(pahoClient MQTT.Client) {
	client.deviceMutex.Lock()
	defer client.deviceMutex.Unlock()

	client.device = nil //  a new subscription and processing of the edge thing local info must be done

	var err error
	if err = client.subscribe(mqttTopicEdgeThingRsp, client.handleThingResponse, 1); err == nil {
		fmt.Println("subscribed for ", mqttTopicEdgeThingRsp)
		if err = client.publish(mqttTopicEdgeThingReq, protocol.Envelope{}, 1, false); err == nil {
			return
		}
	}

	client.notifyClientInitialized(err)
}

func (client *Client) handleThingResponse(mqttClient MQTT.Client, message MQTT.Message) {
	client.deviceMutex.Lock()
	defer client.deviceMutex.Unlock()

	fmt.Printf("Topic: %s\n", message.Topic())
	fmt.Printf("Message: %s\n", message.Payload())

	lc := &clientLocalConfig{}
	if err := json.Unmarshal(message.Payload(), lc); err != nil {
		client.notifyClientInitialized(err)
		return
	}
	subscribe := true
	if client.device != nil {
		if lc.tenantID == client.device.tenantID && lc.id.String() == client.device.viaGateway.String() { // already subscribed for edge containers of this device
			subscribe = false
			if lc.policyID == client.things[client.device.id].policyID { // no changes
				return
			}
		} else { // unsubscribe from the old thing
			topic := fmt.Sprintf(mqttTopicSubscribeCommandsBase, client.device.id)
			if err := client.unsubscribe(topic, true); err != nil {
				fmt.Printf("error unsubscribing from topic = %s \n%v\n", topic, err)
			}
		}
	}

	client.device = &device{}
	client.device.viaGateway = lc.id
	client.device.tenantID = lc.tenantID
	client.cfg.WithGatewayDeviceID(lc.id.String()).WithDeviceTenantID(lc.tenantID)
	client.device.id = NewNamespacedIDFromString(NewNamespacedID(client.device.viaGateway.String(), client.cfg.deviceName).String())

	client.device.connection = &deviceConnection{client: client}
	rootThing := &thing{
		id:         client.device.id,
		hubDevice:  client.device,
		policyID:   lc.policyID,
		attributes: make(map[string]interface{}),
		features:   make(map[string]model.Feature),
	}
	client.things[rootThing.id] = rootThing

	client.notifyClientInitialized(nil)

	if subscribe {
		client.pahoClient.AddRoute(fmt.Sprintf(mqttTopicSubscribeEventsCreatedBase, rootThing.id), client.handleEventCreated)
		client.pahoClient.AddRoute(fmt.Sprintf(mqttTopicSubscribeEventsModifiedBase, rootThing.id), client.handleEventModified)
		client.pahoClient.AddRoute(fmt.Sprintf(mqttTopicSubscribeEventsDeletedBase, rootThing.id), client.handleEventDeleted)
		client.pahoClient.AddRoute(fmt.Sprintf(mqttTopicSubscribeErrorsBase, rootThing.id), client.handleTwinErrors)
		if err := client.subscribe(fmt.Sprintf(mqttTopicSubscribeCommandsBase, rootThing.id), client.handleCommand, 1); err != nil {
			fmt.Printf("error subscribing for topic = %s \n%v\n", fmt.Sprintf(mqttTopicSubscribeCommandsBase, rootThing.id), err)
		}
	}
}

func (client *Client) notifyClientInitialized(err error) {
	notifyChan := make(chan error, 1)
	var notifyOnce sync.Once
	go func() {
		notifyOnce.Do(func() {
			if client.cfg.initHook != nil {
				client.cfg.initHook(client, client.cfg, err)
			}
		})
		notifyChan <- nil
	}()

	select {
	case <-notifyChan:
		fmt.Println("notified for client initialization successfully")
	case <-time.After(client.cfg.connectTimeout):
		fmt.Printf("[ERROR] %v\n", errors.New("timed out waiting for initialization notification to be handled"))
	}
}

func (client *Client) createThing(thing *thing) {
	thing.hubDevice = client.device

	// add to cache - some day move it to Registry.Create
	client.things[thing.id] = thing
}
func (client *Client) updateThing(thing *thing) {
	client.things[thing.GetID()] = thing
}

func (client *Client) removeThing(thingID model.NamespacedID) model.Thing {
	res := client.things[thingID]
	delete(client.things, thingID)
	return res
}

func (client *Client) updateThingDefinition(thingID model.NamespacedID, defID model.DefinitionID) {
	client.things[thingID].definitionID = defID
}
func (client *Client) removeThingDefinition(thingID model.NamespacedID) model.DefinitionID {
	res := client.things[thingID].definitionID
	client.things[thingID].definitionID = nil
	return res
}
func (client *Client) updateThingAttributes(thingID model.NamespacedID, attributes map[string]interface{}) {
	client.things[thingID].setAttributesSafe(attributes)
}
func (client *Client) removeThingAttributes(thingID model.NamespacedID) map[string]interface{} {
	return client.things[thingID].removeAttributesSafe()
}

func (client *Client) updateThingAttribute(thingID model.NamespacedID, attributeID string, attributeValue interface{}) {
	client.things[thingID].setAttributeSafe(attributeID, attributeValue)
}
func (client *Client) removeThingAttribute(thingID model.NamespacedID, attributeID string) interface{} {
	return client.things[thingID].removeAttributeSafe(attributeID)
}

func (client *Client) updateThingFeatures(thingID model.NamespacedID, features map[string]model.Feature) {
	client.things[thingID].setFeaturesSafe(features)
}
func (client *Client) removeThingFeatures(thingID model.NamespacedID) map[string]model.Feature {
	return client.things[thingID].removeFeaturesSafe()
}
func (client *Client) updateThingFeature(thingID model.NamespacedID, featureID string, featureVal model.Feature) {
	client.things[thingID].setFeatureSafe(featureID, featureVal)
}
func (client *Client) removeThingFeature(thingID model.NamespacedID, featureID string) model.Feature {
	return client.things[thingID].removeFeatureSafe(featureID)
}

func (client *Client) updateThingFeatureDefinition(thingID model.NamespacedID, featureID string, definition []model.DefinitionID) {
	if feature := client.things[thingID].GetFeature(featureID); feature != nil {
		feature.SetDefinition(definition)
	}
}

func (client *Client) removeThingFeatureDefinition(thingID model.NamespacedID, featureID string) []model.DefinitionID {
	if feature := client.things[thingID].GetFeature(featureID); feature != nil {
		defer feature.RemoveDefinition()
		return feature.GetDefinition()
	}
	return nil
}

func (client *Client) updateThingFeatureProperties(thingID model.NamespacedID, featureID string, props map[string]interface{}) {
	client.things[thingID].GetFeature(featureID).SetProperties(props)
}
func (client *Client) removeThingFeatureProperties(thingID model.NamespacedID, featureID string) map[string]interface{} {
	if feature := client.things[thingID].GetFeature(featureID); feature != nil {
		defer feature.RemoveProperties()
		return feature.GetProperties()
	}
	return nil
}

func (client *Client) updateThingFeatureProperty(thingID model.NamespacedID, featureID string, propID string, propValue interface{}) {
	if feature := client.things[thingID].GetFeature(featureID); feature != nil {
		feature.SetProperty(propID, propValue)
	}
}
func (client *Client) removeThingFeatureProperty(thingID model.NamespacedID, featureID string, propID string) interface{} {
	if feature := client.things[thingID].GetFeature(featureID); feature != nil {
		defer feature.RemoveProperty(propID)
		return feature.GetProperty(propID)
	}
	return nil
}

func (client *Client) subscribe(topic string, msgHandler MQTT.MessageHandler, qos byte) error {
	token := client.pahoClient.Subscribe(topic, qos, msgHandler)
	if !token.WaitTimeout(client.cfg.subscribeTimeout) {
		return errors.New("subscribe timeout")
	}
	return token.Error()
}

func (client *Client) unsubscribe(topic string, wait bool) error {
	token := client.pahoClient.Unsubscribe(topic)
	if !wait {
		return nil
	}
	if !token.WaitTimeout(client.cfg.unsubscribeTimeout) {
		return errors.New("unsubscribe timeout")
	}
	return token.Error()
}

func (client *Client) publish(topic string, envelope protocol.Envelope, qos byte, reatained bool) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	token := client.pahoClient.Publish(topic, qos, reatained, payload)
	if !token.WaitTimeout(client.cfg.acknowledgeTimeout) {
		return errors.New("publish timeout")
	}
	return token.Error()
}
