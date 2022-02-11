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

package ctr

import (
	"os"
	"path/filepath"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
	"github.com/eclipse-kanto/container-management/containerm/logger/jsonfile"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

type containerLogsManager interface {
	GetLogDriver(c *types.Container) (logger.LogDriver, error)
}

func newContainerLogsManager(metaPath string) containerLogsManager {
	return &ctrLogsMgr{containerLogsDirRoot: metaPath}
}

type ctrLogsMgr struct {
	containerLogsDirRoot string
}

func (mgr *ctrLogsMgr) GetLogDriver(container *types.Container) (logger.LogDriver, error) {
	cfg := container.HostConfig.LogConfig
	if cfg == nil || cfg.DriverConfig.Type == types.LogConfigDriverNone {
		return nil, nil
	}

	ctrLogsRootDir, err := mgr.initContainerLogsRootDir(container)
	if err != nil {
		return nil, err
	}

	if cfg.DriverConfig.Type == types.LogConfigDriverJSONFile {
		logDriverInfo := mgr.prepareLogDriverInfo(container, ctrLogsRootDir)
		logDriverCfg, cfgErr := mgr.prepareLogDriverConfig(container.HostConfig.LogConfig.DriverConfig)
		if cfgErr != nil {
			log.ErrorErr(cfgErr, "error processing log info for container id = %s", container.ID)
			return nil, cfgErr
		}
		return jsonfile.NewJSONFileLog(logDriverInfo, logDriverCfg...)
	}

	log.Warn("unsupported log driver %s", cfg.DriverConfig.Type)
	return nil, nil
}

func (mgr *ctrLogsMgr) prepareLogDriverConfig(driverCfg *types.LogDriverConfiguration) ([]logger.LogConfigOption, error) {
	var logConfigs []logger.LogConfigOption
	if driverCfg.MaxFiles != 0 {
		logConfigs = append(logConfigs, jsonfile.WithMaxFiles(driverCfg.MaxFiles))
	}
	if driverCfg.MaxSize != "" {
		bytes, err := util.SizeToBytes(driverCfg.MaxSize)
		if err != nil {
			return nil, err
		}
		logConfigs = append(logConfigs, jsonfile.WithMaxSize(bytes))
	}
	return logConfigs, nil
}

func (mgr *ctrLogsMgr) prepareLogDriverInfo(container *types.Container, ctrLogsDir string) logger.LogDriverInfo {
	return logger.LogDriverInfo{
		ContainerID:      container.ID,
		ContainerName:    container.Name,
		ContainerImageID: container.Image.Name,
		ContainerRootDir: ctrLogsDir,
		DaemonName:       "container-management",
	}
}

func (mgr *ctrLogsMgr) initContainerLogsRootDir(container *types.Container) (string, error) {
	// <containers-meta-path>/containers/<cid> as default root dir of container log
	rootDir := filepath.Join(mgr.containerLogsDirRoot, container.ID)

	cfg := container.HostConfig.LogConfig.DriverConfig
	if cfg.RootDir == "" {
		return rootDir, nil
	}

	specificRootDir := cfg.RootDir

	if !filepath.IsAbs(specificRootDir) {
		return "", log.NewErrorf("root dir for container log, %s should be an absolute path", specificRootDir)
	}
	// set <specificRootDir>/<cid> as root dir of container log.
	rootDir = filepath.Join(specificRootDir, container.ID)

	err := os.MkdirAll(rootDir, 0644)
	if err != nil {
		return "", log.NewErrorf("failed to create root log dir %s: %v", rootDir, err)
	}

	return rootDir, nil
}
