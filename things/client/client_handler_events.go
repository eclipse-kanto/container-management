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
	specialCharactersSubString           string = "/([^/]+)"
	specialCharactersDeepPathSubString   string = "((/[^/]+)+)"
	mqttTopicSubscribeEventsBase         string = "command//%s/req//"
	matchPathThing                              = "/"
	matchPathThingDefinition                    = "/definition"
	matchPathThingFeatures                      = "/features"
	matchPathThingAttributes                    = "/attributes"
	matchPathThingAttribute                     = matchPathThingAttributes + specialCharactersDeepPathSubString
	matchPathThingFeature                       = matchPathThingFeatures + specialCharactersSubString
	matchPathThingFeatureDefinition             = matchPathThingFeature + "/definition"
	matchPathThingFeatureProperties             = matchPathThingFeature + "/properties"
	matchPathThingFeatureProperty               = matchPathThingFeatureProperties + specialCharactersDeepPathSubString
	mqttTopicSubscribeEventsCreatedBase         = mqttTopicSubscribeEventsBase + string(protocol.ActionCreated)
	mqttTopicSubscribeEventsModifiedBase        = mqttTopicSubscribeEventsBase + string(protocol.ActionModified)
	mqttTopicSubscribeEventsDeletedBase         = mqttTopicSubscribeEventsBase + string(protocol.ActionDeleted)
)

var (
	regexPathThing                  = regexp.MustCompile("^" + matchPathThing + "$")
	regexPathThingDefinition        = regexp.MustCompile("^" + matchPathThingDefinition + "$")
	regexPathThingFeatures          = regexp.MustCompile("^" + matchPathThingFeatures + "$")
	regexPathThingAttributes        = regexp.MustCompile("^" + matchPathThingAttributes + "$")
	regexPathThingAttribute         = regexp.MustCompile("^" + matchPathThingAttribute + "$")
	regexPathThingFeature           = regexp.MustCompile("^" + matchPathThingFeature + "$")
	regexPathThingFeatureDefinition = regexp.MustCompile("^" + matchPathThingFeatureDefinition + "$")
	regexPathThingFeatureProperties = regexp.MustCompile("^" + matchPathThingFeatureProperties + "$")
	regexPathThingFeatureProperty   = regexp.MustCompile("^" + matchPathThingFeatureProperty + "$")
)

func (client *Client) handleEventCreated(mqttClient MQTT.Client, message MQTT.Message) {
	mqttTopic := message.Topic()
	fmt.Printf("Event Created TOPIC: %s\n", mqttTopic)
	fmt.Printf("Event Created MSG: %s\n", message.Payload())

	if !regexTopicSubscribeEvents.MatchString(mqttTopic) {
		fmt.Println("received an unexpected message - will be routed to the default MQTT messages handler")
		client.handleDefault(mqttClient, message)
		return
	}

	honoDevID := extractHonoDeviceIDFromEventsTopic(mqttTopic)

	event, err := getEnvelope(message.Payload())
	if err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("error parsing Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	if err = validateCommandPayload(event); err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	thingID, err := extractThingID(event.Topic)
	if err != nil {
		fmt.Printf("invalid Ditto message topic received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message topic received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	path := event.Path
	thing := client.things[thingID]
	if regexPathThing.MatchString(path) && thing != nil {
		fmt.Printf("thing with id %s already locally exists - discarding event created \n", thingID)
		err := NewMessagesParameterInvalidError("thing with id %s already locally exists - discarding event created", thingID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	for regex, createdEventHandlerFunc := range createdEventHandlers {
		if regex.MatchString(path) {
			createdEventHandlerFunc(path, client, thing, event.Value)
			fmt.Printf("handled created event for path : %s\n", path)
			return
		}
	}
	fmt.Printf("invalid created event path for event modified: %s", path)
}

func (client *Client) handleEventModified(mqttClient MQTT.Client, message MQTT.Message) {
	mqttTopic := message.Topic()
	fmt.Printf("Event Modified TOPIC: %s\n", message.Topic())
	fmt.Printf("Event Modified MSG: %s\n", message.Payload())

	if !regexTopicSubscribeEvents.MatchString(mqttTopic) {
		fmt.Println("received an unexpected message - will be routed to the default MQTT messages handler")
		client.handleDefault(mqttClient, message)
		return
	}

	honoDevID := extractHonoDeviceIDFromEventsTopic(mqttTopic)

	event, err := getEnvelope(message.Payload())
	if err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("error parsing Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	if err = validateCommandPayload(event); err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	thingID, err := extractThingID(event.Topic)
	if err != nil {
		fmt.Printf("invalid Ditto message topic received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message topic received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	path := event.Path
	thing := client.things[thingID]

	if thing == nil {
		fmt.Println("no thing locally exists")
		err := NewMessagesSubjectNotFound("no thing with id [%s] locally exists - discarding event modified", thingID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	for regex, modifiedEventHandlerFunc := range modifiedEventHandlers {
		if regex.MatchString(path) {
			modifiedEventHandlerFunc(path, client, thing, event.Value)
			fmt.Printf("handled modified event for path : %s\n", path)
			return
		}
	}
	fmt.Printf("invalid modified event path for event modified: %s", path)
}

func (client *Client) handleEventDeleted(mqttClient MQTT.Client, message MQTT.Message) {
	mqttTopic := message.Topic()
	fmt.Printf("Event Deleted TOPIC: %s\n", message.Topic())
	fmt.Printf("Event Deleted MSG: %s\n", message.Payload())

	if !regexTopicSubscribeEvents.MatchString(mqttTopic) {
		fmt.Println("received an unexpected message - will be routed to the default MQTT messages handler")
		client.handleDefault(mqttClient, message)
		return
	}

	honoDevID := extractHonoDeviceIDFromEventsTopic(mqttTopic)

	event, err := getEnvelope(message.Payload())
	if err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("error parsing Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	if err = validateCommandPayload(event); err != nil {
		fmt.Printf("invalid message payload received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	thingID, err := extractThingID(event.Topic)
	if err != nil {
		fmt.Printf("invalid Ditto message topic received: %v\n", err)
		err := NewMessagesParameterInvalidError("invalid Ditto message topic received for device id [%s]", honoDevID)
		client.publishErrorResponse(err, nil, honoDevID.String(), mqttTopic)
		return
	}

	path := event.Path
	thing := client.things[thingID]

	if thing == nil {
		fmt.Println("no thing locally exists")
		err := NewMessagesSubjectNotFound("no thing with id [%s] locally exists - discarding event deleted", thingID)
		client.publishErrorResponse(err, event, honoDevID.String(), mqttTopic)
		return
	}

	for regex, deletedEventHandlerFunc := range deletedEventHandlers {
		if regex.MatchString(path) {
			deletedEventHandlerFunc(path, client, thing)
			fmt.Printf("handled deleted event for path : %s\n", path)
			return
		}
	}
	fmt.Printf("invalid deleted event path for event created: %s", path)
}

var (
	createdEventHandlers = map[*regexp.Regexp]func(path string, client *Client, thing *thing, value interface{}){
		regexPathThing: func(path string, client *Client, thing *thing, value interface{}) { //create the entire thing
			th, _ := marshalToThing(value)
			client.createThing(th)
			//send event
			if rh := client.GetThingsRegistryChangedHandler(); rh != nil {
				rh(handlers.Added, th)
			}
		},
		regexPathThingDefinition: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's definition only
			defID, _ := marshalToDefinitionID(value)
			client.updateThingDefinition(thing.id, defID)
			//send event
			if dh := thing.GetDefinitionChangedHandler(); dh != nil {
				dh(handlers.Created, defID)
			}
		},
		regexPathThingAttributes: func(path string, client *Client, thing *thing, value interface{}) { //create all the thing's attributes
			attributes, _ := marshalToAttributes(value)
			client.updateThingAttributes(thing.id, attributes)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				for id, val := range attributes {
					ah(handlers.Created, id, val)
				}
			}
		},
		regexPathThingAttribute: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's attribute
			attrID := regexPathThingAttribute.FindStringSubmatch(path)[1]
			client.updateThingAttribute(thing.id, attrID, value)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				ah(handlers.Created, attrID, value)
			}
		},
		regexPathThingFeatures: func(path string, client *Client, thing *thing, value interface{}) { //create all the thing's features
			features, _ := marshalToFeatures(value)
			client.updateThingFeatures(thing.id, features)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				for id, val := range features {
					fh(handlers.Created, id, val)
				}
			}
		},
		regexPathThingFeature: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's feature
			featureID := regexPathThingFeature.FindStringSubmatch(path)[1]
			feature, _ := marshalToFeature(featureID, value)
			client.updateThingFeature(thing.id, featureID, feature)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				fh(handlers.Created, featureID, feature)
			}
		},
		regexPathThingFeatureDefinition: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's feature definition
			featureID := regexPathThingFeatureDefinition.FindStringSubmatch(path)[1]
			definition, _ := marshalToFeatureDefinition(value)
			client.updateThingFeatureDefinition(thing.id, featureID, definition)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				dh := (feature.(handlers.FeatureHandler)).GetDefinitionChangedHandler()
				if dh != nil {
					dh(handlers.Created, definition)
				}
			}
		},
		regexPathThingFeatureProperties: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's feature properties
			featureID := regexPathThingFeatureProperties.FindStringSubmatch(path)[1]
			props, _ := marshalToFeatureProperties(value)
			client.updateThingFeatureProperties(thing.id, featureID, props)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					for id, val := range props {
						ph(handlers.Created, id, val)
					}
				}
			}
		},
		regexPathThingFeatureProperty: func(path string, client *Client, thing *thing, value interface{}) { //create the thing's feature property
			featureID := regexPathThingFeatureProperty.FindStringSubmatch(path)[1]
			propertyID := regexPathThingFeatureProperty.FindStringSubmatch(path)[2]
			client.updateThingFeatureProperty(thing.id, featureID, propertyID, value)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					ph(handlers.Created, propertyID, value)
				}
			}
		},
	}
	modifiedEventHandlers = map[*regexp.Regexp]func(path string, client *Client, thing *thing, value interface{}){
		regexPathThing: func(path string, client *Client, thing *thing, value interface{}) { //update the entire thing
			th, _ := marshalToThing(value)
			client.updateThing(th)
			//send event
			if rh := client.GetThingsRegistryChangedHandler(); rh != nil {
				rh(handlers.Updated, th)
			}
		},
		regexPathThingDefinition: func(path string, client *Client, thing *thing, value interface{}) { //update the thing's definition only
			defID, _ := marshalToDefinitionID(value)
			client.updateThingDefinition(thing.id, defID)
			//send event
			if dh := thing.GetDefinitionChangedHandler(); dh != nil {
				dh(handlers.Modified, defID)
			}
		},
		regexPathThingAttributes: func(path string, client *Client, thing *thing, value interface{}) { //update all the thing's attributes
			attributes, _ := marshalToAttributes(value)
			client.updateThingAttributes(thing.id, attributes)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				for id, val := range attributes {
					ah(handlers.Modified, id, val)
				}
			}
		},
		regexPathThingAttribute: func(path string, client *Client, thing *thing, value interface{}) { //update the thing's attribute
			attrID := regexPathThingAttribute.FindStringSubmatch(path)[1]
			client.updateThingAttribute(thing.id, attrID, value)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				ah(handlers.Modified, attrID, value)
			}
		},
		regexPathThingFeatures: func(path string, client *Client, thing *thing, value interface{}) { //update all the thing's features
			features, _ := marshalToFeatures(value)
			client.updateThingFeatures(thing.id, features)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				for id, val := range features {
					fh(handlers.Modified, id, val)
				}
			}
		},
		regexPathThingFeature: func(path string, client *Client, thing *thing, value interface{}) { //update the thing's feature
			featureID := regexPathThingFeature.FindStringSubmatch(path)[1]
			feature, _ := marshalToFeature(featureID, value)
			client.updateThingFeature(thing.id, featureID, feature)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				fh(handlers.Modified, featureID, feature)
			}
		},
		regexPathThingFeatureDefinition: func(path string, client *Client, thing *thing, value interface{}) { // update the thing's feature definition
			featureID := regexPathThingFeatureDefinition.FindStringSubmatch(path)[1]
			definition, _ := marshalToFeatureDefinition(value)
			client.updateThingFeatureDefinition(thing.id, featureID, definition)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				dh := (feature.(handlers.FeatureHandler)).GetDefinitionChangedHandler()
				if dh != nil {
					dh(handlers.Modified, definition)
				}
			}
		},
		regexPathThingFeatureProperties: func(path string, client *Client, thing *thing, value interface{}) { // update the thing's feature properties
			featureID := regexPathThingFeatureProperties.FindStringSubmatch(path)[1]
			props, _ := marshalToFeatureProperties(value)
			client.updateThingFeatureProperties(thing.id, featureID, props)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					for id, val := range props {
						ph(handlers.Modified, id, val)
					}
				}
			}
		},
		regexPathThingFeatureProperty: func(path string, client *Client, thing *thing, value interface{}) { // update the thing's feature property
			featureID := regexPathThingFeatureProperty.FindStringSubmatch(path)[1]
			propertyID := regexPathThingFeatureProperty.FindStringSubmatch(path)[2]
			client.updateThingFeatureProperty(thing.id, featureID, propertyID, value)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					ph(handlers.Modified, propertyID, value)
				}
			}
		},
	}
	deletedEventHandlers = map[*regexp.Regexp]func(path string, client *Client, thing *thing){
		regexPathThing: func(path string, client *Client, thing *thing) { //remove the entire thing
			th := client.removeThing(thing.id)
			//send event
			if rh := client.GetThingsRegistryChangedHandler(); rh != nil {
				rh(handlers.Removed, th)
			}
		},
		regexPathThingDefinition: func(path string, client *Client, thing *thing) { //remove the thing's definition only
			defID := client.removeThingDefinition(thing.id)
			//send event
			if dh := thing.GetDefinitionChangedHandler(); dh != nil {
				dh(handlers.Deleted, defID)
			}
		},
		regexPathThingAttributes: func(path string, client *Client, thing *thing) { //remove all the thing's attributes
			attributes := client.removeThingAttributes(thing.id)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				for id, val := range attributes {
					ah(handlers.Deleted, id, val)
				}
			}
		},
		regexPathThingAttribute: func(path string, client *Client, thing *thing) { //remove the thing's attribute
			attrID := regexPathThingAttribute.FindStringSubmatch(path)[1]
			attr := client.removeThingAttribute(thing.id, attrID)
			//send event
			if ah := thing.GetAttributeChangedHandler(); ah != nil {
				ah(handlers.Deleted, attrID, attr)
			}
		},
		regexPathThingFeatures: func(path string, client *Client, thing *thing) { //remove all the thing's features
			features := client.removeThingFeatures(thing.id)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				for id, val := range features {
					fh(handlers.Deleted, id, val)
				}
			}
		},
		regexPathThingFeature: func(path string, client *Client, thing *thing) { //remove the thing's feature
			featureID := regexPathThingFeature.FindStringSubmatch(path)[1]
			feature := client.removeThingFeature(thing.id, featureID)
			//send event
			if fh := thing.GetFeatureChangedHandler(); fh != nil {
				fh(handlers.Deleted, featureID, feature)
			}
		},
		regexPathThingFeatureDefinition: func(path string, client *Client, thing *thing) { // remove the thing's feature definition
			featureID := regexPathThingFeatureDefinition.FindStringSubmatch(path)[1]
			definition := client.removeThingFeatureDefinition(thing.id, featureID)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				dh := (feature.(handlers.FeatureHandler)).GetDefinitionChangedHandler()
				if dh != nil {
					dh(handlers.Deleted, definition)
				}
			}
		},
		regexPathThingFeatureProperties: func(path string, client *Client, thing *thing) { // remove the thing's feature properties
			featureID := regexPathThingFeatureProperties.FindStringSubmatch(path)[1]
			props := client.removeThingFeatureProperties(thing.id, featureID)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					for id, val := range props {
						ph(handlers.Deleted, id, val)
					}
				}
			}
		},
		regexPathThingFeatureProperty: func(path string, client *Client, thing *thing) { // remove the thing's feature property
			featureID := regexPathThingFeatureProperty.FindStringSubmatch(path)[1]
			propertyID := regexPathThingFeatureProperty.FindStringSubmatch(path)[2]
			propVal := client.removeThingFeatureProperty(thing.id, featureID, propertyID)
			//send event
			feature := thing.GetFeature(featureID)
			if feature != nil {
				ph := (feature.(handlers.FeatureHandler)).GetPropertyChangedHandler()
				if ph != nil {
					ph(handlers.Deleted, propertyID, propVal)
				}
			}
		},
	}
)
