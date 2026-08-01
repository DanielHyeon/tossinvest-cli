//go:build unix

package httpapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSecurityStoreRequiresOwnedMode0700Directory(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	groupWritable := filepath.Join(root, "group-writable")
	if err := os.Mkdir(groupWritable, 0o770); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(groupWritable, 0o770); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenSecurityStore(filepath.Join(groupWritable, "security.db")); err == nil {
		t.Fatal("group-writable security database directory accepted")
	}

	secure := filepath.Join(root, "secure")
	if err := os.Mkdir(secure, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenSecurityStore(filepath.Join(secure, "security.db"))
	if err != nil {
		t.Fatalf("private security database directory rejected: %v", err)
	}
	_ = store.Close()
}

func TestSecurityStoreDirectoryRejectsForeignServiceIdentityAndSymlinkComponent(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	secure := filepath.Join(root, "secure")
	if err := os.Mkdir(secure, 0o700); err != nil {
		t.Fatal(err)
	}
	foreignUID := uint32(os.Geteuid()) + 1
	if foreignUID == 0 {
		foreignUID++
	}
	if err := validateSecurityStoreDirectoryForUID(secure, foreignUID); err == nil {
		t.Fatal("foreign-owned security directory accepted for another service identity")
	}
	linked := filepath.Join(root, "linked")
	if err := os.Symlink(secure, linked); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := validateSecurityStoreDirectory(linked); err == nil {
		t.Fatal("symlinked security database directory accepted")
	}
}
