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
	"fmt"

	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client/protocol"
)

var defaultHeaders = []HeaderOpt{WithResponseRequired(false)}

const topicFormat = "%s/%s/%s/%s/%s/%s"

const (
	pathThing                            = "/"
	pathThingDefinition                  = "/definition"
	pathThingFeatures                    = "/features"
	pathThingAttributes                  = "/attributes"
	pathThingAttributeFormat             = pathThingAttributes + "/%s"
	pathThingFeatureFormat               = pathThingFeatures + "/%s"
	pathThingFeatureDefinitionFormat     = pathThingFeatureFormat + "/definition"
	pathThingFeaturePropertiesFormat     = pathThingFeatureFormat + "/properties"
	pathThingFeaturePropertyFormat       = pathThingFeaturePropertiesFormat + "/%s"
	pathThingMessagesOutboxFormat        = "/outbox/messages/%s"
	pathThingFeatureMessagesOutboxFormat = "/features/%s" + pathThingMessagesOutboxFormat
)

func generateCommandCreateThing(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionCreate),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThing,
		Value:   thing,
	}
	return env
}

func generateCommandDeleteThing(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThing,
	}
	return env
}

func generateCommandModifyThing(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThing,
		Value:   thing,
	}
	return env
}

func generateCommandModifyThingDefinition(thing model.Thing, definition model.DefinitionID, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingDefinition,
		Value:   definition.String(),
	}
	return env
}

func generateCommandDeleteThingDefinition(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingDefinition,
	}
	return env
}

func generateCommandModifyThingAttributes(thing model.Thing, attributes map[string]interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingAttributes,
		Value:   attributes,
	}
	return env
}

func generateCommandDeleteThingAttributes(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingAttributes,
	}
	return env
}

func generateCommandModifyThingAttributeSingle(thing model.Thing, attributeID string, attributeValue interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingAttributeFormat, attributeID),
		Value:   attributeValue,
	}
	return env
}

func generateCommandDeleteThingAttributeSingle(thing model.Thing, attributeID string, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingAttributeFormat, attributeID),
	}
	return env
}

func generateCommandModifyThingFeatures(thing model.Thing, features map[string]model.Feature, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingFeatures,
		Value:   features,
	}
	return env
}

func generateCommandDeleteThingFeatures(thing model.Thing, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    pathThingFeatures,
	}
	return env
}

func generateCommandModifyThingFeatureSingle(thing model.Thing, featureID string, feature model.Feature, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeatureFormat, featureID),
		Value:   feature,
	}
	return env
}

func generateCommandDeleteThingFeatureSingle(thing model.Thing, featureID string, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeatureFormat, featureID),
	}
	return env
}

func generateCommandModifyThingFeatureDefinitionSingle(thing model.Thing, featureID string, definition []model.DefinitionID, headerOpts ...HeaderOpt) protocol.Envelope {
	defsAsString := []string{}
	for _, def := range definition {
		defsAsString = append(defsAsString, def.String())
	}
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeatureDefinitionFormat, featureID),
		Value:   defsAsString,
	}
	return env
}

func generateCommandDeleteThingFeatureDefinitionSingle(thing model.Thing, featureID string, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeatureDefinitionFormat, featureID),
	}
	return env
}

func generateCommandModifyThingFeatureProperties(thing model.Thing, featureID string, props map[string]interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeaturePropertiesFormat, featureID),
		Value:   props,
	}
	return env
}

func generateCommandModifyThingFeaturePropertySingle(thing model.Thing, featureID string, propID string, propValue interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionModify),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeaturePropertyFormat, featureID, propID),
		Value:   propValue,
	}
	return env
}

func generateCommandDeleteThingFeatureProperties(thing model.Thing, featureID string, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeaturePropertiesFormat, featureID),
	}
	return env
}

func generateCommandDeleteThingFeaturePropertySingle(thing model.Thing, featureID string, propID string, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Twin, protocol.Commands, protocol.ActionDelete),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeaturePropertyFormat, featureID, propID),
	}
	return env
}

func generateOutboxMessageFromThing(thing model.Thing, action string, value interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Live, protocol.Messages, protocol.TopicAction(action)),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingMessagesOutboxFormat, action),
		Value:   value,
	}
	return env
}
func generateOutboxMessageFromThingFeature(thing model.Thing, featureID string, action string, value interface{}, headerOpts ...HeaderOpt) protocol.Envelope {
	env := protocol.Envelope{
		Topic:   fmt.Sprintf(topicFormat, thing.GetNamespace(), thing.GetID().GetName(), protocol.Group, protocol.Live, protocol.Messages, protocol.TopicAction(action)),
		Headers: NewHeaders(headerOpts...),
		Path:    fmt.Sprintf(pathThingFeatureMessagesOutboxFormat, featureID, action),
		Value:   value,
	}
	return env
}
