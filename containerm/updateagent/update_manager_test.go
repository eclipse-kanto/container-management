// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package updateagent

import (
	"context"
	"strconv"
	"testing"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	eventmocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/events"
	mgrmocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	uamocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/updateagent"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
	ummocks "github.com/eclipse-kanto/update-manager/test/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

const (
	domainName = "containers"

	testContainerName    = "test-container"
	testContainerVersion = "1.2.3"

	testContainerName2    = "test-container2"
	testContainerVersion2 = "11.22.33"

	sysContainerName    = "syslib"
	sysContainerCurrent = "1.1.0"
	sysContainerNext    = "1.2.0"
)

func TestNewUpdateManager(t *testing.T) {
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	mockContainerManager := mgrmocks.NewMockContainerManager(mockCtr)
	mockEventsManager := eventmocks.NewMockContainerEventsManager(mockCtr)

	updateManager := newUpdateManager(mockContainerManager, mockEventsManager, domainName, []string{"syslib"}, false)
	ctrUpdManager := updateManager.(*containersUpdateManager)

	testutil.AssertEqual(t, domainName, updateManager.Name())
	testutil.AssertEqual(t, mockContainerManager, ctrUpdManager.mgr)
	testutil.AssertEqual(t, mockEventsManager, ctrUpdManager.eventsMgr)
	testutil.AssertEqual(t, []string{"syslib"}, ctrUpdManager.systemContainers)
	testutil.AssertFalse(t, ctrUpdManager.verboseInventoryReport)

	updateManager.WatchEvents(context.Background())
	testutil.AssertNil(t, updateManager.Dispose())
}

func TestApplyInvalidDesiredState(t *testing.T) {
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	testActivityID := "test-apply-invalid-desired-state"
	updateManager := newUpdateManager(nil, nil, domainName, nil, false)

	mockCallback := ummocks.NewMockUpdateManagerCallback(mockCtr)
	updateManager.SetCallback(mockCallback)
	mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentificationFailed, gomock.Any(), []*types.Action{})

	updateManager.Apply(context.Background(), testActivityID, &types.DesiredState{})
}

func TestApplyNoDesiredContainers(t *testing.T) {
	testCases := map[string]struct {
		currentContainer  string
		errListContainers error
	}{
		"test-apply-desired-state-identify-no-actions": {},
		"test-apply-desired-state-identify-error":      {errListContainers: errors.New("cannot list current containers")},
		"test-apply-desired-state-identify-actions":    {currentContainer: testContainerName},
	}
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	for testActivityID, testCase := range testCases {
		t.Run(testActivityID, func(t *testing.T) {
			t.Log(testActivityID)
			mockContainerManager := mgrmocks.NewMockContainerManager(mockCtr)
			updateManager := newUpdateManager(mockContainerManager, nil, domainName, nil, false)
			ctrUpdManager := updateManager.(*containersUpdateManager)

			mockCallback := ummocks.NewMockUpdateManagerCallback(mockCtr)
			updateManager.SetCallback(mockCallback)

			var listContainers []*ctrtypes.Container
			var expActions []*types.Action
			if len(testCase.currentContainer) > 0 {
				listContainers = []*ctrtypes.Container{
					{Name: testCase.currentContainer, Image: ctrtypes.Image{Name: testCase.currentContainer + ":" + testContainerVersion}},
				}
				expActions = []*types.Action{
					{
						Component: &types.Component{ID: ctrUpdManager.domainName + ":" + listContainers[0].Name, Version: testContainerVersion},
						Status:    types.ActionStatusIdentified,
						Message:   util.GetActionMessage(util.ActionDestroy),
					},
				}
			} else {
				expActions = []*types.Action{}
			}
			mockContainerManager.EXPECT().List(gomock.Any()).Return(listContainers, testCase.errListContainers)
			mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentifying, "", nil)

			if testCase.errListContainers == nil {
				mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentified, "", expActions)
			} else {
				mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentificationFailed, testCase.errListContainers.Error(), nil)
			}

			if listContainers == nil && testCase.errListContainers == nil {
				mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusCompleted, "", expActions)
			}

			updateManager.Apply(context.Background(), testActivityID, &types.DesiredState{Domains: []*types.Domain{{ID: domainName}}})
			if listContainers == nil || testCase.errListContainers != nil {
				testutil.AssertNil(t, ctrUpdManager.operation)
			} else {
				testutil.AssertNotNil(t, ctrUpdManager.operation)
				testutil.AssertEqual(t, testActivityID, ctrUpdManager.operation.GetActivityID())
			}
		})
	}
}

func TestApplyWithDesiredContainers(t *testing.T) {
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	testActivityID := "test-identify-with-desired-containers"
	testDesiredState := &types.DesiredState{
		Domains: []*types.Domain{{
			ID: domainName,
			Components: []*types.ComponentWithConfig{
				// sys container shall be skipped and no actions identified
				createSimpleDesiredComponent(sysContainerName, sysContainerNext),
				// test container is not existing and to be created
				createSimpleDesiredComponent(testContainerName, testContainerVersion),
				// test container 2 is existing and to be checked if running only
				createSimpleDesiredComponent(testContainerName2, testContainerVersion2),
			},
		}},
	}

	sysContainer := createSimpleContainer(sysContainerName, sysContainerCurrent)
	appContainer := createSimpleContainer(testContainerName2, testContainerVersion2)

	expActions := []*types.Action{
		{
			Component: &types.Component{ID: domainName + ":" + testContainerName, Version: testContainerVersion},
			Status:    types.ActionStatusIdentified,
			Message:   util.GetActionMessage(util.ActionCreate),
		},
		{
			Component: &types.Component{ID: domainName + ":" + testContainerName2, Version: testContainerVersion2},
			Status:    types.ActionStatusIdentified,
			Message:   util.GetActionMessage(util.ActionCheck),
		},
	}

	mockContainerManager := mgrmocks.NewMockContainerManager(mockCtr)
	updateManager := newUpdateManager(mockContainerManager, nil, domainName, []string{sysContainerName}, false)
	ctrUpdManager := updateManager.(*containersUpdateManager)
	mockCallback := ummocks.NewMockUpdateManagerCallback(mockCtr)
	updateManager.SetCallback(mockCallback)

	mockContainerManager.EXPECT().List(gomock.Any()).Return([]*ctrtypes.Container{sysContainer, appContainer}, nil)
	mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentifying, "", nil)
	mockCallback.EXPECT().HandleDesiredStateFeedbackEvent(domainName, testActivityID, "", types.StatusIdentified, "", expActions)

	updateManager.Apply(context.Background(), testActivityID, testDesiredState)

	testutil.AssertNotNil(t, ctrUpdManager.operation)
	testutil.AssertEqual(t, testActivityID, ctrUpdManager.operation.GetActivityID())
}

func TestCommand(t *testing.T) {
	testCases := map[string]struct {
		command        *types.DesiredStateCommand
		setupOperation func(*uamocks.MockUpdateOperation)
	}{
		"test-command-missing-parameter":        {},
		"test-command-no-operation-in-progress": {command: &types.DesiredStateCommand{Command: types.CommandDownload}},
		"test-command-activity-id-mismatch": {
			command: &types.DesiredStateCommand{Command: types.CommandCleanup},
			setupOperation: func(mockOperation *uamocks.MockUpdateOperation) {
				mockOperation.EXPECT().GetActivityID().Return("test-command").Times(2)
			},
		},
		"test-command-without-baseline": {
			command: &types.DesiredStateCommand{Command: types.CommandUpdate},
			setupOperation: func(mockOperation *uamocks.MockUpdateOperation) {
				mockOperation.EXPECT().GetActivityID().Return("test-command-without-baseline")
				mockOperation.EXPECT().Execute(types.CommandUpdate, "")
			},
		},
		"test-command-with-baseline": {
			command: &types.DesiredStateCommand{Command: types.CommandActivate, Baseline: "test-baseline"},
			setupOperation: func(mockOperation *uamocks.MockUpdateOperation) {
				mockOperation.EXPECT().GetActivityID().Return("test-command-with-baseline")
				mockOperation.EXPECT().Execute(types.CommandActivate, "test-baseline")
			},
		},
	}
	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	for testActivityID, testCase := range testCases {
		t.Run(testActivityID, func(t *testing.T) {
			updateManager := newUpdateManager(nil, nil, domainName, nil, false)
			if testCase.setupOperation != nil {
				mockOperation := uamocks.NewMockUpdateOperation(mockCtr)
				updateManager.(*containersUpdateManager).operation = mockOperation
				testCase.setupOperation(mockOperation)
			}
			updateManager.Command(context.Background(), testActivityID, testCase.command)
		})
	}

}

func TestGet(t *testing.T) {
	testCases := map[string]int{
		"test-get-error-list-containers": -1,
		"test-get-no-containers":         0,
		"test-get-single-container":      1,
		"test-get-multiple-containers":   3,
	}

	mockCtr := gomock.NewController(t)
	defer mockCtr.Finish()

	for testActivityID, numberOfContainers := range testCases {
		t.Run(testActivityID, func(t *testing.T) {
			mockContainerManager := mgrmocks.NewMockContainerManager(mockCtr)
			updateManager := newUpdateManager(mockContainerManager, nil, domainName, nil, false)

			expSoftwareNodes := 1
			var errListContainers error
			var listContainers []*ctrtypes.Container
			if numberOfContainers < 0 {
				errListContainers = errors.New("cannot list containers")
			} else {
				expSoftwareNodes = 1 + numberOfContainers
				listContainers = make([]*ctrtypes.Container, numberOfContainers)
				for i := 0; i < numberOfContainers; i++ {
					listContainers[i] = createSimpleContainer(testContainerName+"-"+strconv.Itoa(i), testContainerVersion)
				}
			}
			mockContainerManager.EXPECT().List(context.Background()).Return(listContainers, errListContainers)

			inventory, err := updateManager.Get(context.Background(), testActivityID)
			testutil.AssertNil(t, err)
			testutil.AssertNotNil(t, inventory)
			testutil.AssertEqual(t, expSoftwareNodes, len(inventory.SoftwareNodes))

			expUpdateAgentID := domainName + "-update-agent"
			testutil.AssertEqual(t, types.SoftwareTypeApplication, inventory.SoftwareNodes[0].Type)
			testutil.AssertEqual(t, expUpdateAgentID, inventory.SoftwareNodes[0].ID)
			testutil.AssertNotEqual(t, "", inventory.SoftwareNodes[0].Name)
			testutil.AssertNotEqual(t, "", inventory.SoftwareNodes[0].Version)
			testutil.AssertEqual(t, 1, len(inventory.SoftwareNodes[0].Parameters))
			testutil.AssertEqual(t, "domain", inventory.SoftwareNodes[0].Parameters[0].Key)
			testutil.AssertEqual(t, domainName, inventory.SoftwareNodes[0].Parameters[0].Value)

			testutil.AssertEqual(t, expSoftwareNodes-1, len(inventory.Associations))
			for i := 1; i < expSoftwareNodes; i++ {
				expContainerID := domainName + ":" + listContainers[i-1].Name
				testutil.AssertEqual(t, types.SoftwareTypeContainer, inventory.SoftwareNodes[i].Type)
				testutil.AssertEqual(t, expContainerID, inventory.SoftwareNodes[i].ID)
				testutil.AssertEqual(t, "", inventory.SoftwareNodes[i].Name)
				testutil.AssertEqual(t, testContainerVersion, inventory.SoftwareNodes[i].Version)
				testutil.AssertTrue(t, len(inventory.SoftwareNodes[i].Parameters) > 0)

				testutil.AssertEqual(t, expUpdateAgentID, inventory.Associations[i-1].SourceID)
				testutil.AssertEqual(t, expContainerID, inventory.Associations[i-1].TargetID)
			}
		})
	}
}

func createSimpleContainer(name, version string) *ctrtypes.Container {
	ctr := &ctrtypes.Container{
		Name:  name,
		Image: ctrtypes.Image{Name: name + ":" + version},
	}
	util.FillDefaults(ctr)
	return ctr
}

func createSimpleDesiredComponent(name, version string) *types.ComponentWithConfig {
	return &types.ComponentWithConfig{
		Component: types.Component{ID: name, Version: version},
	}
}
