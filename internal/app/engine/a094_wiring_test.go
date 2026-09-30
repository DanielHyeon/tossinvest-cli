package engine_test

// a094 배선 — 생산 조립(openProtectedGateEngine)이 새 critical 입구와 기동 따라잡기를 실제로 꽂는지(역할 핀).

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// D−6.1 — exit 관측기의 Critical 은 엔진 알림기(창 0 단일 입구)이고, 호출자가 넘긴 값을 덮음.
func TestA094TheExitObserverRecordsCriticalsThroughTheNotifier(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	if eng.Notifier == nil {
		t.Fatal("assembly without a notifier — this test measures nothing")
	}
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(), Critical: a094NoCritical{}})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	if got, ok := observer.OptionsForTest().Critical.(*obs.Notifier); !ok || got != eng.Notifier {
		t.Errorf("Critical = %T, want the engine *obs.Notifier", observer.OptionsForTest().Critical)
	}
}

// D−2.5 — 생산 조립의 Recovery 는 기동 따라잡기를 가짐: attempt 없는 무장 발의가 CatchUpExitProposals 로 풀림.
func TestA094TheRecoveryCarriesTheBootCatchUp(t *testing.T) {
	h := newExitHarness(t, nil)
	p := h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.a094ArmOn(p, "005930", "exit-orphan", string(exitpolicy.ActionBaselineBreach))
	released := engine.CatchUpExitProposalsForTest(context.Background(), h.journal, exitAccount, nil)
	if released != 1 || h.state(p.ID).Pending() {
		t.Fatalf("released = %d, armed = %v — an armed proposal with no attempt must be released at boot", released, h.state(p.ID).Pending())
	}

	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	rec, err := eng.Recovery(reconcile.Options{})
	if err != nil {
		t.Fatalf("Recovery: %v", err)
	}
	if !rec.CatchUpWired() {
		t.Fatal("the production recovery carries no boot catch-up — a proposal left armed over a settled attempt would stay armed forever")
	}
	// 조립된 클로저 자체를 돌림(C P2#3): 엔진 원장 · 엔진 계좌에 attempt 없는 무장 발의를 심고 풀리는지.
	positionID := a094SeedOrphanProposal(t, eng.Journal, eng.AccountRef)
	rec.CatchUpExitProposals(context.Background())
	st, err := eng.Journal.ExitState(context.Background(), positionID)
	if err != nil {
		t.Fatal(err)
	}
	if st.Pending() {
		t.Fatal("the production catch-up closure did not release an armed proposal with no attempt")
	}
}

// a094SeedOrphanProposal 은 체결된 진입 · 열린 exit 상태 · attempt 없는 무장 발의 하나를 원장 API 로 만듦.
func a094SeedOrphanProposal(t *testing.T, j *journal.Journal, account string) string {
	t.Helper()
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	limits, err := execgw.EncodeLimits(execgw.Limits{MaxQuantity: execgw.Bound(1000), MaxNotional: execgw.Bound(1e9),
		MaxTotalExposure: execgw.Bound(1e9), MaxDailyLossAmount: execgw.Bound(1e6), MaxDailyLossRatio: execgw.Bound(0.02), Currency: "KRW"})
	must(err)
	now := time.Now().UTC()
	_, err = j.RecordDecision(ctx, journal.DecisionRequest{ID: "d-orphan", AccountRef: account, SafetyClass: journal.SafetyClassExposureRaising,
		Kind: journal.KindPlace, Preimage: journal.RiskIntent{AccountRef: account, Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "10",
			EntryPrice: "70000", StopPrice: "68000", TargetPrice: "999999", PolicyVersion: "test/v1"},
		LimitsJSON: limits, Nonce: "nonce-d-orphan", IssuedAt: now, ExpiresAt: now.Add(time.Hour)})
	must(err)
	entry, err := j.Prepare(ctx, journal.PrepareRequest{Intent: journal.Intent{ID: "i-orphan-entry", Market: "kr",
		TradingDay: "2026-03-30", AccountRef: account, Symbol: "005930", Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY",
		Quantity: "10", Price: "70000", Currency: "KRW", Source: "engine/test", Fingerprint: "fp-orphan"},
		Kind: journal.KindPlace, AttemptID: "a-orphan-entry", AccountRef: account, DecisionID: "d-orphan",
		SafetyClass: journal.SafetyClassExposureRaising, ClientOrderID: journal.DeriveClientOrderID("d-orphan", 0)})
	must(err)
	must(entry.MarkDispatchStarted(ctx))
	must(entry.MarkAcked(ctx, "O-orphan-entry"))
	must(entry.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""))
	_, err = j.RecordFill(ctx, journal.FillObservation{OrderID: "O-orphan-entry", Symbol: "005930", Market: "kr", AccountRef: account,
		TradingDay: "2026-03-30", Side: "BUY", State: "CLOSED_FILLED", Terminal: true, Quantity: "10", FilledQuantity: "10",
		AveragePrice: "70000", ObservedAt: now.Format(time.RFC3339)})
	must(err)
	p, err := j.CurrentPosition(ctx, account, "kr", "005930")
	must(err)
	_, err = j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000"})
	must(err)
	must(j.RecordExitJudgement(ctx, journal.ExitJudgement{PositionID: p.ID, ObservedPrice: "67900", HighWater: "70000",
		Baseline: "68000", RatchetLevel: journal.RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &journal.ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: "NONE", IntentID: "exit-orphan"}}))
	return p.ID
}

type a094NoCritical struct{}

func (a094NoCritical) RecordCritical(context.Context, obs.Event, time.Duration) error {
	return nil
}
