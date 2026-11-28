package engine_test

// a192_order_terminal_release_log_test.go 는 대사 구동 루프 끝까지(수집 → 비교 → Observe → 해제 → 로그) missing-order
// 종목 차단의 종결 증명 해제를 돌리고, 해제 로그에 계좌·주문번호·수량·가격이 없음을 고정함(a192 design D7).

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA192TheReleaseLogCarriesNoAccountOrderQuantityOrPrice(t *testing.T) {
	var logs bytes.Buffer
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Log = obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true})
	})
	h.tracker.Terminal = h.journal
	ctx := context.Background()

	attempt, err := h.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-o-a192-1", AccountRef: reconcileAccount, Market: "kr", TradingDay: "2026-03-30",
			Symbol: "SYN001", Side: "SELL", OrderType: "LIMIT", Quantity: "7", Price: "4321",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-o-a192-1",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-o-a192-1",
	})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, "o-a192-1"); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}

	// 1주기: 계좌 open 목록에 없음 → missing-order 블록(이 프로세스가 삽입).
	if c := h.cycle(); c.Err != nil || c.Blocked != 1 {
		t.Fatalf("first cycle = %+v, want the missing-order block", c)
	}
	h.clk.Advance(time.Minute)
	if _, err := h.journal.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-a192-1", Symbol: "SYN001", Market: "kr", TradingDay: "2026-03-30", Side: "SELL",
		AccountRef: reconcileAccount, State: "CANCELLED", Terminal: true, Quantity: "7", FilledQuantity: "0",
		ObservedAt: h.clk.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	h.clk.Advance(time.Minute)
	if c := h.cycle(); c.Err != nil || c.Released != 0 || c.Blocked != 1 {
		t.Fatalf("arming cycle = %+v, want the block kept", c)
	}
	h.clk.Advance(time.Minute)
	if c := h.cycle(); c.Err != nil || c.Released != 1 || c.Blocked != 0 {
		t.Fatalf("releasing cycle = %+v, want the block released", c)
	}

	var line string
	for _, l := range strings.Split(logs.String(), "\n") {
		if strings.Contains(l, "reconcile.block_released") {
			if line != "" {
				t.Fatalf("more than one release line:\n%s", logs.String())
			}
			line = l
		}
	}
	if line == "" {
		t.Fatalf("no reconcile.block_released line in:\n%s", logs.String())
	}
	for _, want := range []string{"SYN001", journal.ReconcileReleaseOrderTerminalConfirmed, `"bound_orders":1`} {
		if !strings.Contains(line, want) {
			t.Errorf("release line lacks %q: %s", want, line)
		}
	}
	for _, leak := range []string{reconcileAccount, "o-a192-1", "4321", `"7"`} {
		if strings.Contains(line, leak) {
			t.Errorf("release line carries %q: %s", leak, line)
		}
	}

	// 무장 주기의 유지 사유가 한 번 기록됨(M5) — 같은 사유를 매 주기 다시 쓰지 않음.
	var held []string
	for _, l := range strings.Split(logs.String(), "\n") {
		if strings.Contains(l, "reconcile.block_release_held") {
			held = append(held, l)
		}
	}
	if len(held) != 1 || !strings.Contains(held[0], `"reason":"armed"`) || !strings.Contains(held[0], "SYN001") {
		t.Fatalf("hold lines = %q, want one armed line for SYN001", held)
	}
	for _, leak := range []string{reconcileAccount, "o-a192-1", "4321"} {
		if strings.Contains(held[0], leak) {
			t.Errorf("hold line carries %q: %s", leak, held[0])
		}
	}
}
