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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
)

const (
	testStringToHash     = "ateststring"
	testStringUnmatching = "notmatchingsting"
	//hashed
	testHashedMD5    = "c2572289c78add0e3192262cfd6b85ef"
	testHashedSHA1   = "0c959e814f2d673c46b5d6db5b91f490023738a9"
	testHashedSHA256 = "be1b3ce3b8ceb307b81b515608ed0439f6959089850c31788a008f8c066849f4"
)

var (
	// MD5
	testMapMd5 = map[datatypes.Hash]string{
		datatypes.MD5: testHashedMD5,
	}
	testMapMd5NotMatching = map[datatypes.Hash]string{
		datatypes.MD5: testHashedMD5,
	}
	testMapMd5Invalid = map[datatypes.Hash]string{
		datatypes.MD5: "invalid",
	}

	// SHA1
	testMapSha1 = map[datatypes.Hash]string{
		datatypes.SHA1: testHashedSHA1,
	}
	testMapSha1NotMatching = map[datatypes.Hash]string{
		datatypes.SHA1: testHashedSHA1,
	}
	testMapSha1Invalid = map[datatypes.Hash]string{
		datatypes.SHA1: "invalid",
	}

	// SHA256
	testMapSha256 = map[datatypes.Hash]string{
		datatypes.SHA256: testHashedSHA256,
	}
	testMapSha256NotMatching = map[datatypes.Hash]string{
		datatypes.SHA256: testHashedSHA256,
	}
	testMapSha256Invalid = map[datatypes.Hash]string{
		datatypes.SHA256: "invalid",
	}

	// None
	testMapHashesNone = map[datatypes.Hash]string{}
)

func TestValdiateHash(t *testing.T) {
	tests := map[string]struct {
		value       []byte
		hashes      map[datatypes.Hash]string
		expectedErr error
	}{
		"test_validate_hash_md5": {
			value:       []byte(testStringToHash),
			hashes:      testMapMd5,
			expectedErr: nil,
		},
		"test_validate_hash_md5_not_matching": {
			value:       []byte(testStringUnmatching),
			hashes:      testMapMd5NotMatching,
			expectedErr: log.NewError("md5 checksum does not match"),
		},
		"test_validate_hash_md5_invalid": {
			value:       []byte(testStringToHash),
			hashes:      testMapMd5Invalid,
			expectedErr: log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 16 bytes"),
		},
		"test_validate_hash_sha1": {
			value:       []byte(testStringToHash),
			hashes:      testMapSha1,
			expectedErr: nil,
		},
		"test_validate_hash_sha1_not_matching": {
			value:       []byte(testStringUnmatching),
			hashes:      testMapSha1NotMatching,
			expectedErr: log.NewError("sha1 checksum does not match"),
		},
		"test_validate_hash_sha1_invalid": {
			value:       []byte(testStringToHash),
			hashes:      testMapSha1Invalid,
			expectedErr: log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 20 bytes"),
		},
		"test_validate_hash_sha256": {
			value:       []byte(testStringToHash),
			hashes:      testMapSha256,
			expectedErr: nil,
		},
		"test_validate_hash_sha256_not_matching": {
			value:       []byte(testStringUnmatching),
			hashes:      testMapSha256NotMatching,
			expectedErr: log.NewError("sha256 checksum does not match"),
		},
		"test_validate_hash_sha256_invalid": {
			value:       []byte(testStringToHash),
			hashes:      testMapSha256Invalid,
			expectedErr: log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 32 bytes"),
		},
		"test_validate_hash_none": {
			value:       []byte(testStringToHash),
			hashes:      testMapHashesNone,
			expectedErr: nil,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			resultErr := validateSoftareArtifactHash(testCase.value, testCase.hashes)
			testutil.AssertError(t, testCase.expectedErr, resultErr)
		})
	}
}
