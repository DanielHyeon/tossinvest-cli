//go:build linux

package journal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const (
	crashModeExitSnapshotBeforeCommit = "exit_snapshot_before_commit"
	crashModeExitSnapshotAfterCommit  = "exit_snapshot_after_commit"
)

func TestExitSnapshotTransactionSurvivesSIGKILLAtomically(t *testing.T) {
	switch os.Getenv(crashEnvMode) {
	case crashModeExitSnapshotBeforeCommit:
		exitSnapshotCrashChild(t, true)
		return
	case crashModeExitSnapshotAfterCommit:
		exitSnapshotCrashChild(t, false)
		return
	}

	for _, test := range []struct {
		name        string
		mode        string
		observation string
		pending     bool
		eventCount  int
	}{
		{"before_commit", crashModeExitSnapshotBeforeCommit, "obs-crash-prior", false, 2},
		{"after_commit", crashModeExitSnapshotAfterCommit, "obs-crash-next", true, 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.db")
			runCrashChild(t, "TestExitSnapshotTransactionSurvivesSIGKILLAtomically", test.mode, path)
			assertCrashJournalArtifacts(t, path)

			j := openTestJournalAt(t, path)
			results, err := j.OpenExitStateResults(context.Background(), "acct-1")
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || results[0].Corruption != nil || results[0].State.Snapshot.Snapshot == nil {
				t.Fatalf("recovered state = %+v", results)
			}
			state := results[0].State
			if state.Snapshot.Snapshot.Line.ObservationID != test.observation || state.Pending() != test.pending {
				t.Fatalf("recovered snapshot/pending = %s/%v, want %s/%v",
					state.Snapshot.Snapshot.Line.ObservationID, state.Pending(), test.observation, test.pending)
			}
			events, err := j.ExitEvents(context.Background(), state.PositionID)
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != test.eventCount {
				t.Fatalf("events = %d, want %d complete commits", len(events), test.eventCount)
			}
			last := events[len(events)-1]
			if last.Evaluation.Effective.Snapshot == nil ||
				last.Evaluation.Effective.Snapshot.Line.ObservationID != test.observation {
				t.Fatalf("event/state snapshot coherence = %+v", last)
			}
			if test.pending && (last.ProposedIntentID != "exit-crash" ||
				state.PendingIntentID != "exit-crash") {
				t.Fatalf("committed arm did not survive with event/state: state=%+v event=%+v", state, last)
			}
		})
	}
}

func exitSnapshotCrashChild(t *testing.T, killBeforeCommit bool) {
	t.Helper()
	j := openCrashChildJournal()
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	_, seed := openedPosition(t, j, "10")
	prior := ratchetSnapshotForState(t, seed, "obs-crash-prior", "70500", "70500", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(prior)); err != nil {
		t.Fatalf("prior snapshot: %v", err)
	}

	next := ratchetSnapshotForState(t, seed, "obs-crash-next", "67900", "70500", "68000")
	judgement := judgementForSnapshot(next)
	judgement.Proposal = &ExitProposal{
		Action: string(next.Action), Level: next.Level, IntentID: "exit-crash",
		Provenance: judgement.Provenance,
	}
	if killBeforeCommit {
		j.exitWriteHook = func(stage string) error {
			if stage == "after_event" {
				kill()
			}
			return nil
		}
	}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatalf("next snapshot: %v", err)
	}
	if !killBeforeCommit {
		kill()
	}
}

func assertCrashJournalArtifacts(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("crash left no main database: %v", err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(path + suffix); err != nil {
			t.Logf("no %s artifact after crash (%v); SQLite checkpointed it into the main database", suffix, err)
		}
	}
	// TestDetachedWALLosesCommittedRecord separately pins the backup/recovery
	// rule: when these sidecars exist, the database, WAL, and SHM are one unit.
}
