package main

// runlock_test.go covers the wiring of the soak/verify rate-budget marker
// (task 1.7 ③).
//
// internal/runlock owns the file and internal/soak owns the pause; what only this
// layer can get wrong is which path the two of them agree on. A verification that
// wrote its marker somewhere the soak never looks would be worse than no marker at
// all — it would look like it was working.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// TestTheVerifyAndSoakSidesAgreeOnTheMarkerPath. Both derive it from the verify
// record's location, so an isolated --config-dir profile gets its own.
func TestTheVerifyAndSoakSidesAgreeOnTheMarkerPath(t *testing.T) {
	dir := t.TempDir()
	root := &rootOptions{configDir: dir}

	record, err := resolveVerifyRecord(root, "")
	if err != nil {
		t.Fatalf("resolveVerifyRecord: %v", err)
	}
	if got, want := record, filepath.Join(dir, verifylive.FileName); got != want {
		t.Fatalf("the evidence record resolved to %s, want %s", got, want)
	}
	if got, want := verifyRunLockPath(record), filepath.Join(dir, runlock.FileName); got != want {
		t.Errorf("the marker resolved to %s, want %s — it lives beside the record", got, want)
	}
}

func TestThePauseReaderIgnoresAMissingAndAStaleMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), runlock.FileName)
	pause := verifyRunLockPause(path)

	if paused, _ := pause(); paused {
		t.Error("the soak paused with no verification running")
	}

	lock, err := runlock.Acquire(path, time.Now().Add(-2*runlock.StaleAfter))
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer lock.Release()
	if paused, _ := pause(); paused {
		t.Error("the soak paused on a stale marker; a crashed verification must not wedge the survey")
	}
}

func TestThePauseReaderYieldsToAFreshMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), runlock.FileName)
	lock, err := runlock.Acquire(path, time.Now())
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer lock.Release()

	paused, reason := verifyRunLockPause(path)()
	if !paused {
		t.Fatal("the soak did not yield to a live verification")
	}
	if !strings.Contains(reason, "verification") {
		t.Errorf("the pause reason does not say what it is waiting for: %q", reason)
	}
}

// TestHoldingTheMarkerTakesItAndGivesItBack.
func TestHoldingTheMarkerTakesItAndGivesItBack(t *testing.T) {
	dir := t.TempDir()
	record := filepath.Join(dir, verifylive.FileName)
	var out strings.Builder

	release := holdVerifyRunLock(context.Background(), &out, record)
	if paused, _ := verifyRunLockPause(verifyRunLockPath(record))(); !paused {
		t.Error("the marker was not taken; a concurrent soak would keep spending the rate budget")
	}
	if !strings.Contains(out.String(), "soak pause") {
		t.Errorf("the run did not tell the operator about the marker:\n%s", out.String())
	}

	release()
	if _, err := os.Stat(verifyRunLockPath(record)); !os.IsNotExist(err) {
		t.Errorf("the marker outlived the run: %v", err)
	}
}

// TestAnUnwritableMarkerDoesNotStopAVerification. It is advisory in both
// directions; a live procedure a person is standing over is not blocked by a lock
// file.
func TestAnUnwritableMarkerDoesNotStopAVerification(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing the blocker: %v", err)
	}
	var out strings.Builder
	release := holdVerifyRunLock(context.Background(), &out, filepath.Join(blocker, verifylive.FileName))
	if release == nil {
		t.Fatal("holdVerifyRunLock returned no release; the caller defers it unconditionally")
	}
	release()
	if !strings.Contains(out.String(), "could not be written") {
		t.Errorf("the failure was silent:\n%s", out.String())
	}
}

// --- end to end ------------------------------------------------------------------

// TestVerifyRunTakesAndReleasesTheMarker. The run has no terminal, so it stops at
// the approval — but the marker has to be taken before that and gone after it.
func TestVerifyRunTakesAndReleasesTheMarker(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newVerifyServer(t)
	pointVerifyAt(t, srv, filepath.Join(configDir, "token.json"))

	out, _, _ := runCLI(t, "--config-dir", configDir, "verify", "run")

	lockPath := filepath.Join(configDir, runlock.FileName)
	if !strings.Contains(out, lockPath) {
		t.Errorf("the run did not name the marker it took:\n%s", out)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Errorf("the marker survived the run: %v", err)
	}
}

// TestSoakRunAnnouncesTheMarkerItWatches, and runs straight through a stale one.
//
// The fresh-marker pause is internal/soak's test, driven by a fake clock. What is
// worth checking here is the pair of things only the wiring can break: that the
// survey is watching the file the verification writes, and that a leftover marker
// from a crashed run does not stop it.
func TestSoakRunAnnouncesTheMarkerAndIgnoresAStaleOne(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	lockPath := filepath.Join(configDir, runlock.FileName)
	lock, err := runlock.Acquire(lockPath, time.Now().Add(-2*runlock.StaleAfter))
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer lock.Release()

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0")
	if err != nil {
		t.Fatalf("soak run: %v", err)
	}
	if !strings.Contains(out, lockPath) {
		t.Errorf("the soak does not say which marker it pauses on:\n%s", out)
	}
	if strings.Contains(out, "paused") {
		t.Errorf("the soak paused on a stale marker:\n%s", out)
	}

	cycles, err := soak.LoadCycles(filepath.Join(configDir, soak.FileName))
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("recorded %d cycle(s), want 1 — a stale marker must not hold the survey", len(cycles))
	}
}
