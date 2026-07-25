//go:build linux

package journal

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test's working directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod above the test directory")
		}
		dir = parent
	}
}

// TestSystemFSProberRejectsThisRepositoryMount is the real-syscall half of the
// guard. This repository lives on a fuseblk (NTFS) mount, so the production
// prober must refuse it — that is the concrete case the spec's allowlist exists
// for, and it proves the guard is wired to statfs rather than only to fakes.
//
// On a host where the checkout happens to sit on an allowlisted filesystem there
// is nothing to prove, so the test skips instead of failing.
func TestSystemFSProberRejectsThisRepositoryMount(t *testing.T) {
	root := repoRoot(t)
	prober := SystemFSProber()

	info, err := prober.Probe(root)
	if err != nil {
		t.Fatalf("Probe(%s): %v", root, err)
	}
	t.Logf("repository mount: %s (magic %#x)", info.Name, info.Magic)

	if _, err := CheckFilesystem(prober, root); err == nil {
		if !isAllowedFS(info) {
			t.Fatalf("filesystem %q is outside the allowlist but CheckFilesystem accepted it", info.Name)
		}
		t.Skipf("repository is on %s, which is allowlisted — nothing to prove here", info.Name)
	} else {
		if !errors.Is(err, ErrFilesystemNotAllowed) && !errors.Is(err, ErrFilesystemUnknown) {
			t.Fatalf("unexpected refusal reason: %v", err)
		}
		if !strings.Contains(err.Error(), root) {
			t.Errorf("refusal must name the directory: %v", err)
		}
		for _, name := range AllowedFilesystems() {
			if !strings.Contains(err.Error(), name) {
				t.Errorf("refusal must list the allowlist (%s missing): %v", name, err)
			}
		}
	}
}

// TestSystemFSProberReportsTempDir exercises the syscall path on a directory that
// is normally a journaling local filesystem, so a probe regression (wrong magic
// masking, wrong field) shows up as a rejected temp dir.
func TestSystemFSProberReportsTempDir(t *testing.T) {
	dir := t.TempDir()
	info, err := SystemFSProber().Probe(dir)
	if err != nil {
		t.Fatalf("Probe(%s): %v", dir, err)
	}
	if info.Magic == 0 {
		t.Fatalf("Probe returned no magic for %s: %+v", dir, info)
	}
	t.Logf("temp dir mount: %s (magic %#x)", info.Name, info.Magic)
	if !isAllowedFS(info) {
		t.Skipf("TMPDIR is on %s (magic %#x); the journal's own tests inject a prober instead",
			info.Name, info.Magic)
	}
	if _, err := CheckFilesystem(SystemFSProber(), dir); err != nil {
		t.Fatalf("CheckFilesystem on allowlisted %s: %v", info.Name, err)
	}
}
