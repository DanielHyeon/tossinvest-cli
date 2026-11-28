package journal

// a208 F3 — 원장은 편입 보호선 사본 ID 하나만 더 받고(편입 기록 · 편입 포지션 exit state), 다른 미등록 ID 와
// 편입이 아닌 포지션의 사본 ID 는 지금처럼 거부함.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestA208JournalRefusesOtherUnregisteredPolicies(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	adoptable(t, j, "p-a208-unknown")
	bad := sampleAdoption("p-a208-unknown")
	bad.ExitPolicyID = "COMMON_LADDER_SOMETHING_ELSE"
	if _, err := j.AdoptPosition(ctx, bad); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("AdoptPosition(unregistered) error = %v, want invalid request", err)
	}

	adoptable(t, j, "p-a208-floor")
	good := sampleAdoption("p-a208-floor")
	good.ExitPolicyID = exitpolicy.CommonLadderHybrid50AdoptedFloorOnly
	if _, err := j.AdoptPosition(ctx, good); err != nil {
		t.Fatalf("AdoptPosition(floor-only copy): %v", err)
	}
	state, err := j.OpenAdoptedExitState(ctx, "p-a208-floor")
	if err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	want, err := exitpolicy.LegacyLadderPolicyIdentity(exitpolicy.CommonLadderHybrid50AdoptedFloorOnly, true)
	if err != nil {
		t.Fatal(err)
	}
	if state.PolicyKind != ExitPolicyLadder || state.PolicyIdentity != want {
		t.Fatalf("exit state = %s %+v, want LADDER %+v", state.PolicyKind, state.PolicyIdentity, want)
	}
}

func TestA208FloorOnlyCopyNeedsAnAdoptedPosition(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-a208-floor", attemptID: "a-a208-floor", orderID: "o-a208-floor", decisionID: "d-a208-floor"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, o)
	_, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50AdoptedFloorOnly,
		EntryPrice: "70000", InitialStop: "68000",
	})
	if !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("OpenExitState(engine-entered, floor-only copy) error = %v, want policy identity conflict", err)
	}
}
