package exitpolicy

import (
	"errors"
	"testing"
)

func TestRecoverySelectsOneWholeCoherentSnapshot(t *testing.T) {
	identity, err := LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	saved := recoveryFixture(identity, "els_saved", "100", "110", LevelBreakeven)
	recomputed := recoveryFixture(identity, "els_recomputed", "105", "115", LevelPartialLock)

	got, source, err := SelectRecoverySnapshot(&saved, recomputed)
	if err != nil {
		t.Fatal(err)
	}
	if got.SnapshotID != recomputed.SnapshotID || source != RecoveryRecomputed {
		t.Fatalf("selected = %s/%s, want whole recomputed", got.SnapshotID, source)
	}

	lower := recoveryFixture(identity, "els_lower", "90", "105", LevelHalfRisk)
	got, source, err = SelectRecoverySnapshot(&saved, lower)
	if err != nil {
		t.Fatal(err)
	}
	if got.SnapshotID != saved.SnapshotID || source != RecoverySavedMonotone {
		t.Fatalf("selected = %s/%s, want whole saved", got.SnapshotID, source)
	}
}

func TestRecoveryRefusesCrossedAxesAndPolicyDrift(t *testing.T) {
	identity, err := LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	saved := recoveryFixture(identity, "els_saved", "105", "110", LevelBreakeven)
	crossed := recoveryFixture(identity, "els_crossed", "100", "120", LevelPartialLock)
	if _, _, err := SelectRecoverySnapshot(&saved, crossed); !errors.Is(err, ErrRecoveryAmbiguous) {
		t.Fatalf("crossed error = %v, want ambiguity", err)
	}

	drifted := crossed
	drifted.Policy.Digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, _, err := SelectRecoverySnapshot(&saved, drifted); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("drift error = %v, want identity mismatch", err)
	}

	immutable := recoveryFixture(identity, "els_immutable", "105", "115", LevelPartialLock)
	immutable.EntryPrice = "101"
	immutable.finishIDs()
	if _, _, err := SelectRecoverySnapshot(&saved, immutable); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("entry-axis error = %v, want identity mismatch", err)
	}

	sameStage := recoveryFixture(identity, "els_same_stage", "106", "116", LevelBreakeven)
	sameStage.NextTarget = "999"
	sameStage.finishIDs()
	if _, _, err := SelectRecoverySnapshot(&saved, sameStage); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("derived-line error = %v, want identity mismatch", err)
	}
}

func recoveryFixture(identity PolicyIdentity, id, protection, high string, level Level) ExitLineSnapshot {
	snapshot := ExitLineSnapshot{
		InputDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Policy: identity,
		PositionID: "p-1", PositionGeneration: 1, ObservationID: "obs_" + id,
		EntryPrice: "100", InitialStop: "90", ObservedPrice: high,
		CurrentProtection: protection, HighWater: high, RatchetLevel: level,
		ActiveRung: NoRung, NextTarget: "130", NextProtection: "110",
		Action: ActionNone, ProjectedQuantity: "0",
	}
	snapshot.finishIDs()
	return snapshot
}
