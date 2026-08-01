//go:build unix

package positionpolicyrpc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePrivateTestDescriptor(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
}

func privateTestDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "position-policy-rpc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func privateDescriptorTree(t *testing.T) (string, string) {
	t.Helper()
	engineDir := privateTestDir(t)
	controlDir := ControlDirectory(engineDir)
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := DescriptorPath(engineDir)
	writePrivateTestDescriptor(t, path, `{"address":"127.0.0.1:1","token":"01234567890123456789012345678901","pid":1}`)
	return engineDir, path
}

func TestDecodeDescriptorRejectsUnknownTrailingAndOversizedContent(t *testing.T) {
	valid := `{"address":"127.0.0.1:1","token":"01234567890123456789012345678901","pid":1}`
	tests := []struct {
		name string
		body string
	}{
		{name: "unknown field", body: strings.TrimSuffix(valid, "}") + `,"extra":true}`},
		{name: "trailing value", body: valid + ` {}`},
		{name: "oversized", body: string(bytes.Repeat([]byte("x"), maxDescriptorBytes+1))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeDescriptor(strings.NewReader(test.body)); err == nil {
				t.Fatal("decodeDescriptor accepted malformed descriptor")
			}
		})
	}
}

func TestOpenPrivateDescriptorRejectsInsecureFilesystemObjects(t *testing.T) {
	t.Run("group writable engine directory", func(t *testing.T) {
		engineDir, path := privateDescriptorTree(t)
		if err := os.Chmod(engineDir, 0o770); err != nil {
			t.Fatal(err)
		}
		if _, err := openPrivateDescriptor(path, openDescriptorNoFollow); err == nil {
			t.Fatal("opened descriptor beneath group-writable engine directory")
		}
	})

	t.Run("wrong control directory mode", func(t *testing.T) {
		engineDir, path := privateDescriptorTree(t)
		if err := os.Chmod(ControlDirectory(engineDir), 0o750); err != nil {
			t.Fatal(err)
		}
		if _, err := openPrivateDescriptor(path, openDescriptorNoFollow); err == nil {
			t.Fatal("opened descriptor beneath non-private control directory")
		}
	})

	t.Run("symlink traversal", func(t *testing.T) {
		engineDir := privateTestDir(t)
		target := privateTestDir(t)
		if err := os.Symlink(target, ControlDirectory(engineDir)); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(target, DescriptorFileName)
		writePrivateTestDescriptor(t, path, `{}`)
		if _, err := openPrivateDescriptor(DescriptorPath(engineDir), openDescriptorNoFollow); err == nil {
			t.Fatal("opened descriptor through symlinked control directory")
		}
	})

	t.Run("descriptor symlink", func(t *testing.T) {
		engineDir := privateTestDir(t)
		if err := os.Mkdir(ControlDirectory(engineDir), 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(engineDir, "target.json")
		writePrivateTestDescriptor(t, target, `{}`)
		if err := os.Symlink(target, DescriptorPath(engineDir)); err != nil {
			t.Fatal(err)
		}
		if _, err := openPrivateDescriptor(DescriptorPath(engineDir), openDescriptorNoFollow); err == nil {
			t.Fatal("opened symlinked descriptor")
		}
	})

	t.Run("descriptor hard link", func(t *testing.T) {
		engineDir, path := privateDescriptorTree(t)
		if err := os.Link(path, filepath.Join(engineDir, "second-link.json")); err != nil {
			t.Fatal(err)
		}
		if _, err := openPrivateDescriptor(path, openDescriptorNoFollow); err == nil {
			t.Fatal("opened multiply-linked descriptor")
		}
	})
}

func TestOpenPrivateDescriptorRejectsInodeReplacementDuringOpen(t *testing.T) {
	_, path := privateDescriptorTree(t)
	replaced := false
	opener := func(path string) (*os.File, error) {
		original := path + ".original"
		if err := os.Rename(path, original); err != nil {
			return nil, err
		}
		writePrivateTestDescriptor(t, path, `{}`)
		replaced = true
		return openDescriptorNoFollow(path)
	}
	if f, err := openPrivateDescriptor(path, opener); err == nil {
		f.Close()
		t.Fatal("opened inode replacement")
	}
	if !replaced {
		t.Fatal("replacement opener did not run")
	}
}
