package journal

// a094 §4 (R3) — 종결된 attempt 가 그것이 무장한 발의를 푼다. 판정 하나(ReleaseUnacceptedExitProposal)와 청소의 해제
// (ReleaseClearedExitProposal)가 같은 분류기(exitIntentAttemptsQ)에 선다는 것을 원장 끝에서 잰다.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// a094ArmedPosition 은 엔트리가 체결돼 열린 포지션에 intent 를 단 발의를 하나 무장함.
func a094ArmedPosition(t *testing.T, j *Journal, intentID string) string {
	t.Helper()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	if err := j.RecordExitJudgement(context.Background(), ExitJudgement{
		PositionID: p.ID, ObservedPrice: "68000", HighWater: "70000",
		Baseline: "68000", RatchetLevel: RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: "BASELINE", IntentID: intentID},
	}); err != nil {
		t.Fatalf("arming: %v", err)
	}
	return p.ID
}

// a094SellAttempt 는 intent 의 매도 PLACE attempt 하나를 원하는 상태까지 몰아감(원장 전이 API 만 씀).
func a094SellAttempt(t *testing.T, j *Journal, intentID, attemptID, orderID string, to AttemptState) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "005930", Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10",
			Price: "68000", Currency: "KRW", Source: "engine/exit", Fingerprint: "fp-" + intentID,
		},
		Kind: KindPlace, AttemptID: attemptID,
	})
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	step := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("driving %s to %s: %v", attemptID, to, err)
		}
	}
	switch to {
	case StateRecorded:
		return
	case StateNotDispatched:
		step(attempt.Settle(ctx, StateNotDispatched, ReasonDispatchNotSent, "never left"))
		return
	}
	step(attempt.MarkDispatchStarted(ctx))
	switch to {
	case StateDispatchStarted:
	case StateFailedConfirmed:
		step(attempt.Settle(ctx, StateFailedConfirmed, "opposite_pending_order_exists", "refused"))
	case StateAcked:
		step(attempt.MarkAcked(ctx, orderID))
	case StateConfirmed:
		step(attempt.MarkAcked(ctx, orderID))
		step(attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"))
	case StateInDoubt:
		step(attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "409"))
	case StateUnresolvedInDoubt:
		step(attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "409"))
		step(attempt.ResolveUnresolved(ctx, "in_doubt_unresolved", "parked"))
	default:
		t.Fatalf("a094SellAttempt: unsupported state %s", to)
	}
}

// 4.1 · 4.2 · 4.3 · 3.N1b — 상태별 판정과 해제(판정 하나의 표).
func TestA094TheReleaseJudgementFollowsTheAttemptState(t *testing.T) {
	for _, tc := range []struct {
		state   AttemptState
		verdict ExitIntentVerdict
		release bool
	}{
		{StateFailedConfirmed, ExitIntentUnaccepted, true},
		{StateNotDispatched, ExitIntentUnaccepted, true},
		{StateConfirmed, ExitIntentAwaitingClose, false},  // 4.2 — 접수된 주문의 발의는 유지
		{StateUnresolvedInDoubt, ExitIntentParked, false}, // 4.3 — park 는 해제하지 않음
		{StateRecorded, ExitIntentLive, false},
		{StateDispatchStarted, ExitIntentLive, false},
		{StateAcked, ExitIntentLive, false},
		{StateInDoubt, ExitIntentLive, false},
	} {
		t.Run(string(tc.state), func(t *testing.T) {
			j := exitFixture(t)
			ctx := context.Background()
			positionID := a094ArmedPosition(t, j, "exit-1")
			a094SellAttempt(t, j, "exit-1", "a-exit-1", "o-exit-1", tc.state)

			verdict, released, err := j.ReleaseUnacceptedExitProposal(ctx, positionID, "exit-1", ProposalRefused)
			if err != nil {
				t.Fatalf("ReleaseUnacceptedExitProposal: %v", err)
			}
			if verdict != tc.verdict || released != tc.release {
				t.Fatalf("verdict/released = %s/%v, want %s/%v", verdict, released, tc.verdict, tc.release)
			}
			if got := exitStateOf(t, j, positionID).Pending(); got == tc.release {
				t.Errorf("armed after the call = %v, want %v", got, !tc.release)
			}
		})
	}
}

// 4.3b 의 원장 절반 — attempt 가 하나도 없으면(무장 뒤 Prepare 전 충돌) 해제함.
func TestA094AProposalWithNoAttemptIsReleased(t *testing.T) {
	j := exitFixture(t)
	positionID := a094ArmedPosition(t, j, "exit-none")
	verdict, released, err := j.ReleaseUnacceptedExitProposal(context.Background(), positionID, "exit-none", ProposalRefused)
	if err != nil || verdict != ExitIntentNoAttempts || !released {
		t.Fatalf("verdict/released/err = %s/%v/%v, want NO_ATTEMPTS/true/nil", verdict, released, err)
	}
}

// 4.3a — 해제는 무장된 발의의 intent 가 기대 intent 와 같을 때만.
func TestA094ALateReleaseDoesNotClearAnotherProposal(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	positionID := a094ArmedPosition(t, j, "exit-new")
	a094SellAttempt(t, j, "exit-old", "a-old", "o-old", StateFailedConfirmed)

	_, released, err := j.ReleaseUnacceptedExitProposal(ctx, positionID, "exit-old", ProposalRefused)
	if err != nil || released {
		t.Fatalf("released/err = %v/%v, want false/nil — the old intent must not clear the new proposal", released, err)
	}
	if st := exitStateOf(t, j, positionID); !st.Pending() || st.PendingIntentID != "exit-new" {
		t.Fatalf("pending = %v/%s, want the new proposal untouched", st.Pending(), st.PendingIntentID)
	}
	if err := j.ResolveExitProposal(ctx, positionID, "exit-old", ProposalCancelled); err != nil {
		t.Fatalf("ResolveExitProposal: %v", err)
	}
	if !exitStateOf(t, j, positionID).Pending() {
		t.Fatal("ResolveExitProposal with another intent cleared the armed proposal")
	}
	if err := j.ResolveExitProposal(ctx, positionID, "", ProposalCancelled); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an empty expected intent = %v, want ErrInvalidRequest", err)
	}
}

// 4.5 · 4.6 — 해제는 손절 가격을 바꾸지 않고, 두 번째 호출은 아무것도 바꾸지 않음(멱등).
func TestA094ReleaseKeepsThePricesAndIsIdempotent(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	positionID := a094ArmedPosition(t, j, "exit-1")
	a094SellAttempt(t, j, "exit-1", "a-1", "", StateFailedConfirmed)
	before := exitStateOf(t, j, positionID)

	if _, released, err := j.ReleaseUnacceptedExitProposal(ctx, positionID, "exit-1", ProposalRefused); err != nil || !released {
		t.Fatalf("first release = %v/%v", released, err)
	}
	after := exitStateOf(t, j, positionID)
	if after.EntryPrice != before.EntryPrice || after.InitialStop != before.InitialStop || after.Baseline != before.Baseline {
		t.Errorf("prices moved: entry %s→%s, stop %s→%s, baseline %s→%s", before.EntryPrice, after.EntryPrice,
			before.InitialStop, after.InitialStop, before.Baseline, after.Baseline)
	}
	events := a094CountEvents(t, j, positionID, ExitEventProposalRefused)
	if _, released, err := j.ReleaseUnacceptedExitProposal(ctx, positionID, "exit-1", ProposalRefused); err != nil || released {
		t.Fatalf("second release = %v/%v, want false/nil", released, err)
	}
	if got := a094CountEvents(t, j, positionID, ExitEventProposalRefused); got != events {
		t.Errorf("refusal events = %d after the repeat, want %d", got, events)
	}
}

// 4.5a — 음수 pending_level(080220 의 모양)은 rung 되돌림 없이 해제됨(RungIndex 거부).
func TestA094ANegativeLadderLevelReleasesWithoutARungRollback(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-l", attemptID: "a-l", orderID: "o-l", decisionID: "d-l"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	p := currentPosition(t, j, o)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{PositionID: p.ID, PolicyKind: ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000"}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "67000", HighWater: "70000", Baseline: "68000",
		RatchetLevel: RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionLadderStop), Level: "-1", IntentID: "exit-l"},
	}); err != nil {
		t.Fatalf("arming: %v", err)
	}
	rungBefore := exitStateOf(t, j, p.ID).ActiveRung
	a094SellAttempt(t, j, "exit-l", "a-exit-l", "", StateFailedConfirmed)
	if _, released, err := j.ReleaseUnacceptedExitProposal(ctx, p.ID, "exit-l", ProposalRefused); err != nil || !released {
		t.Fatalf("release = %v/%v", released, err)
	}
	if got := exitStateOf(t, j, p.ID); got.Pending() || got.ActiveRung != rungBefore {
		t.Errorf("pending %v rung %d, want released with the rung %d untouched", got.Pending(), got.ActiveRung, rungBefore)
	}
}

// 3.R7 · 3.R7a 의 원장 절반 — 접수 확정 매도는 그 주문 번호의 종결 체결 기록이 올 때까지 청소도 풀지 않음.
func TestA094AConfirmedSellIsReleasedOnlyOnItsClosingRecord(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	positionID := a094ArmedPosition(t, j, "exit-tp")
	a094SellAttempt(t, j, "exit-tp", "a-tp", "o-tp", StateConfirmed)

	for name, release := range map[string]func() (ExitIntentVerdict, bool, error){
		"unaccepted": func() (ExitIntentVerdict, bool, error) {
			return j.ReleaseUnacceptedExitProposal(ctx, positionID, "exit-tp", ProposalCancelled)
		},
		"cleared": func() (ExitIntentVerdict, bool, error) {
			return j.ReleaseClearedExitProposal(ctx, positionID, "exit-tp")
		},
	} {
		verdict, released, err := release()
		if err != nil || released || verdict != ExitIntentAwaitingClose {
			t.Fatalf("%s before the closing record = %s/%v/%v, want AWAITING_CLOSE/false/nil", name, verdict, released, err)
		}
	}
	facts, err := j.ExitIntentAttempts(ctx, "exit-tp")
	if err != nil || len(facts.Awaiting) != 1 || facts.Awaiting[0].OrderID != "o-tp" {
		t.Fatalf("awaiting = %+v/%v, want order o-tp", facts.Awaiting, err)
	}

	// 취소 종결 스냅숏 — 체결 감지가 쓰는 증거 그대로(체결 0 · 종결). 발의 자신의 주문이 종결되면 체결 적용 훅이 같은
	// 트랜잭션에서 발의를 PROPOSAL_FILLED 로 끝냄 — 청소의 해제는 그 뒤에 찾을 것이 없음(멱등).
	a094CloseOrder(t, j, "o-tp")
	if exitStateOf(t, j, positionID).Pending() {
		t.Fatal("the closing record of the proposal's own order did not end the proposal")
	}
	if verdict, released, err := j.ReleaseClearedExitProposal(ctx, positionID, "exit-tp"); err != nil || released || verdict != ExitIntentOrdersClosed {
		t.Fatalf("cleared after the closing record = %s/%v/%v, want ORDERS_CLOSED/false (nothing left to release)", verdict, released, err)
	}
}

// 청소 해제의 ORDERS_CLOSED 갈래 — 종결 기록이 발의 무장보다 먼저 온 경우(적용 훅이 끝낼 발의가 없던 때)에도 청소가 풂.
// 판정 하나(unaccepted)는 접수 확정 주문을 비수용으로 읽지 않으므로 풀지 않음.
func TestA094TheClearReleasesAProposalWhoseOrdersClosedFirst(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	a094SellAttempt(t, j, "exit-tp", "a-tp", "o-tp", StateConfirmed)
	a094CloseOrder(t, j, "o-tp")
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "68000", HighWater: "70000", Baseline: "68000",
		RatchetLevel: RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: "BASELINE", IntentID: "exit-tp"},
	}); err != nil {
		t.Fatalf("arming: %v", err)
	}
	if verdict, released, err := j.ReleaseUnacceptedExitProposal(ctx, p.ID, "exit-tp", ProposalCancelled); err != nil || released || verdict != ExitIntentOrdersClosed {
		t.Fatalf("unaccepted = %s/%v/%v, want ORDERS_CLOSED/false — a confirmed order is not an unaccepted one", verdict, released, err)
	}
	if verdict, released, err := j.ReleaseClearedExitProposal(ctx, p.ID, "exit-tp"); err != nil || !released || verdict != ExitIntentOrdersClosed {
		t.Fatalf("cleared = %s/%v/%v, want ORDERS_CLOSED/true", verdict, released, err)
	}
}

// a094CloseOrder 는 체결 0 의 취소 종결 스냅숏을 기록함(체결 감지가 쓰는 모양).
func a094CloseOrder(t *testing.T, j *Journal, orderID string) {
	t.Helper()
	sell := order{orderID: orderID, side: "SELL", quantity: "10"}.withDefaults()
	closing := fillOf(sell, "0", "0")
	closing.State, closing.Terminal = "CLOSED_CANCELED", true
	if _, err := j.RecordFill(context.Background(), closing); err != nil {
		t.Fatalf("RecordFill(closing %s): %v", orderID, err)
	}
}

// ConfirmedCancelOf 는 계좌 · 시장 · 종목으로 한정한 엔진 취소 중 첫 접수 확정을 돌려줌.
func TestA094TheConfirmedCancelOfAnOrderIsScoped(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	for _, c := range []struct {
		id, account, symbol string
		confirm             bool
	}{
		{"c-other-account", "acct-2", "005930", true},
		{"c-other-symbol", "acct-1", "000660", true},
		{"c-in-doubt", "acct-1", "005930", false},
		{"c-ok", "acct-1", "005930", true},
	} {
		attempt, err := j.Prepare(ctx, PrepareRequest{
			Intent: Intent{ID: "i-" + c.id, Market: "kr", TradingDay: "2026-03-30", AccountRef: c.account,
				Symbol: c.symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10",
				Price: "68000", Currency: "KRW", Source: "engine/exit", Fingerprint: "fp-" + c.id},
			Kind: KindCancel, AttemptID: c.id, TargetOrderID: "o-target",
		})
		if err != nil {
			t.Fatalf("Prepare(%s): %v", c.id, err)
		}
		if err := attempt.MarkDispatchStarted(ctx); err != nil {
			t.Fatal(err)
		}
		if c.confirm {
			if err := attempt.MarkAcked(ctx, "o-target"); err != nil {
				t.Fatal(err)
			}
			if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "cancelled"); err != nil {
				t.Fatal(err)
			}
		} else if err := attempt.MarkInDoubt(ctx, "dispatch_outcome_unknown", "x"); err != nil {
			t.Fatal(err)
		}
	}
	rec, found, err := j.ConfirmedCancelOf(ctx, "acct-1", "KR", "005930", "o-target")
	if err != nil || !found || rec.ID != "c-ok" {
		t.Fatalf("ConfirmedCancelOf = %s/%v/%v, want c-ok", rec.ID, found, err)
	}
	if _, found, err := j.ConfirmedCancelOf(ctx, "acct-1", "kr", "005930", "o-none"); err != nil || found {
		t.Fatalf("an order with no cancel = %v/%v, want not found", found, err)
	}
}

func a094CountEvents(t *testing.T, j *Journal, positionID, action string) int {
	t.Helper()
	var n int
	if err := j.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM exit_events WHERE position_id = ? AND action = ?`, positionID, action).Scan(&n); err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("counting events: %v", err)
	}
	return n
}

// 3.R7a 의 원장 절반 — 소유가 모호해 미체결 목록에서 빠진 접수 확정 주문도 발의 intent 의 주문 번호로는 「종결 증거 없음」임.
func TestA094AnAmbiguouslyOwnedOrderIsStillAwaitingClose(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	positionID := a094ArmedPosition(t, j, "exit-tp")
	a094SellAttempt(t, j, "exit-tp", "a-tp", "o-dup", StateConfirmed)
	a094SellAttempt(t, j, "other-intent", "a-other", "o-dup", StateConfirmed)

	// 대조: 같은 범위의 소유 모호는 미체결 목록에서 조용히 빠지지 않고 오류다(guardTrackedFillIdentity) — 청소는 그 오류를
	// 돌려주고 해제 · 제출하지 않는다. 해제 판정은 그 목록에 기대지 않고 intent 의 주문 번호로 따로 본다.
	if _, err := j.LiveOrdersForSymbol(ctx, "acct-1", "kr", "005930"); err == nil {
		t.Fatal("control: the working list accepted an order confirmed for two intents — the fixture no longer names the gap")
	}
	verdict, released, err := j.ReleaseClearedExitProposal(ctx, positionID, "exit-tp")
	if err != nil || released || verdict != ExitIntentAwaitingClose {
		t.Fatalf("cleared = %s/%v/%v, want AWAITING_CLOSE/false — absence from the list is not a closing record", verdict, released, err)
	}
}
