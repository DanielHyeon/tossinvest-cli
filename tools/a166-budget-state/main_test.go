//go:build linux

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestA166OfflineCLIProvisionAndCheck(t *testing.T) {
	base := t.TempDir()
	if err := os.Chmod(base, 0700); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(base, "manifest")
	var out bytes.Buffer
	if err := run(context.Background(), []string{"provision", "--root", root, "--manifest", path, "--scope", "scope", "--instance", "instance"}, &out); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"check", "--manifest", path}, &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "provisioned\nverified\n" || strings.Contains(got, "scope") || strings.Contains(got, "instance") {
		t.Fatalf("unexpected output %q", got)
	}
}
func TestA166OfflineCLIUsageRejects(t *testing.T) {
	for _, args := range [][]string{nil, {"other"}, {"provision"}, {"check"}, {"check", "--manifest", "/tmp/x", "extra"}} {
		var out bytes.Buffer
		if err := run(context.Background(), args, &out); err == nil || out.Len() != 0 {
			t.Fatalf("accepted %v", args)
		}
	}
}
