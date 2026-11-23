package journal

// a198 보강 2(D10) — IN_DOUBT 취소를 대상 주문의 종결 기록으로 닫는 원장 함수와 판정 표의 단위 시험.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
)

const a198Source = "engine/test"

// a198Owner 는 testIntent 범위(acct-1·us·AAPL·BUY)의 소유 PLACE 를 번호 orderID 로 접수 확정함(거래일·intent 접미를 정함).
func a198Owner(t *testing.T, j *Journal, key, orderID, day string) {
	t.Helper()
	intent := testIntent()
	intent.ID, intent.Fingerprint, intent.TradingDay = "owner-"+key, "fp-owner-"+key, day
	a, err := j.Prepare(context.Background(), PrepareRequest{Intent: intent, Kind: KindPlace, AttemptID: "owner-att-" + key})
	if err != nil {
		t.Fatalf("Prepare(owner %s): %v", key, err)
	}
	a165Confirm(t, a, orderID)
}

// a198InDoubtCancel 은 그 번호를 겨눈 취소를 IN_DOUBT 로 남김. mutate 로 intent 를 바꿀 수 있음.
func a198InDoubtCancel(t *testing.T, j *Journal, key, target string, mutate func(*Intent)) *Attempt {
	t.Helper()
	ctx := context.Background()
	intent := testIntent()
	intent.ID, intent.Fingerprint = "cancel-"+key, "fp-cancel-"+key
	if mutate != nil {
		mutate(&intent)
	}
	a, err := j.Prepare(ctx, PrepareRequest{Intent: intent, Kind: KindCancel, AttemptID: "cancel-att-" + key, TargetOrderID: target})
	if err != nil {
		t.Fatalf("Prepare(cancel %s): %v", key, err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.MarkInDoubt(ctx, ReasonDispatchAmbiguous, "5xx"); err != nil {
		t.Fatal(err)
	}
	return a
}

// a198Fill 은 체결 감지 입구로 종결(또는 비종결) 기록을 남김.
func a198Fill(t *testing.T, j *Journal, orderID, day, side, state string, terminal bool, filled string) {
	t.Helper()
	if _, err := j.RecordFill(context.Background(), FillObservation{
		OrderID: orderID, AccountRef: "acct-1", Symbol: "AAPL", Market: "us", TradingDay: day, Side: side,
		State: state, Terminal: terminal, Quantity: "10", FilledQuantity: filled,
		ObservedAt: time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("RecordFill(%s %s): %v", orderID, state, err)
	}
}

func a198State(t *testing.T, j *Journal, id string) AttemptRecord {
	t.Helper()
	rec, err := j.LookupAttempt(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// 판정 표 — brokerstate 의 State 전부(11 값)를 덮음.
func TestA198CancelVerdictCoversEveryBrokerState(t *testing.T) {
	want := map[brokerstate.State]AttemptState{
		brokerstate.StateCancelled:                StateConfirmed,
		brokerstate.StateCancelledPartiallyFilled: StateConfirmed,
		brokerstate.StateFilled:                   StateFailedConfirmed,
		brokerstate.StateRejected:                 StateFailedConfirmed,
		brokerstate.StateRejectedPartiallyFilled:  StateFailedConfirmed,
	}
	all := []brokerstate.State{
		brokerstate.StateOpenUnfilled, brokerstate.StateOpenPartiallyFilled, brokerstate.StateFilled,
		brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled, brokerstate.StateReplaced,
		brokerstate.StateRejected, brokerstate.StateRejectedPartiallyFilled, brokerstate.StateCancelPending,
		brokerstate.StateReplacePending, brokerstate.StateUnknown,
	}
	for _, s := range append(all, "", "SOMETHING_NEW") {
		got, reason, ok := CancelVerdictForClosedTarget(string(s))
		w, closes := want[s]
		if ok != closes || got != w {
			t.Errorf("%q: got (%s, %v), want (%s, %v)", s, got, ok, w, closes)
		}
		switch {
		case ok && got == StateConfirmed && reason != ReasonResolvedFound,
			ok && got == StateFailedConfirmed && reason != ReasonResolvedAbsent:
			t.Errorf("%q: reason %q", s, reason)
		}
	}
}

// 상수 분리 뒤 결합 술어가 분리 전(base e1eaf720c 의 exit_proposal_release.go:281-294)과 바이트 동일.
func TestA198TheTerminalPredicateIsUnchangedBySplittingIt(t *testing.T) {
	const before = "EXISTS (\n" +
		"\t\t   SELECT 1 FROM all_fill_snapshots f\n" +
		"\t\t    WHERE f.order_id = c.order_id AND f.terminal = 1\n" +
		"\t\t      AND c.ownership_at < f.committed_at\n" +
		"\t\t      AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(c.symbol))\n" +
		"\t\t      AND UPPER(TRIM(f.market)) = UPPER(TRIM(c.market))\n" +
		"\t\t      AND (TRIM(f.side) = '' OR UPPER(TRIM(f.side)) = UPPER(TRIM(c.side)))\n" +
		"\t\t      AND ((TRIM(f.account_ref) = TRIM(c.account_ref)\n" +
		"\t\t            AND TRIM(f.trading_day) = TRIM(c.trading_day))\n" +
		"\t\t           OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = ''\n" +
		"\t\t               AND TRIM(f.side) = '' AND EXISTS (\n" +
		"\t\t                 SELECT 1 FROM legacy_snapshot_owner owner\n" +
		"\t\t                  WHERE owner.order_id = f.order_id AND owner.intent_id = c.intent_id)))\n" +
		"\t\t )"
	if confirmedOrderTerminalEvidence != before {
		t.Fatalf("the shared terminal predicate changed:\n%q\nwant\n%q", confirmedOrderTerminalEvidence, before)
	}
}

func TestA198AnEngineCancelClosesOnItsTargetsClosingRecord(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		state, filled string
		want          AttemptState
		reason        string
		target        string
	}{
		{"CANCELLED", "0", StateConfirmed, ReasonResolvedFound, "O-1"},
		{"FILLED", "10", StateFailedConfirmed, ReasonResolvedAbsent, ""},
	} {
		t.Run(tc.state, func(t *testing.T) {
			j := openTestJournal(t)
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", tc.state, true, tc.filled)
			got, err := j.CloseInDoubtCancelOnTargetClose(ctx, c.ID(), a198Source)
			if err != nil || !got.Closed || got.State != tc.want || got.Reason != tc.reason {
				t.Fatalf("close = %+v err %v, want closed %s|%s", got, err, tc.want, tc.reason)
			}
			rec := a198State(t, j, c.ID())
			if rec.State != tc.want || rec.ReasonCode != tc.reason || rec.BrokerOrderID != tc.target || rec.SettledAt == "" {
				t.Fatalf("stored %+v", rec)
			}
		})
	}
}

// 닫지 않는 모양 전부 — 무변·오류 없음.
func TestA198ACancelIsNotClosedOutsideItsTargetsOwnScope(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, j *Journal) *Attempt
		src   string
	}{
		{"no closing record", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			return a198InDoubtCancel(t, j, "a", "O-1", nil)
		}, a198Source},
		{"open record", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "OPEN_PARTIALLY_FILLED", false, "3")
			return c
		}, a198Source},
		{"closing state without the terminal flag", func(t *testing.T, j *Journal) *Attempt {
			// 미체결 목록이 살아 있다고 보는 행(terminal 0) — state 문자열만으로 닫지 않음(같은 정의).
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", false, "0")
			return c
		}, a198Source},
		{"replaced with a successor", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "REPLACED", true, "3")
			return c
		}, a198Source},
		{"terminal row that later failed closed", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED_PARTIALLY_FILLED", true, "3")
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED_PARTIALLY_FILLED", true, "2")
			return c
		}, a198Source},
		{"closing record of another trading day", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-31", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"closing record of the other side", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "SELL", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"two owners across trading days", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			a198Owner(t, j, "b", "O-1", "2026-03-27")
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"no owner (an external order)", func(t *testing.T, j *Journal) *Attempt {
			c := a198InDoubtCancel(t, j, "a", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"owner in another account", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", func(i *Intent) { i.AccountRef = "acct-2" })
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"owner on the other side", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", func(i *Intent) { i.Side = "SELL" })
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"owner of another symbol", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", func(i *Intent) { i.Symbol = "MSFT" })
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
		{"a cancel from another producer", func(t *testing.T, j *Journal) *Attempt {
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			c := a198InDoubtCancel(t, j, "a", "O-1", func(i *Intent) { i.Source = "flatten-all" })
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
			return c
		}, a198Source},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			c := tc.setup(t, j)
			got, err := j.CloseInDoubtCancelOnTargetClose(ctx, c.ID(), tc.src)
			if err != nil || got.Closed {
				t.Fatalf("close = %+v err %v, want untouched", got, err)
			}
			if rec := a198State(t, j, c.ID()); rec.State != StateInDoubt {
				t.Fatalf("cancel moved to %s", rec.State)
			}
		})
	}
}

// 전일 주문 — 소유 intent 의 거래일로 찾음(취소 intent 는 오늘).
func TestA198APreviousDayOrdersCancelClosesOnThatDaysRecord(t *testing.T) {
	j := openTestJournal(t)
	a198Owner(t, j, "a", "O-1", "2026-03-27")
	c := a198InDoubtCancel(t, j, "a", "O-1", nil)
	a198Fill(t, j, "O-1", "2026-03-27", "BUY", "CANCELLED", true, "0")
	got, err := j.CloseInDoubtCancelOnTargetClose(context.Background(), c.ID(), a198Source)
	if err != nil || !got.Closed || got.State != StateConfirmed {
		t.Fatalf("close = %+v err %v, want CONFIRMED on the owner's trading day", got, err)
	}
}

// 닫을 대상이 아닌 attempt(발주·이미 종결) — 무변·오류 없음(상태 조건 쓰기의 패배 포함).
func TestA198OnlyAnInDoubtCancelIsClosed(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	a198Owner(t, j, "a", "O-1", "2026-03-30")
	c := a198InDoubtCancel(t, j, "a", "O-1", nil)
	a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
	if err := c.ResolveFailed(ctx, ReasonResolvedAbsent, "someone else closed it first"); err != nil {
		t.Fatal(err)
	}
	got, err := j.CloseInDoubtCancelOnTargetClose(ctx, c.ID(), a198Source)
	if err != nil || got.Closed {
		t.Fatalf("close of a settled cancel = %+v err %v", got, err)
	}
	if rec := a198State(t, j, c.ID()); rec.State != StateFailedConfirmed {
		t.Fatalf("the settled cancel moved to %s", rec.State)
	}
	place := a198State(t, j, "owner-att-a")
	if got, err := j.CloseInDoubtCancelOnTargetClose(ctx, place.ID, a198Source); err != nil || got.Closed {
		t.Fatalf("close of a PLACE = %+v err %v", got, err)
	}
	if _, err := j.CloseInDoubtCancelOnTargetClose(ctx, "no-such-attempt", a198Source); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("unknown attempt: err %v, want ErrAttemptNotFound", err)
	}
}

// 쓰기 동치(Manager Q1) — 같은 모양의 취소를 새 함수와 기존 API 로 닫은 행이 시각 칸을 빼고 같음.
func TestA198TheCloseWritesWhatTheResolveAPIWrites(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		state, filled string
		want          AttemptState
	}{
		{"CANCELLED", "0", StateConfirmed},
		{"REJECTED", "0", StateFailedConfirmed},
	} {
		t.Run(tc.state, func(t *testing.T) {
			j := openTestJournal(t)
			a198Owner(t, j, "a", "O-1", "2026-03-30")
			viaClose := a198InDoubtCancel(t, j, "a", "O-1", nil)
			viaAPI := a198InDoubtCancel(t, j, "b", "O-1", nil)
			a198Fill(t, j, "O-1", "2026-03-30", "BUY", tc.state, true, tc.filled)
			if got, err := j.CloseInDoubtCancelOnTargetClose(ctx, viaClose.ID(), a198Source); err != nil || !got.Closed {
				t.Fatalf("close = %+v err %v", got, err)
			}
			closed := a198State(t, j, viaClose.ID())
			var err error
			if tc.want == StateConfirmed {
				err = viaAPI.ResolveConfirmed(ctx, closed.BrokerOrderID, closed.ReasonCode, closed.Detail)
			} else {
				err = viaAPI.ResolveFailed(ctx, closed.ReasonCode, closed.Detail)
			}
			if err != nil {
				t.Fatal(err)
			}
			if a, b := a198Row(t, j, viaClose.ID()), a198Row(t, j, viaAPI.ID()); a != b {
				t.Fatalf("attempt rows differ:\nclose %s\napi   %s", a, b)
			}
			if a, b := a198Transitions(t, j, viaClose.ID()), a198Transitions(t, j, viaAPI.ID()); a != b {
				t.Fatalf("transition rows differ:\nclose %s\napi   %s", a, b)
			}
		})
	}
}

// a198Row 는 attempt 행에서 id·시각을 뺀 칸들(settled_at 은 있음/없음만).
func a198Row(t *testing.T, j *Journal, id string) string {
	t.Helper()
	var state, order, reason, detail string
	var settled sql.NullString
	if err := j.db.QueryRow(`SELECT state, broker_order_id, reason_code, detail, settled_at FROM mutation_attempts WHERE id = ?`, id).
		Scan(&state, &order, &reason, &detail, &settled); err != nil {
		t.Fatal(err)
	}
	return state + "|" + order + "|" + reason + "|" + detail + "|settled=" + boolText(settled.Valid && settled.String != "")
}

func a198Transitions(t *testing.T, j *Journal, id string) string {
	t.Helper()
	rows, err := j.db.Query(`SELECT from_state, to_state, coalesce(reason_code,''), coalesce(detail,'') FROM attempt_transitions WHERE attempt_id = ? ORDER BY rowid`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := ""
	for rows.Next() {
		var from, to, reason, detail string
		if err := rows.Scan(&from, &to, &reason, &detail); err != nil {
			t.Fatal(err)
		}
		out += from + ">" + to + "|" + reason + "|" + detail + ";"
	}
	return out
}

func boolText(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// 한 트랜잭션 — 전이 행 쓰기가 실패하면 상태도 그대로(전체 rollback).
func TestA198AFailedTransitionRowRollsTheCloseBack(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	a198Owner(t, j, "a", "O-1", "2026-03-30")
	c := a198InDoubtCancel(t, j, "a", "O-1", nil)
	a198Fill(t, j, "O-1", "2026-03-30", "BUY", "CANCELLED", true, "0")
	if _, err := j.db.Exec(`CREATE TRIGGER a198_refuse_transition BEFORE INSERT ON attempt_transitions
		WHEN NEW.to_state = 'CONFIRMED' BEGIN SELECT RAISE(ABORT, 'a198 fixture: refuse the transition row'); END`); err != nil {
		t.Fatal(err)
	}
	if got, err := j.CloseInDoubtCancelOnTargetClose(ctx, c.ID(), a198Source); err == nil || got.Closed {
		t.Fatalf("close = %+v err %v, want the write error", got, err)
	}
	if rec := a198State(t, j, c.ID()); rec.State != StateInDoubt {
		t.Fatalf("the cancel moved to %s despite the failed transition row", rec.State)
	}
}
