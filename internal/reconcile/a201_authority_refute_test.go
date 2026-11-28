package reconcile_test

// a201_authority_refute_test.go 는 design V4b(§3-12)의 조정 credit 판을 고정함 — 체결 credit 판은 a201_fill_credit_test.go.
//
// 입력: 계좌 범위 영구 승격이 쓰기 실패로 대기 중 + 그 뒤의 깨끗한 비교가 credit 종목을 외부 포지션으로만 봄 + Observe 안
// 권위 확인(ActiveReconcileStates) 실패. 깨끗한 갈래는 그 오류로 일찍 돌아가 Awaiting 처분도 끝 루프 반박도 없으므로, 편집
// 전에는 명시적으로 관측된 분쟁이 credit 을 소멸시키지 못하고 뒤의 일치가 그 credit 으로 풂. 편집 뒤에는 반환 직전 반박.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a201AuthorityStore 는 계좌 범위 행 삽입을 커밋 전에 실패시키고(대기 중 영구 승격), failReads 동안 권위 읽기를 실패시킴.
type a201AuthorityStore struct {
	reconcile.ReconcileStore
	failReads bool
}

func (s *a201AuthorityStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if req.Symbol == "" {
		return journal.ReconcileState{}, false, errors.New("a201 injected account-wide enter failure before commit")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *a201AuthorityStore) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	if s.failReads {
		return nil, errors.New("a201 injected journal authority read failure")
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

// a201PendingPermanentWorld 는 AAPL 수량 불일치를 임계까지 반복해 계좌 범위 승격을 대기 중(쓰기 실패)으로 만든 추적기임.
// 반환하는 문자열은 임계 비교의 as-of — credit 은 그 비교에 대해 세움(그 뒤 비교가 엄격히 나중).
func a201PendingPermanentWorld(t *testing.T) (*clock.Fake, *journal.Journal, *a201AuthorityStore, *reconcile.Tracker, string) {
	t.Helper()
	clk := clock.NewFake(asOf)
	j := openJournal(t)
	store := &a201AuthorityStore{ReconcileStore: j}
	tracker := a110Tracker(clk, noStaleGate(clk), store)
	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))
		a110Advance(clk)
	}
	threshold := mismatchDiffAt(clk, "AAPL", "10", "4")
	a110ObserveError(t, tracker, threshold)
	if !tracker.Permanent() {
		t.Fatal("fixture did not leave a pending account-wide promotion")
	}
	return clk, j, store, tracker, threshold.AsOf
}

// a201ExternalOnly 는 AAPL 을 외부 포지션으로만 보는 깨끗한 비교임(BlocksEntry 거짓 — 깨끗한 갈래로 감).
func a201ExternalOnly(clk clock.Clock) reconcile.Diff {
	diff := cleanDiffAt(clk)
	diff.ExternalPos = []reconcile.ExternalPosition{{Holding: reconcile.Holding{Symbol: "AAPL", Quantity: "3", Market: "us"}}}
	return diff
}

func a201ReleaseCause(t *testing.T, j *journal.Journal, symbol string) (string, bool) {
	t.Helper()
	history, err := j.ReconcileStateHistory(context.Background(), "acct-7")
	if err != nil {
		t.Fatalf("ReconcileStateHistory: %v", err)
	}
	for _, s := range history {
		if s.Symbol == symbol && !s.AccountWide() {
			return s.ReleaseCause, s.Active()
		}
	}
	t.Fatalf("no %s row in %+v", symbol, history)
	return "", false
}

// TestA201AnAuthorityErrorOnACleanComparisonRefutesAnAdjustmentCredit: V4b 조정 credit 판 — 권위 오류 주기에 관측된
// 외부 포지션 분쟁이 credit 을 반박해, 권위가 돌아온 뒤의 일치 비교가 그 credit 으로 풀지 못함(Awaiting).
func TestA201AnAuthorityErrorOnACleanComparisonRefutesAnAdjustmentCredit(t *testing.T) {
	clk, j, store, tracker, threshold := a201PendingPermanentWorld(t)
	tracker.AdjustmentApplied(threshold, "AAPL")

	a110Advance(clk)
	store.failReads = true
	if out, err := tracker.Observe(context.Background(), a201ExternalOnly(clk)); err == nil {
		t.Fatalf("external-only Observe succeeded despite the authority outage: %+v", out)
	}
	store.failReads = false

	a110Advance(clk)
	out := a110Observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared %+v — the credit the external position disputed must not release", out.Cleared)
	}
	if len(out.AwaitingAdjustment) != 1 || out.AwaitingAdjustment[0].Symbol != "AAPL" {
		t.Fatalf("awaiting %+v, want AAPL", out.AwaitingAdjustment)
	}
	if cause, active := a201ReleaseCause(t, j, "AAPL"); !active {
		t.Fatalf("AAPL row released with %q, want it active", cause)
	}
}

// TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged: V4b 대조 — 같은 입력에서 권위 읽기가 성공하면 오늘도 그 credit
// 은 B10(분쟁 → Awaiting)과 끝 루프(분쟁 → 삭제)로 소멸하고, 다음 일치에서 풀리지 않음(편집 전후 같음).
func TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged(t *testing.T) {
	clk, j, _, tracker, threshold := a201PendingPermanentWorld(t)
	tracker.AdjustmentApplied(threshold, "AAPL")

	a110Advance(clk)
	out := a110Observe(t, tracker, a201ExternalOnly(clk))
	if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("external-only outcome cleared=%+v awaiting=%+v, want AAPL awaiting", out.Cleared, out.AwaitingAdjustment)
	}

	a110Advance(clk)
	out = a110Observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 || len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("clean outcome cleared=%+v awaiting=%+v, want AAPL awaiting", out.Cleared, out.AwaitingAdjustment)
	}
	if cause, active := a201ReleaseCause(t, j, "AAPL"); !active {
		t.Fatalf("AAPL row released with %q, want it active", cause)
	}
}
