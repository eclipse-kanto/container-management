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

// TestCommand defines all required and optional features that are needed for CLI integration test.
type TestCommand struct {
	GoldenFile   string       `yaml:"goldenFile"`
	Name         string       `yaml:"name"`
	Command      Command      `yaml:"command"`
	Expected     Expected     `yaml:"expected"`
	CustomResult CustomResult `yaml:"customResult"`
	Setup        *[]Command   `yaml:"setupCmd"`
	OnExit       *[]Command   `yaml:"onExit"`
}

// Command is contains the binary and arguments that will run in the test case.
type Command struct {
	Binary string   `yaml:"binary"`
	Args   []string `yaml:"args"`
}

// Expected is the expected output from a Command.
type Expected struct {
	ExitCode int    `yaml:"exitCode"`
	Timeout  bool   `yaml:"timeout"`
	Error    string `yaml:"error"`
	Out      string `yaml:"out"`
	Err      string `yaml:"err"`
}

// CustomResult is a representation of a "func(result icmd.Result, args ...string) assert.BoolOrComparison"
// where Name is the name of the function and Args... are optional string arguments that the function may use.
// If CustomResult function is defined the result of the icmd.Cmd that is executed will be asserted to that CustomResult function.
// CustomResult functions can be passed as an argument to the runCmdTestCases(...) or one of the default functions can be used.
//
// Default functions:
//   - REGEX - this function will compare the stdout or stderr outputs from the result with the regex provided to args[0].
//     example:
//     customResult:
//     name: REGEX
//     args: ["([A-Za-z0-9]+(-[A-Za-z0-9]+)+)"]
type CustomResult struct {
	Type string   `yaml:"type"`
	Args []string `yaml:"args"`
}
