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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client/protocol"
)

var regexTopic = regexp.MustCompile("^([^/]+)/([^/]+)/things/live/messages/([^/]+)$")
var regexHonoOperationRequest = regexp.MustCompile("^command//([^/]+)/req/([^/]*)/([^/]+)$")
var regexFeatureIDOp = regexp.MustCompile("^/features/([^/]+)/inbox/messages/([^/]+)$")
var regexThingOp = regexp.MustCompile("^/inbox/messages/([^/]+)$")

const (
	commandResponseHonoTopicFormat   = "command//%s/res/%s/%d"
	commandResponsePathThingFormat   = "/outbox/messages/%s"
	commandResponsePathFeatureFormat = "/features/%s/outbox/messages/%s"
)

func extractThingID(topic string) (model.NamespacedID, error) {
	matches := strings.Split(topic, "/")
	if len(matches) < 3 {
		return nil, errors.New("provided topic [" + topic + "] does not conform the Ditto specification")
	}
	return NewNamespacedID(matches[0], matches[1]), nil
}

func extractHonoRequestID(honoTopic string) string {
	reqIDInfo := regexHonoOperationRequest.FindStringSubmatch(honoTopic)
	return reqIDInfo[2]
}
func extractMessageFeatureID(path string) string {
	if !regexFeatureIDOp.MatchString(path) {
		return ""
	}
	return regexFeatureIDOp.FindStringSubmatch(path)[1]
}

func extractMessageFeatureOperation(path string) string {
	if !regexFeatureIDOp.MatchString(path) {
		return ""
	}
	return regexFeatureIDOp.FindStringSubmatch(path)[2]
}

func extractMessageThingOperation(path string) string {
	if !regexThingOp.MatchString(path) {
		return ""
	}
	return regexThingOp.FindStringSubmatch(path)[1]
}

func generateResponseThingOperation(topic string, operationName string, status int, result interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	return protocol.Envelope{
		Topic:   topic,
		Headers: generateHeaders(status, headerOpts...),
		Path:    fmt.Sprintf(commandResponsePathThingFormat, operationName),
		Value:   result,
		Status:  status,
	}
}

func generateResponseFeatureOperation(topic string, featureID string, operationName string, status int, result interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	return protocol.Envelope{
		Topic:   topic,
		Headers: generateHeaders(status, headerOpts...),
		Path:    fmt.Sprintf(commandResponsePathFeatureFormat, featureID, operationName),
		Value:   result,
		Status:  status,
	}
}

func generateResponseError(topic, path string, error *ThingError, headerOpts ...HeaderOpt) protocol.Envelope {
	return protocol.Envelope{
		Topic:   topic,
		Headers: generateHeaders(error.Status, headerOpts...),
		Path:    path,
		Value:   error,
		Status:  error.Status,
	}
}

func generateHeaders(status int, headerOpts ...HeaderOpt) protocol.Headers {
	if status != responseStatusOKEmptyPayload {
		// prepend json content type, content type from above is with priority
		headerOpts = append([]HeaderOpt{WithContentType(jsonContent)}, headerOpts...)
	}
	return NewHeaders(headerOpts...)
}

func generateHonoResponseTopic(deviceID string, requestTopic string, status int) string {
	requestID := extractHonoRequestID(requestTopic)
	return fmt.Sprintf(commandResponseHonoTopicFormat, deviceID, requestID, status)
}

func isOneWayCommand(mqttRequesTopic string, requestEnv *protocol.Envelope) bool {
	return (extractHonoRequestID(mqttRequesTopic) == "") || (requestEnv.Headers != nil && !requestEnv.Headers.IsResponseRequired())
}

func extractHonoDeviceIDFromCommandTopic(topic string) model.NamespacedID {
	honoDevIDElements := regexHonoOperationRequest.FindStringSubmatch(topic)
	return NewNamespacedIDFromString(honoDevIDElements[1])
}

func extractHonoDeviceIDFromEventsTopic(topic string) model.NamespacedID {
	honoDevIDElements := regexTopicSubscribeEvents.FindStringSubmatch(topic)
	return NewNamespacedIDFromString(honoDevIDElements[1])
}

func getEnvelope(mqttPayload []byte) (*protocol.Envelope, error) {
	env := jsonEnvelope{}
	if err := json.Unmarshal(mqttPayload, &env); err != nil {
		return nil, err
	}
	return &protocol.Envelope{
		Topic:     env.Topic,
		Headers:   &headers{values: env.Headers},
		Path:      env.Path,
		Value:     env.Value,
		Status:    env.Status,
		Revision:  env.Revision,
		Timestamp: env.Timestamp,
	}, nil
}

func validateCommandPayload(env *protocol.Envelope) error {
	if env.Topic == "" {
		return errors.New("message topic is missing")
	}
	if env.Path == "" {
		return errors.New("message path is missing")
	}
	return nil
}

func validateHeaders(env *protocol.Envelope) error {
	if env.Headers.GetCorrelationID() == "" {
		return errors.New("correlation id header is missing")
	}
	contentType := env.Headers.GetContentType()
	if env.Value != nil && contentType != jsonContent {
		if contentType == "" {
			return errors.New("missing content type header")
		}
		return errors.New("unsupported content type header " + contentType)
	} else if env.Value == nil && contentType != "" {
		return errors.New("unexpected content type header " + contentType)
	}
	return nil
}
