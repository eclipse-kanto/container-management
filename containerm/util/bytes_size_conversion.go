// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package util

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	kb = 1024
	mb = 1024 * kb
	gb = 1024 * mb

	symbolKilo = "k"
	symbolMega = "m"
	symbolGiga = "g"
)

var (
	units             = map[string]int64{symbolKilo: kb, symbolMega: mb, symbolGiga: gb}
	sizeAsStringRegex = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgG])$`)
)

// SizeToBytes converts size string representation to a number
func SizeToBytes(sizeStr string) (int64, error) {
	size, unit, err := parseSize(sizeStr)
	if err != nil {
		return -1, err
	}

	if multiplier, ok := units[strings.ToLower(unit)]; ok {
		size *= float64(multiplier)
	}

	return int64(size), nil
}

// SizeRecalculate takes size string representation and returns a new recalculated size string representation
func SizeRecalculate(sizeStr string, eval func(float64) float64) (string, error) {
	size, unit, err := parseSize(sizeStr)
	if err != nil {
		return "", err
	}

	size = math.Round(eval(size)*10000) / 10000
	return strconv.FormatFloat(size, 'f', -1, 64) + unit, nil
}

func parseSize(sizeStr string) (size float64, unit string, err error) {
	groupMatches := sizeAsStringRegex.FindStringSubmatch(sizeStr)
	if len(groupMatches) == 4 {
		unit = groupMatches[3]
		size, err = strconv.ParseFloat(groupMatches[1], 64)
	} else {
		err = log.NewErrorf("invalid size provided %s", sizeStr)
	}
	return
}
