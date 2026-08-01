package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpenHTTPAPIRegularNoFollowAcceptsOwnedPrivatePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix ownership and mode semantics are not available on Windows")
	}
	path := writeHTTPAPISecureFileFixture(t, 0o700, 0o600)

	file, err := openHTTPAPIRegularNoFollow(path)
	if err != nil {
		t.Fatalf("open secure trust anchor: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close secure trust anchor: %v", err)
	}
}

func TestOpenHTTPAPIRegularNoFollowRejectsGroupWritableParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix ownership and mode semantics are not available on Windows")
	}
	path := writeHTTPAPISecureFileFixture(t, 0o770, 0o600)

	if file, err := openHTTPAPIRegularNoFollow(path); err == nil {
		_ = file.Close()
		t.Fatal("group-writable trust-anchor directory accepted")
	}
}

func TestOpenHTTPAPIRegularNoFollowRejectsForeignOwner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix ownership semantics are not available on Windows")
	}
	path := writeHTTPAPISecureFileFixture(t, 0o700, 0o600)
	foreignUID := uint32(os.Geteuid()) + 1
	if foreignUID == 0 {
		foreignUID++
	}

	if file, err := openHTTPAPIRegularNoFollowForUID(path, foreignUID); err == nil {
		_ = file.Close()
		t.Fatal("trust-anchor path owned by another service identity accepted")
	}
}

func TestOpenHTTPAPIRegularNoFollowRejectsSymlinkedPathComponent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix openat path-walk semantics are not available on Windows")
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	realParent := filepath.Join(root, "real")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(realParent, "capability-public.key")
	if err := os.WriteFile(path, []byte("test-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	linkedParent := filepath.Join(root, "linked")
	if err := os.Symlink(realParent, linkedParent); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	if file, err := openHTTPAPIRegularNoFollow(filepath.Join(linkedParent, filepath.Base(path))); err == nil {
		_ = file.Close()
		t.Fatal("symlinked trust-anchor path component accepted")
	}
}

func writeHTTPAPISecureFileFixture(t *testing.T, directoryMode, fileMode os.FileMode) string {
	t.Helper()
	temporary := t.TempDir()
	if err := os.Chmod(temporary, 0o700); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(temporary, "trust")
	if err := os.Mkdir(parent, directoryMode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, directoryMode); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "capability-public.key")
	if err := os.WriteFile(path, []byte("test-key"), fileMode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, fileMode); err != nil {
		t.Fatal(err)
	}
	return path
}
