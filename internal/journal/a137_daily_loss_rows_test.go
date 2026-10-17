package journal

// a137 원장 읽기 시험 — 관리 · 조정 종결(decimal 0) · 수명과 겹친 엔진 매도(SELL PLACE · 보호 CREATE/REPLACE) · 성과 조인 ·
// 계좌 한정 · 형식이 어긋난 closed_at.

import (
	"context"
	"errors"
	"testing"
	"time"
)

func a137Exec(t *testing.T, j *Journal, query string, args ...any) {
	t.Helper()
	if _, err := j.db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("%v\n%s", err, query)
	}
}

// a137Closed 는 CLOSED 관리 포지션 하나를 만듦(성과 금액이 비면 성과 행 없음).
func a137Closed(t *testing.T, j *Journal, id, account, market, symbol string, seq int, opened, closed, pnl string) {
	t.Helper()
	a137Exec(t, j, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		quantity, avg_price, opened_at, closed_at) VALUES (?,?,?,?,?,'d-1','CLOSED','0','0',?,?)`,
		id, account, market, symbol, seq, opened, closed)
	if pnl != "" {
		a137Exec(t, j, `INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
			initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at) VALUES (?,?, '0','1','1',0,NULL,NULL,?)`,
			id, pnl, closed)
	}
}

func a137Adjust(t *testing.T, j *Journal, id, positionID, newQuantity, at string) {
	t.Helper()
	a137Exec(t, j, `INSERT INTO position_adjustments (id, position_id, kind, expected_prev_quantity, prev_quantity, new_quantity,
		broker_as_of, evidence, created_at) VALUES (?,?,'MANUAL','5','5',?,?, 'holdings read',?)`, id, positionID, newQuantity, at, at)
}

func a137Window() (time.Time, time.Time) {
	return time.Date(2026, 3, 29, 22, 0, 0, 0, time.UTC), time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
}

func a137Rows(t *testing.T, j *Journal, account string) map[string]DailyLossRow {
	t.Helper()
	from, to := a137Window()
	rows, err := j.DailyLossRows(context.Background(), account, from, to)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]DailyLossRow{}
	for _, r := range rows {
		out[r.PositionID] = r
	}
	return out
}

func TestA137DailyLossRowsJoinOutcomesAndScopeToTheManagedAccount(t *testing.T) {
	j := openTestJournal(t)
	insertDecision(t, j, "d-1", "nonce-1")
	a137Closed(t, j, "p-loss", "acct-1", "kr", "005930", 1, "2026-03-30T00:00:00Z", "2026-03-30T01:00:00Z", "-5000")
	a137Closed(t, j, "p-noout", "acct-1", "us", "AAPL", 1, "2026-03-30T00:00:00Z", "2026-03-30T02:00:00Z", "")
	a137Closed(t, j, "p-other", "acct-2", "kr", "005930", 1, "2026-03-30T00:00:00Z", "2026-03-30T01:00:00Z", "-1")
	a137Closed(t, j, "p-old", "acct-1", "kr", "000660", 1, "2026-03-20T00:00:00Z", "2026-03-20T01:00:00Z", "-9")
	a137Exec(t, j, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price, closed_at)
		VALUES ('p-external','acct-1','kr','035420',1,'CLOSED','0','0','2026-03-30T01:00:00Z')`)

	got := a137Rows(t, j, "acct-1")
	if len(got) != 2 {
		t.Fatalf("rows = %v, want the two managed acct-1 closes in the window", got)
	}
	if r := got["p-loss"]; r.Outcome == nil || *r.Outcome != "-5000" || r.AdjustmentClosed || r.Market != "kr" {
		t.Fatalf("p-loss = %+v", r)
	}
	if r := got["p-noout"]; r.Outcome != nil {
		t.Fatalf("p-noout must carry no outcome, got %+v", r)
	}
}

func TestA137AnAdjustmentToDecimalZeroIsAnAdjustmentClose(t *testing.T) {
	j := openTestJournal(t)
	insertDecision(t, j, "d-1", "nonce-1")
	a137Closed(t, j, "p-adj", "acct-1", "kr", "005930", 1, "2026-03-30T00:00:00Z", "2026-03-30T03:00:00Z", "")
	a137Adjust(t, j, "adj-1", "p-adj", "2", "2026-03-30T01:00:00Z")
	a137Adjust(t, j, "adj-2", "p-adj", "0.000", "2026-03-30T03:00:00Z")
	a137Closed(t, j, "p-fill", "acct-1", "kr", "000660", 1, "2026-03-30T00:00:00Z", "2026-03-30T03:00:00Z", "-1")
	a137Adjust(t, j, "adj-3", "p-fill", "3", "2026-03-30T01:00:00Z")

	got := a137Rows(t, j, "acct-1")
	if !got["p-adj"].AdjustmentClosed || got["p-adj"].EngineSellDispatched {
		t.Fatalf("p-adj = %+v, want an adjustment close with no engine sell", got["p-adj"])
	}
	if got["p-fill"].AdjustmentClosed {
		t.Fatalf("a last adjustment to 3 is not an adjustment close: %+v", got["p-fill"])
	}
}

// 조정 종결 + 수명과 겹친 발송 엔진 SELL PLACE → 엔진 매도가 섞였을 수 있음. NOT_DISPATCHED · 수명 밖 · BUY 는 세지 않음.
func TestA137AnOverlappingDispatchedEngineSellIsSeen(t *testing.T) {
	cases := map[string]struct {
		side, kind, state, recorded string
		want                        bool
	}{
		"dispatched sell in lifetime":  {"SELL", "PLACE", "CONFIRMED", "2026-03-30T01:00:00Z", true},
		"in-doubt sell in lifetime":    {"SELL", "PLACE", "IN_DOUBT", "2026-03-30T01:00:00Z", true},
		"not dispatched":               {"SELL", "PLACE", "NOT_DISPATCHED", "2026-03-30T01:00:00Z", false},
		"failed confirmed":             {"SELL", "PLACE", "FAILED_CONFIRMED", "2026-03-30T01:00:00Z", false},
		"cancel attempt":               {"SELL", "CANCEL", "CONFIRMED", "2026-03-30T01:00:00Z", false},
		"buy":                          {"BUY", "PLACE", "CONFIRMED", "2026-03-30T01:00:00Z", false},
		"after the close":              {"SELL", "PLACE", "CONFIRMED", "2026-03-30T05:00:00Z", false},
		"unparseable recorded instant": {"SELL", "PLACE", "CONFIRMED", "garbage", true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			j := openTestJournal(t)
			insertDecision(t, j, "d-1", "nonce-1")
			a137Closed(t, j, "p-adj", "acct-1", "kr", "005930", 1, "2026-03-30T00:00:00Z", "2026-03-30T03:00:00Z", "")
			a137Adjust(t, j, "adj-0", "p-adj", "0", "2026-03-30T03:00:00Z")
			a137Exec(t, j, `INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,
				time_in_force,quantity,price,currency,source,fingerprint,notes)
				VALUES ('i-1','2026-03-30T01:00:00Z','kr','2026-03-30','acct-1','005930',?,'LIMIT','DAY','1','1','KRW','exit-policy','fp-1','')`, c.side)
			a137Exec(t, j, `INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,fingerprint,recorded_at)
				VALUES ('a-1','i-1',?,?,1,'fp-1',?)`, c.kind, c.state, c.recorded)
			if got := a137Rows(t, j, "acct-1")["p-adj"].EngineSellDispatched; got != c.want {
				t.Fatalf("EngineSellDispatched = %v, want %v", got, c.want)
			}
		})
	}
}

// 브로커에 놓였을 수 있는 보호 주문(CREATE · REPLACE, PLANNED 아님)이 수명과 겹치면 엔진 매도가 섞였을 수 있음(§4.2 P3-3 — 수명 안 생성만이 아님).
func TestA137AnOverlappingProtectionOrderIsSeen(t *testing.T) {
	cases := map[string]struct {
		kind, state, created, sagaState, sagaUpdated string
		want                                         bool
	}{
		"create acknowledged in lifetime":   {"CREATE", "ACKNOWLEDGED", "2026-03-30T01:00:00Z", "CLOSED", "2026-03-30T03:00:00Z", true},
		"created before, still live":        {"CREATE", "ACKNOWLEDGED", "2026-03-29T23:00:00Z", "ACTIVE", "2026-03-29T23:00:00Z", true},
		"replace dispatched":                {"REPLACE", "DISPATCHED", "2026-03-30T01:00:00Z", "CLOSED", "2026-03-30T03:00:00Z", true},
		"planned only":                      {"CREATE", "PLANNED", "2026-03-30T01:00:00Z", "ACTIVE", "2026-03-30T01:00:00Z", false},
		"cancel":                            {"CANCEL", "DISPATCHED", "2026-03-30T01:00:00Z", "CLOSED", "2026-03-30T03:00:00Z", false},
		"ended before the instance opened":  {"CREATE", "ACKNOWLEDGED", "2026-03-29T20:00:00Z", "CLOSED", "2026-03-29T21:00:00Z", false},
		"created after the instance closed": {"CREATE", "ACKNOWLEDGED", "2026-03-30T04:00:00Z", "ACTIVE", "2026-03-30T04:00:00Z", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			j := openTestJournal(t)
			insertDecision(t, j, "d-1", "nonce-1")
			a137Closed(t, j, "p-adj", "acct-1", "kr", "005930", 1, "2026-03-30T00:00:00Z", "2026-03-30T03:00:00Z", "")
			a137Adjust(t, j, "adj-0", "p-adj", "0", "2026-03-30T03:00:00Z")
			a137Exec(t, j, `INSERT INTO protection_sagas(saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,client_order_id,updated_at)
				VALUES ('s-1','acct-1','default','KR','005930',1,1,?,1,1,'c-1',?)`, c.sagaState, c.sagaUpdated)
			a137Exec(t, j, `INSERT INTO protection_mutation_attempts(attempt_id,saga_id,generation,kind,state,serializer_version,canonical_body,created_at,updated_at)
				VALUES ('pa-1','s-1',1,?,?,1,'{}',?,?)`, c.kind, c.state, c.created, c.created)
			if got := a137Rows(t, j, "acct-1")["p-adj"].EngineSellDispatched; got != c.want {
				t.Fatalf("EngineSellDispatched = %v, want %v", got, c.want)
			}
		})
	}
}

// 사전 필터(문자열 범위)가 떨어뜨릴 수 있는 형식 결함 행은 따로 세어 측정 불가로 함 — 오늘의 손실일 수 있음.
func TestA137AMalformedClosedAtMakesTheReadFail(t *testing.T) {
	for name, closedAt := range map[string]any{
		"null":   nil,
		"empty":  "",
		"offset": "2026-03-30T10:00:00+09:00",
		"words":  "yesterday",
	} {
		t.Run(name, func(t *testing.T) {
			j := openTestJournal(t)
			insertDecision(t, j, "d-1", "nonce-1")
			a137Exec(t, j, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
				quantity, avg_price, closed_at) VALUES ('p-bad','acct-1','kr','005930',1,'d-1','CLOSED','0','0',?)`, closedAt)
			from, to := a137Window()
			if _, err := j.DailyLossRows(context.Background(), "acct-1", from, to); !errors.Is(err, ErrDailyLossLedgerMalformed) {
				t.Fatalf("err = %v, want ErrDailyLossLedgerMalformed", err)
			}
		})
	}
	// 다른 계좌 · 관리 대상이 아닌 행의 형식 결함은 이 계좌의 측정을 막지 않음.
	j := openTestJournal(t)
	insertDecision(t, j, "d-1", "nonce-1")
	a137Exec(t, j, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price, closed_at)
		VALUES ('p-ext','acct-1','kr','035420',1,'CLOSED','0','0','junk')`)
	a137Exec(t, j, `INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state, quantity, avg_price, closed_at)
		VALUES ('p-acct2','acct-2','kr','035420',1,'d-1','CLOSED','0','0','junk')`)
	from, to := a137Window()
	if _, err := j.DailyLossRows(context.Background(), "acct-1", from, to); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}
