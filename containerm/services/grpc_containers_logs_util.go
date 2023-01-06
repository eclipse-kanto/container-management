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

package services

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/logger/jsonfile"
)

func sendAllLogs(file *os.File, srv pbcontainers.Containers_LogsServer) error {
	const maxBuffSize = 3 << 20
	scanner := bufio.NewScanner(file)
	buff := bytes.NewBufferString("")
	for {
		if !scanner.Scan() {
			if scanner.Err() != nil {
				return scanner.Err()
			}
			if err := srv.Send(&pbcontainers.GetLogsResponse{Log: buff.String()}); err != nil {
				return err
			}
			break
		}

		buff.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
		if len(buff.Bytes()) > maxBuffSize {
			if err := srv.Send(&pbcontainers.GetLogsResponse{Log: buff.String()}); err != nil {
				return err
			}
			buff.Reset()
		}
	}
	return nil
}

func tailLogs(file *os.File, srv pbcontainers.Containers_LogsServer, tail int) error {
	var (
		cursor   int64 = 0
		buff           = bytes.NewBufferString("")
		logLines       = make([]string, 0)
	)

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return nil
	}

	for {
		cursor--
		if _, err := file.Seek(cursor, io.SeekEnd); err != nil {
			return err
		}

		char := make([]byte, 1)
		if _, err := file.Read(char); err != nil {
			return err
		}

		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			logLines = append(logLines, reverse(buff.String()))
			buff.Reset()
		} else {
			buff.WriteString(string(char))
		}

		if cursor == -fileSize || len(logLines) == tail {
			break
		}
	}
	for i := len(logLines) - 1; i >= 0; i-- {
		if err := srv.Send(&pbcontainers.GetLogsResponse{Log: logLines[i]}); err != nil {
			return err
		}
	}
	return nil
}

func reverse(s string) (result string) {
	for _, r := range s {
		result = string(r) + result
	}
	return
}

func getLogFilePath(container *types.Container) (string, error) {
	if container.HostConfig == nil {
		return "", fmt.Errorf("no host config for container %s", container.ID)
	}
	if container.HostConfig.LogConfig == nil {
		return "", fmt.Errorf("no log config for container %s", container.ID)
	}
	if container.HostConfig.LogConfig.DriverConfig == nil {
		return "", fmt.Errorf("log driver config is not set for container %s", container.ID)
	}
	if container.HostConfig.LogConfig.DriverConfig.Type == types.LogConfigDriverNone {
		return "", fmt.Errorf("there are not any logs for container %s with log type %s", container.ID, types.LogConfigDriverNone)
	}
	if container.HostConfig.LogConfig.DriverConfig.Type == types.LogConfigDriverJSONFile {
		if container.HostConfig.LogConfig.DriverConfig.RootDir != "" {
			return filepath.Join(container.HostConfig.LogConfig.DriverConfig.RootDir, jsonfile.JSONLogFileName), nil
		}
		logFileDir, _ := filepath.Split(container.HostsPath)
		return filepath.Join(logFileDir, jsonfile.JSONLogFileName), nil
	}
	return "", fmt.Errorf("unknown log type %s", container.HostConfig.LogConfig.DriverConfig.Type)
}
