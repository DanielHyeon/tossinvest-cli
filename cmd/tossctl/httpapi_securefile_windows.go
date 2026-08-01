//go:build windows

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func openHTTPAPIRegularNoFollow(path string) (*os.File, error) {
	return openHTTPAPIRegularNoFollowForUID(path, 0)
}

// Windows does not expose Unix uid/mode ownership. It still rejects every
// symlinked directory component and verifies that the opened file is the same
// regular file inspected before the open.
func openHTTPAPIRegularNoFollowForUID(path string, _ uint32) (*os.File, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	volume := filepath.VolumeName(abs)
	current := volume + string(filepath.Separator)
	relative := strings.TrimPrefix(abs, current)
	components := strings.Split(relative, string(filepath.Separator))
	if len(components) < 1 || components[len(components)-1] == "" {
		return nil, errors.New("securely opening file: regular file name is required")
	}
	for _, component := range components[:len(components)-1] {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, errors.New("securely opening file: path contains a non-directory or linked component")
		}
	}
	before, err := os.Lstat(abs)
	if err != nil || before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, errors.New("securely opening file: path must name a regular file, not a link")
	}
	file, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		_ = file.Close()
		return nil, errors.New("securely opening file: path changed while opening")
	}
	return file, nil
}
