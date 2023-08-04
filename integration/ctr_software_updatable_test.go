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

//go:build integration

package integration

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/things"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/things/client"

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type softwareUpdatableSuite struct {
	ctrManagementSuite
	suFeatureURL string
	suFilter     string
}

const (
	actionInstall     = "install"
	paramCorID        = "correlationId"
	paramForced       = "forced"
	validContainerURL = "https://raw.githubusercontent.com/eclipse-kanto/container-management/main/containerm/pkg/testutil/config/container/valid.json"
)

func (suite *softwareUpdatableSuite) SetupSuite() {
	suite.SetupCtrManagementSuite()

	suite.suFeatureURL = util.GetFeatureURL(suite.ctrThingURL, things.SoftwareUpdatableFeatureID)
	suite.suFilter = fmt.Sprintf("like(resource:path,'/features/%s*')", things.SoftwareUpdatableFeatureID)

	def := client.NewDefinitionID(things.SoftwareUpdatableDefinitionNamespace,
		things.SoftwareUpdatableDefinitionName,
		things.SoftwareUpdatableDefinitionVersion)
	suite.assertCtrFeatureDefinition(suite.suFeatureURL, fmt.Sprintf("[\"%s\"]", def))
}

func (suite *softwareUpdatableSuite) TearDownSuite() {
	suite.TearDown()
}

func TestSoftwareUpdatableSuite(t *testing.T) {
	suite.Run(t, new(softwareUpdatableSuite))
}

func (suite *softwareUpdatableSuite) installContainer(params map[string]interface{}) (string, error) {
	wsConnection := suite.createWSConnection()
	defer suite.closeUnsubscribe(wsConnection)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, suite.suFilter)
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	_, err = util.ExecuteOperation(suite.Cfg, suite.suFeatureURL, actionInstall, params)
	suite.closeOnError(wsConnection, err, "failed to execute software updatable install for containers with params %v", params)

	var (
		featureID      string
		eventValue     map[string]interface{}
		propertyStatus string
		isStarted      bool
		isDownloaded   bool
		isCtrInstalled bool
	)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		var err error
		if event.Topic.String() == suite.topicModified {
			eventValue, err = parseMap(event.Value)
			require.NoError(suite.T(), err, "failed to parse event value")

			if event.Path != suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "lastOperation") {
				if event.Path == suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "installedDependencies") {
					for _, element := range eventValue {
						dockerValue, err := parseMap(element)
						require.NoError(suite.T(), err, "failed to parse docker value")
						featureID, err = parseString(dockerValue["name"])
						require.NoError(suite.T(), err, "failed to parse property name")
					}
					return false, nil
				}

				propertyStatus, err = parseString(eventValue["status"])
				require.NoError(suite.T(), err, "failed to parse property status")

				if propertyStatus == string(datatypes.FinishedError) {
					propertyMessage, err := parseString(eventValue["message"])
					require.NoError(suite.T(), err, "failed to parse property message")
					return true, fmt.Errorf(propertyMessage)
				}
				return true, fmt.Errorf("received event is not expected")
			}

			propertyStatus, err = parseString(eventValue["status"])
			require.NoError(suite.T(), err, "failed to parse property status")

			if propertyStatus == string(datatypes.Started) {
				isStarted = true
				return false, nil
			}
			if propertyStatus == string(datatypes.Downloading) {
				if isStarted {
					return false, nil
				}
				return true, nil
			}
			if propertyStatus == string(datatypes.Downloaded) {
				if isStarted {
					isDownloaded = true
					return false, nil
				}
				return true, nil
			}
			if propertyStatus == string(datatypes.Installing) {
				if isStarted && isDownloaded {
					return false, nil
				}
				return true, nil
			}
			if propertyStatus == string(datatypes.Installed) {
				if isStarted && isDownloaded {
					isCtrInstalled = true
					return false, nil
				}
				return true, nil
			}
			if propertyStatus == string(datatypes.FinishedSuccess) {
				if isStarted && isDownloaded && isCtrInstalled {
					return true, nil
				}
				return true, fmt.Errorf("container status is not expected")
			}
			return true, fmt.Errorf("event for an unexpected container status is received")
		}
		return true, fmt.Errorf("unknown message is received")
	})
	if err != nil {
		wsConnection.Close()
	}
	return featureID, err
}

func (suite *softwareUpdatableSuite) removeContainer(params map[string]interface{}) error {
	wsConnection := suite.createWSConnection()
	defer suite.closeUnsubscribe(wsConnection)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, suite.suFilter)
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	_, err = util.ExecuteOperation(suite.Cfg, suite.suFeatureURL, "remove", params)
	suite.closeOnError(wsConnection, err, "failed to execute software updatable install for containers with params %v", params)

	var (
		eventValue      map[string]interface{}
		propertyStatus  string
		isRemoving      bool
		isInstalledDeps bool
		isRemoved       bool
	)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		var err error
		if event.Topic.String() == suite.topicModified {
			if event.Path == suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "lastOperation") {
				eventValue, err = parseMap(event.Value)
				require.NoError(suite.T(), err, "failed to parse event value")

				propertyStatus, err = parseString(eventValue["status"])
				require.NoError(suite.T(), err, "failed to parse property status")

				if propertyStatus == string(datatypes.Removing) {
					isRemoving = true
					return false, nil
				}
				if propertyStatus == string(datatypes.Removed) {
					if isRemoving && isInstalledDeps {
						isRemoved = true
						return false, nil
					}
					return true, nil
				}
				if propertyStatus == string(datatypes.FinishedSuccess) {
					if isRemoving && isInstalledDeps && isRemoved {
						return true, nil
					}
					return true, fmt.Errorf("container status is not expected")
				}
				return true, nil
			} else if event.Path == suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "installedDependencies") {
				if isRemoving {
					isInstalledDeps = true
					return false, nil
				}
				return true, nil
			} else if event.Path == suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "lastFailedOperation") {
				eventValue, err = parseMap(event.Value)
				require.NoError(suite.T(), err, "failed to parse event value")

				propertyMessage, err := parseString(eventValue["message"])
				require.NoError(suite.T(), err, "failed to parse property message")
				return true, fmt.Errorf(propertyMessage)
			}

			return true, fmt.Errorf("event for an unexpected container status while removing is received")
		}
		return true, fmt.Errorf("unknown message is received")
	})
	if err != nil {
		wsConnection.Close()
	}
	return err
}

func (suite *softwareUpdatableSuite) TestSoftwareInstallRemove() {
	ctrID, err := suite.installContainer(suite.createParameters(false))
	require.NoError(suite.T(), err, "failed to process installing the container")

	require.NoError(suite.T(), suite.removeContainer(createRemoveParams(ctrID)), "failed to process removing the container")
}

func (suite *softwareUpdatableSuite) TestSoftwareInstallWithInvalidParameters() {
	wsConnection := suite.createWSConnection()
	defer suite.closeUnsubscribe(wsConnection)

	params := make(map[string]interface{})
	_, err := util.ExecuteOperation(suite.Cfg, suite.suFeatureURL, actionInstall, params)
	if err != nil {
		wsConnection.Close()
	}
	require.Errorf(suite.T(), err, "failed to execute software updatable install for containers with params %v", params)

}

func (suite *softwareUpdatableSuite) TestSoftwareInstallWithWrongChecksum() {
	_, err := suite.installContainer(suite.createParameters(true))
	require.ErrorContains(suite.T(), err, "internal runtime error")
}

func (suite *softwareUpdatableSuite) TestSoftwareRemoveNonExistingContainer() {
	ctrID := "NonExistingContainer"
	require.ErrorContains(suite.T(), suite.removeContainer(createRemoveParams(ctrID)), "container with ID = "+ctrID+" does not exist")
}

func (suite *softwareUpdatableSuite) createParameters(wrongChecksum bool) map[string]interface{} {
	src, err := download(validContainerURL)
	require.NoError(suite.T(), err, "unable to download file from url "+validContainerURL)
	splitStr := strings.Split(validContainerURL, "/")
	filePath := "/tmp/" + splitStr[len(splitStr)-1]
	require.NoError(suite.T(), os.WriteFile(filePath, src, 7777), "unable to write file with path "+filePath)

	defer require.NoError(suite.T(), os.Remove(filePath), "unable to remove file with path "+filePath)
	return createInstallParams(filePath, src, wrongChecksum)
}

func createInstallParams(filePath string, src []byte, wrongChecksum bool) map[string]interface{} {
	fileInfo, ctrStruct, err := getCtrImageStructure(filePath)
	if err != nil {
		return make(map[string]interface{})
	}
	splitStr := strings.Split(ctrStruct.Image.Name, "/")
	swModuleStr := strings.Split(splitStr[len(splitStr)-1], ":")

	swModule := make(map[string]interface{})
	swModule["softwareModule"] = map[string]string{
		"name":    swModuleStr[0],
		"version": swModuleStr[1],
	}
	checksum := fmt.Sprintf("%x", md5.Sum(src))
	if wrongChecksum {
		checksum = fmt.Sprintf("%x", sha1.Sum(src))
	}
	swModule["artifacts"] = []*datatypes.SoftwareArtifactAction{
		{
			Download: map[datatypes.Protocol]*datatypes.Links{
				datatypes.HTTPS: {
					URL:    validContainerURL,
					MD5URL: validContainerURL,
				},
			},
			Checksums: map[datatypes.Hash]string{
				datatypes.MD5:    checksum,
				datatypes.SHA1:   fmt.Sprintf("%x", sha1.Sum(src)),
				datatypes.SHA256: fmt.Sprintf("%x", sha256.Sum256(src)),
			},
			Size:     uint64(fileInfo.Size()),
			FileName: fileInfo.Name(),
		},
	}
	return map[string]interface{}{
		paramForced:       true,
		paramCorID:        uuid.NewString(),
		"softwareModules": [1]map[string]interface{}{swModule},
	}
}

func createRemoveParams(ctrID string) map[string]interface{} {
	return map[string]interface{}{
		paramForced: true,
		paramCorID:  uuid.NewString(),
		"software": []*datatypes.DependencyDescription{
			&datatypes.DependencyDescription{
				Name: ctrID,
			},
		},
	}
}

// CtrImageStruct represents the container image structure which will be used to software update
type CtrImageStruct struct {
	ContainerName string      `json:"container_name"`
	Image         ImageStruct `json:"image"`
}

// ImageStruct represents the image structure which will be used to software update of container
type ImageStruct struct {
	Name string `json:"name"`
}

func getCtrImageStructure(filePath string) (fs.FileInfo, CtrImageStruct, error) {
	ctrImage := CtrImageStruct{}
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return fileStat, ctrImage, err
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return fileStat, ctrImage, err
	}
	return fileStat, ctrImage, json.Unmarshal(bytes, &ctrImage)
}

func download(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
