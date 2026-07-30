package localupdate

import (
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func copyTestExecutable(t *testing.T, dst string, suffix []byte) []byte {
	t.Helper()
	src, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, suffix...)
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatal(err)
	}
	return data
}

func newFixture(t *testing.T) (*Updater, string, []byte) {
	t.Helper()
	dir := t.TempDir()
	current := filepath.Join(dir, "tossctl")
	currentBytes := copyTestExecutable(t, current, nil)
	info, err := buildinfo.ReadFile(current)
	if err != nil {
		t.Fatalf("reading test executable build info: %v", err)
	}
	updater, err := newUpdater(
		current, info.Main.Path, info.Path, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return updater, current, currentBytes
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestInspectReportsFixedCandidateMetadataWithoutExecutingIt(t *testing.T) {
	updater, current, _ := newFixture(t)
	candidateBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew-build\n"))

	got := updater.Inspect()
	if !got.Installable {
		t.Fatalf("candidate not installable: %s", got.Reason)
	}
	if got.Current.Path != current || got.Candidate.Path != current+".candidate" {
		t.Fatalf("paths = current %q candidate %q", got.Current.Path, got.Candidate.Path)
	}
	if got.Candidate.SHA256 != digest(candidateBytes) {
		t.Fatalf("candidate SHA256 = %s, want %s", got.Candidate.SHA256, digest(candidateBytes))
	}
	if got.Candidate.ModulePath == "" || got.Candidate.CommandPath == "" ||
		got.Candidate.GOOS == "" || got.Candidate.GOARCH == "" {
		t.Fatalf("missing build metadata: %+v", got.Candidate)
	}
}

func TestInspectRefusesMissingSymlinkNonExecutableAndNeverExecutesCandidate(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		updater, _, _ := newFixture(t)
		got := updater.Inspect()
		if got.Installable || !strings.Contains(got.Reason, "없") {
			t.Fatalf("inspection = %+v", got)
		}
	})

	t.Run("symlink", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		if err := os.Symlink(current, current+".candidate"); err != nil {
			t.Fatal(err)
		}
		got := updater.Inspect()
		if got.Installable || !strings.Contains(strings.ToLower(got.Reason), "symbolic") {
			t.Fatalf("inspection = %+v", got)
		}
	})

	t.Run("non executable", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		copyTestExecutable(t, current+".candidate", nil)
		if err := os.Chmod(current+".candidate", 0o600); err != nil {
			t.Fatal(err)
		}
		got := updater.Inspect()
		if got.Installable || !strings.Contains(got.Reason, "실행") {
			t.Fatalf("inspection = %+v", got)
		}
	})

	t.Run("non regular", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		if err := os.Mkdir(current+".candidate", 0o755); err != nil {
			t.Fatal(err)
		}
		got := updater.Inspect()
		if got.Installable || !strings.Contains(got.Reason, "regular") {
			t.Fatalf("inspection = %+v", got)
		}
	})

	t.Run("script is not executed", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		marker := filepath.Join(t.TempDir(), "executed")
		script := "#!/bin/sh\n: > " + marker + "\n"
		if err := os.WriteFile(current+".candidate", []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		if got := updater.Inspect(); got.Installable {
			t.Fatalf("script candidate was accepted: %+v", got)
		}
		if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("candidate code ran; marker stat = %v", err)
		}
	})
}

func TestInspectRefusesWrongModuleAndPlatform(t *testing.T) {
	t.Run("module", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		copyTestExecutable(t, current+".candidate", nil)
		updater.expectedModule = "example.invalid/not-tossctl"
		got := updater.Inspect()
		if got.Installable || !strings.Contains(got.Reason, "module") {
			t.Fatalf("inspection = %+v", got)
		}
	})

	t.Run("platform", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		copyTestExecutable(t, current+".candidate", nil)
		updater.goarch = "definitely-not-this-architecture"
		got := updater.Inspect()
		if got.Installable || !strings.Contains(got.Reason, "platform") {
			t.Fatalf("inspection = %+v", got)
		}
	})
}

func TestInspectRefusesDifferentCommandFromTheSameModule(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(t.TempDir(), "tossctl")
	for output, target := range map[string]string{
		"./cmd/tossctl":       current,
		"./tools/boundarymap": current + ".candidate",
	} {
		cmd := exec.Command("go", "build", "-o", target, output)
		cmd.Dir = root
		if buildOutput, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("building %s: %v\n%s", output, err, buildOutput)
		}
	}
	updater, err := New(current)
	if err != nil {
		t.Fatalf("New rejected an actual tossctl executable: %v", err)
	}

	got := updater.Inspect()
	if got.Installable || !strings.Contains(strings.ToLower(got.Reason), "command") {
		t.Fatalf("same-module non-tossctl candidate was accepted: %+v", got)
	}
}

func TestInstallBindsReviewedHashAndCurrentStartupFingerprint(t *testing.T) {
	t.Run("candidate changed after review", func(t *testing.T) {
		updater, current, currentBytes := newFixture(t)
		copyTestExecutable(t, current+".candidate", []byte("\nreviewed\n"))
		reviewed := updater.Inspect().Candidate.SHA256
		copyTestExecutable(t, current+".candidate", []byte("\nchanged\n"))

		if _, err := updater.Install(reviewed, nil); !errors.Is(err, ErrCandidateChanged) {
			t.Fatalf("Install = %v, want ErrCandidateChanged", err)
		}
		if got, _ := os.ReadFile(current); digest(got) != digest(currentBytes) {
			t.Fatal("current bytes changed on candidate mismatch")
		}
		if _, err := os.Stat(current + ".rollback"); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("rollback changed on candidate mismatch: %v", err)
		}
	})

	t.Run("current drifted", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		candidate := copyTestExecutable(t, current+".candidate", []byte("\nnew\n"))
		if err := os.WriteFile(current, append(candidate, []byte("manual drift")...), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := updater.Install(digest(candidate), nil); !errors.Is(err, ErrCurrentChanged) {
			t.Fatalf("Install = %v, want ErrCurrentChanged", err)
		}
	})
}

func TestInstallPublishesRollbackThenAtomicallyReplacesCurrent(t *testing.T) {
	updater, current, oldBytes := newFixture(t)
	newBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew-build\n"))
	var guarded bool
	result, err := updater.Install(digest(newBytes), func() error {
		guarded = true
		return nil
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !guarded {
		t.Fatal("commit guard was not called")
	}
	if got, _ := os.ReadFile(current); digest(got) != digest(newBytes) {
		t.Fatal("current path does not contain candidate bytes")
	}
	if got, _ := os.ReadFile(current + ".rollback"); digest(got) != digest(oldBytes) {
		t.Fatal("rollback does not contain previous current bytes")
	}
	if result.OldSHA256 != digest(oldBytes) || result.NewSHA256 != digest(newBytes) ||
		result.RollbackPath != current+".rollback" {
		t.Fatalf("result = %+v", result)
	}
}

func TestInstallPublishesAndSyncsRollbackBeforeReplacingCurrent(t *testing.T) {
	updater, current, _ := newFixture(t)
	newBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew-build\n"))
	realRename := updater.rename
	realSync := updater.syncDir
	var events []string
	updater.rename = func(old, new string) error {
		switch new {
		case current + ".rollback":
			events = append(events, "rename-rollback")
		case current:
			events = append(events, "rename-current")
		}
		return realRename(old, new)
	}
	updater.syncDir = func(dir string) error {
		events = append(events, "sync-directory")
		return realSync(dir)
	}

	if _, err := updater.Install(digest(newBytes), nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	want := []string{"rename-rollback", "sync-directory", "rename-current", "sync-directory"}
	if strings.Join(events, ",") != strings.Join(want, ",") {
		t.Fatalf("publish events = %v, want %v", events, want)
	}
}

func TestInstallCommitGuardRefusalLeavesCurrentAndRollbackUntouched(t *testing.T) {
	updater, current, oldBytes := newFixture(t)
	newBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew-build\n"))
	guardErr := errors.New("verification became active")

	if _, err := updater.Install(digest(newBytes), func() error { return guardErr }); !errors.Is(err, guardErr) {
		t.Fatalf("Install = %v, want commit guard error", err)
	}
	if got, err := os.ReadFile(current); err != nil || digest(got) != digest(oldBytes) {
		t.Fatalf("current changed on guard refusal: err=%v", err)
	}
	if _, err := os.Stat(current + ".rollback"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rollback was published on guard refusal: %v", err)
	}
}

func TestInstallFailureNeverLeavesCurrentAbsentAndSyncFailureRestoresOldBytes(t *testing.T) {
	t.Run("candidate rename fails", func(t *testing.T) {
		updater, current, oldBytes := newFixture(t)
		newBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew\n"))
		realRename := updater.rename
		updater.rename = func(old, new string) error {
			if new == current && strings.Contains(filepath.Base(old), ".candidate-") {
				return errors.New("injected candidate rename failure")
			}
			return realRename(old, new)
		}
		if _, err := updater.Install(digest(newBytes), nil); err == nil {
			t.Fatal("Install succeeded")
		}
		got, err := os.ReadFile(current)
		if err != nil {
			t.Fatalf("current path disappeared: %v", err)
		}
		if digest(got) != digest(oldBytes) {
			t.Fatal("current bytes changed after candidate rename failure")
		}
	})

	t.Run("post replacement directory sync fails", func(t *testing.T) {
		updater, current, oldBytes := newFixture(t)
		newBytes := copyTestExecutable(t, current+".candidate", []byte("\nnew\n"))
		realSync := updater.syncDir
		calls := 0
		updater.syncDir = func(dir string) error {
			calls++
			if calls == 2 {
				return errors.New("injected post-replacement sync failure")
			}
			return realSync(dir)
		}
		if _, err := updater.Install(digest(newBytes), nil); err == nil {
			t.Fatal("Install succeeded")
		}
		got, err := os.ReadFile(current)
		if err != nil {
			t.Fatalf("current path disappeared: %v", err)
		}
		if digest(got) != digest(oldBytes) {
			t.Fatal("old current bytes were not restored")
		}
	})
}
