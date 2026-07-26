package enginelock_test

// enginelock_test.go pins the two halves apart: the flock is an exclusion and the
// marker is not, and each has to fail in its own direction.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

var lockNow = time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)

func unixOnly(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" || runtime.GOOS == "js" {
		t.Skipf("the file lock is Unix-only by design; %s refuses honestly", runtime.GOOS)
	}
}

// --- the exclusion ----------------------------------------------------------------

// TestASecondAcquireIsRefused is the scenario "두 번째 인스턴스 기동": the second
// caller is told, immediately, that an engine already holds this journal.
func TestASecondAcquireIsRefused(t *testing.T) {
	unixOnly(t)
	dir := t.TempDir()

	first, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	t.Cleanup(first.Release)

	second, err := enginelock.Acquire(dir)
	if !errors.Is(err, enginelock.ErrAlreadyRunning) {
		t.Fatalf("second Acquire returned (%v, %v), want ErrAlreadyRunning", second, err)
	}
	if second != nil {
		t.Fatal("a refused Acquire handed back a lock")
	}
}

// TestReleasingLetsTheNextInstanceIn, and releasing twice is safe — the command
// defers a release it may also have run on a refusal path.
func TestReleasingLetsTheNextInstanceIn(t *testing.T) {
	unixOnly(t)
	dir := t.TempDir()

	first, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	first.Release()
	first.Release()

	second, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire after Release: %v", err)
	}
	second.Release()
}

// TestTheLockIsTakenBeforeTheDirectoryExists. The lock is the first action of a
// start, which on a first run is before the journal directory has been created.
func TestTheLockIsTakenBeforeTheDirectoryExists(t *testing.T) {
	unixOnly(t)
	dir := filepath.Join(t.TempDir(), "not", "created", "yet")

	lock, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	defer lock.Release()
	if lock.Path() != enginelock.LockPath(dir) {
		t.Errorf("lock path = %s, want %s", lock.Path(), enginelock.LockPath(dir))
	}
	if _, err := os.Stat(lock.Path()); err != nil {
		t.Errorf("the lock file was not created: %v", err)
	}
}

// TestTheLockSitsInTheJournalDirectory. It protects the journal, so it lives with
// it — an isolated --config-dir profile gets its own and cannot exclude the real
// engine, and the real engine cannot exclude a test.
func TestTheLockSitsInTheJournalDirectory(t *testing.T) {
	if got, want := enginelock.LockPath("/data/tossos"), "/data/tossos/engine.lock"; got != want {
		t.Errorf("LockPath = %s, want %s", got, want)
	}
	if got, want := enginelock.MarkerPath("/data/tossos"), "/data/tossos/engine-run.json"; got != want {
		t.Errorf("MarkerPath = %s, want %s", got, want)
	}
}

// --- the advisory marker ------------------------------------------------------------

// TestAHeldMarkerReadsAsRunning, and the reading carries the pid and the build so
// the console can draw both.
func TestAHeldMarkerReadsAsRunning(t *testing.T) {
	path := enginelock.MarkerPath(t.TempDir())

	release, err := enginelock.Hold(context.Background(), path, lockNow)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer release()

	status := enginelock.Read(path, lockNow.Add(time.Second))
	if !status.Running {
		t.Fatal("a marker written a second ago does not read as running")
	}
	if status.Marker.PID != os.Getpid() {
		t.Errorf("marker pid = %d, want %d", status.Marker.PID, os.Getpid())
	}
	if !status.Marker.StartedAt.Equal(lockNow) {
		t.Errorf("started_at = %s, want %s", status.Marker.StartedAt, lockNow)
	}
	if status.Marker.Note == "" {
		t.Error("the marker carries no note; somebody who finds it with cat has to be told what it is")
	}
}

// TestAStaleMarkerIsNotARunningEngine: five minutes without a refresh and a
// crashed engine stops being believed. The cost of the mistake is a status line,
// which is why the window can be minutes at all.
func TestAStaleMarkerIsNotARunningEngine(t *testing.T) {
	path := enginelock.MarkerPath(t.TempDir())
	release, err := enginelock.Hold(context.Background(), path, lockNow)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer release()

	if status := enginelock.Read(path, lockNow.Add(enginelock.StaleAfter-time.Second)); !status.Running {
		t.Error("a marker inside the staleness window does not read as running")
	}
	if status := enginelock.Read(path, lockNow.Add(enginelock.StaleAfter+time.Second)); status.Running {
		t.Error("a stale marker still reads as a running engine")
	}
}

// TestTheMarkerNumbersAreRunlocksPrecedent. The spec fixes them (갱신 1분·stale
// 5분 — runlock 선례 수치) and a drift here would make the console, the autostart
// script and the verification's own marker disagree about what "recently" means.
func TestTheMarkerNumbersAreRunlocksPrecedent(t *testing.T) {
	if enginelock.StaleAfter != runlock.StaleAfter {
		t.Errorf("StaleAfter = %s, want runlock's %s", enginelock.StaleAfter, runlock.StaleAfter)
	}
	if enginelock.RefreshEvery != runlock.RefreshEvery {
		t.Errorf("RefreshEvery = %s, want runlock's %s", enginelock.RefreshEvery, runlock.RefreshEvery)
	}
	if enginelock.StaleAfter != 5*time.Minute || enginelock.RefreshEvery != time.Minute {
		t.Errorf("the numbers moved: refresh %s, stale %s (spec: 1m / 5m)",
			enginelock.RefreshEvery, enginelock.StaleAfter)
	}
}

// TestReleasingRemovesTheMarker: a clean stop leaves no signal behind, so the
// console says "stopped" at once rather than for the next five minutes saying
// "running".
func TestReleasingRemovesTheMarker(t *testing.T) {
	path := enginelock.MarkerPath(t.TempDir())
	release, err := enginelock.Hold(context.Background(), path, lockNow)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	release()
	release() // idempotent

	if status := enginelock.Read(path, lockNow); status.Running {
		t.Error("the marker survived a clean stop")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the marker file is still there: %v", err)
	}
}

// TestAnUnparseableMarkerIsStillARunningEngine. The reader is a dashboard: a
// format it does not recognise must degrade to "running, build unknown", never to
// an error page or to "stopped".
func TestAnUnparseableMarkerIsStillARunningEngine(t *testing.T) {
	path := enginelock.MarkerPath(t.TempDir())
	if err := os.WriteFile(path, []byte("pid 12345\nrefreshed whenever\n"), 0o600); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}
	if err := os.Chtimes(path, lockNow, lockNow); err != nil {
		t.Fatalf("timestamping: %v", err)
	}

	status := enginelock.Read(path, lockNow.Add(time.Second))
	if !status.Running {
		t.Error("an unparseable marker read as a stopped engine")
	}
	if status.Marker.Binary.Known() {
		t.Error("an unparseable marker claimed to know the build")
	}
	if status.Stale(mustStamp(t)) {
		t.Error("an unknown build read as stale; 'we could not look' must not become 'it moved on'")
	}
}

// TestAMissingMarkerIsNotAnError.
func TestAMissingMarkerIsNotAnError(t *testing.T) {
	status := enginelock.Read(filepath.Join(t.TempDir(), "nothing.json"), lockNow)
	if status.Running {
		t.Error("a missing marker reads as a running engine")
	}
	if status := enginelock.Read("", lockNow); status.Running {
		t.Error("an unset marker path reads as a running engine")
	}
}

// TestStaleReportsAnEngineOlderThanTheInstalledBuild — the dashboard's
// stale-binary warning, answered from the engine's own statement about itself.
func TestStaleReportsAnEngineOlderThanTheInstalledBuild(t *testing.T) {
	running := binstamp.Stamp{Path: "/usr/local/bin/tossctl", Size: 100, ModTime: lockNow}
	installed := binstamp.Stamp{Path: "/usr/local/bin/tossctl", Size: 200, ModTime: lockNow.Add(time.Hour)}

	status := enginelock.Status{Running: true, Marker: enginelock.Marker{Binary: running}}
	if !status.Stale(installed) {
		t.Error("an engine started from an older executable does not read as stale")
	}
	if status.Stale(running) {
		t.Error("an engine running the installed build reads as stale")
	}

	stopped := enginelock.Status{Marker: enginelock.Marker{Binary: running}}
	if stopped.Stale(installed) {
		t.Error("a stopped engine reads as a stale running one")
	}
}

func mustStamp(t *testing.T) binstamp.Stamp {
	t.Helper()
	s, err := binstamp.Self()
	if err != nil {
		t.Fatalf("binstamp.Self: %v", err)
	}
	return s
}
