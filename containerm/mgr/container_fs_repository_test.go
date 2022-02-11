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

package mgr

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const (
	containerJSON      = "config.json"
	baseTestPath       = "../pkg/testutil/metapath/empty"
	baseContainersPath = "../pkg/testutil/metapath/empty/containers"
)

func TestLoadContainers(t *testing.T) {
	baseTestPath := "../pkg/testutil/metapath/valid"
	baseContainersPath := filepath.Join(baseTestPath, "containers")
	containerOnePath := filepath.Join(baseContainersPath, "test-id", containerJSON)
	containerTwoPath := filepath.Join(baseContainersPath, "61aff3dc-1f31-420b-883a-686165e1b06b", containerJSON)
	containerThreePath := filepath.Join(baseContainersPath, "dead-container", containerJSON)
	containerFourPath := filepath.Join(baseContainersPath, "paused-container", containerJSON)
	containerFivePath := filepath.Join(baseContainersPath, "stopped-container", containerJSON)

	c1 := readContainerFormFS(containerOnePath)
	c2 := readContainerFormFS(containerTwoPath)
	c3 := readContainerFormFS(containerThreePath)
	c4 := readContainerFormFS(containerFourPath)
	c5 := readContainerFormFS(containerFivePath)
	expectedResult := &map[string]*types.Container{
		c1.ID: c1,
		c2.ID: c2,
		c3.ID: c3,
		c4.ID: c4,
		c5.ID: c5,
	}

	locksCache := util.NewLocksCache()
	unitUnderTest := containerFsRepository{metaPath: baseTestPath, locksCache: &locksCache}

	containers, err := unitUnderTest.ReadAll()
	testutil.AssertNil(t, err)

	result := containerArrayToMap(containers)
	testutil.AssertEqual(t, 5, len(containers))
	testutil.AssertEqual(t, expectedResult, result)
}

func TestGetContainerById(t *testing.T) {
	containerID := "test-id"
	defer util.RemoveChildren(baseContainersPath) // Clean up after ourselves

	unitUnderTest, pathToContainer := setup(containerID)

	expected := readContainerFormFS(pathToContainer)

	actual, err := unitUnderTest.Read(containerID)
	testutil.AssertNotNil(t, actual)
	testutil.AssertEqual(t, expected, actual)
	testutil.AssertNil(t, err) //
}

func TestUpdateContainer(t *testing.T) {
	containerID := "test-id"
	defer util.RemoveChildren(baseContainersPath) // Clean up after ourselves

	unitUnderTest, pathToContainer := setup(containerID)

	containerPreUpdate := readContainerFormFS(pathToContainer)
	containerToBeChanged := readContainerFormFS(pathToContainer)
	containerToBeChanged.DomainName = "Changed-Domain-Name"

	unitUnderTest.Save(containerToBeChanged)

	allContainers, allErr := unitUnderTest.ReadAll()
	target, err := unitUnderTest.Read(containerID)

	testutil.AssertNil(t, allErr)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, 1, len(allContainers))
	testutil.AssertEqual(t, containerToBeChanged, target)
	testutil.AssertNotEqual(t, containerPreUpdate, target)
}

func TestDeleteContainer(t *testing.T) {
	containerID := "test-id"
	pathToContainer := filepath.Join(baseContainersPath, containerID)
	defer util.RemoveChildren(baseContainersPath) // Clean up after ourselves

	unitUnderTest, _ := setup(containerID)

	allContainers, allErr := unitUnderTest.ReadAll()
	testutil.AssertNil(t, allErr)
	testutil.AssertEqual(t, 1, len(allContainers))

	err := unitUnderTest.Delete(containerID)
	testutil.AssertNil(t, err)

	postDeleteContainers, paErr := unitUnderTest.ReadAll()
	testutil.AssertNil(t, paErr)
	testutil.AssertEqual(t, 0, len(postDeleteContainers))

	exists, _ := util.IsDirectory(pathToContainer)

	deleted, err := unitUnderTest.Read(containerID)
	testutil.AssertFalse(t, exists)
	testutil.AssertNil(t, deleted)
	testutil.AssertNotNil(t, err) //
}

func TestPruneContainers(t *testing.T) {
	containerID := "test-id"
	unitUnderTest, _ := setup(containerID)
	os.Mkdir(filepath.Join(baseContainersPath, "invalidly-deleted"), 0777)
	defer util.RemoveChildren(baseContainersPath) // Clean up after ourselves

	childrenBeforePrune, _ := util.GetDirChildrenNames(baseContainersPath)
	testutil.AssertEqual(t, 2, len(childrenBeforePrune))

	err := unitUnderTest.Prune()
	testutil.AssertNil(t, err)

	childrenAfterPrune, _ := util.GetDirChildrenNames(baseContainersPath)
	testutil.AssertEqual(t, 1, len(childrenAfterPrune))
	testutil.AssertEqual(t, childrenAfterPrune[0], containerID)
}

func TestPruneContainerOnInvalidError(t *testing.T) {
	locksCache := util.NewLocksCache()
	unitUnderTest := containerFsRepository{metaPath: "/invalid-path", locksCache: &locksCache}

	err := unitUnderTest.Prune()
	testutil.AssertNotNil(t, err)
}

func setup(containerID string) (containerFsRepository, string) {
	// containerFolderPath := filepath.Join(baseContainersPath, containerId)
	pathToContatiner := filepath.Join("../pkg/testutil/metapath/valid/containers", containerID, containerJSON)
	expectedPathToNewContaianer := filepath.Join(baseContainersPath, containerID, containerJSON)
	locksCache := util.NewLocksCache()

	// Clean up if needed
	util.MkDirs(baseContainersPath)
	util.RemoveChildren(baseContainersPath)

	// Create if needed
	util.MkDirs(filepath.Join(baseContainersPath, containerID))
	util.Copy(pathToContatiner, expectedPathToNewContaianer, 2048)

	return containerFsRepository{metaPath: baseTestPath, locksCache: &locksCache}, pathToContatiner
}

func containerArrayToMap(arr []*types.Container) *map[string]*types.Container {
	mp := make(map[string]*types.Container)

	for indx := range arr {
		mp[arr[indx].ID] = arr[indx]
	}

	return &mp
}

func readContainerFormFS(path string) *types.Container {
	jsonFile, _ := os.Open(path)
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var container types.Container
	json.Unmarshal(byteValue, &container)

	return &container
}
