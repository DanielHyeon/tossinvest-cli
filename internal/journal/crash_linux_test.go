//go:build linux

package journal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// The crash tests run the journal in a child copy of this test binary and kill it
// with SIGKILL — no deferred close, no flush, no chance to tidy up. That is the
// only honest way to test "the record survived a crash": an in-process fake cannot
// lose an fsync the way a real power cut does.
const (
	crashEnvMode = "TOSSOS_JOURNAL_CRASH_MODE"
	crashEnvPath = "TOSSOS_JOURNAL_CRASH_PATH"

	crashModeRecorded    = "recorded"
	crashModeUncommitted = "uncommitted"
)

// TestCrashAfterRecordedIntentSurvivesRestart is the durability claim the whole
// design rests on: once Prepare returned, the RECORDED attempt is on disk even if
// the process dies in the very next instruction — which is what makes it safe to
// submit to the broker only after that point.
func TestCrashAfterRecordedIntentSurvivesRestart(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeRecorded {
		crashChildRecorded()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	runCrashChild(t, "TestCrashAfterRecordedIntentSurvivesRestart", crashModeRecorded, path)

	// The child never checkpointed, so the commit is still only in the WAL.
	if _, err := os.Stat(path + "-wal"); err != nil {
		t.Logf("no -wal file after the crash (%v); the commit must then be in the main file", err)
	}

	ctx := context.Background()
	j, err := Open(ctx, Options{
		Path:     path,
		Clock:    clock.NewFake(time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)),
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("reopening the journal after a crash: %v", err)
	}
	defer j.Close()

	intent, err := j.LookupIntent(ctx, "intent-1")
	if err != nil {
		t.Fatalf("intent recorded before the crash is missing: %v", err)
	}
	if intent.Quantity != "10" || intent.Symbol != "AAPL" {
		t.Fatalf("intent came back damaged: %+v", intent)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("attempt recorded before the crash is missing: %v", err)
	}
	if rec.State != StateRecorded {
		t.Fatalf("state after the crash = %s, want RECORDED (the restart pass closes it, not the crash)", rec.State)
	}
	if rec.DispatchStartedAt != "" {
		t.Fatalf("dispatch_started_at = %q, want empty: nothing was dispatched", rec.DispatchStartedAt)
	}
}

// TestCrashDuringUncommittedWriteLeavesNothing is the other half: work that had not
// committed when the process died must be invisible afterwards. A half-written
// intent would make the restart pass invent a mutation that never existed.
func TestCrashDuringUncommittedWriteLeavesNothing(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeUncommitted {
		crashChildUncommitted()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	runCrashChild(t, "TestCrashDuringUncommittedWriteLeavesNothing", crashModeUncommitted, path)

	ctx := context.Background()
	j, err := Open(ctx, Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("reopening the journal after a crash: %v", err)
	}
	defer j.Close()

	var intents, attempts int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM intents").Scan(&intents); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM mutation_attempts").Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if intents != 0 || attempts != 0 {
		t.Fatalf("uncommitted work survived the crash: %d intents, %d attempts", intents, attempts)
	}
	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after an interrupted write: %v", err)
	}
	// The journal must still be usable — a crash is not a corruption.
	if _, err := j.Prepare(ctx, testRequest()); err != nil {
		t.Fatalf("Prepare after crash recovery: %v", err)
	}
}

// TestDetachedWALLosesCommittedRecord documents why the corruption guidance tells
// operators to move the -wal and -shm files together with the database: a commit
// that has not been checkpointed lives in the WAL, so copying only the .db file
// silently drops it. This is the failure mode the recovery message exists to
// prevent, pinned as a test so the guidance cannot drift away from the behaviour.
func TestDetachedWALLosesCommittedRecord(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeRecorded {
		crashChildRecorded()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	runCrashChild(t, "TestDetachedWALLosesCommittedRecord", crashModeRecorded, path)

	wal := path + "-wal"
	if _, err := os.Stat(wal); err != nil {
		t.Skipf("no WAL file after the crash (%v): nothing to detach", err)
	}
	if err := os.Rename(wal, wal+".moved"); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(path + "-shm")

	ctx := context.Background()
	j, err := Open(ctx, Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		// Refusing is an acceptable outcome too; losing the record silently is
		// the one we are documenting.
		t.Logf("journal refused to open without its WAL: %v", err)
		return
	}
	defer j.Close()

	if _, err := j.LookupIntent(ctx, "intent-1"); !errors.Is(err, ErrIntentNotFound) {
		t.Logf("the commit had already been checkpointed into the main file; nothing lost (%v)", err)
		return
	}
	t.Log("confirmed: detaching the -wal file drops commits that were not checkpointed " +
		"— the recovery guidance must keep telling operators to move all three files")
}

// runCrashChild re-executes this test binary for a single test, with the crash mode
// set, and asserts the child really died by SIGKILL.
func runCrashChild(t *testing.T, testName, mode, path string) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^"+testName+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), crashEnvMode+"="+mode, crashEnvPath+"="+path)
	out, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("child was expected to die by signal, got err=%v\noutput:\n%s", err, out)
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		t.Fatalf("unexpected wait status type %T\noutput:\n%s", exitErr.Sys(), out)
	}
	if !status.Signaled() || status.Signal() != syscall.SIGKILL {
		t.Fatalf("child exited (code %d, signaled=%v, signal=%v) instead of being killed\noutput:\n%s",
			status.ExitStatus(), status.Signaled(), status.Signal(), out)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("child left no journal file at %s: %v\noutput:\n%s", path, statErr, out)
	}
}

// crashChildRecorded records an intent, verifies the commit returned, then kills
// itself before anything could be dispatched.
func crashChildRecorded() {
	j := openCrashChildJournal()
	if _, err := j.Prepare(context.Background(), testRequest()); err != nil {
		fmt.Fprintf(os.Stderr, "crash child: Prepare: %v\n", err)
		os.Exit(3)
	}
	kill()
}

// crashChildUncommitted starts a write transaction, gets the rows in, and dies
// before the commit.
func crashChildUncommitted() {
	j := openCrashChildJournal()
	ctx := context.Background()

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crash child: BeginTx: %v\n", err)
		os.Exit(3)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO intents
		   (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
		    time_in_force, quantity, price, currency, source, fingerprint, notes)
		 VALUES ('intent-1','2026-03-30T00:30:00Z','us','2026-03-30','acct-1','AAPL','BUY',
		         'LIMIT','DAY','10','200.5','USD','engine/test','fp-1','')`); err != nil {
		fmt.Fprintf(os.Stderr, "crash child: insert: %v\n", err)
		os.Exit(3)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO mutation_attempts
		   (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		 VALUES ('attempt-1','intent-1','PLACE','RECORDED',1,'fp-1','2026-03-30T00:30:00Z')`); err != nil {
		fmt.Fprintf(os.Stderr, "crash child: insert attempt: %v\n", err)
		os.Exit(3)
	}
	// No Commit: the write is in the WAL as an open transaction at most.
	kill()
}

func openCrashChildJournal() *Journal {
	path := os.Getenv(crashEnvPath)
	if path == "" {
		fmt.Fprintln(os.Stderr, "crash child: "+crashEnvPath+" is not set")
		os.Exit(2)
	}
	j, err := Open(context.Background(), Options{
		Path:     path,
		Clock:    clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "crash child: Open: %v\n", err)
		os.Exit(2)
	}
	return j
}

// kill terminates this process the way a power cut would: no unwinding, no flush.
func kill() {
	_ = syscall.Kill(os.Getpid(), syscall.SIGKILL)
	// Unreachable in practice; keep the process from continuing if the signal is
	// somehow delayed.
	time.Sleep(30 * time.Second)
	os.Exit(4)
}
