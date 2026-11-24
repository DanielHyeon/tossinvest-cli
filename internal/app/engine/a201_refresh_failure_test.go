package engine_test

// a201_refresh_failure_test.go 는 design V4c(gstack r1 G1)를 실제 RunOnce 로 고정함.
//
// RunOnce 는 Tracker.Refresh 가 실패하면 Observe 앞에서 돌아감 — 그 주기 비교가 credit 종목을 분쟁으로 봐도 반박할 자리가
// 없어 credit 이 살아남고, 다음 깨끗한 비교가 그 credit 으로 풀었음(V4b 와 같은 원리의 둘째 조기 반환). 반환 직전에 그
// 비교의 분쟁 종목에 대해 「엄격히 나중」 반박을 남김. 같은 as-of(그 주기 수렴이 준 조정 credit)는 지우지 않음.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a201RefreshFailOnce 는 추적기 장부의 다음 ActiveReconcileStates(Refresh 의 첫 읽기)를 한 번 실패시킴.
type a201RefreshFailOnce struct {
	reconcile.ReconcileStore
	fail bool
}

func (s *a201RefreshFailOnce) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	if s.fail {
		s.fail = false
		return nil, errors.New("a201 injected refresh read failure")
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

func newA201RefreshWorld(t *testing.T) (*a201World, *a201RefreshFailOnce) {
	t.Helper()
	w := newA201World(t)
	fs := &a201RefreshFailOnce{ReconcileStore: w.tracker.Journal}
	w.tracker.Journal = fs
	return w, fs
}

// refreshFails 는 다음 주기의 Refresh 를 실패시켜 돌리고, 그 주기가 Refresh 오류로 끝났음을 단언함.
func (w *a201World) refreshFails(fs *a201RefreshFailOnce, label, symbol string) {
	w.t.Helper()
	fs.fail = true
	c := w.next()
	w.logCycle(label, symbol, c)
	if fs.fail || c.Err == nil || c.Released != 0 {
		w.t.Fatalf("%s = %+v, want the injected Refresh failure before any observation", label, c)
	}
}

// stillBlocked 는 다음 깨끗한 주기에 해제가 없고(Awaiting) 진입이 거절됨을 단언함.
func (w *a201World) stillBlocked(symbol string) {
	w.t.Helper()
	c := w.next()
	w.logCycle("다음 깨끗한 주기", symbol, c)
	if c.Err != nil || c.Released != 0 || c.Awaiting != 1 {
		w.t.Fatalf("next clean cycle = %+v, want the block awaiting and nothing released", c)
	}
	w.assertStuck(symbol)
}

// TestA201ARefreshFailureCycleStillRefutesADisputedFillCredit: V4c 체결 credit 판 — k 의 W2 체결 credit 뒤, k+1 에
// 사람이 10 주를 사(외부 포지션 — 분쟁) 그 주기 Refresh 가 실패해도 credit 은 반박되고, k+2 깨끗한 비교에서 안 풀림.
func TestA201ARefreshFailureCycleStillRefutesADisputedFillCredit(t *testing.T) {
	const symbol = "SYN201"
	w, fs := newA201RefreshWorld(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("k W2", symbol, c)
	w.assertStuck(symbol)

	w.holds(symbol, "10", "100", 100)
	w.refreshFails(fs, "k+1 외부 매수 · Refresh 실패", symbol)
	w.stillBlocked(symbol)
}

// TestA201ARefreshFailureCycleStillRefutesADisputedAdjustmentCredit: V4c 조정 credit 판 — 같은 입력에서 오늘의 조정
// credit 경로도 같이 바뀜(반박이 늘어나는 보수 방향, design §2).
func TestA201ARefreshFailureCycleStillRefutesADisputedAdjustmentCredit(t *testing.T) {
	const symbol = "SYN201"
	w, fs := newA201RefreshWorld(t)
	w.hold(symbol, "31")
	w.holds(symbol, "31", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	if c := w.next(); c.Err != nil {
		t.Fatalf("start = %+v", c)
	}
	w.holdings.items = nil // 사람이 앱으로 전량 매도(장부 체결 없음) — 수렴 31→0 · 조정 credit · QM
	c := w.next()
	w.logCycle("k 외부 매도", symbol, c)
	if c.Err != nil || c.Converged != 1 {
		t.Fatalf("k = %+v, want the convergence and its credit", c)
	}
	w.assertStuck(symbol)

	w.holds(symbol, "10", "100", 100)
	w.refreshFails(fs, "k+1 외부 매수 · Refresh 실패", symbol)
	w.stillBlocked(symbol)
}

// TestA201ARefreshFailureCycleKeepsACreditItsComparisonDoesNotDispute: V4c 대조 — Refresh 가 실패한 주기의 비교가 그
// 종목을 분쟁으로 보지 않으면 credit 은 남고, 다음 깨끗한 비교에서 FILL_APPLIED 로 풀림.
func TestA201ARefreshFailureCycleKeepsACreditItsComparisonDoesNotDispute(t *testing.T) {
	const symbol = "SYN201"
	w, fs := newA201RefreshWorld(t)
	w.start(symbol)
	w.converge.at = a201BeforePositions
	w.converge.fire = func() { w.stopFills(symbol, "31") }
	c := w.next()
	w.logCycle("k W2", symbol, c)
	w.assertStuck(symbol)

	w.refreshFails(fs, "k+1 일치 · Refresh 실패", symbol)
	w.releasedByFill(symbol)
}

// TestA201ARefreshFailureDoesNotRefuteTheCreditItsOwnConvergenceGave: V4c — 같은 주기 수렴이 준 조정 credit(같은
// as-of)은 그 주기의 Refresh 실패 반박에 안 지워짐(엄격히 나중만). 다음 깨끗한 비교에서 ADJUSTMENT_APPLIED 로 풀림.
func TestA201ARefreshFailureDoesNotRefuteTheCreditItsOwnConvergenceGave(t *testing.T) {
	const symbol = "SYN201"
	w, fs := newA201RefreshWorld(t)
	w.hold(symbol, "31")
	w.holds(symbol, "31", "100", 100)
	w.stopOpenAtBroker(symbol, "31")
	if c := w.next(); c.Err != nil {
		t.Fatalf("start = %+v", c)
	}
	w.holdings.items = nil
	w.holds(symbol, "20", "100", 100) // 사람의 앱 매도 11 주 — 수렴 31→20 · credit · QM(블록이 섬)
	c := w.next()
	w.logCycle("k-1 외부 매도 11", symbol, c)
	if c.Err != nil || c.Converged != 1 {
		t.Fatalf("k-1 = %+v, want the convergence", c)
	}
	w.assertStuck(symbol)

	w.holdings.items = nil // 사람의 앱 매도 20 주 — 그 주기 수렴 20→0 이 같은 as-of 의 credit 을 줌, Refresh 실패
	fs.fail = true
	c = w.next()
	w.logCycle("k 외부 매도 20 · Refresh 실패", symbol, c)
	if fs.fail || c.Err == nil || c.Converged != 1 {
		t.Fatalf("k = %+v, want the convergence then the injected Refresh failure", c)
	}
	c = w.next()
	w.logCycle("k+1 깨끗한 주기", symbol, c)
	if c.Err != nil || c.Released != 1 {
		t.Fatalf("k+1 = %+v, want the same-cycle adjustment credit to release", c)
	}
	w.releasedAs(symbol, journal.ReconcileReleaseAdjustmentApplied)
}
