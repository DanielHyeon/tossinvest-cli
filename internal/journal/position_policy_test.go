package journal

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func seedPolicyPosition(t *testing.T, j *Journal, pending bool) string {
	t.Helper()
	ctx := context.Background()
	_, err := j.db.ExecContext(ctx, `INSERT INTO positions
		(id, account_ref, market, symbol, instance_seq, state, quantity, avg_price, adoption_id)
		VALUES ('p-policy','acct-1','kr','005930',7,'OPEN','10','70000',NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.AdoptPosition(ctx, AdoptionRequest{
		PositionID: "p-policy", Symbol: "005930", Market: "kr", Quantity: "10",
		CostBasis: "60000", ObservedPrice: "70000", SyntheticStop: "66500",
		ObservedAt: "2026-08-01T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	pendingAction := any(nil)
	pendingLevel := any(nil)
	if pending {
		pendingAction, pendingLevel = "FULL_EXIT", "STOP"
	}
	_, err = j.db.ExecContext(ctx, `INSERT INTO exit_states
		(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,high_water,
		 ratchet_level,active_rung,pending_action,pending_level,pending_intent_id,completed,updated_at,
		 policy_id,snapshot_status,policy_version,policy_digest,position_generation)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,0,?,?,?,?,?,?)`,
		"p-policy", "RATCHET", "70000", "66500", "3500", "66500", "70000", "NONE", nil,
		pendingAction, pendingLevel, nil, "2026-08-01T00:00:00Z", nil, "SEED", "1", "digest", 7)
	if err != nil {
		t.Fatal(err)
	}
	return "p-policy"
}

func policyRequest(action positionpolicy.Action, version int64) positionpolicy.Request {
	reason := map[positionpolicy.Action]positionpolicy.Reason{
		positionpolicy.ActionOverride: positionpolicy.ReasonPolicyOverride,
		positionpolicy.ActionInherit:  positionpolicy.ReasonPolicyInherit,
		positionpolicy.ActionRelease:  positionpolicy.ReasonRelease,
		positionpolicy.ActionReadopt:  positionpolicy.ReasonReadopt,
	}[action]
	return positionpolicy.Request{
		PositionID: "p-policy", ExpectedGeneration: 1, ExpectedVersion: version,
		Action: action, Actor: positionpolicy.ActorLocalOperator, Reason: reason,
		At: time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC),
	}
}

func readoptRequest(version int64, at time.Time) positionpolicy.Request {
	req := policyRequest(positionpolicy.ActionReadopt, version)
	req.At = at
	req.ReAdoption = &positionpolicy.ReAdoptionObservation{
		ObservedPrice: "72000", SyntheticStop: "68400", ObservedAt: at.UTC().Format(time.RFC3339Nano),
		PolicyID: exitpolicy.CommonLadderBalanced,
	}
	return req
}

func TestPositionPolicyOverrideCASAndAuditAreAtomic(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	req := policyRequest(positionpolicy.ActionOverride, 0)
	req.PolicyID = "COMMON_LADDER_HYBRID_50"

	preview, err := j.PreviewPositionPolicy(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Before.Version != 0 || preview.After.Version != 1 || preview.RestartRequired {
		t.Fatalf("preview = %+v", preview)
	}
	got, err := j.ApplyPositionPolicy(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if got.AdoptionGeneration != 1 || got.Version != 1 || got.DesiredPolicyID != req.PolicyID {
		t.Fatalf("state = %+v", got)
	}
	events, err := j.PositionPolicyAudit(ctx, "p-policy")
	if err != nil || len(events) != 1 {
		t.Fatalf("audit = %+v, err=%v", events, err)
	}
	if events[0].BeforeJSON == "" || events[0].AfterJSON == "" || events[0].Actor != positionpolicy.ActorLocalOperator {
		t.Fatalf("audit evidence incomplete: %+v", events[0])
	}

	stale := req
	stale.PolicyID = "COMMON_LADDER_BALANCED"
	if _, err := j.ApplyPositionPolicy(ctx, stale); !errors.Is(err, positionpolicy.ErrVersionMismatch) {
		t.Fatalf("stale apply error = %v", err)
	}
	after, _ := j.PositionPolicy(ctx, "p-policy")
	if after.Version != 1 || after.DesiredPolicyID != req.PolicyID {
		t.Fatalf("stale apply changed state: %+v", after)
	}
	events, _ = j.PositionPolicyAudit(ctx, "p-policy")
	if len(events) != 1 {
		t.Fatalf("stale apply appended audit: %d", len(events))
	}
}

func TestPositionPolicyReleaseRejectsPendingExitWithoutMutation(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, true)
	req := policyRequest(positionpolicy.ActionRelease, 0)
	if _, err := j.ApplyPositionPolicy(context.Background(), req); !errors.Is(err, positionpolicy.ErrExitConflict) {
		t.Fatalf("release error = %v", err)
	}
	if events, _ := j.PositionPolicyAudit(context.Background(), "p-policy"); len(events) != 0 {
		t.Fatalf("refused release wrote %d audit events", len(events))
	}
}

func TestPositionPolicyReleaseAndReadoptCreateFreshGeneration(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	released, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0))
	if err != nil {
		t.Fatal(err)
	}
	if released.Status != positionpolicy.StatusReleased || released.Version != 1 {
		t.Fatalf("released = %+v", released)
	}
	if released.ObservedAt == "" {
		t.Fatal("first durable lifecycle has no observation boundary")
	}
	readopt := readoptRequest(1, policyRequest(positionpolicy.ActionReadopt, 1).At.Add(time.Minute))
	readopt.ReAdoption.PolicyID = ""
	got, err := j.ApplyPositionPolicy(ctx, readopt)
	if err != nil {
		t.Fatal(err)
	}
	if got.AdoptionGeneration != 2 || got.Version != 1 || got.Status != positionpolicy.StatusManaged {
		t.Fatalf("readopted = %+v", got)
	}
	if got.EffectivePolicyID != exitpolicy.RatchetPolicyID {
		t.Fatalf("readopted effective policy=%q, want actual reset policy %q",
			got.EffectivePolicyID, exitpolicy.RatchetPolicyID)
	}
	if got.ObservedAt == "" || got.ObservedAt == released.ObservedAt {
		t.Fatalf("re-adopt did not establish fresh t0: before=%q after=%q", released.ObservedAt, got.ObservedAt)
	}

	late := policyRequest(positionpolicy.ActionInherit, 1)
	late.ExpectedGeneration = 1
	late.At = late.At.Add(2 * time.Minute)
	if _, err := j.ApplyPositionPolicy(ctx, late); !errors.Is(err, positionpolicy.ErrVersionMismatch) {
		t.Fatalf("late old-generation event error = %v", err)
	}
	current, _ := j.PositionPolicy(ctx, "p-policy")
	if current.AdoptionGeneration != 2 || current.Version != 1 {
		t.Fatalf("old generation changed current: %+v", current)
	}
}

func TestReadOnlyPositionPoliciesPreservesReleasedLifecycle(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	if _, err := j.ApplyPositionPolicy(context.Background(), policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	states, err := openTestReadOnly(t, j.Path()).PositionPolicies(context.Background())
	if err != nil || len(states) != 1 || states[0].PositionID != "p-policy" ||
		states[0].Status != positionpolicy.StatusReleased {
		t.Fatalf("read-only states=%+v err=%v", states, err)
	}
}

func TestPositionPolicyReleaseRemovesExactGenerationFromWorkingSet(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	if states, err := j.OpenExitStateResults(ctx, "acct-1"); err != nil || len(states) != 1 {
		t.Fatalf("before release results=%d err=%v", len(states), err)
	}
	if _, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	if states, err := j.OpenExitStateResults(ctx, "acct-1"); err != nil || len(states) != 0 {
		t.Fatalf("released generation remains in observer results: %d err=%v", len(states), err)
	}
	if states, err := j.OpenExitStates(ctx, "acct-1"); err != nil || len(states) != 0 {
		t.Fatalf("released generation remains in working set: %d err=%v", len(states), err)
	}
}

func TestPositionPolicyReadoptResetsExitStateAtFreshT0(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `UPDATE exit_states SET baseline_price='71000',high_water='75000',
		ratchet_level='BREAKEVEN',active_rung=1,taken_ratio_total='0.5' WHERE position_id='p-policy'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC)
	state, err := j.ApplyPositionPolicy(ctx, readoptRequest(1, at))
	if err != nil {
		t.Fatal(err)
	}
	if state.AdoptionGeneration != 2 || state.Status != positionpolicy.StatusManaged {
		t.Fatalf("lifecycle=%+v", state)
	}
	if state.EffectivePolicyID != exitpolicy.CommonLadderBalanced {
		t.Fatalf("effective policy=%q, want reset observation policy %q",
			state.EffectivePolicyID, exitpolicy.CommonLadderBalanced)
	}
	exit, err := j.ExitState(ctx, "p-policy")
	if err != nil {
		t.Fatal(err)
	}
	if exit.LifecycleGeneration != 2 || exit.EntryPrice != "72000" || exit.InitialStop != "68400" ||
		exit.HighWater != "72000" || exit.Baseline != "68400" || exit.ActiveRung != exitpolicy.NoRung ||
		exit.TakenRatioTotal != "0" || exit.Pending() || exit.Completed {
		t.Fatalf("fresh exit state reused old progress: %+v", exit)
	}
	if states, err := j.OpenExitStates(ctx, "acct-1"); err != nil || len(states) != 1 || states[0].LifecycleGeneration != 2 {
		t.Fatalf("readopt working set=%+v err=%v", states, err)
	}
}

func TestPositionPolicyReleaseAndReadoptRequireExternalAdoptionAtJournalBoundary(t *testing.T) {
	j := exitFixture(t)
	opened, _ := openedPosition(t, j, "10")
	position := currentPosition(t, j, opened)
	state, err := j.PositionPolicy(context.Background(), position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Provenance != positionpolicy.ProvenanceEngineEntry ||
		state.Eligibility != positionpolicy.EligibilityExitOnly {
		t.Fatalf("engine position projection=%+v", state)
	}
	for _, action := range []positionpolicy.Action{positionpolicy.ActionRelease, positionpolicy.ActionReadopt} {
		req := policyRequest(action, state.Version)
		req.PositionID = position.ID
		req.ExpectedGeneration = state.AdoptionGeneration
		if action == positionpolicy.ActionReadopt {
			req.ReAdoption = readoptRequest(state.Version, req.At).ReAdoption
		}
		if _, err := j.ApplyPositionPolicy(context.Background(), req); !errors.Is(err, positionpolicy.ErrIneligible) {
			t.Errorf("%s error=%v", action, err)
		}
	}
}

func TestReadoptedGenerationFillAndCompletionEventsStayGenerationTwo(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	if _, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionPolicy(ctx, readoptRequest(1,
		time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC))); err != nil {
		t.Fatal(err)
	}
	sqlTx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	applyTx := &ApplyTx{tx: sqlTx, now: "2026-08-01T02:01:00Z"}
	if err := appendExitEventFromHook(ctx, applyTx, exitEventRow{
		PositionID: "p-policy", Action: ExitEventProposalFilled,
		ProposedIntentID: "intent-gen-2", CreatedAt: applyTx.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := markExitCompletedTx(ctx, applyTx, "p-policy"); err != nil {
		t.Fatal(err)
	}
	if err := sqlTx.Commit(); err != nil {
		t.Fatal(err)
	}
	applyTx.invalidate()
	events, err := j.ExitEvents(ctx, "p-policy")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 2 {
		t.Fatalf("events=%+v", events)
	}
	filled, completed := events[len(events)-2], events[len(events)-1]
	if filled.Action != ExitEventProposalFilled || completed.Action != ExitEventCompleted ||
		filled.LifecycleGeneration != 2 || completed.LifecycleGeneration != 2 {
		t.Fatalf("generation-2 hook events=%+v / %+v", filled, completed)
	}
}

func TestPositionPolicyReadoptRejectsIneligibleHolding(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price)
		VALUES('p-ineligible','acct-1','kr','000660',8,'OPEN','1','100000')`); err != nil {
		t.Fatal(err)
	}
	req := readoptRequest(0, time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC))
	req.PositionID = "p-ineligible"
	if _, err := j.ApplyPositionPolicy(ctx, req); !errors.Is(err, positionpolicy.ErrIneligible) {
		t.Fatalf("ineligible re-adopt error=%v", err)
	}
	state, err := j.PositionPolicy(ctx, "p-ineligible")
	if err != nil || state.Status == positionpolicy.StatusManaged || state.ExitEligible {
		t.Fatalf("ineligible state mislabeled: %+v err=%v", state, err)
	}
}

func TestLateOldGenerationJudgementIsQuarantined(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	if _, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionPolicy(ctx, readoptRequest(1, time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC))); err != nil {
		t.Fatal(err)
	}
	before, _ := j.ExitState(ctx, "p-policy")
	err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: "p-policy", LifecycleGeneration: 1, ObservedPrice: "76000",
		HighWater: "76000", Baseline: "70000", RatchetLevel: string(exitpolicy.LevelBreakeven),
	})
	if !errors.Is(err, ErrExitLifecycleStale) {
		t.Fatalf("late judgement error=%v", err)
	}
	after, _ := j.ExitState(ctx, "p-policy")
	if after.HighWater != before.HighWater || after.Baseline != before.Baseline || after.Pending() {
		t.Fatalf("late generation changed state: before=%+v after=%+v", before, after)
	}
}

func TestExitEventsCarryLifecycleGeneration(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	if _, err := j.ApplyPositionPolicy(ctx, policyRequest(positionpolicy.ActionRelease, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionPolicy(ctx, readoptRequest(1, time.Date(2026, 8, 1, 2, 0, 0, 0, time.UTC))); err != nil {
		t.Fatal(err)
	}
	events, err := j.ExitEvents(ctx, "p-policy")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 || events[len(events)-1].LifecycleGeneration != 2 {
		t.Fatalf("exit events lack current lifecycle generation: %+v", events)
	}
}

func TestPositionPolicyCommandsNeverRebindSavedExitSnapshot(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	var before struct{ kind, id, version, digest, baseline, high string }
	if err := j.db.QueryRowContext(ctx, `SELECT policy_kind,coalesce(policy_id,''),coalesce(policy_version,''),
		coalesce(policy_digest,''),baseline_price,high_water FROM exit_states WHERE position_id='p-policy'`).
		Scan(&before.kind, &before.id, &before.version, &before.digest, &before.baseline, &before.high); err != nil {
		t.Fatal(err)
	}
	req := policyRequest(positionpolicy.ActionOverride, 0)
	req.PolicyID = "COMMON_LADDER_BALANCED"
	if _, err := j.ApplyPositionPolicy(ctx, req); err != nil {
		t.Fatal(err)
	}
	var after struct{ kind, id, version, digest, baseline, high string }
	if err := j.db.QueryRowContext(ctx, `SELECT policy_kind,coalesce(policy_id,''),coalesce(policy_version,''),
		coalesce(policy_digest,''),baseline_price,high_water FROM exit_states WHERE position_id='p-policy'`).
		Scan(&after.kind, &after.id, &after.version, &after.digest, &after.baseline, &after.high); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("saved exit snapshot rebound: before=%+v after=%+v", before, after)
	}
}

func TestPositionPolicyRejectsFreeActorReasonAndUnknownState(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	req := policyRequest(positionpolicy.ActionRelease, 0)
	req.Actor = "typed user"
	if _, err := j.ApplyPositionPolicy(ctx, req); !errors.Is(err, positionpolicy.ErrInvalidRequest) {
		t.Fatalf("free actor error = %v", err)
	}
	req = policyRequest(positionpolicy.ActionRelease, 0)
	req.Reason = "typed reason"
	if _, err := j.ApplyPositionPolicy(ctx, req); !errors.Is(err, positionpolicy.ErrInvalidRequest) {
		t.Fatalf("free reason error = %v", err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE positions SET state='CLOSING' WHERE id='p-policy'`); err != nil {
		t.Fatal(err)
	}
	req = policyRequest(positionpolicy.ActionRelease, 0)
	if _, err := j.ApplyPositionPolicy(ctx, req); !errors.Is(err, positionpolicy.ErrExitConflict) {
		t.Fatalf("closing state release error = %v", err)
	}
}

func TestPositionPolicyConcurrentCASCommitsExactlyOneWinner(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	ctx := context.Background()
	requests := []positionpolicy.Request{
		policyRequest(positionpolicy.ActionOverride, 0),
		policyRequest(positionpolicy.ActionOverride, 0),
	}
	requests[0].PolicyID = "COMMON_LADDER_BALANCED"
	requests[1].PolicyID = "COMMON_LADDER_RUNNER"
	var wg sync.WaitGroup
	errs := make([]error, len(requests))
	for index := range requests {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, errs[index] = j.ApplyPositionPolicy(ctx, requests[index])
		}(index)
	}
	wg.Wait()
	wins, stale := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, positionpolicy.ErrVersionMismatch):
			stale++
		default:
			t.Fatalf("unexpected race result: %v", err)
		}
	}
	if wins != 1 || stale != 1 {
		t.Fatalf("race winners=%d stale=%d errors=%v", wins, stale, errs)
	}
	events, err := j.PositionPolicyAudit(ctx, "p-policy")
	if err != nil || len(events) != 1 {
		t.Fatalf("race audit events=%d err=%v", len(events), err)
	}
}

func TestPositionPolicyLifecycleAndAuditRollbackTogetherOnInjectedCrashPoint(t *testing.T) {
	j := openTestJournal(t)
	seedPolicyPosition(t, j, false)
	j.exitWriteHook = func(stage string) error {
		if stage == "position_policy_after_state" {
			return errors.New("injected crash point")
		}
		return nil
	}
	req := policyRequest(positionpolicy.ActionOverride, 0)
	req.PolicyID = exitpolicy.CommonLadderBalanced
	if _, err := j.ApplyPositionPolicy(context.Background(), req); err == nil {
		t.Fatal("injected failure unexpectedly committed")
	}
	j.exitWriteHook = nil
	state, err := j.PositionPolicy(context.Background(), "p-policy")
	if err != nil || state.Version != 0 || state.DesiredPolicyID != "" {
		t.Fatalf("partial lifecycle survived: %+v err=%v", state, err)
	}
	if events, err := j.PositionPolicyAudit(context.Background(), "p-policy"); err != nil || len(events) != 0 {
		t.Fatalf("partial audit survived: %+v err=%v", events, err)
	}
}
