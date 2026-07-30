package localupdate

import (
	"bytes"
	"debug/buildinfo"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInspectReportsVCSBuildSettings(t *testing.T) {
	updater, current, _ := newFixture(t)
	candidate := copyTestExecutable(t, current+".candidate", []byte("\ninspect-vcs\n"))
	info, err := buildinfo.ReadFile(current + ".candidate")
	if err != nil {
		t.Fatal(err)
	}
	wantRevision := ""
	wantModified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			wantRevision = setting.Value
		case "vcs.modified":
			wantModified = setting.Value == "true"
		}
	}

	got := updater.Inspect().Candidate
	if got.VCSRevision != wantRevision || got.VCSModified != wantModified {
		t.Fatalf("VCS metadata = revision %q modified %t, want %q/%t",
			got.VCSRevision, got.VCSModified, wantRevision, wantModified)
	}
	if got.SHA256 != digest(candidate) {
		t.Fatalf("candidate digest = %s, want %s", got.SHA256, digest(candidate))
	}
}

func TestStageCandidateValidatesAndAtomicallyPublishesFixedSibling(t *testing.T) {
	updater, current, _ := newFixture(t)
	candidate := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nsigned-release\n"))

	got, err := updater.StageCandidate(bytes.NewReader(candidate), "")
	if err != nil {
		t.Fatalf("StageCandidate: %v", err)
	}
	if got.Metadata.Path != current+".candidate" ||
		got.Metadata.SHA256 != digest(candidate) || got.RecoveryPath != "" {
		t.Fatalf("stage result = %+v", got)
	}
	staged, err := os.ReadFile(current + ".candidate")
	if err != nil || digest(staged) != digest(candidate) {
		t.Fatalf("fixed candidate = %v digest=%s", err, digest(staged))
	}
}

func TestStageCandidateRejectsWrongRevisionAndPreservesEarlierCandidate(t *testing.T) {
	updater, current, _ := newFixture(t)
	oldCandidate := copyTestExecutable(t, current+".candidate", []byte("\nold-candidate\n"))
	next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))

	if _, err := updater.StageCandidate(bytes.NewReader(next), strings.Repeat("f", 40)); !errors.Is(err, ErrProvenanceMismatch) {
		t.Fatalf("StageCandidate = %v, want ErrProvenanceMismatch", err)
	}
	got, err := os.ReadFile(current + ".candidate")
	if err != nil || digest(got) != digest(oldCandidate) {
		t.Fatalf("old candidate changed: err=%v digest=%s", err, digest(got))
	}
}

func TestStageCandidatePublishFailuresPreserveOrRecoverEarlierCandidate(t *testing.T) {
	t.Run("publish rename", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		oldCandidate := copyTestExecutable(t, current+".candidate", []byte("\nold\n"))
		next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))
		realRename := updater.rename
		updater.rename = func(old, new string) error {
			if new == current+".candidate" {
				return errors.New("injected publish rename failure")
			}
			return realRename(old, new)
		}
		if _, err := updater.StageCandidate(bytes.NewReader(next), ""); err == nil {
			t.Fatal("StageCandidate succeeded")
		}
		got, _ := os.ReadFile(current + ".candidate")
		if digest(got) != digest(oldCandidate) {
			t.Fatal("old candidate changed")
		}
	})

	t.Run("directory sync restores", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		oldCandidate := copyTestExecutable(t, current+".candidate", []byte("\nold\n"))
		next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))
		realSync := updater.syncDir
		calls := 0
		updater.syncDir = func(dir string) error {
			calls++
			if calls == 2 {
				return errors.New("injected publish sync failure")
			}
			return realSync(dir)
		}
		if _, err := updater.StageCandidate(bytes.NewReader(next), ""); err == nil {
			t.Fatal("StageCandidate succeeded")
		}
		got, err := os.ReadFile(current + ".candidate")
		if err != nil || digest(got) != digest(oldCandidate) {
			t.Fatalf("old candidate was not restored: err=%v", err)
		}
	})

	t.Run("restore failure retains recovery", func(t *testing.T) {
		updater, current, _ := newFixture(t)
		oldCandidate := copyTestExecutable(t, current+".candidate", []byte("\nold\n"))
		next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))
		realRename := updater.rename
		realSync := updater.syncDir
		syncCalls := 0
		updater.syncDir = func(dir string) error {
			syncCalls++
			if syncCalls == 2 {
				return errors.New("injected publish sync failure")
			}
			return realSync(dir)
		}
		updater.rename = func(old, new string) error {
			if new == current+".candidate" && strings.Contains(filepath.Base(old), ".restore-") {
				return errors.New("injected restore rename failure")
			}
			return realRename(old, new)
		}
		result, err := updater.StageCandidate(bytes.NewReader(next), "")
		if err == nil || result.RecoveryPath == "" {
			t.Fatalf("StageCandidate result=%+v err=%v", result, err)
		}
		recovery, readErr := os.ReadFile(result.RecoveryPath)
		if readErr != nil || digest(recovery) != digest(oldCandidate) {
			t.Fatalf("recovery missing old candidate: err=%v", readErr)
		}
	})
}

func TestStageCandidateWithoutPriorCandidateCleansUpAfterSyncFailure(t *testing.T) {
	updater, current, _ := newFixture(t)
	next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))
	realSync := updater.syncDir
	calls := 0
	updater.syncDir = func(dir string) error {
		calls++
		if calls == 1 {
			return errors.New("injected publish sync failure")
		}
		return realSync(dir)
	}
	if _, err := updater.StageCandidate(bytes.NewReader(next), ""); err == nil {
		t.Fatal("StageCandidate succeeded")
	}
	if _, err := os.Stat(current + ".candidate"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("uncertain new candidate remained: %v", err)
	}
}

func TestUpdaterSerializesInspectStageAndInstall(t *testing.T) {
	updater, current, _ := newFixture(t)
	next := copyTestExecutable(t, filepath.Join(t.TempDir(), "downloaded"), []byte("\nnext\n"))
	entered := make(chan struct{})
	release := make(chan struct{})
	realSync := updater.syncDir
	var once sync.Once
	updater.syncDir = func(dir string) error {
		once.Do(func() {
			close(entered)
			<-release
		})
		return realSync(dir)
	}

	stageDone := make(chan error, 1)
	go func() {
		_, err := updater.StageCandidate(bytes.NewReader(next), "")
		stageDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("stage did not reach publish sync")
	}

	inspectDone := make(chan Inspection, 1)
	go func() { inspectDone <- updater.Inspect() }()
	installDone := make(chan error, 1)
	go func() {
		_, err := updater.Install(digest(next), nil)
		installDone <- err
	}()
	select {
	case <-inspectDone:
		t.Fatal("Inspect observed a half-published stage")
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case err := <-installDone:
		t.Fatalf("Install interleaved with stage: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	if err := <-stageDone; err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-inspectDone:
		if got.Candidate.SHA256 != digest(next) {
			t.Fatalf("Inspect saw digest %s", got.Candidate.SHA256)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Inspect did not resume")
	}
	select {
	case err := <-installDone:
		if err != nil {
			t.Fatalf("Install after stage: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Install did not resume")
	}

	if got, err := os.ReadFile(current); err != nil || digest(got) != digest(next) {
		t.Fatalf("current after serialized install: err=%v", err)
	}
}
