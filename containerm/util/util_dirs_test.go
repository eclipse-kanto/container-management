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

package util

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	dirName       = "testdir"
	dirNameSecond = "testdir2"
	fileName      = "testfile"
)

func TestMkDir(t *testing.T) {
	err := MkDir(dirName)
	defer os.RemoveAll(dirName)
	if err != nil {
		t.Log(err)
		t.Error("couldn't create dir")
	} else {
		fileInfo, err := os.Stat(dirName)
		if fileInfo == nil || err != nil {
			t.Error("error during file stat")
		} else {
			if !fileInfo.IsDir() {
				t.Error("created file not dir")
			}
			if fileInfo.Mode().Perm() != 0711 {
				t.Errorf("dir file permissions expected %v, but were %v", 0711, fileInfo.Mode().Perm())
			}
		}
	}
}

func TestMkDirAlreadyExists(t *testing.T) {
	err := MkDir(dirName)
	defer os.RemoveAll(dirName)
	if err != nil {
		t.Log(err)
		t.Error("couldn't create dir")
	} else {
		fileInfo, err := os.Stat(dirName)
		if fileInfo == nil || err != nil {
			t.Error("error during file stat")
		} else {
			if fileInfo.Mode().Perm() != 0711 {
				t.Errorf("dir file permissions expected %v, but were %v", 0711, fileInfo.Mode().Perm())
			}
		}
	}
}

func TestMkDirAlreadyExistsAsNonDir(t *testing.T) {
	// create a non-directory file with the same name as the MkDir() directory argument
	_, err := os.Create(fileName)
	defer os.RemoveAll(fileName)

	if err != nil {
		t.Error("could create test file")
	}

	err = MkDir(fileName)
	if err == nil {
		t.Error("error expected when file exists and is not dir")
	}
}

func TestMkDirs(t *testing.T) {
	err := MkDirs(dirName, dirNameSecond)
	defer func() {
		os.RemoveAll(dirName)
		os.RemoveAll(dirNameSecond)
	}()
	if err != nil {
		t.Log(err)
		t.Error("couldn't create dir")
	} else {
		dir, err := os.Stat(dirName)
		dirSecond, err := os.Stat(dirNameSecond)
		if err != nil {
			t.Error("error during stat")
		} else {
			if dir == nil || dirSecond == nil {
				t.Error("multiple dirs not created")
			}
		}
	}
}

func TestFileExistsNotEmptyOrDir(t *testing.T) {
	_, err := os.Create(fileName)
	defer os.RemoveAll(fileName)
	if err != nil {
		t.Error("could create test file")
	}

	result := FileNotExistEmptyOrDir(fileName)
	testutil.AssertError(t, log.NewErrorf("file %s is empty", fileName), result)

}

func TestFileExistsNotEmptyOrDirEmptyEmpty(t *testing.T) {
	_, err := os.Create(fileName)
	defer os.RemoveAll(fileName)

	if err != nil {
		t.Error("could create test file")
	}

	result := FileNotExistEmptyOrDir(fileName)
	testutil.AssertError(t, log.NewErrorf("file %s is empty", fileName), result)
}

func TestFileExistsNotEmptyOrDirEmptyDir(t *testing.T) {
	err := os.MkdirAll(fileName, 0711)
	defer os.RemoveAll(fileName)

	if err != nil {
		t.Error("could create test file")
	}

	result := FileNotExistEmptyOrDir(fileName)
	testutil.AssertError(t, log.NewErrorf("the provided path %s is a dir path - file is required", fileName), result)
}

func TestIsDirectoryTrue(t *testing.T) {
	path := "../pkg/testutil"

	result, err := IsDirectory(path)

	if err != nil {
		t.Errorf("Error for IsDeirectory is not nil")
		t.Fail()
	}
	testutil.AssertTrue(t, result)
}

func TestIsDirectoryFalse(t *testing.T) {
	path := "../pkg/testutil/assertions.go"

	result, err := IsDirectory(path)

	if err != nil {
		t.Errorf("Error for IsDeirectory is not nil")
		t.Fail()
	}
	testutil.AssertFalse(t, result)
}

func TestIsDirectoryOnNotExisting(t *testing.T) {
	path := "../pkg/testutils"

	result, err := IsDirectory(path)

	if err == nil {
		t.Errorf("Error for IsDeirectory is nil")
		t.Fail()
	}
	testutil.AssertFalse(t, result)
}

func TestGetChildrenOfEmpty(t *testing.T) {
	path := "../pkg/testutil/metapath/empty"

	MkDir(path)
	defer os.Remove(path)
	RemoveChildren(path)
	expected := make([]string, 0)
	result, err := GetDirChildrenNames(path)

	if err != nil {
		t.Errorf("Error for GetChildrenNames is not nil %v", err)
		t.Fail()
	}
	testutil.AssertEqual(t, result, expected)
}

func TestGetChildren(t *testing.T) {
	path := "../pkg/testutil/metapath"
	emptyPath := "../pkg/testutil/metapath/empty"
	tmpPath := "../pkg/testutil/metapath/tmp"

	MkDir(emptyPath)
	defer os.Remove(emptyPath)
	MkDir("../pkg/testutil/metapath/invalid") // This not be needed when we merge with mgr_tests branch
	MkDir(tmpPath)
	defer os.Remove(tmpPath)
	expected := []string{"empty", "invalid", "tmp", "valid"}

	result, err := GetDirChildrenNames(path)

	if err != nil {
		t.Errorf("Error for GetChildrenNames is not nil %v", err)
		t.Fail()
	}

	sort.Strings(result)
	sort.Strings(expected)

	testutil.AssertEqual(t, expected, result)
}

func TestDeleteChildren(t *testing.T) {
	path := "baseDir"
	dOne := filepath.Join(path, dirName)
	dTwo := filepath.Join(path, dirNameSecond)

	MkDir(path)
	defer os.Remove(path)
	MkDir(dOne)
	MkDir(dTwo)

	childrenBefore, _ := GetDirChildrenNames(path)
	testutil.AssertEqual(t, 2, len(childrenBefore))

	removedChildren, err := RemoveChildren(path)
	if err != nil {
		t.Errorf("Error for remove children is not nil")
		t.Fail()
	}
	testutil.AssertEqual(t, 0, len(removedChildren))

	childrenAfter, _ := GetDirChildrenNames(path)
	testutil.AssertEqual(t, 0, len(childrenAfter))
}
