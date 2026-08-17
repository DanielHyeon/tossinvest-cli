//go:build unix

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

// snapshotGoDistribution takes the only acceptable GOROOT from the explicit
// selected `<root>/bin/go` layout.  All subsequent Go commands receive only
// privateRoot; sourceRoot is never used as a command environment or reopened.
func snapshotGoDistribution(selectedPath, selectedDigest string) (*goDistributionSnapshot, error) {
	return snapshotGoDistributionContext(context.Background(), selectedPath, selectedDigest)
}

type goDistributionInput struct {
	path string
	fd   int
	stat unix.Stat_t
}

// snapshotGoDistributionContext pre-enumerates a no-follow descriptor-backed
// source inventory, then bounds copy/fsync work to a small fixed pool.  Source
// regular FDs remain open from inventory through their post-copy digest check;
// no worker can resolve the original distribution pathname.
func snapshotGoDistributionContext(ctx context.Context, selectedPath, selectedDigest string) (_ *goDistributionSnapshot, err error) {
	if ctx == nil {
		return nil, errors.New("Go distribution snapshot context is nil")
	}
	if filepath.Base(selectedPath) != "go" || filepath.Base(filepath.Dir(selectedPath)) != "bin" {
		return nil, errors.New("selected Go binary is not <go-root>/bin/go")
	}
	sourceRoot := filepath.Dir(filepath.Dir(selectedPath))
	if !filepath.IsAbs(sourceRoot) || filepath.Clean(sourceRoot) != sourceRoot || sourceRoot == string(filepath.Separator) {
		return nil, errors.New("selected Go distribution root is invalid")
	}
	sourceFD, err := unix.Open(sourceRoot, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(sourceFD)
	var sourceBefore unix.Stat_t
	if err := unix.Fstat(sourceFD, &sourceBefore); err != nil || !safeSourceGoDirectory(sourceBefore) {
		return nil, errors.New("selected Go distribution root is unsafe")
	}
	inputs := make([]goDistributionInput, 0)
	directories := make([]string, 0)
	entries := 0
	var total int64
	if err := enumerateGoDistribution(ctx, sourceFD, "", &inputs, &directories, &entries, &total); err != nil {
		closeGoDistributionInputs(inputs)
		return nil, err
	}
	defer closeGoDistributionInputs(inputs)
	var sourceAfter unix.Stat_t
	if err := unix.Fstat(sourceFD, &sourceAfter); err != nil || !safeSourceGoDirectory(sourceAfter) || sourceAfter.Dev != sourceBefore.Dev || sourceAfter.Ino != sourceBefore.Ino || sourceAfter.Size != sourceBefore.Size {
		return nil, errors.New("selected Go distribution root changed during inventory")
	}

	privateRoot, err := os.MkdirTemp("/tmp", "a112-mb-us-goroot-")
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(privateRoot, 0o700); err != nil {
		_ = os.Remove(privateRoot)
		return nil, err
	}
	rootFD, err := unix.Open(privateRoot, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		_ = os.Remove(privateRoot)
		return nil, err
	}
	snapshot := &goDistributionSnapshot{sourceRoot: sourceRoot, privateRoot: privateRoot, rootFD: rootFD, entries: map[string]goDistributionEntry{}}
	complete := false
	defer func() {
		if !complete {
			_ = snapshot.close()
		}
	}()
	if err := validatePrivateDirectory(rootFD); err != nil {
		return nil, err
	}
	if err := createPrivateGoDirectories(rootFD, directories); err != nil {
		return nil, err
	}
	fileEntries, err := copyGoDistributionInputs(ctx, rootFD, inputs)
	if err != nil {
		return nil, err
	}
	for _, directory := range directories {
		snapshot.entries[directory] = goDistributionEntry{path: directory, mode: 0o500, dir: true}
	}
	for _, entry := range fileEntries {
		snapshot.entries[entry.path] = entry
	}
	if len(snapshot.entries) != entries {
		return nil, errors.New("private Go distribution inventory is incomplete")
	}
	if err := sealPrivateGoDirectories(rootFD, directories); err != nil {
		return nil, err
	}
	goEntry, ok := snapshot.entries["bin/go"]
	if !ok || goEntry.dir || goEntry.digest != selectedDigest {
		return nil, errors.New("selected Go binary does not match private distribution manifest")
	}
	snapshot.digest = goDistributionManifestDigest(snapshot.entries)
	if err := unix.Fsync(rootFD); err != nil {
		return nil, err
	}
	if err := validateGoDistributionSnapshot(snapshot); err != nil {
		return nil, err
	}
	complete = true
	return snapshot, nil
}

func enumerateGoDistribution(ctx context.Context, directoryFD int, parent string, inputs *[]goDistributionInput, directories *[]string, entries *int, total *int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	names, err := privateDirectoryEntries(directoryFD)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		if name == "." || name == ".." || strings.Contains(name, string(filepath.Separator)) {
			return errors.New("selected Go distribution has invalid entry name")
		}
		path := name
		if parent != "" {
			path = parent + "/" + name
		}
		*entries++
		if *entries > maxGoDistributionEntries {
			return errors.New("selected Go distribution exceeds entry limit")
		}
		var before unix.Stat_t
		if err := unix.Fstatat(directoryFD, name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		switch before.Mode & unix.S_IFMT {
		case unix.S_IFDIR:
			if !safeSourceGoDirectory(before) {
				return errors.New("selected Go distribution contains unsafe directory")
			}
			child, err := unix.Openat(directoryFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			if err != nil {
				return err
			}
			err = enumerateGoDistribution(ctx, child, path, inputs, directories, entries, total)
			var after unix.Stat_t
			statErr := unix.Fstat(child, &after)
			closeErr := unix.Close(child)
			if err != nil {
				return err
			}
			if statErr != nil || closeErr != nil || !safeSourceGoDirectory(after) || after.Dev != before.Dev || after.Ino != before.Ino || after.Size != before.Size {
				return errors.New("selected Go distribution directory changed during inventory")
			}
			*directories = append(*directories, path)
		case unix.S_IFREG:
			if !safeSourceGoFile(before) || before.Size > maxIdentityInputBytes || *total > maxGoDistributionBytes-before.Size {
				return errors.New("selected Go distribution file exceeds safety limit")
			}
			file, err := unix.Openat(directoryFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
			if err != nil {
				return err
			}
			var opened unix.Stat_t
			if err := unix.Fstat(file, &opened); err != nil || !safeSourceGoFile(opened) || opened.Dev != before.Dev || opened.Ino != before.Ino || opened.Size != before.Size {
				_ = unix.Close(file)
				return errors.New("selected Go distribution file changed before inventory")
			}
			*inputs = append(*inputs, goDistributionInput{path: path, fd: file, stat: opened})
			*total += opened.Size
		default:
			return errors.New("selected Go distribution contains symlink or special file")
		}
	}
	return nil
}

func createPrivateGoDirectories(rootFD int, directories []string) error {
	sort.Slice(directories, func(i, j int) bool {
		return strings.Count(directories[i], "/") < strings.Count(directories[j], "/") || (strings.Count(directories[i], "/") == strings.Count(directories[j], "/") && directories[i] < directories[j])
	})
	for _, path := range directories {
		parent, name := filepath.ToSlash(filepath.Dir(path)), filepath.Base(path)
		if parent == "." {
			parent = ""
		}
		parentFD, err := privateGoDirectoryAt(rootFD, parent)
		if err != nil {
			return err
		}
		err = unix.Mkdirat(parentFD, name, 0o700)
		closeErr := unix.Close(parentFD)
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func copyGoDistributionInputs(ctx context.Context, rootFD int, inputs []goDistributionInput) ([]goDistributionEntry, error) {
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > 16 {
		workers = 16
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan goDistributionInput)
	entries := make(chan goDistributionEntry, len(inputs))
	var firstErr error
	var once sync.Once
	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for input := range jobs {
				entry, err := copyGoDistributionInput(workCtx, rootFD, input)
				if err != nil {
					fail(err)
					return
				}
				entries <- entry
			}
		}()
	}
	for _, input := range inputs {
		select {
		case <-workCtx.Done():
			break
		case jobs <- input:
		}
		if workCtx.Err() != nil {
			break
		}
	}
	close(jobs)
	group.Wait()
	close(entries)
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]goDistributionEntry, 0, len(inputs))
	for entry := range entries {
		out = append(out, entry)
	}
	if len(out) != len(inputs) {
		return nil, errors.New("private Go distribution copy was incomplete")
	}
	return out, nil
}

func copyGoDistributionInput(ctx context.Context, rootFD int, input goDistributionInput) (goDistributionEntry, error) {
	if err := ctx.Err(); err != nil {
		return goDistributionEntry{}, err
	}
	parent, name := filepath.ToSlash(filepath.Dir(input.path)), filepath.Base(input.path)
	if parent == "." {
		parent = ""
	}
	parentFD, err := privateGoDirectoryAt(rootFD, parent)
	if err != nil {
		return goDistributionEntry{}, err
	}
	defer unix.Close(parentFD)
	destination, err := unix.Openat(parentFD, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return goDistributionEntry{}, err
	}
	digest, copied, copyErr := copyFDExactContext(ctx, input.fd, destination, input.stat.Size)
	var after unix.Stat_t
	statErr := unix.Fstat(input.fd, &after)
	postDigest, postSize, postDigestErr := digestFD(input.fd, input.stat.Size)
	if copyErr == nil && (statErr != nil || !safeSourceGoFile(after) || after.Dev != input.stat.Dev || after.Ino != input.stat.Ino || after.Size != input.stat.Size || copied != input.stat.Size || postDigestErr != nil || postSize != input.stat.Size || postDigest != digest) {
		copyErr = errors.New("selected Go distribution file changed during snapshot")
	}
	mode := uint32(0o400)
	if input.stat.Mode&0o111 != 0 {
		mode = 0o500
	}
	if copyErr == nil {
		copyErr = unix.Fchmod(destination, mode)
	}
	if copyErr == nil {
		copyErr = unix.Fsync(destination)
	}
	closeErr := unix.Close(destination)
	if copyErr != nil {
		return goDistributionEntry{}, copyErr
	}
	if closeErr != nil {
		return goDistributionEntry{}, closeErr
	}
	return goDistributionEntry{path: input.path, mode: mode, size: input.stat.Size, digest: digest}, nil
}

func sealPrivateGoDirectories(rootFD int, directories []string) error {
	sort.Slice(directories, func(i, j int) bool {
		return strings.Count(directories[i], "/") > strings.Count(directories[j], "/") || (strings.Count(directories[i], "/") == strings.Count(directories[j], "/") && directories[i] > directories[j])
	})
	for _, path := range directories {
		fd, err := privateGoDirectoryAt(rootFD, path)
		if err != nil {
			return err
		}
		err = unix.Fchmod(fd, 0o500)
		if err == nil {
			err = unix.Fsync(fd)
		}
		closeErr := unix.Close(fd)
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return unix.Fsync(rootFD)
}

func privateGoDirectoryAt(rootFD int, path string) (int, error) {
	fd, err := unix.Dup(rootFD)
	if err != nil {
		return -1, err
	}
	if path == "" {
		return fd, nil
	}
	for _, part := range strings.Split(path, "/") {
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		closeErr := unix.Close(fd)
		if err != nil {
			return -1, err
		}
		if closeErr != nil {
			_ = unix.Close(next)
			return -1, closeErr
		}
		fd = next
	}
	return fd, nil
}

func closeGoDistributionInputs(inputs []goDistributionInput) {
	for _, input := range inputs {
		if input.fd >= 0 {
			_ = unix.Close(input.fd)
		}
	}
}

func validateGoDistributionSnapshot(snapshot *goDistributionSnapshot) error {
	if snapshot == nil || snapshot.rootFD < 0 || snapshot.privateRoot == "" || snapshot.digest == "" || len(snapshot.entries) == 0 || len(snapshot.entries) > maxGoDistributionEntries {
		return errors.New("private Go distribution snapshot is unavailable")
	}
	if err := validatePrivateDirectory(snapshot.rootFD); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(snapshot.entries))
	if err := validatePrivateGoDirectory(snapshot.rootFD, "", snapshot.entries, seen); err != nil {
		return err
	}
	if len(seen) != len(snapshot.entries) || goDistributionManifestDigest(snapshot.entries) != snapshot.digest {
		return errors.New("private Go distribution manifest drift")
	}
	return nil
}

func validatePrivateGoDirectory(directoryFD int, parent string, manifest map[string]goDistributionEntry, seen map[string]struct{}) error {
	names, err := privateDirectoryEntries(directoryFD)
	if err != nil {
		return err
	}
	for _, name := range names {
		path := name
		if parent != "" {
			path = parent + "/" + name
		}
		expected, ok := manifest[path]
		if !ok {
			return errors.New("private Go distribution has unexpected entry")
		}
		if _, duplicate := seen[path]; duplicate {
			return errors.New("private Go distribution duplicate entry")
		}
		var stat unix.Stat_t
		if err := unix.Fstatat(directoryFD, name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if expected.dir {
			if stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Uid != uint32(os.Geteuid()) || uint32(stat.Mode&0o777) != expected.mode {
				return errors.New("private Go distribution directory drift")
			}
			child, err := unix.Openat(directoryFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			if err != nil {
				return err
			}
			err = validatePrivateGoDirectory(child, path, manifest, seen)
			closeErr := unix.Close(child)
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		} else {
			if !safePrivateGoFile(stat, expected.mode) || stat.Size != expected.size {
				return errors.New("private Go distribution file drift")
			}
			file, err := unix.Openat(directoryFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
			if err != nil {
				return err
			}
			digest, size, digestErr := digestFD(file, expected.size)
			closeErr := unix.Close(file)
			if digestErr != nil || closeErr != nil || size != expected.size || digest != expected.digest {
				return errors.New("private Go distribution file content drift")
			}
		}
		seen[path] = struct{}{}
	}
	return nil
}

func closeGoDistributionSnapshot(snapshot *goDistributionSnapshot) error {
	var errs []error
	if snapshot.rootFD >= 0 {
		if err := removePrivateGoTree(snapshot.rootFD); err != nil {
			errs = append(errs, err)
		}
		if err := unix.Close(snapshot.rootFD); err != nil {
			errs = append(errs, err)
		}
		snapshot.rootFD = -1
	}
	if snapshot.privateRoot != "" {
		if err := os.Remove(snapshot.privateRoot); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// removePrivateGoTree never follows an entry supplied after snapshot creation.
// It first restores owner-write on each private directory, then removes every
// entry descriptor-relatively so both successful and HOLD paths release all
// 0500/0400 snapshot material.
func removePrivateGoTree(directoryFD int) error {
	if err := unix.Fchmod(directoryFD, 0o700); err != nil {
		return err
	}
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
			err = removePrivateGoTree(child)
			closeErr := unix.Close(child)
			if err != nil {
				return err
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

func safeSourceGoDirectory(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFDIR && stat.Mode&0o022 == 0
}

func safeSourceGoFile(stat unix.Stat_t) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG && stat.Size >= 0 && stat.Mode&0o022 == 0
}

func safePrivateGoFile(stat unix.Stat_t, mode uint32) bool {
	return stat.Mode&unix.S_IFMT == unix.S_IFREG && stat.Uid == uint32(os.Geteuid()) && uint32(stat.Mode&0o777) == mode && stat.Size >= 0 && stat.Size <= maxIdentityInputBytes
}

func goDistributionManifestDigest(entries map[string]goDistributionEntry) string {
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var manifest strings.Builder
	for _, path := range paths {
		entry := entries[path]
		kind := "file"
		if entry.dir {
			kind = "dir"
		}
		fmt.Fprintf(&manifest, "%s\x00%s\x00%o\x00%d\x00%s\n", entry.path, kind, entry.mode, entry.size, entry.digest)
	}
	return digestBytes([]byte(manifest.String()))
}
