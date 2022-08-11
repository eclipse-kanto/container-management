// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package matchers

import (
	"github.com/golang/mock/gomock"
	"strings"
)

type containsStringMatcher struct{ sub string }

// ContainsString returns a Matcher interface for checking whether the actual argument contains the provided string
func ContainsString(s string) gomock.Matcher {
	return &containsStringMatcher{sub: s}
}

func (o *containsStringMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case string:
		return strings.Contains(x.(string), o.sub)
	case []byte:
		return strings.Contains(string(x.([]byte)), o.sub)
	default:
		return false
	}
}

func (o *containsStringMatcher) String() string {
	return "no such substring: " + o.sub
}
