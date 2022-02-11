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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	testAPILogDriverType = types.LogConfigDriverJSONFile
	testLogDriverType    = jsonFile

	testLogMaxFiles = 2
	testLogMaxSize  = "100M"

	testAPILogMode = types.LogModeBlocking
	testLogMode    = blocking

	testLogBufferSize = "50M"
)

func TestToAPILogConfiguration(t *testing.T) {
	logConfig := &logConfiguration{
		Type:          testLogDriverType,
		MaxFiles:      testLogMaxFiles,
		MaxSize:       testLogMaxSize,
		Mode:          testLogMode,
		MaxBufferSize: testLogBufferSize,
	}

	result := toAPILogConfiguration(logConfig)

	t.Run("test_to_api_log_configuration_driver_type", func(t *testing.T) {
		testutil.AssertEqual(t, toAPILogDriver(logConfig.Type), result.DriverConfig.Type)
	})
	t.Run("test_to_api_log_configuration_driver_max_files", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.MaxFiles, result.DriverConfig.MaxFiles)
	})
	t.Run("test_to_api_log_configuration_driver_max_size", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.MaxSize, result.DriverConfig.MaxSize)
	})
	t.Run("test_to_api_log_configuration_mode", func(t *testing.T) {
		testutil.AssertEqual(t, toAPILogMode(logConfig.Mode), result.ModeConfig.Mode)
	})
	t.Run("test_to_api_log_configuration_buffer_size", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.MaxBufferSize, result.ModeConfig.MaxBufferSize)
	})

}

func TestFromAPILogConfiguration(t *testing.T) {
	logConfig := &types.LogConfiguration{
		DriverConfig: &types.LogDriverConfiguration{
			Type:     testAPILogDriverType,
			MaxFiles: testLogMaxFiles,
			MaxSize:  testLogMaxSize,
		},
		ModeConfig: &types.LogModeConfiguration{
			Mode:          testAPILogMode,
			MaxBufferSize: testLogBufferSize,
		},
	}
	result := fromAPILogConfiguration(logConfig)

	t.Run("test_from_api_log_configuration_driver_type", func(t *testing.T) {
		testutil.AssertEqual(t, fromAPILogDriver(logConfig.DriverConfig.Type), result.Type)
	})
	t.Run("test_from_api_log_configuration_driver_max_files", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.DriverConfig.MaxFiles, result.MaxFiles)
	})
	t.Run("test_from_api_log_configuration_driver_max_size", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.DriverConfig.MaxSize, result.MaxSize)
	})
	t.Run("test_from_api_log_configuration_mode", func(t *testing.T) {
		testutil.AssertEqual(t, fromAPILogMode(logConfig.ModeConfig.Mode), result.Mode)
	})
	t.Run("test_from_api_log_configuration_buffer_size", func(t *testing.T) {
		testutil.AssertEqual(t, logConfig.ModeConfig.MaxBufferSize, result.MaxBufferSize)
	})
}

func TestToAPILogDriver(t *testing.T) {
	t.Run("test_to_api_log_driver_driver_type_json", func(t *testing.T) {
		testutil.AssertEqual(t, types.LogConfigDriverJSONFile, toAPILogDriver(jsonFile))
	})

	t.Run("test_to_api_log_driver_driver_type_none", func(t *testing.T) {
		testutil.AssertEqual(t, types.LogConfigDriverNone, toAPILogDriver(none))
	})
}

func TestTFromAPILogDriver(t *testing.T) {

	t.Run("test_from_api_log_driver_driver_type_json", func(t *testing.T) {
		testutil.AssertEqual(t, jsonFile, fromAPILogDriver(types.LogConfigDriverJSONFile))
	})

	t.Run("test_form_api_log_driver_driver_type_none", func(t *testing.T) {
		testutil.AssertEqual(t, none, fromAPILogDriver(types.LogConfigDriverNone))
	})
}

func TestToAPILogMode(t *testing.T) {

	t.Run("test_to_api_log_mode_blocking", func(t *testing.T) {
		testutil.AssertEqual(t, types.LogModeBlocking, toAPILogMode(blocking))
	})

	t.Run("test_to_api_log_mode_non_blocking", func(t *testing.T) {
		testutil.AssertEqual(t, types.LogModeNonBlocking, toAPILogMode(nonBlocking))
	})

}

func TestTFromAPILogMode(t *testing.T) {

	t.Run("test_from_api_log_mode_blocking", func(t *testing.T) {
		testutil.AssertEqual(t, blocking, fromAPILogMode(types.LogModeBlocking))
	})

	t.Run("test_from_api_log_mode_non_blocking", func(t *testing.T) {
		testutil.AssertEqual(t, nonBlocking, fromAPILogMode(types.LogModeNonBlocking))
	})
}
