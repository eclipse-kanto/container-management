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

package framework

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
)

// TestData constant for the testdata directory
const TestData = "testdata"

var customResultDefaultFns = map[string]func(result icmd.Result, args ...string) assert.BoolOrComparison{
	"REGEX":     regex,
	"LOGS_JSON": logs,
}

// TestCaseCMD represents a command and expected result
type TestCaseCMD struct {
	name             string
	icmd             icmd.Cmd
	expected         icmd.Expected
	goldenFile       string
	customResult     string
	customResultArgs []string
	setupCmd         *[]icmd.Cmd
	onExit           *[]icmd.Cmd
}

// GetTestCaseFromYamlFile parses yaml file to TestCaseCMD array.
func GetTestCaseFromYamlFile(filePath string) ([]TestCaseCMD, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(f)
	cmdList := []TestCaseCMD{}
	for {
		cmTestCommand := new(TestCommand)
		err := decoder.Decode(cmTestCommand)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		cmdList = append(cmdList, fromAPITestCommand(*cmTestCommand))
	}

	return cmdList, nil
}

// GetAllTestCasesFromTestdataDir reads all files that matches the "*-test.yaml" and parses them to TestCaseCMD map.
func GetAllTestCasesFromTestdataDir() (map[string][]TestCaseCMD, error) {
	files, err := filepath.Glob(filepath.Join(TestData, "*-test.yaml"))
	if err != nil {
		return nil, err
	}

	testCases := make(map[string][]TestCaseCMD)
	for _, file := range files {
		testCase, err := GetTestCaseFromYamlFile(file)
		if err != nil {
			return nil, err
		}
		testCases[file] = testCase
	}
	return testCases, nil
}

// RunCmdTestCases runs the provided test cases and asserts the result according to the provided parameters.
func RunCmdTestCases(t *testing.T, cmdList []TestCaseCMD) {
	for _, cmd := range cmdList {
		t.Run(cmd.name, func(t *testing.T) {
			if cmd.setupCmd != nil {
				runMultipleCommands(t, *cmd.setupCmd)
			}
			result := icmd.RunCmd(cmd.icmd)
			if cmd.goldenFile != "" {
				assert.Assert(t, golden.String(result.Stdout(), cmd.goldenFile))
			}
			if cmd.customResult != "" {
				assertCustomResult(t, *result, cmd.customResult, cmd.customResultArgs...)
			}
			result.Assert(t, cmd.expected)
		})
		if cmd.onExit != nil {
			t.Run(cmd.name+"_on_exit", func(t *testing.T) {
				runMultipleCommands(t, *cmd.onExit)
			})
		}
	}
}

func runMultipleCommands(t *testing.T, cmdArr []icmd.Cmd) {
	for _, cmd := range cmdArr {
		result := icmd.RunCmd(cmd)
		result.Assert(t, icmd.Expected{ExitCode: 0})
	}
}

func buildCmd(binary string, args ...string) *icmd.Cmd {
	envArgs := []string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "$") {
			if val, ok := os.LookupEnv(strings.TrimPrefix(arg, "$")); ok {
				arg = val
			}
		}
		envArgs = append(envArgs, arg)
	}
	cmd := icmd.Command(binary, envArgs...)
	return &cmd
}

func assertCustomResult(t *testing.T, result icmd.Result, name string, args ...string) {
	f, ok := customResultDefaultFns[name]
	assert.Equal(t, ok, true, fmt.Sprintf("function %s not found", name))
	assert.Assert(t, f(result, args...))
}

func fromAPITestCommand(cmd TestCommand) TestCaseCMD {
	return TestCaseCMD{
		name: cmd.Name,
		icmd: *buildCmd(cmd.Command.Binary, cmd.Command.Args...),
		expected: icmd.Expected{
			ExitCode: cmd.Expected.ExitCode,
			Timeout:  cmd.Expected.Timeout,
			Error:    cmd.Expected.Error,
			Out:      cmd.Expected.Out,
			Err:      cmd.Expected.Err,
		},
		goldenFile:       cmd.GoldenFile,
		customResult:     cmd.CustomResult.Type,
		customResultArgs: cmd.CustomResult.Args,
		setupCmd:         buildCmdArrFromCommand(cmd.Setup),
		onExit:           buildCmdArrFromCommand(cmd.OnExit),
	}
}

func buildCmdArrFromCommand(cmd *[]Command) *[]icmd.Cmd {
	if cmd == nil {
		return nil
	}
	cmds := make([]icmd.Cmd, 0)
	for _, cmd := range *cmd {
		cmds = append(cmds, *buildCmd(cmd.Binary, cmd.Args...))
	}
	return &cmds
}
