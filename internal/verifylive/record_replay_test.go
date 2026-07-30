package verifylive

// record_replay_test.go replays the operator's real evidence records through the
// new terminal-state logic and checks that nothing about them moved.
//
// The unit tests above prove the rule on records this file wrote. This proves it
// on the ones a real account produced over five days of live runs, which is the
// data every entry in measurements.md is derived from and the data the next run's
// cleanup prologue will actually read.
//
// It is an A/B rather than a snapshot. The change replaced "cancelled" with
// "cancelled or filled" in the one function that decides what is live, so the test
// recomputes both answers — the new predicate and a local copy of the old one —
// and requires them to agree everywhere. A snapshot of expected identifiers would
// have to be updated by hand and would pass whatever it was updated to.
//
// The records live in the operator's data directory and are not in the repository:
// they carry the identifiers of real orders on a real account. Absent, the test
// skips, and the unit tests are what protect a machine that does not have them.

import (
	"os"
	"path/filepath"
	"testing"
)

// realRecords returns the paths of the live evidence records, if this machine has
// them.
func realRecords(t *testing.T) []string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	dir := filepath.Join(home, ".local", "share", "tossos")
	var out []string
	for _, name := range []string{FileName, USFileName} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			out = append(out, path)
		}
	}
	return out
}

// legacyOutstanding is the rule exactly as it stood before a fill could end an
// object: the last line per identifier wins, a cancel is monotone, and a
// non-cancelled artifact is live.
func legacyOutstanding(entries []Entry) []Artifact {
	order := []string{}
	latest := map[string]Artifact{}
	for _, e := range entries {
		for _, a := range e.Artifacts {
			key := a.Kind + "\x00" + a.ID
			if _, seen := latest[key]; !seen {
				order = append(order, key)
			}
			if prev, seen := latest[key]; seen && prev.Cancelled && !a.Cancelled {
				continue
			}
			latest[key] = a
		}
	}
	var out []Artifact
	for _, key := range order {
		if a := latest[key]; !a.Cancelled {
			out = append(out, a)
		}
	}
	return out
}

func TestTheRealRecordsAreJudgedExactlyAsBefore(t *testing.T) {
	paths := realRecords(t)
	if len(paths) == 0 {
		t.Skip("no live evidence record on this machine; the unit tests cover the rule")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			entries, err := LoadEntries(path)
			if err != nil {
				t.Fatalf("LoadEntries(%s): %v", path, err)
			}
			if len(entries) == 0 {
				t.Fatalf("%s decoded to nothing; the replay would prove not much", path)
			}

			// Every line on these records predates the field, so the new predicate
			// has to reduce to the old one by construction. If this ever fails the
			// A/B below is comparing two different questions.
			for i, e := range entries {
				for _, a := range e.Artifacts {
					if a.Filled || !a.FilledAt.IsZero() {
						t.Fatalf("entry %d already carries a fill on %s %s; this record was written by a "+
							"newer build than the one this regression is about", i, a.Kind, a.ID)
					}
				}
			}

			// Both records happen to end clean, so comparing only the survivors
			// would compare two empty lists and prove nothing. Every artifact line
			// on the record is classified by both rules instead, which reaches the
			// cancels, the amend supersessions and the deliberate holds as well.
			lines := 0
			for _, e := range entries {
				for _, a := range e.Artifacts {
					lines++
					if a.terminal() != a.Cancelled {
						t.Errorf("%s %s: the new rule calls it terminal=%v and the old one %v",
							a.Kind, a.ID, a.terminal(), a.Cancelled)
					}
				}
			}
			if lines == 0 {
				t.Fatal("this record names no artifact at all, so it exercises nothing")
			}

			want := legacyOutstanding(entries)
			got := Outstanding(entries)
			if len(got) != len(want) {
				t.Fatalf("Outstanding = %d artifact(s), the pre-change rule says %d\n new: %s\n old: %s",
					len(got), len(want), describeArtifacts(got), describeArtifacts(want))
			}
			for i := range got {
				if got[i].Kind != want[i].Kind || got[i].ID != want[i].ID {
					t.Errorf("outstanding[%d] = %s %s, the pre-change rule says %s %s",
						i, got[i].Kind, got[i].ID, want[i].Kind, want[i].ID)
				}
			}

			// The set that decides what a person is asked to approve, which is the
			// one a regression here would actually cost something.
			pending := PendingCleanup(entries)
			for _, a := range pending {
				found := false
				for _, o := range got {
					if o.Kind == a.Kind && o.ID == a.ID {
						found = true
					}
				}
				if !found {
					t.Errorf("PendingCleanup offers %s %s, which is not outstanding at all", a.Kind, a.ID)
				}
			}
			t.Logf("%d entries, %d artifact lines classified identically, %d outstanding, "+
				"%d offered for cleanup — unchanged", len(entries), lines, len(got), len(pending))
		})
	}
}
