// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

package integration

const (
	influxdbImageRef    = "docker.io/library/influxdb:1.8.4"
	httpdImageRef       = "docker.io/library/httpd:latest"
	paramImageRef       = "imageRef"
	paramStart          = "start"
	paramConfig         = "config"
	ctrStatusCreated    = "CREATED"
	ctrStatusRunning    = "RUNNING"
	ctrStatusStopped    = "STOPPED"
	ctrStatusPaused     = "PAUSED"
	unknownMessageError = "unknown message is received with topic: %s"
)
