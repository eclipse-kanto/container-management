package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// WriteAtomicFile writes data to a temporary file, then renames it into filename. If the
// target filename already exists but is not a regular file, WriteAtomicFile will return an error.
func WriteAtomicFile(filename string, data []byte, perm os.FileMode) (err error) {
	fi, err := os.Stat(filename)
	if err == nil && !fi.Mode().IsRegular() {
		log.WarnErr(err, "file already exists and is not a regular file.")
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}

// MakeAtomicCopy makes a copy of an existing file given a source and destionation file.
func MakeAtomicCopy(sourcepath, destinationpath string) (err error) {
	_, err = os.Stat(sourcepath)
	if os.IsNotExist(err) {
		log.Debug("configuration file does not exist to create a backup", err)
		return err
	}
	if err != nil {
		log.WarnErr(err, "error reading from previous configuration")
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(destinationpath), "tmp_")
	if err != nil {
		log.WarnErr(err, "error creating temporary file")
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	src, err := os.Open(sourcepath)
	if err != nil {
		log.WarnErr(err, "error opening current state configuration file")
		return err
	}
	defer src.Close()

	_, err = io.Copy(tmpFile, src)
	if err != nil {
		log.WarnErr(err, "error copying the contents to the temporary file")
		return err
	}

	tmpFile.Close()

	err = os.Rename(tmpName, destinationpath)
	if err != nil {
		log.WarnErr(err, "error renaming file")
		return err
	}

	fmt.Printf("File copied successfully from %s to %s\n", sourcepath, destinationpath)
	return nil
}
