package journal

// a087 tasks 3.8 — 청산 슬리피지 읽기 전용 질의(openspec/changes/…a087…/analysis/exit-slippage-3.8/exit_slippage.sql)를
// 이 저장소의 마이그레이션이 만든 임시 저널 위에서 돌려 결속과 산식을 고정함. 실제 저널(~/.config/tossctl)은 열지 않음.
// 질의 파일은 change 가 아카이브돼도 찾도록 활성·아카이브 두 경로를 glob 으로 찾고, 정확히 하나가 아니면 실패함
// (못 찾았다고 조용히 건너뛰면 질의가 깨져도 초록임).

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func a087SlippageQuery(t *testing.T) string {
	t.Helper()
	const rel = "a087-a-protective-exit-is-a-market-order/analysis/exit-slippage-3.8/exit_slippage.sql"
	var found []string
	for _, pattern := range []string{
		filepath.Join("..", "..", "openspec", "changes", rel),
		filepath.Join("..", "..", "openspec", "changes", "archive", "*-"+rel),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		found = append(found, matches...)
	}
	if len(found) != 1 {
		t.Fatalf("exit_slippage.sql: found %v, want exactly one", found)
	}
	body, err := os.ReadFile(found[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// a087ConfirmSell 은 확정된 KR 시장가 매도 intent 하나를 남김(가격 NULL).
func a087ConfirmSell(t *testing.T, j *Journal, intentID, orderID, day string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: intentID, Market: "kr", TradingDay: day, AccountRef: "acct-1",
			Symbol: "005930", Side: "SELL", OrderType: "MARKET", Quantity: "10",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-" + intentID,
		},
		Kind: KindPlace, AttemptID: "attempt-" + intentID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
}

func a087Fill(t *testing.T, j *Journal, orderID, day, avg string) {
	t.Helper()
	if _, err := j.RecordFill(context.Background(), FillObservation{
		OrderID: orderID, AccountRef: "acct-1", Symbol: "005930", Market: "kr", TradingDay: day,
		Side: "SELL", State: "FILLED", Terminal: true, Quantity: "10", FilledQuantity: "10",
		AveragePrice: avg, ObservedAt: "2026-03-30T00:31:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

func a087ExitEvent(t *testing.T, j *Journal, positionID string, observed any, action, intentID string) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(), `
		INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after,
		   action, proposed_intent_id, created_at)
		VALUES (?,?,NULL,NULL,NULL,?,?,'2026-03-30T00:30:30Z')`,
		positionID, observed, action, intentID); err != nil {
		t.Fatalf("exit event: %v", err)
	}
}

func TestA087ExitSlippageQueryBindsTheDecisionPriceToItsOwnFill(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "pos-1", nil)

	// 보호 청산 1 — 70000 에서 결정, 브로커 평균 69930 에 체결 → (70000-69930)/70000×1e4 = 10 bps.
	a087ConfirmSell(t, j, "intent-stop", "o-reused", "2026-03-30")
	a087Fill(t, j, "o-reused", "2026-03-30", "69930")
	a087ExitEvent(t, j, "pos-1", "70000", "BASELINE_BREACH_EXIT", "intent-stop")
	// 같은 발의의 체결 반영 행(관측가 없음) — 산출에서 빠져야 함.
	a087ExitEvent(t, j, "pos-1", nil, ExitEventProposalFilled, "intent-stop")
	// 보호 행동이지만 관측가가 없는 행 — 결정 가격이 없으면 슬리피지도 없음(행 자체를 뺌).
	a087ExitEvent(t, j, "pos-1", nil, "BASELINE_BREACH_EXIT", "intent-stop")
	// 미끼: 다음 날 같은 브로커 번호를 재사용한 다른 주문 — order_id 만으로 이으면 이 50000 이 섞임.
	a087ConfirmSell(t, j, "intent-other-day", "o-reused", "2026-03-31")
	a087Fill(t, j, "o-reused", "2026-03-31", "50000")
	// 보호 청산 2 — 아직 체결 스냅숏 없음 → bps NULL(0 으로 읽어 10000 bps 를 만들지 않음).
	a087ConfirmSell(t, j, "intent-unfilled", "o-unfilled", "2026-03-30")
	a087ExitEvent(t, j, "pos-1", "70000", "STOP_LOSS_LADDER", "intent-unfilled")

	rows, err := j.db.QueryContext(ctx, a087SlippageQuery(t))
	if err != nil {
		t.Fatalf("running the slippage query: %v", err)
	}
	defer rows.Close()
	type row struct {
		intent, avg string
		bps         sql.NullFloat64
	}
	var got []row
	for rows.Next() {
		var (
			eventID, terminal                                                     sql.NullInt64
			position, decided, action, intent, orderType, symbol, market, orderID string
			observed, avg, filled, state                                          sql.NullString
			bps                                                                   sql.NullFloat64
		)
		if err := rows.Scan(&eventID, &position, &decided, &action, &intent, &orderType, &symbol, &market,
			&orderID, &observed, &avg, &filled, &state, &terminal, &bps); err != nil {
			t.Fatal(err)
		}
		got = append(got, row{intent: intent, avg: avg.String, bps: bps})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("rows = %+v, want the two protective decisions only", got)
	}
	if got[0].intent != "intent-stop" || got[0].avg != "69930" || !got[0].bps.Valid || got[0].bps.Float64 != 10 {
		t.Fatalf("filled stop = %+v, want intent-stop @ 69930 = 10 bps (not the reused id's 50000)", got[0])
	}
	if got[1].intent != "intent-unfilled" || got[1].bps.Valid {
		t.Fatalf("unfilled stop = %+v, want NULL bps", got[1])
	}
}
