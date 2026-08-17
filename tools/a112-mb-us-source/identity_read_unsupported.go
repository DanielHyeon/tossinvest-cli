//go:build !unix

package main

import (
	"errors"
	"io"
	"os"
)

func readRegularNoFollow(path string) ([]byte, error) {
	return readRegularNoFollowWithLimit(path, maxIdentityInputBytes)
}

func readRegularNoFollowWithLimit(path string, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, errors.New("identity input limit is invalid")
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() < 0 || before.Size() > limit {
		return nil, errors.New("identity input is not a regular no-follow file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if int64(len(data)) > limit {
		return nil, errors.New("identity input exceeds size limit")
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) || !after.Mode().IsRegular() || after.Mode()&os.ModeSymlink != 0 || before.Size() != after.Size() || int64(len(data)) != after.Size() {
		return nil, errors.New("identity input changed during read")
	}
	return data, nil
}
