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
	"github.com/golang/mock/gomock"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	mockseventsspb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/events"
	mocksmgrpb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	mocksthingspb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/things"
)

const (
	testContainerID = "some-container-id"

	testContainerImage     = "image:latest"
	testContainerName      = "name"
	testContainerCreatedAt = "2020/20/20"
)

var (
	testThingsFeaturesDefaultSet = []string{ContainerFactoryFeatureID, SoftwareUpdatableFeatureID, MetricsFeatureID}
	testContainer                = &types.Container{
		ID:    testContainerID,
		Name:  testContainerName,
		State: &types.State{},
	}
	mockContainerManager   *mocksmgrpb.MockContainerManager
	mockEventsManager      *mockseventsspb.MockContainerEventsManager
	mockThing              *mocksthingspb.MockThing
	mockContainerStorage   *mocksthingspb.MockcontainerStorage
	testThingsMgr          *containerThingsMgr
	testContainerFeatureID = generateContainerFeatureID(testContainerID)
)

func copyTestContainer(testContainer *types.Container) types.Container {
	return types.Container{
		ID:    testContainer.ID,
		Name:  testContainer.Name,
		State: testContainer.State,
	}
}

func setupManagerMock(controller *gomock.Controller) {
	mockContainerManager = mocksmgrpb.NewMockContainerManager(controller)
}

func setupEventsManagerMock(controller *gomock.Controller) {
	mockEventsManager = mockseventsspb.NewMockContainerEventsManager(controller)
}

func setupThingMock(controller *gomock.Controller) {
	mockThing = mocksthingspb.NewMockThing(controller)
}

func setupContainerFactoryStorageMock(controller *gomock.Controller) {
	mockContainerStorage = mocksthingspb.NewMockcontainerStorage(controller)
}

const (
	testThingsStoragePath = "../pkg/testutil/metapath/valid/things"
)

func setupThingsContainerManager(controller *gomock.Controller) {
	testThingsMgr = newThingsContainerManager(mockContainerManager, mockEventsManager,
		"",
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
		&tlsConfig{})
}
