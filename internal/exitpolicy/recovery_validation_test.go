package exitpolicy

import (
	"errors"
	"strconv"
	"testing"
)

func ladderRecoverySnapshot(t *testing.T, price, quantity string) (ExitLineSnapshot, LadderPolicy) {
	t.Helper()
	policy := DefaultLadderPolicy()
	snapshot, err := EvaluateLadderSnapshot(LadderSnapshotInput{
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
	})
	if err != nil {
		t.Fatalf("EvaluateLadderSnapshot: %v", err)
	}
	return snapshot, policy
}

func TestRecoveryAllowsLadderBeforeFirstRung(t *testing.T) {
	snapshot, policy := ladderRecoverySnapshot(t, "10010", "10")
	if snapshot.ActiveRung != NoRung || !snapshot.Changed || snapshot.Action != ActionNone {
		t.Fatalf("pre-first-rung snapshot = %+v", snapshot)
	}
	if err := ValidateRecoveryDerivation(snapshot, NewLadderRecoveryPolicy(policy, "10")); err != nil {
		t.Fatalf("valid high-water-only ladder recovery: %v", err)
	}
}

func TestRecoveryRefusesInvalidLadderRungBounds(t *testing.T) {
	snapshot, policy := ladderRecoverySnapshot(t, "10010", "10")
	for _, rung := range []int{-2, 999} {
		t.Run(strconv.Itoa(rung), func(t *testing.T) {
			forged := snapshot
			forged.ActiveRung = rung
			forged.finishIDs()
			if err := ValidateRecoveryDerivation(forged, NewLadderRecoveryPolicy(policy, "10")); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("active rung %d error = %v, want identity refusal", rung, err)
			}
		})
	}
}

func TestRecoveryRefusesSemanticallyInvalidLadderOutputs(t *testing.T) {
	orderable, policy := ladderRecoverySnapshot(t, "10260", "10")
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
			if err := ValidateRecoveryDerivation(forged, NewLadderRecoveryPolicy(policy, "10")); !errors.Is(err, ErrRecoveryIdentity) {
				t.Fatalf("error = %v, want identity refusal", err)
			}
		})
	}
}
