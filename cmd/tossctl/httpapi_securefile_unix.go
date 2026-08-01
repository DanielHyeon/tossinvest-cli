//go:build unix

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func openHTTPAPIRegularNoFollow(path string) (*os.File, error) {
	return openHTTPAPIRegularNoFollowForUID(path, uint32(os.Geteuid()))
}

// openHTTPAPIRegularNoFollowForUID resolves a trust anchor from a directory FD
// instead of re-resolving a pathname after validation. Every directory
// component must be a real directory owned by root or the service identity. The
// immediate parent is additionally required to reject group/other writes, so a
// second local account cannot replace an otherwise read-only public key.
func openHTTPAPIRegularNoFollowForUID(path string, serviceUID uint32) (*os.File, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("securely resolving file path: %w", err)
	}
	components := strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator))
	if len(components) < 1 || components[len(components)-1] == "" {
		return nil, errors.New("securely resolving file path: regular file name is required")
	}

	directoryFD, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, fmt.Errorf("securely opening filesystem root: %w", err)
	}
	defer func() { _ = unix.Close(directoryFD) }()

	var rootStat unix.Stat_t
	if err := unix.Fstat(directoryFD, &rootStat); err != nil {
		return nil, fmt.Errorf("securely inspecting filesystem root: %w", err)
	}
	rootUID := rootStat.Uid
	if err := validateHTTPAPIPathDirectory(&rootStat, serviceUID, rootUID, len(components) == 1); err != nil {
		return nil, err
	}

	for index, component := range components[:len(components)-1] {
		if component == "" || component == "." || component == ".." {
			return nil, errors.New("securely resolving file path: ambiguous directory component")
		}
		nextFD, openErr := unix.Openat(directoryFD, component,
			unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return nil, fmt.Errorf("securely opening directory component %q: %w", component, openErr)
		}
		var stat unix.Stat_t
		if statErr := unix.Fstat(nextFD, &stat); statErr != nil {
			_ = unix.Close(nextFD)
			return nil, fmt.Errorf("securely inspecting directory component %q: %w", component, statErr)
		}
		immediateParent := index == len(components)-2
		if validationErr := validateHTTPAPIPathDirectory(&stat, serviceUID, rootUID, immediateParent); validationErr != nil {
			_ = unix.Close(nextFD)
			return nil, fmt.Errorf("securely validating directory component %q: %w", component, validationErr)
		}
		_ = unix.Close(directoryFD)
		directoryFD = nextFD
	}

	basename := components[len(components)-1]
	fd, err := unix.Openat(directoryFD, basename,
		unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("securely opening file: %w", err)
	}
	var fileStat unix.Stat_t
	if err := unix.Fstat(fd, &fileStat); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("securely inspecting file: %w", err)
	}
	if err := validateHTTPAPITrustAnchor(&fileStat, serviceUID, rootUID); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	file := os.NewFile(uintptr(fd), abs)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("securely opening file: invalid descriptor")
	}
	return file, nil
}

func validateHTTPAPIPathDirectory(stat *unix.Stat_t, serviceUID, rootUID uint32, immediateParent bool) error {
	if stat == nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return errors.New("securely validating directory: path component is not a directory")
	}
	if stat.Uid != rootUID && stat.Uid != serviceUID {
		return errors.New("securely validating directory: path component is owned by another account")
	}
	writableByOthers := stat.Mode&0o022 != 0
	if immediateParent && writableByOthers {
		return errors.New("securely validating directory: trust-anchor parent is group/other writable")
	}
	// A root-owned sticky ancestor such as /tmp cannot have another account's
	// entries renamed or removed. No other writable ancestor is safe to traverse.
	if !immediateParent && writableByOthers && (stat.Uid != rootUID || stat.Mode&unix.S_ISVTX == 0) {
		return errors.New("securely validating directory: writable ancestor is not root-owned and sticky")
	}
	return nil
}

func validateHTTPAPITrustAnchor(stat *unix.Stat_t, serviceUID, rootUID uint32) error {
	if stat == nil || stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return errors.New("securely validating file: trust anchor is not a regular file")
	}
	if stat.Uid != rootUID && stat.Uid != serviceUID {
		return errors.New("securely validating file: trust anchor is owned by another account")
	}
	if stat.Mode&0o022 != 0 {
		return errors.New("securely validating file: trust anchor is group/other writable")
	}
	return nil
}
