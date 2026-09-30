package engine_test

// a094 배선 — 생산 조립(openProtectedGateEngine)이 새 critical 입구와 기동 따라잡기를 실제로 꽂는지(역할 핀).

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
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
}

type a094NoCritical struct{}

func (a094NoCritical) RecordCritical(context.Context, obs.Event, time.Duration) error {
	return nil
}
