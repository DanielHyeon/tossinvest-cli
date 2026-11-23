package engine_test

// a198 D4 재결정(gstack review r2 G1, codex P1) — 길 치우기 취소 실패 줄도 그 주기의 모든 판정·제출·해제 뒤에 씀
// (정본 exit-policy 「엔진이 낸 청산 주문은 운영 로그에 … 줄을 남긴다」의 쓰기 시점 규칙, a129 대기열과 같은 자리).
// 포지션 둘: 앞(000660 — 위치 순서가 종목 순)은 엔진 매수 미체결의 취소가 거절되고, 뒤(005930)는 손절만 냄.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// entryAs 는 e2eStack.entry 와 같은 진입(결정·intent·attempt·전량 체결)을 key 로 식별자를 갈라 남김 — 포지션 둘을 세우려고.
func (s *e2eStack) entryAs(key, symbol, quantity, limit, stop string) journal.Position {
	s.t.Helper()
	ctx := context.Background()
	limits, err := execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(100), MaxNotional: execgw.Bound(1_000_000),
		MaxTotalExposure: execgw.Bound(10_000_000), MaxDailyLossAmount: execgw.Bound(100_000),
		MaxDailyLossRatio: execgw.Bound(0.01), Currency: "KRW",
	})
	if err != nil {
		s.t.Fatalf("EncodeLimits: %v", err)
	}
	decisionID := "d-e2e-entry-" + key
	if _, err := s.engine.Journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: s.engine.AccountRef,
		SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: s.engine.AccountRef, Market: "kr", Symbol: symbol, Side: "BUY",
			Quantity: quantity, EntryPrice: limit, StopPrice: stop, TargetPrice: "90000",
			PolicyVersion: "add-core-domain/7.6",
		},
		LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: s.clk.Now(), ExpiresAt: s.clk.Now().Add(time.Hour),
	}); err != nil {
		s.t.Fatalf("RecordDecision: %v", err)
	}
	attempt, err := s.engine.Journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "i-e2e-entry-" + key, Market: "kr", TradingDay: "2026-03-30",
			AccountRef: s.engine.AccountRef, Symbol: symbol, Side: "BUY",
			OrderType: "LIMIT", TimeInForce: "DAY", Quantity: quantity, Price: limit,
			Currency: "KRW", Source: "engine", Fingerprint: "fp-e2e-entry-" + key,
		},
		Kind: journal.KindPlace, AttemptID: "a-e2e-entry-" + key,
		AccountRef: s.engine.AccountRef, DecisionID: decisionID,
		SafetyClass:   journal.SafetyClassExposureRaising,
		ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		s.t.Fatalf("Prepare(entry): %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		s.t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, "BO-entry-"+key); err != nil {
		s.t.Fatalf("MarkAcked: %v", err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		s.t.Fatalf("Settle: %v", err)
	}
	s.fill("BO-entry-"+key, symbol, "BUY", quantity, quantity, limit, true)
	p, err := s.engine.Journal.CurrentPosition(ctx, s.engine.AccountRef, "kr", symbol)
	if err != nil {
		s.t.Fatalf("CurrentPosition: %v", err)
	}
	return p
}

// twoPositions 는 앞 포지션(000660, 매수 미체결 W-BUY-A — 그 취소는 400 거절)과 뒤 포지션(005930, 미체결 없음)을 세움.
func twoPositions(t *testing.T) *a198Fixture {
	t.Helper()
	f := newA198Fixture(t)
	f.s.entryAs("a", "000660", "10", "70000", "68000")
	a198WorkingOrder(t, f.s, f.broker, "000660", "BUY", "W-BUY-A", "3", "69000")
	f.p = f.s.entryAs("b", "005930", "10", "70000", "68000")
	f.broker.quote("000660", "70100")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.refuseCancels(http.StatusBadRequest)
	f.broker.quote("000660", "67900")
	f.broker.quote("005930", "67900")
	return f
}

// (1) 취소 실패 줄이 쓰이는 순간, 뒤 포지션의 손절 제출은 이미 끝나 있다.
func TestA198TheCancelFailureLineIsWrittenAfterTheCyclesSubmissions(t *testing.T) {
	f := twoPositions(t)
	sellsAtLine := -1
	f.sink.onWrite = func(line []byte) {
		if sellsAtLine < 0 && strings.Contains(string(line), `"kind":"cancel"`) {
			sellsAtLine = len(f.sells())
		}
	}
	f.s.observe()
	f.sink.onWrite = nil
	if sellsAtLine < 0 {
		t.Fatalf("no cancel failure line was written (sink %s)", f.sink.String())
	}
	if sellsAtLine != 1 {
		t.Fatalf("the cancel failure line was written with %d stop(s) submitted, want the later position's stop already out (1)", sellsAtLine)
	}
}

// (2) 쓰기가 막힌 writer 에서도 그 주기의 뒤 포지션 손절은 나간다.
func TestA198ABlockedLogWriterDoesNotHoldTheNextPositionsStop(t *testing.T) {
	f := twoPositions(t)
	release := make(chan struct{})
	f.sink.onWrite = func(line []byte) {
		if strings.Contains(string(line), `"kind":"cancel"`) {
			<-release
		}
	}
	done := make(chan struct{})
	go func() { f.s.observe(); close(done) }()
	deadline := time.Now().Add(5 * time.Second)
	for len(f.sells()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	got := len(f.sells())
	close(release)
	<-done
	f.sink.onWrite = nil
	if got != 1 {
		t.Fatalf("broker sells = %d while the log writer was blocked, want the later position's stop (1)", got)
	}
}

// (4) 대기열은 ObserveOnce 의 모든 return·패닉 경로에서 비워짐 — 그 자리는 ObserveOnce 첫 문장의 defer(endExitOrderCycle)이고,
// endExitOrderCycle 이 취소 실패 줄 대기열을 비움(구조: 판정 경로 안에서 바로 쓰는 호출이 없음).
func TestA198TheCancelFailureQueueIsFlushedByTheCycleEnd(t *testing.T) {
	fset := token.NewFileSet()
	calls := func(path, fn string) map[string]bool {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					if s, ok := c.Fun.(*ast.SelectorExpr); ok {
						out[s.Sel.Name] = true
					}
				}
				return true
			})
		}
		return out
	}
	if !calls("exit_order_log.go", "endExitOrderCycle")["flushClearCancelFailures"] {
		t.Fatal("endExitOrderCycle does not flush the clearing-cancel failure queue")
	}
	if calls("a198_clear_cancel_log.go", "noteClearCancelFailure")["Warn"] {
		t.Fatal("noteClearCancelFailure writes the line inside the judgement path")
	}
}
