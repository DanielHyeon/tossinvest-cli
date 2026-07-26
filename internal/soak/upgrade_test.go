package soak_test

// upgrade_test.go covers the survey noticing that the binary under it was replaced
// (task 1.8 ②).
//
// The claim that matters is not "it re-executes" — that is one function call — but
// "the record and the streak do not care that it did". The soak's whole value is
// that it ran for three consecutive days, and a self-upgrade that reset the streak
// would be a feature that destroys the thing it is helping to produce. So the test
// that carries the weight runs a survey, upgrades it, runs the successor against
// the same file, and reads the streak back out of the file rather than out of
// either process.
//
// Nothing here executes anything. ReExec is a seam; the fake records the call.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// build fabricates a fingerprint for a notional install.
func build(minute int) binstamp.Stamp {
	return binstamp.Stamp{
		Path:    "/opt/tossos/bin/tossctl",
		Size:    int64(9_000_000 + minute),
		ModTime: soakStart.Add(time.Duration(minute) * time.Minute),
	}
}

// installer is the Binary seam: it reports whatever is "installed" right now. It
// is mutex-guarded because a test reinstalls from one goroutine while the survey
// reads from another, which is exactly the situation the feature is about.
type installer struct {
	mu    sync.Mutex
	stamp binstamp.Stamp
	err   error
}

func newInstaller(s binstamp.Stamp) *installer { return &installer{stamp: s} }

func (i *installer) reinstall(s binstamp.Stamp) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.stamp = s
}

func (i *installer) breaks(err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.err = err
}

func (i *installer) fn() func() (binstamp.Stamp, error) {
	return func() (binstamp.Stamp, error) {
		i.mu.Lock()
		defer i.mu.Unlock()
		return i.stamp, i.err
	}
}

// execSeam records a re-exec instead of performing one.
type execSeam struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (e *execSeam) fn() func() error {
	return func() error {
		e.mu.Lock()
		defer e.mu.Unlock()
		e.calls++
		return e.err
	}
}

func (e *execSeam) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.calls
}

// TestEveryCycleStampsTheBuildThatWroteIt.
//
// This is what the dashboard reads to answer "is the running soak the installed
// soak" — the simplest honest source, because it is the survey's own statement
// about itself rather than a guess made from outside.
func TestEveryCycleStampsTheBuildThatWroteIt(t *testing.T) {
	inst := newInstaller(build(0))
	rec, err := soak.OpenRecorder(t.TempDir() + "/soak.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 2
		o.Binary = inst.fn()
	})
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	cycles, err := soak.LoadCycles(rec.Path())
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	for i, c := range cycles {
		if !c.Binary.Same(build(0)) || !c.Binary.Known() {
			t.Errorf("cycle %d recorded %+v, want the build it was running", i, c.Binary)
		}
	}
	if got := soak.Summarize(cycles).Binary; !got.Same(build(0)) {
		t.Errorf("Summarize reported %+v as the running build", got)
	}
}

// TestASurveyWithNoFingerprintSeamRecordsNoneAndUpgradesNothing — the pre-1.8
// behaviour, still a valid way to run.
func TestASurveyWithNoFingerprintSeamRecordsNoneAndUpgradesNothing(t *testing.T) {
	seam := &execSeam{}
	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Cycles = 1
		o.ReExec = seam.fn()
	})
	cycle, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
	if cycle.Binary.Known() {
		t.Errorf("a survey with no Binary seam stamped %+v", cycle.Binary)
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if seam.count() != 0 {
		t.Errorf("the re-exec seam was called %d time(s) with nothing to compare against", seam.count())
	}
}

// TestTheUpgradeHappensAtACycleBoundaryAndNotInsideOne.
//
// A cycle that was interrupted is a cycle that did not happen, and the streak is
// counted out of recorded cycles. The handover therefore waits until the current
// cycle is written and synced.
func TestTheUpgradeHappensAtACycleBoundaryAndNotInsideOne(t *testing.T) {
	inst := newInstaller(build(0))
	seam := &execSeam{}
	rec, err := soak.OpenRecorder(t.TempDir() + "/soak.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	progress := &strings.Builder{}
	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 0 // until interrupted, which is how a real soak runs
		o.Binary = inst.fn()
		o.ReExec = seam.fn()
		o.Progress = progress
	})

	// Reinstall between the first cycle and the second.
	inst.reinstall(build(30))

	err = r.Run(context.Background())
	if !errors.Is(err, soak.ErrUpgraded) {
		t.Fatalf("Run returned %v, want ErrUpgraded", err)
	}
	if seam.count() != 1 {
		t.Fatalf("the re-exec seam was called %d time(s), want exactly 1", seam.count())
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	cycles, err := soak.LoadCycles(rec.Path())
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("%d cycle(s) recorded before the handover, want the one that was in flight", len(cycles))
	}
	if !cycles[0].Binary.Same(build(0)) {
		t.Errorf("the last cycle of the old process is stamped %+v, want the build it was running",
			cycles[0].Binary)
	}
	if !strings.Contains(progress.String(), "바이너리가 바뀌었다") {
		t.Errorf("the handover was silent:\n%s", progress.String())
	}
}

// TestTheRecordAndTheStreakSurviveAnUpgrade is the claim task 1.8 asks to be shown.
//
// Two processes, one file. The second is a different build and knows nothing about
// the first except what is on disk — which is the same situation as a reboot, and
// the reason the record is append-only JSON Lines in the first place.
//
// The clock is the fake, so a "day" between cycles costs nothing and the reinstall
// lands at a point the test chooses rather than a point it hopes for.
func TestTheRecordAndTheStreakSurviveAnUpgrade(t *testing.T) {
	path := t.TempDir() + "/soak.jsonl"
	const day = 24 * time.Hour

	// --- the old build: three daily cycles, then a reinstall --------------------
	inst := newInstaller(build(0))
	seam := &execSeam{}

	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	first, fake := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 0 // until interrupted, which is how a real soak runs
		o.Interval = day
		o.Binary = inst.fn()
		o.ReExec = seam.fn()
	})

	done := make(chan error, 1)
	go func() { done <- first.Run(context.Background()) }()

	// Day one lands, then the survey parks on its interval.
	waitForSleeper(t, fake)
	fake.Advance(day) // day two
	waitForSleeper(t, fake)
	// The operator reinstalls while the survey is between cycles.
	inst.reinstall(build(45))
	fake.Advance(day) // day three, then the boundary notices the new build

	select {
	case err := <-done:
		if !errors.Is(err, soak.ErrUpgraded) {
			t.Fatalf("the first process ended with %v, want ErrUpgraded", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the first process never handed over")
	}
	if seam.count() != 1 {
		t.Fatalf("the re-exec seam was called %d time(s), want exactly 1", seam.count())
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	before, err := soak.LoadCycles(path)
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(before) != 3 {
		t.Fatalf("the first process recorded %d cycle(s), want 3", len(before))
	}
	summaryBefore := soak.Summarize(before)
	if summaryBefore.StreakDays != 3 {
		t.Fatalf("the streak before the upgrade is %d, want 3", summaryBefore.StreakDays)
	}

	// --- the successor: same path, new build, appends ---------------------------
	rec2, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("re-opening the record: %v", err)
	}
	defer rec2.Close()

	successor, fake2 := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec2
		o.Cycles = 2
		o.Interval = day
		o.Binary = newInstaller(build(45)).fn()
		o.ReExec = func() error { t.Error("the successor tried to upgrade itself again"); return nil }
	})
	// The successor starts where the predecessor stopped: two more days.
	fake2.Advance(3 * day)
	successorDone := make(chan error, 1)
	go func() { successorDone <- successor.Run(context.Background()) }()
	waitForSleeper(t, fake2)
	fake2.Advance(day)

	select {
	case err := <-successorDone:
		if err != nil {
			t.Fatalf("the successor's Run: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the successor never finished")
	}
	if err := rec2.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	after, err := soak.LoadCycles(path)
	if err != nil {
		t.Fatalf("LoadCycles after the upgrade: %v", err)
	}
	if len(after) != len(before)+2 {
		t.Fatalf("the successor appended %d cycle(s), want 2", len(after)-len(before))
	}
	// Every line the old process wrote is still there, unchanged.
	for i := range before {
		if !after[i].StartedAt.Equal(before[i].StartedAt) {
			t.Fatalf("cycle %d changed across the upgrade: %s vs %s",
				i, before[i].StartedAt, after[i].StartedAt)
		}
		if !after[i].Binary.Same(build(0)) {
			t.Errorf("cycle %d lost the build that wrote it: %+v", i, after[i].Binary)
		}
	}

	summary := soak.Summarize(after)
	if summary.StreakDays < summaryBefore.StreakDays {
		t.Errorf("the streak went backwards across an upgrade: %d then %d",
			summaryBefore.StreakDays, summary.StreakDays)
	}
	if summary.StreakDays != 5 {
		t.Errorf("the streak across the upgrade is %d, want 5 — the record is one file and the process "+
			"that wrote a day is not part of the claim", summary.StreakDays)
	}
	if summary.All.Cycles != len(after) {
		t.Errorf("the summary counts %d cycles over a %d-line record", summary.All.Cycles, len(after))
	}
	// And the record now reports the build that is actually running.
	if !summary.Binary.Same(build(45)) {
		t.Errorf("the record reports %+v as the running build, want the successor's", summary.Binary)
	}
}

// waitForSleeper blocks until the survey is parked on its interval, which is the
// only moment a test may move the clock without racing a cycle.
func waitForSleeper(t *testing.T, fake *clock.Fake) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if fake.Sleepers() > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the survey never parked on its interval")
}

// TestAFailedReExecKeepsTheSurveyRunning.
//
// Being one build behind is a note on a dashboard. Stopping the survey over it
// would cost a day of a three-day streak, which is the expensive direction.
func TestAFailedReExecKeepsTheSurveyRunning(t *testing.T) {
	inst := newInstaller(build(0))
	seam := &execSeam{err: errors.New("permission denied")}
	progress := &strings.Builder{}

	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Cycles = 3
		o.Binary = inst.fn()
		o.ReExec = seam.fn()
		o.Progress = progress
	})
	inst.reinstall(build(10))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("a failed re-exec ended the run with %v; it should carry on", err)
	}
	if seam.count() == 0 {
		t.Fatal("the re-exec seam was never tried")
	}
	if !strings.Contains(progress.String(), "자기 재실행 실패") {
		t.Errorf("the failure was not reported:\n%s", progress.String())
	}
}

// TestAnUnreadableInstallIsNotAnUpgrade — an unanswerable question must not become
// a handover.
func TestAnUnreadableInstallIsNotAnUpgrade(t *testing.T) {
	inst := newInstaller(build(0))
	seam := &execSeam{}
	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Cycles = 2
		o.Binary = inst.fn()
		o.ReExec = seam.fn()
	})
	inst.breaks(errors.New("the filesystem went away"))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if seam.count() != 0 {
		t.Errorf("a failed stat produced %d re-exec(s)", seam.count())
	}
}
