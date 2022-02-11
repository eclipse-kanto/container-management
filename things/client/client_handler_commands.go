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
	"fmt"
	"regexp"

	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/client/protocol"

	//import the Paho Go MQTT library
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttTopicSubscribeCommandsBase = "command//%s/req/#"

	responseStatusOKWithPayload  = 200
	responseStatusOKEmptyPayload = 204
	jsonContent                  = "application/json"
)

var (
	regexTopicSubscribeEvents = regexp.MustCompile("^command//([^/]+)/req//(" + string(protocol.ActionCreated) + "|" + string(protocol.ActionModified) + "|" + string(protocol.ActionDeleted) + ")$")
	regexTopicSubscribeErrors = regexp.MustCompile("^command//([^/]+)/req//" + honoCommandTopicSuffixErrorsResponse + "$")
)

func (client *Client) handleDefault(mqttClient MQTT.Client, message MQTT.Message) {
	fmt.Println("received a message matching no known-to-be-processed filters:")
	fmt.Printf("Default TOPIC: %s\n", message.Topic())
	fmt.Printf("Default MSG: %s\n", message.Payload())
}

func (client *Client) handleCommand(mqttClient MQTT.Client, message MQTT.Message) {
	mqttTopic := message.Topic()
	fmt.Printf("Command TOPIC: %s\n", mqttTopic)
	fmt.Printf("Command MSG: %s\n", message.Payload())

	if regexTopicSubscribeEvents.MatchString(mqttTopic) || regexTopicSubscribeErrors.MatchString(mqttTopic) {
		fmt.Println("received a routed message")
		return
	}
	if !regexHonoOperationRequest.MatchString(mqttTopic) {
		fmt.Println("received an unexpected message - will be routed to the default MQTT messages handler")
		client.handleDefault(mqttClient, message)
		return
	}
	honoDevID := extractHonoDeviceIDFromCommandTopic(mqttTopic)

	requestEnv, err := getEnvelope(message.Payload())
	if err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("error parsing Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, requestEnv, honoDevID.String(), mqttTopic)
		return
	}
	if err = validateCommandPayload(requestEnv); err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	featureID := extractMessageFeatureID(requestEnv.Path)
	thingID, err := extractThingID(requestEnv.Topic)
	if err != nil {
		fmt.Printf("invalid Ditto message topic received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message topic received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}
	if err = validateHeaders(requestEnv); err != nil {
		fmt.Printf("invalid message headers received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message received for device id [%s]: %v", honoDevID, err)
		client.publishErrorResponse(err, requestEnv, honoDevID.String(), mqttTopic)
		return
	}

	thing := client.things[thingID]

	if thing == nil {
		fmt.Println("no thing locally exists - will return error ")
		err := NewMessagesSubjectNotFound("thing with ID = %s is not found", honoDevID)
		client.publishErrorResponse(err, requestEnv, honoDevID.String(), mqttTopic)
		return
	}
	if featureID != "" && thing.GetFeature(featureID) == nil {
		fmt.Printf("feature [%s] not found \n", featureID)
		err := NewMessagesParameterInvalidError("feature [%s] not found ", featureID)
		client.publishErrorResponse(err, requestEnv, honoDevID.String(), mqttTopic)
		return
	}
	if featureID == "" {
		client.processThingOperation(requestEnv, thing, mqttTopic)
	} else {
		client.processFeatureOperation(requestEnv, thing, featureID, mqttTopic)
	}
}

func processOperationResult(result interface{}, resultError error) (int, interface{}) {
	if resultError != nil {
		thErr, ok := resultError.(*ThingError)
		if ok {
			return thErr.Status, thErr
		}
		res := NewMessagesInternalError(resultError.Error())
		return res.Status, res
	}
	if result == nil {
		return responseStatusOKEmptyPayload, result
	}
	return responseStatusOKWithPayload, result
}

func (client *Client) publishErrorResponse(err *ThingError, requestEnv *protocol.Envelope, thingID, mqttReqTopic string) {
	if isOneWayCommand(mqttReqTopic, requestEnv) {
		fmt.Println("the command is one way and resulted in an error - no error will be posted")
		return
	}
	var respEnv protocol.Envelope
	if requestEnv == nil {
		respEnv = generateResponseError(fmt.Sprintf(protocol.TopicErrorsFormat, protocol.Twin), pathThing, err)
	} else {
		respEnv = generateResponseError(requestEnv.Topic, requestEnv.Path, err, WithCorrelationID(requestEnv.Headers.GetCorrelationID()))
	}
	if err := client.publish(generateHonoResponseTopic(thingID, mqttReqTopic, err.Status), respEnv, 1, false); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
	}
}

func (client *Client) processThingOperation(requestEnv *protocol.Envelope, thing *thing, mqttTopic string) {
	opName := extractMessageThingOperation(requestEnv.Path)
	if opName == "" {
		fmt.Printf("missing operation name for thing command")
		err := NewMessagesParameterInvalidError("missing operation name for device id [%s]", thing.id.String())
		client.publishErrorResponse(err, requestEnv, thing.id.String(), mqttTopic)
		return
	}
	if tOperationHandler := thing.GetOperationsHandler(); tOperationHandler != nil {
		operationResult, operationError := tOperationHandler(opName, requestEnv.Value)
		if !isOneWayCommand(mqttTopic, requestEnv) {
			status, res := processOperationResult(operationResult, operationError)
			respEnv := generateResponseThingOperation(requestEnv.Topic, opName, status, res, WithCorrelationID(requestEnv.Headers.GetCorrelationID()))
			if err := client.publish(generateHonoResponseTopic(thing.id.String(), mqttTopic, status), respEnv, 1, false); err != nil {
				fmt.Printf("[ERROR] %v\n", err)
			}
		} else {
			fmt.Printf("the provided command %v is one way - will not post the result back {res:%v, err:%v}\n", requestEnv, operationResult, operationError)
		}
	} else {
		err := NewMessagesSubjectNotFound("the target thing ID = %s does not have e registered operations handler - thus, the command will not be processed", thing.id.String())
		fmt.Println(err.Error())
		client.publishErrorResponse(err, requestEnv, thing.id.String(), mqttTopic)
	}
}

func (client *Client) processFeatureOperation(requestEnv *protocol.Envelope, thing *thing, featureID, mqttTopic string) {
	feature := thing.GetFeature(featureID)
	if feature != nil {
		fOperationsHandler := feature.(handlers.FeatureHandler).GetOperationsHandler()
		if fOperationsHandler != nil {
			opName := extractMessageFeatureOperation(requestEnv.Path)
			if opName == "" {
				fmt.Printf("missing operation name for feature [%s] command", featureID)
				err := NewMessagesParameterInvalidError("missing operation name for feature [%s] for device id [%s]", featureID, thing.id.String())
				client.publishErrorResponse(err, requestEnv, thing.id.String(), mqttTopic)
				return
			}
			operationResult, operationError := fOperationsHandler(opName, requestEnv.Value)
			if !isOneWayCommand(mqttTopic, requestEnv) {
				status, res := processOperationResult(operationResult, operationError)
				respEnv := generateResponseFeatureOperation(requestEnv.Topic, featureID, opName, status, res, WithCorrelationID(requestEnv.Headers.GetCorrelationID()))
				if err := client.publish(generateHonoResponseTopic(thing.id.String(), mqttTopic, status), respEnv, 1, false); err != nil {
					fmt.Printf("[ERROR] %v\n", err)
				}
			} else {
				fmt.Printf("the provided command %v is one way - will not post the result back {res:%v, err:%v}\n", requestEnv, operationResult, operationError)
			}
		} else {
			err := NewMessagesSubjectNotFound("the target feature %s does not have e registered operations handler - thus, the command will not be processed", featureID)
			fmt.Println(err.Error())
			client.publishErrorResponse(err, requestEnv, thing.id.String(), mqttTopic)
		}
	} else {
		err := NewMessagesSubjectNotFound("no such feature [id = %s] exists in the local thing representation", featureID)
		fmt.Println(err.Error())
		client.publishErrorResponse(err, requestEnv, thing.id.String(), mqttTopic)
	}
}
