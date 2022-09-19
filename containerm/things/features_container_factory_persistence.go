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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	thingStorageFileNameTemplate = "sup_back_%s.json"
)

type containerStorage interface {
	Restore() (map[string]string, error)
	UpdateContainersInfo(ctrFeaturesInfo map[string]string)
	StoreContainerInfo(ctrID string)
	DeleteContainerInfo(ctrID string)
}
type containerFactoryStorage struct {
	storageRoot              string
	thingID                  string
	ctrsInfoMutex            sync.Mutex
	managedContainerFeatures map[string]string
}

func newContainerFactoryStorage(storageRoot string, thingID string) containerStorage {
	return &containerFactoryStorage{
		storageRoot:              storageRoot,
		thingID:                  thingID,
		managedContainerFeatures: make(map[string]string),
	}
}

func (ctrFactoryStorage *containerFactoryStorage) store(ctrFeaturesInfo map[string]string) error {
	jsonString, err := json.Marshal(ctrFeaturesInfo)
	if err != nil {
		return err
	}

	fileName := ctrFactoryStorage.generateFileName()
	if err := ioutil.WriteFile(fileName, jsonString, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (ctrFactoryStorage *containerFactoryStorage) Restore() (map[string]string, error) {
	if compatErr := ctrFactoryStorage.ensureFileNameCompatible(); compatErr != nil {
		log.ErrorErr(compatErr, "could not update file name to the current version for thing ID [%s]", ctrFactoryStorage.thingID)
		return nil, compatErr
	}
	fileName := ctrFactoryStorage.generateFileName()

	fi, fierr := os.Stat(fileName)
	if fierr != nil {
		if os.IsNotExist(fierr) {
			return nil, nil
		}
		return nil, fierr
	} else if fi.Size() == 0 {
		log.Warn("the file %s is empty", fileName)
		return nil, nil
	} else {
		log.Debug("successfully retrieved file stats for [%s]", fileName)
	}

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	err = json.Unmarshal(file, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}
func (ctrFactoryStorage *containerFactoryStorage) UpdateContainersInfo(ctrFeaturesInfo map[string]string) {
	ctrFactoryStorage.ctrsInfoMutex.Lock()
	defer ctrFactoryStorage.ctrsInfoMutex.Unlock()
	ctrFactoryStorage.managedContainerFeatures = ctrFeaturesInfo
	if err := ctrFactoryStorage.store(ctrFactoryStorage.managedContainerFeatures); err != nil {
		log.ErrorErr(err, "failed to store things service info")
	}
}

func (ctrFactoryStorage *containerFactoryStorage) StoreContainerInfo(ctrID string) {
	ctrFactoryStorage.ctrsInfoMutex.Lock()
	defer ctrFactoryStorage.ctrsInfoMutex.Unlock()
	if ctrFactoryStorage.managedContainerFeatures == nil {
		ctrFactoryStorage.managedContainerFeatures = map[string]string{}
	}
	ctrFactoryStorage.managedContainerFeatures[ctrID] = generateContainerFeatureID(ctrID)
	if err := ctrFactoryStorage.store(ctrFactoryStorage.managedContainerFeatures); err != nil {
		log.ErrorErr(err, "failed to store things service info for new container ID = %s", ctrID)
	}
}
func (ctrFactoryStorage *containerFactoryStorage) DeleteContainerInfo(ctrID string) {
	ctrFactoryStorage.ctrsInfoMutex.Lock()
	defer ctrFactoryStorage.ctrsInfoMutex.Unlock()
	if ctrFactoryStorage.managedContainerFeatures == nil || len(ctrFactoryStorage.managedContainerFeatures) == 0 {
		return
	}
	if _, ok := ctrFactoryStorage.managedContainerFeatures[ctrID]; ok {
		delete(ctrFactoryStorage.managedContainerFeatures, ctrID)
		if err := ctrFactoryStorage.store(ctrFactoryStorage.managedContainerFeatures); err != nil {
			log.ErrorErr(err, "failed to delete things service info for new container ID = %s", ctrID)
		}
	}
}
func (ctrFactoryStorage *containerFactoryStorage) generateFileName() string {
	return filepath.Join(ctrFactoryStorage.storageRoot, fmt.Sprintf(thingStorageFileNameTemplate, strings.Replace(ctrFactoryStorage.thingID, ":", "_", -1)))
}

func (ctrFactoryStorage *containerFactoryStorage) generateFileNameV1() string {
	return filepath.Join(ctrFactoryStorage.storageRoot, fmt.Sprintf(thingStorageFileNameTemplate, strings.Replace(ctrFactoryStorage.thingID, ":", "_", 1)))
}

func (ctrFactoryStorage *containerFactoryStorage) ensureFileNameCompatible() error {
	oldName := ctrFactoryStorage.generateFileNameV1()

	_, fierr := os.Stat(oldName)
	if fierr != nil {
		if os.IsNotExist(fierr) {
			log.Debug("no old file is persisted from previous versions for thing ID [%s]", ctrFactoryStorage.thingID)
			return nil
		}
		return fierr
	}
	log.Debug("updating existing old file from previous versions to the new one for thing ID [%s]", ctrFactoryStorage.thingID)
	return os.Rename(oldName, ctrFactoryStorage.generateFileName())
}
