package journal

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestOpenExitStatePreservesPinnedRuntimePolicyIdentity(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-policy-runtime", attemptID: "a-policy-runtime", orderID: "o-policy-runtime", decisionID: "d-policy-runtime"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	want, err := exitpolicy.LegacyLadderPolicyIdentity(exitpolicy.CommonLadderHybrid50, false)
	if err != nil {
		t.Fatal(err)
	}
	state, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50,
		PolicyIdentity: want, EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.PolicyIdentity != want {
		t.Fatalf("runtime policy identity = %+v, want %+v", state.PolicyIdentity, want)
	}
}

func TestOpenExitStateRejectsASeedWhosePolicyMeaningDoesNotMatchItsID(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-policy-identity", attemptID: "a-policy-identity", orderID: "o-policy-identity", decisionID: "d-policy-identity"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	identity, err := exitpolicy.LegacyLadderPolicyIdentity("default_v1", false)
	if err != nil {
		t.Fatal(err)
	}
	identity.Digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err = j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder, PolicyID: "default_v1",
		PolicyIdentity: identity, EntryPrice: "70000", InitialStop: "68000",
	})
	if !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("OpenExitState error = %v, want policy identity conflict", err)
	}
}

func TestRecordExitJudgementRejectsProposalFromAnotherSnapshot(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-provenance", attemptID: "a-provenance", orderID: "o-provenance", decisionID: "d-provenance"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	state, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatal(err)
	}
	identity, err := exitpolicy.LegacyRatchetPolicyIdentity()
	if err != nil {
		t.Fatal(err)
	}
	provenance := ExitDecisionProvenance{
		ObservationID: "obs_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SnapshotID:    "els_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		DecisionID:    "eld_cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Policy:        identity,
	}
	other := provenance
	other.DecisionID = "eld_dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	err = j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "69000", HighWater: state.HighWater,
		Baseline: state.Baseline, RatchetLevel: state.RatchetLevel,
		ActiveRung: state.ActiveRung, Provenance: provenance,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionBaselineBreach), Level: "INITIAL",
			IntentID: "exit-mismatch", Provenance: other,
		},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("RecordExitJudgement error = %v, want invalid request", err)
	}
}
