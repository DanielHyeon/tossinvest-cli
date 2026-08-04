package exitpolicy

// a078: activating the first ladder rung is a forward move, not an incomparable
// axis.
//
// compareRecoveryStage refused any pair where exactly one side held a rung. The
// guard was meant to stop a ratchet snapshot from being compared against a
// ladder one — but SelectRecoverySnapshot has already proven, before it calls
// this, that both candidates carry the same policy identity, position and
// generation. Inside that guarantee a one-sided NoRung means one thing only: a
// ladder that has not activated a rung yet.
//
// On 2026-08-03 that mistake quarantined a live position at the exact moment it
// crossed its first take-profit line, and the engine stopped judging it — stop
// included. Every ladder holding reaches this transition.

import (
	"errors"
	"testing"
)

// ladderRecoveryFixture is a ladder-policy candidate at a given rung. NoRung is
// "has not activated one yet", which is the state every ladder position starts
// in and the one the defect refused to compare.
func ladderRecoveryFixture(identity PolicyIdentity, id, protection, high string, rung int) ExitLineSnapshot {
	snapshot := ExitLineSnapshot{
		InputDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Policy:      identity,
		PositionID:  "p-ladder", PositionGeneration: 1, ObservationID: "obs_" + id,
		EntryPrice: "100", InitialStop: "90", ObservedPrice: high,
		CurrentProtection: protection, HighWater: high, RatchetLevel: LevelNone,
		ActiveRung: rung, NextTarget: "1" + high, NextProtection: protection,
		Action: ActionNone, ProjectedQuantity: "0",
	}
	snapshot.finishIDs()
	return snapshot
}

func ladderIdentity(t *testing.T) PolicyIdentity {
	t.Helper()
	identity, err := LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

// TestFirstRungActivationIsAForwardMove is a078 task 2.1 — the defect itself.
func TestFirstRungActivationIsAForwardMove(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_no_rung", "100", "110", NoRung)
	activated := ladderRecoveryFixture(identity, "els_rung_zero", "105", "115", 0)

	got, source, err := SelectRecoverySnapshot(&saved, activated)
	if err != nil {
		t.Fatalf("activating the first rung was refused: %v", err)
	}
	if got.SnapshotID != activated.SnapshotID || source != RecoveryRecomputed {
		t.Fatalf("selected = %s/%s, want the whole recomputed candidate", got.SnapshotID, source)
	}
}

// TestARecomputationThatLostItsRungKeepsTheSavedCandidate is task 2.2. Before
// a078 this was refused too, so the spec's "복구된 기준선은 낮아질 수 없다" could
// never actually run on this transition — it quarantined instead.
func TestARecomputationThatLostItsRungKeepsTheSavedCandidate(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_rung_two", "120", "130", 2)
	regressed := ladderRecoveryFixture(identity, "els_lost_rung", "100", "110", NoRung)

	got, source, err := SelectRecoverySnapshot(&saved, regressed)
	if err != nil {
		t.Fatalf("a regressed recomputation was refused instead of kept: %v", err)
	}
	if got.SnapshotID != saved.SnapshotID || source != RecoverySavedMonotone {
		t.Fatalf("selected = %s/%s, want the whole saved candidate", got.SnapshotID, source)
	}
	if got.CurrentProtection != "120" || got.ActiveRung != 2 {
		t.Fatalf("protection/rung fell to %s/%d", got.CurrentProtection, got.ActiveRung)
	}
}

// TestCrossedAxesAcrossRungsAreStillRefused is task 2.3: the genuine ambiguity
// this guard was reaching for survives.
func TestCrossedAxesAcrossRungsAreStillRefused(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_high_protection", "125", "130", NoRung)
	crossed := ladderRecoveryFixture(identity, "els_high_rung", "110", "135", 1)

	if _, _, err := SelectRecoverySnapshot(&saved, crossed); !errors.Is(err, ErrRecoveryAmbiguous) {
		t.Fatalf("crossed axes error = %v, want ambiguity", err)
	}
}

// TestIdentityDriftIsStillRefusedAcrossRungs is task 2.4. The identity block is
// the real ratchet/ladder guard, and a078 leans on it — so it is pinned here
// with a rung present on one side, the shape the removed branch used to catch.
func TestIdentityDriftIsStillRefusedAcrossRungs(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_saved_drift", "100", "110", NoRung)

	for _, c := range []struct {
		name  string
		mutot func(*ExitLineSnapshot)
	}{
		{"policy digest", func(s *ExitLineSnapshot) {
			s.Policy.Digest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		}},
		{"entry price", func(s *ExitLineSnapshot) { s.EntryPrice = "101" }},
		{"initial stop", func(s *ExitLineSnapshot) { s.InitialStop = "91" }},
		{"generation", func(s *ExitLineSnapshot) { s.PositionGeneration = 2 }},
	} {
		t.Run(c.name, func(t *testing.T) {
			drifted := ladderRecoveryFixture(identity, "els_drift", "105", "115", 0)
			c.mutot(&drifted)
			drifted.finishIDs()
			if _, _, err := SelectRecoverySnapshot(&saved, drifted); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("%s drift error = %v, want identity mismatch", c.name, err)
			}
		})
	}
}

// TestAnUnrankedRatchetLevelIsStillRefused is task 2.5: the other error return
// in this function stays exactly as it was.
func TestAnUnrankedRatchetLevelIsStillRefused(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_ranked", "100", "110", NoRung)
	unranked := ladderRecoveryFixture(identity, "els_unranked", "105", "115", NoRung)
	unranked.RatchetLevel = Level("NOT_A_LEVEL")
	unranked.finishIDs()

	if _, _, err := SelectRecoverySnapshot(&saved, unranked); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("unranked level error = %v, want identity mismatch", err)
	}
}

// TestRatchetLevelRankingIsUnchanged is task 2.6: both candidates at NoRung keep
// taking the ratchet-level path, which is what a ratchet policy always does and
// what a ladder does before its first rung.
func TestRatchetLevelRankingIsUnchanged(t *testing.T) {
	identity := ladderIdentity(t)
	saved := recoveryFixture(identity, "els_level_saved", "100", "110", LevelBreakeven)

	ahead := recoveryFixture(identity, "els_level_ahead", "105", "115", LevelPartialLock)
	if got, source, err := SelectRecoverySnapshot(&saved, ahead); err != nil ||
		got.SnapshotID != ahead.SnapshotID || source != RecoveryRecomputed {
		t.Fatalf("forward level = %s/%s err=%v, want the recomputed candidate", got.SnapshotID, source, err)
	}

	behind := recoveryFixture(identity, "els_level_behind", "95", "105", LevelHalfRisk)
	if got, source, err := SelectRecoverySnapshot(&saved, behind); err != nil ||
		got.SnapshotID != saved.SnapshotID || source != RecoverySavedMonotone {
		t.Fatalf("backward level = %s/%s err=%v, want the saved candidate", got.SnapshotID, source, err)
	}
}

// TestEqualStageWithDifferentDerivedLinesIsStillRefused is task 2.7.
func TestEqualStageWithDifferentDerivedLinesIsStillRefused(t *testing.T) {
	identity := ladderIdentity(t)
	saved := ladderRecoveryFixture(identity, "els_same_rung_a", "100", "110", 1)
	divergent := ladderRecoveryFixture(identity, "els_same_rung_b", "100", "110", 1)
	divergent.NextTarget = "999"
	divergent.finishIDs()

	if _, _, err := SelectRecoverySnapshot(&saved, divergent); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("derived-line error = %v, want identity mismatch", err)
	}
}
