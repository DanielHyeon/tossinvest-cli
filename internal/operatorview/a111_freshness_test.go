package operatorview

import (
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func a111OperatorSnapshot(observedAt string) journal.ExitSnapshotView {
	return journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
		Line: exitpolicy.ExitLineSnapshot{
			SnapshotID: "a111-snapshot", DecisionID: "a111-decision", ObservationID: "a111-observation",
			Policy:     exitpolicy.PolicyIdentity{ID: "ratchet-v1", Version: "1", Digest: "sha256:test"},
			PositionID: "a111-position", PositionGeneration: 1,
			EntryPrice: "200", InitialStop: "190", ObservedPrice: "205",
			CurrentProtection: "195", HighWater: "210", RatchetLevel: exitpolicy.LevelNone,
			ActiveRung: exitpolicy.NoRung, NextTarget: "220", NextProtection: "200",
			Action: exitpolicy.ActionNone, ProjectedQuantity: "0",
		},
		ObservationSource: "quote_fetched_at", ObservedAt: observedAt,
	}}
}

func TestA111SharedFreshnessAppliesTheExactThirtySecondBoundToEveryNonStoppedLiveness(t *testing.T) {
	observed := time.Date(2026, 8, 14, 5, 0, 0, 0, time.UTC)
	for _, live := range []ExitLiveness{
		ExitLivenessRunning,
		ExitLivenessUnavailable,
		ExitLivenessUnwired,
	} {
		for _, tc := range []struct {
			name  string
			age   time.Duration
			stale bool
		}{
			{"inside", 29999 * time.Millisecond, false},
			{"exact", 30 * time.Second, false},
			{"outside", 30*time.Second + time.Nanosecond, true},
		} {
			t.Run(fmt.Sprint(live)+"/"+tc.name, func(t *testing.T) {
				got := ApplyExitFreshness(a111OperatorSnapshot(observed.Format(time.RFC3339Nano)),
					observed.Add(tc.age), live)
				if got.Stale != tc.stale {
					t.Fatalf("stale = %v, want %v (%+v)", got.Stale, tc.stale, got)
				}
				if tc.stale && got.StaleReason != "observation_older_than_limit" {
					t.Fatalf("stale reason = %q", got.StaleReason)
				}
			})
		}
	}
}

func TestA111SharedFreshnessMakesStoppedImmediateAndStillChecksIntegrityOtherwise(t *testing.T) {
	now := time.Date(2026, 8, 14, 5, 0, 0, 0, time.UTC)
	stopped := ApplyExitFreshness(a111OperatorSnapshot(now.Format(time.RFC3339Nano)), now,
		ExitLivenessStopped)
	if !stopped.Stale || stopped.StaleReason != "engine_not_running" {
		t.Fatalf("stopped verdict = %+v", stopped)
	}

	for _, tc := range []struct {
		name       string
		observedAt string
		reason     string
	}{
		{"malformed", "not-a-time", "invalid_observed_at"},
		{"future", now.Add(time.Nanosecond).Format(time.RFC3339Nano), "observation_in_future"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyExitFreshness(a111OperatorSnapshot(tc.observedAt), now, ExitLivenessRunning)
			if !got.Stale || got.StaleReason != tc.reason {
				t.Fatalf("integrity verdict = %+v, want %q", got, tc.reason)
			}
		})
	}
}

func TestA111FreshnessIsPerPositionAndCannotBeBorrowedFromAValidSibling(t *testing.T) {
	now := time.Date(2026, 8, 14, 5, 0, 31, 0, time.UTC)
	freshSibling := ApplyExitFreshness(a111OperatorSnapshot(now.Format(time.RFC3339Nano)), now,
		ExitLivenessRunning)
	invalidSibling := ApplyExitFreshness(a111OperatorSnapshot(now.Add(-31*time.Second).Format(time.RFC3339Nano)),
		now, ExitLivenessRunning)
	if freshSibling.Stale {
		t.Fatalf("valid sibling unexpectedly stale: %+v", freshSibling)
	}
	if !invalidSibling.Stale || invalidSibling.StaleReason != "observation_older_than_limit" {
		t.Fatalf("invalid sibling borrowed liveness: %+v", invalidSibling)
	}
}

func TestA111SeedCorruptAndPartialEvidenceRemainNonActionable(t *testing.T) {
	for _, reason := range []string{
		"not_evaluated_yet",
		"invalid_effective_snapshot",
		"partial_evaluated_tuple",
		"flattened_snapshot_mismatch",
	} {
		t.Run(reason, func(t *testing.T) {
			freshness := ApplyExitFreshness(journal.ExitSnapshotView{UnknownReason: reason},
				time.Date(2026, 8, 14, 5, 0, 0, 0, time.UTC), ExitLivenessRunning)
			got := BuildExitLine(Source{UnknownReason: freshness.UnknownReason})
			if !got.Unknown() || got.CurrentProtection != dash || got.NextTarget != dash ||
				got.NextProtection != dash {
				t.Fatalf("%s became actionable: %+v", reason, got)
			}
		})
	}
}
