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

package protocol

// TopicCriterion is a representation of the defined topic creation options
type TopicCriterion string

// constants for the supported topic creation options
const (
	Commands TopicCriterion = "commands"
	Events   TopicCriterion = "events"
	Search   TopicCriterion = "search"
	Messages TopicCriterion = "messages"
	Errors   TopicCriterion = "errors"
)

// TopicChannel is a representation of the defined topic channel options
type TopicChannel string

// constants for the supported topic channel options
const (
	Twin TopicChannel = "twin"
	Live TopicChannel = "live"
)

// TopicAction is a representation of the defined topic action options
type TopicAction string

// contants for the supported topic action options
const (
	ActionCreate   TopicAction = "create"
	ActionCreated  TopicAction = "created"
	ActionModify   TopicAction = "modify"
	ActionModified TopicAction = "modified"
	ActionDelete   TopicAction = "delete"
	ActionDeleted  TopicAction = "deleted"
	ActionRetrieve TopicAction = "retrieve"
)

// Group is a representation of the topic group things
const Group = "things"

// TopicErrorsFormat is a representation of the topic error format
const TopicErrorsFormat = "unknown/unknown/" + string(Group) + "/%s/" + string(Errors)
