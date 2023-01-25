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
	"regexp"

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
