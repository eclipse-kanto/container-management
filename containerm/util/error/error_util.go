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

package error

import (
	"fmt"
	"strings"
)

// CompoundError contains a slice of errors.
type CompoundError struct {
	errs []error
}

// Append adds the errors into the list.
func (m *CompoundError) Append(errs ...error) {
	m.errs = append(m.errs, errs...)
}

// Size returns the count of list of errors.
func (m *CompoundError) Size() int {
	return len(m.errs)
}

// Error returns the combined error messages.
func (m *CompoundError) Error() string {
	if len(m.errs) == 0 {
		return fmt.Sprintf("no error")
	}

	if len(m.errs) == 1 {
		return fmt.Sprintf("%s", m.errs[0])
	}

	serrs := make([]string, len(m.errs))
	for i, err := range m.errs {
		serrs[i] = fmt.Sprintf("* %s", err)
	}
	return fmt.Sprintf("%d errors:\n\n%s", len(m.errs), strings.Join(serrs, "\n"))
}
