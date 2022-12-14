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

package mgr

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
)

const (
	containersRootDir       = "containers"
	containerConfigFilename = "config.json"
)

type containerFsRepository struct {
	metaPath   string
	locksCache *util.LocksCache
}

// WriteHostConfig saves the host configuration on disk for the container,
// and returns a deep copy of the saved object. Callers must hold a Container lock.
func (repository *containerFsRepository) Save(container *types.Container) (*types.Container, error) {
	var (
		buf      bytes.Buffer
		deepCopy types.Container
	)

	if container.ID == "" {
		return nil, log.NewErrorf("container id cannot be empty string: %+v", container)
	}

	lock := repository.locksCache.GetLock(container.ID)
	lock.Lock()
	defer lock.Unlock()

	basePath := repository.getContainerContainerMetaPath(container.ID)
	if valid, err := util.IsDirectory(basePath); valid == false || err != nil {
		util.MkDir(basePath)
	}

	pth := repository.getContainerConfigMetaPath(container.ID)

	f, err := ioutils.NewAtomicFileWriter(pth, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w := io.MultiWriter(&buf, f)
	if err := json.NewEncoder(w).Encode(&container); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(&buf).Decode(&deepCopy); err != nil {
		return nil, err
	}
	return &deepCopy, nil
}

// readAllConfigs reads the configuration on disk for the containers,
// and returns a deep copies of the saved objects
func (repository *containerFsRepository) ReadAll() ([]*types.Container, error) {
	var readCtrConfigs []*types.Container
	path := filepath.Join(repository.metaPath, containersRootDir)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debug("the root containers directory does not exist - will exit loading ")
		return nil, nil
	}

	contDirs, readErr := ioutil.ReadDir(path)

	if readErr != nil {
		return nil, readErr
	}
	var (
		ctr *types.Container
		err error
	)
	if len(contDirs) != 0 {
		for _, cont := range contDirs {
			ctr, err = repository.Read(cont.Name())
			if err != nil {
				log.ErrorErr(err, "error reading configuration for container id = %s", cont.Name())
			} else {
				readCtrConfigs = append(readCtrConfigs, ctr)
			}
		}
	}

	return readCtrConfigs, nil
}

// readConfig reads a container's configuration from disk
// and returns a deep copies of the saved object
func (repository *containerFsRepository) Read(containerID string) (*types.Container, error) {
	lock := repository.locksCache.GetLock(containerID)
	lock.Lock()
	defer lock.Unlock()

	pth := repository.getContainerConfigMetaPath(containerID)

	return util.ReadContainer(pth)
}

func (repository *containerFsRepository) Delete(containerID string) error {
	lock := repository.locksCache.GetLock(containerID)
	lock.Lock()
	defer func() {
		lock.Unlock()
		repository.locksCache.RemoveLock(containerID)
	}()

	pth := repository.getContainerContainerMetaPath(containerID)

	if err := os.RemoveAll(pth); err != nil {
		customErr := log.NewErrorf("failed to Delete local storage for container id = %s, %v", containerID, err)
		return customErr
	}

	return nil
}

// Prune removes all container folders that may have been deleted invalidly
// For example: the config.json is missing
func (repository *containerFsRepository) Prune() error {
	path := filepath.Join(repository.metaPath, containersRootDir)
	var errs = errorUtil.CompoundError{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return log.NewError("root containers directory does not exist")
	}

	contDirs, readErr := ioutil.ReadDir(path)

	if readErr != nil {
		return readErr
	}

	for _, containerDir := range contDirs {
		configDir := repository.getContainerConfigMetaPath(containerDir.Name())

		if _, err := os.Stat(configDir); err == nil {
			// file exists cary on
			continue
		} else if os.IsNotExist(err) {
			// file does *not* exist
			repository.Delete(containerDir.Name())
		} else {
			// something whent wrong
			log.Error(err.Error())
			errs.Append(err)
		}
	}

	if errs.Size() != 0 {
		return &errs
	}
	return nil
}

func (repository *containerFsRepository) getContainerConfigMetaPath(containerID string) string {
	basePath := repository.getContainerContainerMetaPath(containerID)
	return filepath.Join(basePath, containerConfigFilename)
}

func (repository *containerFsRepository) getContainerContainerMetaPath(containerID string) string {
	return filepath.Join(repository.metaPath, containersRootDir, containerID)
}
