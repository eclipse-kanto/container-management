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

package services

// Service IDs used to enable configurability and discovery between the different services using the internal daemon's registry
const (
	// Service ID of the container management gRPC service
	ContainersServiceID = "container-management.grpc.v1.service-containers"
	// Service ID of the system information gRPC service
	SystemInfoServiceID = "container-management.grpc.v1.service-systemInfo"
)
