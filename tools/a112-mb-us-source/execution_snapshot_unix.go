//go:build unix

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/sys/unix"
)

const executionSnapshotName = "tool"

// snapshotExecutionBinary opens the caller-selected source exactly once with
// O_NOFOLLOW, copies those opened bytes into a fresh owner-only temporary
// capability, and never needs to resolve sourcePath again.
func snapshotExecutionBinary(sourcePath string) (_ *executionSnapshot, err error) {
	if sourcePath == "" || !filepath.IsAbs(sourcePath) || filepath.Clean(sourcePath) != sourcePath {
		return nil, errors.New("execution source must be a clean absolute path")
	}
	sourceFD, err := unix.Open(sourcePath, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(sourceFD)
	var source unix.Stat_t
	if err := unix.Fstat(sourceFD, &source); err != nil || source.Mode&unix.S_IFMT != unix.S_IFREG || source.Size < 0 || source.Size > maxIdentityInputBytes || source.Mode&0o022 != 0 {
		return nil, errors.New("execution source is not a safe no-follow regular file")
	}

	directory, err := os.MkdirTemp("/tmp", "a112-mb-us-exec-")
	if err != nil {
		return nil, err
	}
	createdDirectory := true
	defer func() {
		if createdDirectory {
			_ = os.Remove(directory)
		}
	}()
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, err
	}
	directoryFD, err := unix.Open(directory, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	keepDirectoryFD := false
	defer func() {
		if !keepDirectoryFD {
			_ = unix.Close(directoryFD)
		}
	}()
	if err := validatePrivateDirectory(directoryFD); err != nil {
		return nil, err
	}
	fileFD, err := unix.Openat(directoryFD, executionSnapshotName, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o500)
	if err != nil {
		return nil, err
	}
	keepFileFD := false
	defer func() {
		if !keepFileFD {
			_ = unix.Close(fileFD)
			_ = unix.Unlinkat(directoryFD, executionSnapshotName, 0)
		}
	}()
	if err := unix.Fchmod(fileFD, 0o500); err != nil {
		return nil, err
	}
	digest, copied, err := copyFDExact(sourceFD, fileFD, source.Size)
	if err != nil {
		return nil, err
	}
	var sourceAfter unix.Stat_t
	if err := unix.Fstat(sourceFD, &sourceAfter); err != nil || sourceAfter.Mode&unix.S_IFMT != unix.S_IFREG || sourceAfter.Dev != source.Dev || sourceAfter.Ino != source.Ino || sourceAfter.Size != source.Size || copied != source.Size {
		return nil, errors.New("execution source changed during snapshot copy")
	}
	if err := unix.Fsync(fileFD); err != nil {
		return nil, err
	}
	// Keep only a read descriptor after the copy.  Linux refuses to exec a
	// binary while any process holds it writable (ETXTBSY); this descriptor is
	// also the capability used for all later digest validation.
	readFD, err := unix.Openat(directoryFD, executionSnapshotName, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	if err := unix.Close(fileFD); err != nil {
		_ = unix.Close(readFD)
		return nil, err
	}
	fileFD = readFD
	if err := unix.Fsync(directoryFD); err != nil {
		return nil, err
	}
	snapshot := &executionSnapshot{
		sourcePath:    sourcePath,
		executionPath: filepath.Join(directory, executionSnapshotName),
		directory:     directory,
		name:          executionSnapshotName,
		digest:        digest,
		directoryFD:   directoryFD,
		fileFD:        fileFD,
	}
	if err := validateExecutionSnapshot(snapshot); err != nil {
		return nil, err
	}
	keepDirectoryFD = true
	keepFileFD = true
	createdDirectory = false
	return snapshot, nil
}

func validateExecutionSnapshot(snapshot *executionSnapshot) error {
	if snapshot == nil || snapshot.directoryFD < 0 || snapshot.fileFD < 0 || snapshot.digest == "" || snapshot.name != executionSnapshotName {
		return errors.New("execution snapshot is unavailable")
	}
	if err := validatePrivateDirectory(snapshot.directoryFD); err != nil {
		return err
	}
	entries, err := privateDirectoryEntries(snapshot.directoryFD)
	if err != nil || len(entries) != 1 || entries[0] != snapshot.name {
		return fmt.Errorf("execution snapshot directory contains unexpected entries: %v", entries)
	}
	var entry unix.Stat_t
	if err := unix.Fstatat(snapshot.directoryFD, snapshot.name, &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	var opened unix.Stat_t
	if err := unix.Fstat(snapshot.fileFD, &opened); err != nil {
		return err
	}
	if !safePrivateExecutable(entry) || !safePrivateExecutable(opened) || entry.Dev != opened.Dev || entry.Ino != opened.Ino || entry.Size != opened.Size {
		return errors.New("execution snapshot file changed")
	}
	digest, size, err := digestFD(snapshot.fileFD, opened.Size)
	if err != nil || size != opened.Size || digest != snapshot.digest {
		return errors.New("execution snapshot digest changed")
	}
	var after unix.Stat_t
	if err := unix.Fstat(snapshot.fileFD, &after); err != nil || !safePrivateExecutable(after) || after.Dev != opened.Dev || after.Ino != opened.Ino || after.Size != opened.Size {
		return errors.New("execution snapshot changed during validation")
	}
	return nil
}

func closeExecutionSnapshot(snapshot *executionSnapshot) error {
	if snapshot.directoryFD < 0 || snapshot.fileFD < 0 {
		return nil
	}
	var errs []error
	if err := unix.Close(snapshot.fileFD); err != nil {
		errs = append(errs, err)
	}
	snapshot.fileFD = -1
	if err := removePrivateExecutionEntries(snapshot.directoryFD); err != nil {
		errs = append(errs, err)
	}
	if err := unix.Close(snapshot.directoryFD); err != nil {
		errs = append(errs, err)
	}
	snapshot.directoryFD = -1
	if err := os.Remove(snapshot.directory); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// removePrivateExecutionEntries removes every entry through already-opened,
// no-follow descriptors. Cleanup therefore cannot traverse an unexpected
// symlink added after validation fails.
func removePrivateExecutionEntries(directoryFD int) error {
	names, err := privateDirectoryEntries(directoryFD)
	if err != nil {
		return err
	}
	for _, name := range names {
		var stat unix.Stat_t
		if err := unix.Fstatat(directoryFD, name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if stat.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, err := unix.Openat(directoryFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			if err != nil {
				return err
			}
			removeErr := removePrivateExecutionEntries(child)
			closeErr := unix.Close(child)
			if removeErr != nil {
				return removeErr
			}
			if closeErr != nil {
				return closeErr
			}
			if err := unix.Unlinkat(directoryFD, name, unix.AT_REMOVEDIR); err != nil {
				return err
			}
			continue
		}
		if err := unix.Unlinkat(directoryFD, name, 0); err != nil {
			return err
		}
	}
	return unix.Fsync(directoryFD)
}

func validatePrivateDirectory(directoryFD int) error {
	var stat unix.Stat_t
	if err := unix.Fstat(directoryFD, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o777 != 0o700 {
		return errors.New("execution snapshot directory is not private")
	}
	return nil
}

func safePrivateExecutable(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG && stat.Uid == uint32(os.Geteuid()) && stat.Mode&0o777 == 0o500 && stat.Size >= 0 && stat.Size <= maxIdentityInputBytes
}

func privateDirectoryEntries(directoryFD int) ([]string, error) {
	duplicate, err := unix.Dup(directoryFD)
	if err != nil {
		return nil, err
	}
	if _, err := unix.Seek(duplicate, 0, 0); err != nil {
		_ = unix.Close(duplicate)
		return nil, err
	}
	directory := os.NewFile(uintptr(duplicate), "execution-snapshot-directory")
	if directory == nil {
		_ = unix.Close(duplicate)
		return nil, errors.New("execution snapshot directory descriptor is invalid")
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func copyFDExact(sourceFD, destinationFD int, expected int64) (string, int64, error) {
	return copyFDExactContext(context.Background(), sourceFD, destinationFD, expected)
}

// copyFDExactContext preserves the exact-byte copy contract while allowing a
// bounded distribution snapshot to stop promptly between fixed-size I/O
// chunks.  Callers still verify the opened source descriptor both before and
// after copying.
func copyFDExactContext(ctx context.Context, sourceFD, destinationFD int, expected int64) (string, int64, error) {
	if expected < 0 || expected > maxIdentityInputBytes {
		return "", 0, errors.New("execution source has invalid size")
	}
	hash := sha256.New()
	buffer := make([]byte, 32<<10)
	var total int64
	for total < expected {
		if err := ctx.Err(); err != nil {
			return "", total, err
		}
		limit := int64(len(buffer))
		if remaining := expected - total; remaining < limit {
			limit = remaining
		}
		read, err := unix.Read(sourceFD, buffer[:limit])
		if err != nil {
			return "", total, err
		}
		if read == 0 {
			return "", total, errors.New("execution source was truncated")
		}
		if _, err := hash.Write(buffer[:read]); err != nil {
			return "", total, err
		}
		for written := 0; written < read; {
			if err := ctx.Err(); err != nil {
				return "", total, err
			}
			count, err := unix.Write(destinationFD, buffer[written:read])
			if err != nil {
				return "", total, err
			}
			if count == 0 {
				return "", total, errors.New("execution snapshot write was truncated")
			}
			written += count
		}
		total += int64(read)
	}
	if err := ctx.Err(); err != nil {
		return "", total, err
	}
	probe := []byte{0}
	if read, err := unix.Read(sourceFD, probe); err != nil || read != 0 {
		if err != nil {
			return "", total, err
		}
		return "", total, errors.New("execution source grew during snapshot copy")
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), total, nil
}

func digestFD(fd int, expected int64) (string, int64, error) {
	if expected < 0 || expected > maxIdentityInputBytes {
		return "", 0, errors.New("execution snapshot has invalid size")
	}
	hash := sha256.New()
	buffer := make([]byte, 32<<10)
	var offset int64
	for offset < expected {
		limit := int64(len(buffer))
		if remaining := expected - offset; remaining < limit {
			limit = remaining
		}
		count, err := unix.Pread(fd, buffer[:limit], offset)
		if err != nil {
			return "", offset, err
		}
		if count == 0 {
			return "", offset, errors.New("execution snapshot was truncated")
		}
		if _, err := hash.Write(buffer[:count]); err != nil {
			return "", offset, err
		}
		offset += int64(count)
	}
	probe := []byte{0}
	if count, err := unix.Pread(fd, probe, offset); err != nil || count != 0 {
		if err != nil {
			return "", offset, err
		}
		return "", offset, fmt.Errorf("execution snapshot grew during validation")
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), offset, nil
}
