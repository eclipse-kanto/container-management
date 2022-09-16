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

package jsonfile

import (
	"os"
	"strconv"
)

func rotate(logFileName string, maxFiles int) error {
	if maxFiles < 2 {
		return nil
	}
	for i := maxFiles - 1; i > 1; i-- {
		newFileName := logFileName + "." + strconv.Itoa(i)
		oldFileName := logFileName + "." + strconv.Itoa(i-1)
		if err := os.Rename(oldFileName, newFileName); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	if err := os.Rename(logFileName, logFileName+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
