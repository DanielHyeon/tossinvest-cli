package exitpolicy_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

func snapshotContext(quantity, observation string) exitpolicy.SnapshotContext {
	return exitpolicy.SnapshotContext{
		PositionID: "position-7", PositionGeneration: 7,
		ObservationID: observation, RemainingQuantity: quantity,
	}
}

func ladderSnapshotInput(price, quantity string) exitpolicy.LadderSnapshotInput {
	policy := exitpolicy.DefaultLadderPolicy()
	return exitpolicy.LadderSnapshotInput{
		Context: snapshotContext(quantity, "quote-2026-07-31T14:00:00Z"),
		Input: exitpolicy.LadderInput{
			EntryPrice: "10000", ObservedPrice: price, HighWater: "10000", Baseline: "9800",
			State: exitpolicy.LadderState{
				PolicyID: policy.PolicyID, ActivatedRung: exitpolicy.NoRung,
				TakenRatioTotal: "0", PendingRung: exitpolicy.NoRung,
			},
			Policy: policy,
		},
	}
}

func TestOneShareIntermediateLadderTargetIsStateOnly(t *testing.T) {
	snapshot, err := exitpolicy.EvaluateLadderSnapshot(ladderSnapshotInput("10260", "1"))
	if err != nil {
		t.Fatalf("EvaluateLadderSnapshot: %v", err)
	}
	if snapshot.Action != exitpolicy.ActionLadderPartial || snapshot.Ratio != "0.25" {
		t.Fatalf("action/ratio = %s/%s", snapshot.Action, snapshot.Ratio)
	}
	if snapshot.ProjectedQuantity != "0" || snapshot.Orderable || !snapshot.StateOnly {
		t.Fatalf("projection = %+v, want state-only zero", snapshot)
	}
	if snapshot.ActiveRung != 1 || snapshot.CurrentProtection != "10100" {
		t.Fatalf("rung/protection = %d/%s", snapshot.ActiveRung, snapshot.CurrentProtection)
	}
	if snapshot.NextTarget != "10400" || snapshot.NextProtection != "10200" {
		t.Fatalf("next target/protection = %s/%s", snapshot.NextTarget, snapshot.NextProtection)
	}
	if !snapshot.ExecutableProposal().Zero() {
		t.Fatalf("zero projection exposed an executable proposal: %+v", snapshot.ExecutableProposal())
	}
}

func TestOneShareFinalAndBreachProjectExactlyOne(t *testing.T) {
	final, err := exitpolicy.EvaluateLadderSnapshot(ladderSnapshotInput("10700", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if final.Action != exitpolicy.ActionLadderTakeProfit || final.ProjectedQuantity != "1" || !final.Orderable {
		t.Fatalf("final = %+v", final)
	}

	in := ladderSnapshotInput("10000", "1")
	in.Input.HighWater = "10260"
	in.Input.Baseline = "10100"
	in.Input.State.ActivatedRung = 1
	breach, err := exitpolicy.EvaluateLadderSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	if breach.Action != exitpolicy.ActionLadderStop || breach.ProjectedQuantity != "1" || !breach.Orderable {
		t.Fatalf("breach = %+v", breach)
	}
}

func TestPromotionAndBreachUseTheNewProtectionFromOneSnapshot(t *testing.T) {
	in := ladderSnapshotInput("10000", "1")
	in.Input.HighWater = "10700"
	snapshot, err := exitpolicy.EvaluateLadderSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ActiveRung != 3 || snapshot.CurrentProtection != "10350" {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if snapshot.Action != exitpolicy.ActionLadderStop || snapshot.ProjectedQuantity != "1" {
		t.Fatalf("promoted protection did not win: %+v", snapshot)
	}
}

func TestSnapshotIdentityIsDeterministicAndObservationBound(t *testing.T) {
	in := ladderSnapshotInput("10260", "7")
	a, err := exitpolicy.EvaluateLadderSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := exitpolicy.EvaluateLadderSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	if a.SnapshotID == "" || a.DecisionID == "" || a.InputDigest == "" {
		t.Fatalf("missing identity: %+v", a)
	}
	if a != b {
		t.Fatalf("same canonical input drifted:\n%+v\n%+v", a, b)
	}
	in.Context.ObservationID = "quote-2026-07-31T14:00:05Z"
	c, err := exitpolicy.EvaluateLadderSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	if c.SnapshotID == a.SnapshotID || c.DecisionID == a.DecisionID {
		t.Fatalf("different observation reused identity: %s/%s", c.SnapshotID, c.DecisionID)
	}
}

func TestSnapshotIsSafeForConcurrentConsumers(t *testing.T) {
	snapshot, err := exitpolicy.EvaluateLadderSnapshot(ladderSnapshotInput("10260", "7"))
	if err != nil {
		t.Fatal(err)
	}
	const consumers = 32
	var wg sync.WaitGroup
	errs := make(chan string, consumers)
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if snapshot.ExecutableProposal().Action != exitpolicy.ActionLadderPartial || snapshot.DecisionID == "" {
				errs <- "consumer observed a different decision"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestPolicyIdentityRejectsDigestCollision(t *testing.T) {
	a := exitpolicy.DefaultLadderPolicy()
	a.PolicyID = "POLICY_X"
	a.PolicyVersion = "1.0.0"
	a.PolicyDigest = ""
	aID, err := a.Identity()
	if err != nil {
		t.Fatal(err)
	}
	a.PolicyDigest = aID.Digest

	b := a
	b.Rungs = append([]exitpolicy.Rung(nil), a.Rungs...)
	b.Rungs[1].TargetPct = "2.6"
	b.PolicyDigest = ""
	bID, err := b.Identity()
	if err != nil {
		t.Fatal(err)
	}
	b.PolicyDigest = bID.Digest

	registry, err := exitpolicy.NewPolicyRegistry(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(b); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("collision error = %v", err)
	}

	tampered := b
	tampered.PolicyDigest = a.PolicyDigest
	in := ladderSnapshotInput("10260", "7")
	in.Input.Policy = tampered
	in.Input.State.PolicyID = tampered.PolicyID
	if _, err := exitpolicy.EvaluateLadderSnapshot(in); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("evaluation accepted conflicting digest: %v", err)
	}

	stateMismatch := ladderSnapshotInput("10260", "7")
	stateMismatch.Input.State.PolicyVersion = "2.0.0"
	if _, err := exitpolicy.EvaluateLadderSnapshot(stateMismatch); err == nil {
		t.Fatal("snapshot evaluation overwrote a mismatched persisted policy version")
	}
	stateMismatch = ladderSnapshotInput("10260", "7")
	stateMismatch.Input.State.PolicyDigest = "sha256:" + strings.Repeat("0", 64)
	if _, err := exitpolicy.EvaluateLadderSnapshot(stateMismatch); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("snapshot evaluation overwrote a mismatched persisted digest: %v", err)
	}
}

func TestRatchetOneSharePartialIsStateOnlyButBreachIsFull(t *testing.T) {
	base := exitpolicy.RatchetSnapshotInput{
		Context: snapshotContext("1", "quote-ratchet"),
		Input: exitpolicy.RatchetInput{
			Entry: "10000", InitialStop: "9800", ObservedPrice: "10200",
			HighWater: "10000", Baseline: "9800", RealBreakeven: "10010",
			TakenRatioTotal: "0",
		},
	}
	partial, err := exitpolicy.EvaluateRatchetSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}
	if partial.Action != exitpolicy.ActionRatchetPartial || partial.ProjectedQuantity != "0" || partial.Orderable {
		t.Fatalf("partial = %+v", partial)
	}
	base.Input.ObservedPrice = "9799"
	base.Context.ObservationID = "quote-ratchet-breach"
	breach, err := exitpolicy.EvaluateRatchetSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}
	if breach.Action != exitpolicy.ActionBaselineBreach || breach.ProjectedQuantity != "1" || !breach.Orderable {
		t.Fatalf("breach = %+v", breach)
	}
}

func TestCommonPolicyDescriptorsAreServerAuthoritative(t *testing.T) {
	descriptors := exitpolicy.RegisteredCommonPolicyDescriptors()
	if len(descriptors) != 3 {
		t.Fatalf("descriptor count = %d", len(descriptors))
	}
	seen := map[string]bool{}
	for _, d := range descriptors {
		seen[d.Identity.ID] = true
		if d.Identity.Version == "" || d.Identity.Digest == "" || d.Summary == "" || d.Unit.Target != "%" {
			t.Errorf("incomplete descriptor: %+v", d)
		}
		if d.OneShare.Intermediate != "매도 0주 · 보호선 승격" || d.OneShare.Final != "1주 전량" || d.OneShare.ProtectionBreach != "1주 전량" {
			t.Errorf("one-share projection = %+v", d.OneShare)
		}
	}
	for _, id := range []string{exitpolicy.CommonLadderBalanced, exitpolicy.CommonLadderRunner, exitpolicy.CommonLadderHybrid50} {
		if !seen[id] {
			t.Errorf("missing descriptor %s", id)
		}
	}

	field := exitpolicy.CommonPolicyFieldDescriptor()
	if err := field.Validate(); err != nil {
		t.Fatalf("field descriptor: %v", err)
	}
	if field.Effective.Kind != settingmeta.StateUnapproved || field.Effective.Display != "기존 RATCHET 유지" {
		t.Fatalf("effective = %+v", field.Effective)
	}
	var recommended string
	for _, option := range field.Options {
		if option.Recommended {
			recommended = option.ID
		}
	}
	if recommended != exitpolicy.CommonLadderHybrid50 {
		t.Fatalf("recommended = %q", recommended)
	}
}
