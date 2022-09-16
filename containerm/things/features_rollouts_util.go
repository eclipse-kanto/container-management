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
	"bytes"
	cryptoMd5 "crypto/md5"
	cryptoSha1 "crypto/sha1"
	cryptoSha256 "crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
)

func convertToUpdateAction(args interface{}) (datatypes.UpdateAction, error) {
	bytes, err := json.Marshal(args)
	if err != nil {
		return datatypes.UpdateAction{}, err
	}
	var uas datatypes.UpdateAction
	err = json.Unmarshal(bytes, &uas)
	return uas, err
}

func convertToRemoveAction(args interface{}) (datatypes.RemoveAction, error) {
	bytes, err := json.Marshal(args)
	if err != nil {
		return datatypes.RemoveAction{}, err
	}
	var uas datatypes.RemoveAction
	err = json.Unmarshal(bytes, &uas)
	return uas, err
}

func createContainer(saa *datatypes.SoftwareArtifactAction) (*types.Container, bool, error) {
	downloadURL := saa.Download[datatypes.HTTP]

	if downloadURL == nil {
		downloadURL = saa.Download[datatypes.HTTPS]
	}

	ctrDescriptionBytes, err := downloadContainerDescription(downloadURL.URL)
	if err != nil {
		return nil, false, err
	}

	if err := validateSoftareArtifactHash(ctrDescriptionBytes, saa.Checksums); err != nil {
		// status should be FinishedRejected
		return nil, true, err
	}

	ctr := &types.Container{}
	if err := json.Unmarshal(ctrDescriptionBytes, ctr); err != nil {
		return nil, false, err
	}

	// fill in the defaults in order to be able to make an early validation
	util.FillDefaults(ctr)

	// perform an early validation
	if err := util.ValidateContainer(ctr); err != nil {
		log.ErrorErr(err, "configuration for container id = %s is invalid", ctr.ID)
		return nil, true, err
	}

	log.Debug("created container to install from Things : [%s]", ctr)
	return ctr, false, nil
}

func downloadContainerDescription(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)

}

func validateSoftareArtifactHash(value []byte, hashes map[datatypes.Hash]string) error {
	if hashes[datatypes.MD5] != "" {
		return validateHashMd5(value, hashes[datatypes.MD5])
	} else if hashes[datatypes.SHA1] != "" {
		return validateHashSha1(value, hashes[datatypes.SHA1])
	} else if hashes[datatypes.SHA256] != "" {
		return validateHashSha256(value, hashes[datatypes.SHA256])
	} else {
		log.Warn("no hash information is provided to veryfiy the downloaded artifact")
		return nil
	}
}

func validateHashMd5(value []byte, md5Hash string) error {
	md5HashBytes, err := convertStringHashToBytes16(md5Hash)
	if err != nil {
		return err
	}

	md5 := cryptoMd5.Sum(value)
	if md5 != md5HashBytes {
		return log.NewError("md5 checksum does not match")
	}
	return nil
}

func validateHashSha1(value []byte, sha1Hash string) error {
	sha1HashBytes, err := convertStringHashToBytes20(sha1Hash)
	if err != nil {
		return err
	}
	sha1 := cryptoSha1.Sum(value)
	if sha1 != sha1HashBytes {
		return log.NewError("sha1 checksum does not match")
	}
	return nil
}

func validateHashSha256(value []byte, sha256Hash string) error {
	sha256HashBytes, err := convertStringHashToBytes32(sha256Hash)
	if err != nil {
		return err
	}
	sha256 := cryptoSha256.Sum256(value)
	if sha256 != sha256HashBytes {
		return log.NewError("sha256 checksum does not match")
	}
	return nil
}

func convertStringHashToBytes16(checkSum string) ([16]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [16]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 16 bytes")
	}
	return dst, nil
}

func convertStringHashToBytes20(checkSum string) ([20]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [20]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 20 bytes")
	}
	return dst, nil
}

func convertStringHashToBytes32(checkSum string) ([32]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [32]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 32 bytes")
	}
	return dst, nil
}

func validateSoftwareUpdateAction(updateAction datatypes.UpdateAction) error {
	if len(updateAction.SoftwareModules) == 0 {
		return log.NewError("there are no SoftwareModules to be installed")
	}
	for _, softMod := range updateAction.SoftwareModules {
		if len(softMod.Artifacts) == 0 {
			return log.NewErrorf("there are no SoftwareArtifacts referenced for SoftwareModule [Name.version] = [%s.%s]", softMod.SoftwareModule.Name, softMod.SoftwareModule.Version)
		}
	}

	return nil
}

func generateDependencyDescriptionKey(depDescr *datatypes.DependencyDescription) string {
	unencodedKey := fmt.Sprintf(featureSoftwareUpdatableInstalledDependenciesKeyTemplate, depDescr.Group, depDescr.Name, depDescr.Version)
	return strings.ReplaceAll(unencodedKey, installedDependenciesKeysSlash, installedDependenciesKeysSlashEncoding)
}

func dependencyDescription(ctr *types.Container) *datatypes.DependencyDescription {
	elements := strings.Split(ctr.Image.Name, ":")
	return &datatypes.DependencyDescription{
		Group:   elements[0],
		Name:    ctr.ID,
		Version: strings.Join(elements[1:], ":"),
	}
}
