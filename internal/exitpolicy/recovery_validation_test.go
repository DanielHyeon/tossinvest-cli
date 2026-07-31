package exitpolicy

import (
	"errors"
	"strconv"
	"testing"
)

func ladderRecoverySnapshot(t *testing.T, price, quantity string) (ExitLineSnapshot, RecoveryPolicyDefinition) {
	t.Helper()
	policy := DefaultLadderPolicy()
	evaluation := LadderSnapshotInput{
		Context: SnapshotContext{
			PositionID: "position-recovery", PositionGeneration: 3,
			ObservationID: "observation-" + price, RemainingQuantity: quantity,
		},
		Input: LadderInput{
			EntryPrice: "10000", InitialStop: "9800", ObservedPrice: price,
			HighWater: "10000", Baseline: "9800",
			State: LadderState{
				PolicyID: policy.PolicyID, ActivatedRung: NoRung,
				TakenRatioTotal: "0", PendingRung: NoRung,
			},
			Policy: policy,
		},
	}
	snapshot, err := EvaluateLadderSnapshot(evaluation)
	if err != nil {
		t.Fatalf("EvaluateLadderSnapshot: %v", err)
	}
	snapshot = snapshot.ChangedFromState("10000", "9800", LevelNone, NoRung)
	return snapshot, NewLadderRecoveryPolicy(evaluation, "10000", "9800", NoRung)
}

func TestRecoveryAllowsLadderBeforeFirstRung(t *testing.T) {
	snapshot, recovery := ladderRecoverySnapshot(t, "10010", "10")
	if snapshot.ActiveRung != NoRung || !snapshot.Changed || snapshot.Action != ActionNone {
		t.Fatalf("pre-first-rung snapshot = %+v", snapshot)
	}
	if err := ValidateRecoveryDerivation(snapshot, recovery); err != nil {
		t.Fatalf("valid high-water-only ladder recovery: %v", err)
	}
}

func TestRecoveryRefusesInvalidLadderRungBounds(t *testing.T) {
	snapshot, recovery := ladderRecoverySnapshot(t, "10010", "10")
	for _, rung := range []int{-2, 999} {
		t.Run(strconv.Itoa(rung), func(t *testing.T) {
			forged := snapshot
			forged.ActiveRung = rung
			forged.finishIDs()
			if err := ValidateRecoveryDerivation(forged, recovery); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("active rung %d error = %v, want identity refusal", rung, err)
			}
		})
	}
}

func TestRecoveryRefusesSemanticallyInvalidLadderOutputs(t *testing.T) {
	orderable, recovery := ladderRecoverySnapshot(t, "10260", "10")
	nonorderable, _ := ladderRecoverySnapshot(t, "10010", "10")
	tests := []struct {
		name   string
		base   ExitLineSnapshot
		mutate func(*ExitLineSnapshot)
	}{
		{"foreign_action", orderable, func(s *ExitLineSnapshot) { s.Action = ActionRatchetPartial }},
		{"zero_ratio", orderable, func(s *ExitLineSnapshot) { s.Ratio = "0" }},
		{"ratio_above_one", orderable, func(s *ExitLineSnapshot) { s.Ratio = "2" }},
		{"invalid_ratio", orderable, func(s *ExitLineSnapshot) { s.Ratio = "not-a-ratio" }},
		{"fractional_projection", orderable, func(s *ExitLineSnapshot) { s.ProjectedQuantity = "1.5" }},
		{"wrong_projection", orderable, func(s *ExitLineSnapshot) { s.ProjectedQuantity = "9" }},
		{"wrong_level", orderable, func(s *ExitLineSnapshot) { s.Level = "2" }},
		{"orderable_state_only", orderable, func(s *ExitLineSnapshot) { s.StateOnly = true }},
		{"orderable_suppressed", orderable, func(s *ExitLineSnapshot) { s.Suppressed = SuppressedPending }},
		{"unknown_suppression", nonorderable, func(s *ExitLineSnapshot) { s.Suppressed = "unknown" }},
		{"none_state_only", nonorderable, func(s *ExitLineSnapshot) { s.StateOnly = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			forged := test.base
			test.mutate(&forged)
			forged.finishIDs()
			if err := ValidateRecoveryDerivation(forged, recovery); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("error = %v, want identity refusal", err)
			}
		})
	}
}

func TestRecoveryReevaluatesExactInputAndEveryExecutionField(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ExitLineSnapshot, *RecoveryPolicyDefinition)
	}{
		{"current_protection", func(s *ExitLineSnapshot, _ *RecoveryPolicyDefinition) { s.CurrentProtection = "9700" }},
		{"remaining_quantity_same_projection", func(_ *ExitLineSnapshot, r *RecoveryPolicyDefinition) {
			r.Ladder.Evaluation.Context.RemainingQuantity = "11"
		}},
		{"cancel_pending_first", func(s *ExitLineSnapshot, _ *RecoveryPolicyDefinition) {
			s.CancelPendingFirst = !s.CancelPendingFirst
		}},
		{"changed", func(s *ExitLineSnapshot, _ *RecoveryPolicyDefinition) { s.Changed = !s.Changed }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot, recovery := ladderRecoverySnapshot(t, "10260", "10")
			test.mutate(&snapshot, &recovery)
			if err := ValidateRecoveryDerivation(snapshot, recovery); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("error = %v, want exact re-evaluation refusal", err)
			}
		})
	}
}

func TestRecoveryRejectsForgedRatchetLevel(t *testing.T) {
	config := DefaultRatchetConfig()
	evaluation := RatchetSnapshotInput{
		Context: SnapshotContext{PositionID: "ratchet-recovery", PositionGeneration: 2,
			ObservationID: "ratchet-level", RemainingQuantity: "10"},
		Input: RatchetInput{Entry: "10000", InitialStop: "9800", ObservedPrice: "10010",
			HighWater: "10000", Baseline: "9800", RealBreakeven: "10005",
			TakenRatioTotal: "0", Level: LevelNone, Config: &config},
	}
	snapshot, err := EvaluateRatchetSnapshot(evaluation)
	if err != nil {
		t.Fatal(err)
	}
	snapshot = snapshot.ChangedFromState("10000", "9800", LevelNone, NoRung)
	recovery := NewRatchetRecoveryPolicy(evaluation, "10000", "9800", LevelNone)
	snapshot.RatchetLevel = LevelProfitLock
	if err := ValidateRecoveryDerivation(snapshot, recovery); !errors.Is(err, ErrRecoveryIdentity) {
		t.Fatalf("error = %v, want forged level refusal", err)
	}
}
