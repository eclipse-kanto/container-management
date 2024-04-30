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
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/util"
	. "github.com/eclipse-kanto/container-management/integration/framework/cli"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"
)

//go:embed testdata
var testdataFS embed.FS

type cliTestConfiguration struct {
	KantoHost       string `env:"KANTO_HOST" envDefault:"/run/container-management/container-management.sock"`
	ContainerConfig string `env:"CONTAINER_CONFIG" envDefault:"./testdata/container.json"`
}

func init() {
	AddCustomResult("ASSERT_JSON_CONTAINER", assertJSONContainer)
}

func TestCtrMgrCLI(t *testing.T) {
	cliTestConfiguration := &cliTestConfiguration{}
	require.NoError(t, env.Parse(cliTestConfiguration, env.Options{RequiredIfNoDef: true}))
	require.NoError(t, os.Setenv("KANTO_HOST", cliTestConfiguration.KantoHost))
	require.NoError(t, os.Setenv("CONTAINER_CONFIG", cliTestConfiguration.ContainerConfig))

	if exist, _ := util.IsDirectory(TestData); !exist {
		require.NoError(t, dumpTestdata())
		defer os.RemoveAll(TestData)
	}

	testCases, err := GetAllTestCasesFromTestdataDir()
	testutil.AssertNil(t, err)
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			RunCmdTestCases(t, tc)
		})
	}
}

func dumpTestdata() error {
	entries, err := testdataFS.ReadDir(TestData)
	if err != nil {
		return err
	}
	if err = util.MkDir(TestData); err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err = util.MkDir(entry.Name()); err != nil {
				return err
			}
		} else {
			bytes, err := fs.ReadFile(testdataFS, filepath.Join(TestData, entry.Name()))
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(TestData, entry.Name()), bytes, 0711); err != nil {
				return err
			}
		}
	}
	return nil
}

func assertJSONContainer(result icmd.Result, args ...string) assert.BoolOrComparison {
	output := result.Stdout()
	if output == "" {
		return errors.New("stdout result is empty")
	}
	var container *types.Container
	if err := json.Unmarshal([]byte(output), &container); err != nil {
		return err
	}

	if len(args)%2 != 0 {
		return errors.New("there should be even number of arguments")
	}
	for i := 0; i < len(args); i = i + 2 {
		var (
			value     interface{}
			err       error
			byteArray []byte
		)
		if value, err = getValueFromStruct(args[i], container); err != nil {
			return err
		}
		if byteArray, err = json.Marshal(value); err != nil {
			return err
		}
		if string(byteArray) != args[i+1] {
			return false
		}
	}
	return true
}

func getValueFromStruct(keyWithDots string, object interface{}) (interface{}, error) {
	keySlice := strings.Split(keyWithDots, ".")
	v := reflect.ValueOf(object)
	for _, key := range keySlice[1:] {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("only accepts structs; got %T", v)
		}
		v = v.FieldByName(key)
	}
	return v.Interface(), nil
}
