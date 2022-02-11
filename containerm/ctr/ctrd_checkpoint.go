// Copyright The PouchContainer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package name changed also removed not needed logic and added custom code to handle the specific use case, Bosch.IO GmbH, 2020

package ctr

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/containerd"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/pkg/errors"
)

func createCheckpointDescriptor(ctx context.Context, checkpointDir string, client *containerd.Client) (*containerdtypes.Descriptor, error) {
	if checkpointDir == "" {
		return nil, nil
	}

	// create a checkpoint blob
	tar := archive.Diff(ctx, "", checkpointDir)
	checkpoint, err := writeContent(ctx, images.MediaTypeContainerd1Checkpoint, checkpointDir, tar, client)
	if err := tar.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close checkpoint tar stream")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upload checkpoint to containerd")
	}

	return checkpoint, nil
}

func writeContent(ctx context.Context, mediaType, ref string, r io.Reader, client *containerd.Client) (*containerdtypes.Descriptor, error) {
	writer, err := client.ContentStore().Writer(ctx, content.WithRef(ref))
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	size, err := io.Copy(writer, r)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{
		"containerd.io/gc.root": time.Now().UTC().Format(time.RFC3339),
	}
	if err := writer.Commit(ctx, 0, "", content.WithLabels(labels)); err != nil {
		return nil, err
	}
	return &containerdtypes.Descriptor{
		MediaType: mediaType,
		Digest:    writer.Digest(),
		Size_:     size,
	}, nil

}
func withCheckpointOpt(checkpoint *containerdtypes.Descriptor) containerd.NewTaskOpts {
	return func(_ context.Context, _ *containerd.Client, t *containerd.TaskInfo) error {
		t.Checkpoint = checkpoint
		return nil
	}
}

// getCheckpointDir verifies checkpoint directory for create,remove, list options and checks if checkpoint already exists
func getCheckpointDir(checkDir, checkpointID, ctrName, ctrID, ctrCheckpointDir string, create bool) (string, error) {
	var checkpointDir string
	var err2 error
	if checkDir != "" {
		checkpointDir = checkDir
	} else {
		checkpointDir = ctrCheckpointDir
	}
	checkpointAbsDir := filepath.Join(checkpointDir, checkpointID)
	stat, err := os.Stat(checkpointAbsDir)
	if create {
		switch {
		case err == nil && stat.IsDir():
			err2 = fmt.Errorf("checkpoint with name %s already exists for container %s", checkpointID, ctrName)
		case err != nil && os.IsNotExist(err):
			err2 = os.MkdirAll(checkpointAbsDir, 0700)
		case err != nil:
			err2 = err
		case err == nil:
			err2 = fmt.Errorf("%s exists and is not a directory", checkpointAbsDir)
		default:
			// should never get here
		}
	} else {
		switch {
		case err != nil:
			err2 = fmt.Errorf("checkpoint %s does not exist for container %s", checkpointID, ctrName)
		case err == nil && stat.IsDir():
			err2 = nil
		case err == nil:
			err2 = fmt.Errorf("%s exists and is not a directory", checkpointAbsDir)
		default:
			// should never get here
		}
	}
	return checkpointAbsDir, err2
}
