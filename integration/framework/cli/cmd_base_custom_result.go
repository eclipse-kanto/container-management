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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"
)

func regex(result icmd.Result, args ...string) assert.BoolOrComparison {
	r, err := regexp.Compile(args[0])
	if err != nil {
		return err
	}
	if result.Stdout() != "" {
		return checkRegex(r, result.Stdout())
	}
	if result.Stderr() != "" {
		return checkRegex(r, result.Stderr())
	}
	return errors.New("empty stdout and stderr")
}

func checkRegex(r *regexp.Regexp, s string) assert.BoolOrComparison {
	if !r.MatchString(s) {
		return fmt.Errorf("%s does not match regex", s)
	}
	return true
}

func logs(result icmd.Result, args ...string) assert.BoolOrComparison {
	var logEntriesCount int
	if result.Stdout() != "" {
		lines := strings.Split(result.Stdout(), "\n")
		for _, l := range lines {
			if len(l) == 0 {
				continue
			}
			var x map[string]interface{}
			if err := json.Unmarshal([]byte(l), &x); err != nil {
				return err
			}
			logEntriesCount++
		}
	}

	if len(args) > 0 {
		expectedLogEntries, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if logEntriesCount != expectedLogEntries {
			return log.NewErrorf("unexpected number of log entries: %d", logEntriesCount)
		}
	}
	return nil
}
