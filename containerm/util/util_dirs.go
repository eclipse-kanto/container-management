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
	"io"
	"os"
	"path/filepath"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// MkDir creates a directory with the provided name
func MkDir(dirname string) error {
	if fileInfo, err := os.Stat(dirname); err == nil {
		// root current exists; verify the access bits are correct by setting them
		if fileInfo != nil {
			if !fileInfo.IsDir() {
				return log.NewErrorf("non-directory file already exists with name: %s", dirname)
			}
		}
		if err = os.Chmod(dirname, 0711); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		// no root exists yet, create it 0711 with root:root ownership
		if err := os.MkdirAll(dirname, 0711); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// MkDirs creates directories for the given directory names
func MkDirs(dirNames ...string) error {
	if len(dirNames) == 0 {
		return nil
	}
	for _, dirName := range dirNames {
		if err := MkDir(dirName); err != nil {
			return err
		}
	}
	return nil
}

// FileNotExistEmptyOrDir validates whether the provided filename refers to an existing non-empty file
func FileNotExistEmptyOrDir(filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return log.NewErrorf("the provided path %s is a dir path - file is required", filename)
	}
	if fi.Size() == 0 {
		return log.NewErrorf("file %s is empty", filename)
	}
	return nil
}

// RemoveChildren removes all children of a given directory
//
// Returns [], nil if everything went OK
// Returns nil, error if the error happened before or during the retrieval of the children. For
// example the path did not exist was locked or something else similar.
// If an error occurs on some or all of the children it will return a []string, error
// where the list contains the children that where not deleted
func RemoveChildren(dirPath string) ([]string, error) {
	if check, err := IsDirectory(dirPath); check == false || err != nil {
		//TODO: wrap error
		return nil, log.NewErrorf("the provided path: %s is a file or does not exist", dirPath)
	}

	children, childrenError := GetDirChildrenNames(dirPath)

	if childrenError != nil {
		//TODO: wrap error
		return nil, log.NewErrorf("could not remove children for %s", dirPath)
	}

	var notDeletedChildren []string

	for _, fName := range children {
		pathToDelete := filepath.Join(dirPath, fName)

		if err := os.RemoveAll(pathToDelete); err != nil {
			log.ErrorErr(err, "Failed to delete %s", pathToDelete)
			notDeletedChildren = append(notDeletedChildren, pathToDelete)
		}
	}

	if len(notDeletedChildren) > 0 {
		return notDeletedChildren, log.NewErrorf("some subfolders of %s where not deleted", dirPath)
	}

	// DO NOT RETURN nil, because it breaks agains interface {}
	// https://groups.google.com/g/golang-nuts/c/wnH302gBa4I/discussion
	return notDeletedChildren, nil
}

// GetDirChildrenNames returns all child directories of the provided one
func GetDirChildrenNames(dirPath string) ([]string, error) {
	if check, err := IsDirectory(dirPath); check == false || err != nil {
		// TODO: wrap error
		return nil, log.NewErrorf("the provided path: %s is a file or does not exista", dirPath)
	}

	dir, dirErr := os.Open(dirPath)

	if dirErr != nil {
		//TODO: wrap error
		return nil, log.NewErrorf("could not open filepath: %s", dirPath)
	}

	children, childrenErr := dir.Readdirnames(-1) // -1 tells the function to return all

	if childrenErr != nil {
		// TODO: wrap error
		return nil, log.NewErrorf("could not get children for : %s", dirPath)
	}

	if closeErr := dir.Close(); closeErr != nil {
		// TODO: wrap error
		return children, log.NewErrorf("Could not close dir: %s", dirPath)
	}

	// By here there are no errors, but we want to return empty
	if children == nil {
		return make([]string, 0), nil
	}

	return children, nil
}

// Copy copies the resources from a given source path to the given destination path
func Copy(sourcePath string, destPath string, bufferSize int) error {
	// Check source
	isValid, fileValidity := IsFile(sourcePath)

	if fileValidity != nil {
		return fileValidity
	}

	if !isValid {
		return log.NewErrorf("%+v is not a valid file", isValid)
	}

	// Check destination
	isDestValid, _ := IsFile(destPath)
	if isDestValid == true {
		return log.NewErrorf("%s file already exists", destPath)
	}

	// Open source
	source, srcOpenErr := os.Open(sourcePath)
	if srcOpenErr != nil {
		//TODO: wrap error
		return log.NewErrorf("could not open file: %s", sourcePath)
	}
	defer source.Close()

	destination, destError := os.Create(destPath)
	if destError != nil {
		//TODO: wrap error
		return log.NewErrorf("could not create file: %s", destPath)
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)

	for {
		n, err := source.Read(buf)

		if err != nil && err != io.EOF {
			//TODO: wrap error
			return log.NewErrorf("an error occured while copying the file from %s to %s",
				sourcePath, destPath)
		}

		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			// TODO: wrap error
			return log.NewErrorf("an error occured while copying the file from %s to %s",
				sourcePath, destPath)
		}
	}

	return nil
}

// IsFile returns true if the given path is a file
func IsFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)

	if err != nil {
		// TODO: wrap error
		return false, log.NewErrorf("An error occured trying to read %s", path)
	}

	return fileInfo.IsDir() == false && fileInfo.Mode().IsRegular(), nil
}

// IsDirectory returns true if the given path is a directory
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)

	if err != nil {
		// TODOl wrap error
		return false, log.NewErrorf("An error occured trying to read %s", path)
	}

	return fileInfo.IsDir(), nil
}
