package engine_test

// a198 보강 2(D10) — 엔진이 낸 취소가 IN_DOUBT 로 남은 뒤 대상 주문의 종결 기록이 원장에 들어오면, 세션 해소 실행자의 스윕이
// 그 취소를 닫고 그 뒤 관측 주기에 손절이 나가는지를 **실제 주문 경로**(엔진 조립의 게이트웨이·원장·해소 실행자)로 잼.
// 리뷰 r2 탐침 모양: 취소 5xx → IN_DOUBT → 종결 기록 → 스윕 → 손절. 픽스처는 합성 값만.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// sweep 은 운영 조립(Context.InDoubtSettler)의 실행자를 한 스윕만 돌림 — 첫 스윕 뒤 쉬는 자리에 닿으면 멈춤.
func (f *a198Fixture) sweep() {
	f.t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	exec := f.s.engine.InDoubtSettler(f.s.clk)
	before := f.s.clk.Sleepers()
	done := make(chan error, 1)
	go func() { done <- exec.Run(ctx) }()
	if !f.s.clk.WaitForSleepers(before+1, 10*time.Second) {
		cancel()
		f.t.Fatalf("the in-doubt settler never finished its first sweep (stopped with %v)", <-done)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		f.t.Fatalf("settler stopped with %v, want context.Canceled", err)
	}
}

// inDoubtBuyCancel 은 W-BUY-1 위에서 손절을 발의하고 그 취소를 5xx 로 IN_DOUBT 에 남김(브로커 응답은 다시 정상으로).
func (f *a198Fixture) inDoubtBuyCancel() {
	f.t.Helper()
	f.openWithWorkingBuy()
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.breach()
	f.broker.refuseCancels(0)
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateInDoubt)+"|"+journal.ReasonDispatchAmbiguous {
		f.t.Fatalf("cancel attempts = %v, want one IN_DOUBT from the 5xx", got)
	}
	if len(f.sells()) != 0 {
		f.t.Fatalf("broker sells = %v over an IN_DOUBT cancel", f.sells())
	}
}

// fill 은 체결 감지 입구로 그 주문의 체결 기록을 남김(종결 여부·fail-closed 를 시험이 정함).
func (f *a198Fixture) fill(orderID, side, day, state string, terminal, failClosed bool, ordered, filled string) {
	f.t.Helper()
	if _, err := f.s.engine.Journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, AccountRef: f.s.engine.AccountRef, Symbol: "005930", Market: "kr",
		TradingDay: day, Side: side, State: state, Terminal: terminal, FailClosed: failClosed,
		Quantity: ordered, FilledQuantity: filled, AveragePrice: "69000",
		ObservedAt: f.s.clk.Now().Format(time.RFC3339),
	}); err != nil {
		f.t.Fatalf("RecordFill(%s %s): %v", orderID, state, err)
	}
}

// cancelOrderID 는 유일한 CANCEL attempt 의 broker_order_id 임.
func (f *a198Fixture) cancelOrderID() string {
	f.t.Helper()
	var id string
	if err := f.db().QueryRow(`SELECT broker_order_id FROM mutation_attempts WHERE kind = 'CANCEL'`).Scan(&id); err != nil {
		f.t.Fatalf("reading the cancel's broker order id: %v", err)
	}
	return id
}

// §4-18 — RED 의 핵심(리뷰 r2 탐침 모양, 매수).
func TestA198AnInDoubtBuyCancelClosesOnItsClosingRecordAndTheStopFollows(t *testing.T) {
	f := newA198Fixture(t)
	f.inDoubtBuyCancel()
	f.fill("W-BUY-1", "BUY", "2026-03-30", "CANCELLED", true, false, "3", "0")
	f.sweep()
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateConfirmed)+"|"+journal.ReasonResolvedFound {
		t.Fatalf("cancel attempts after the sweep = %v (all %v), want CONFIRMED|%s — the target's closing record settles the cancel",
			got, f.attemptRows(), journal.ReasonResolvedFound)
	}
	if id := f.cancelOrderID(); id != "W-BUY-1" {
		t.Fatalf("the settled cancel names order %q, want the target W-BUY-1", id)
	}
	f.next()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v after the sweep closed the cancel, want the stop — attempts %v", f.sells(), f.attemptRows())
	}
}

// §4-18 — 매도(무장된 익절 지정가) 대상: 취소 5xx → 그 매도의 무체결 취소 종결 기록 → 스윕 → 손절 10.
func TestA198AnInDoubtSellCancelClosesOnItsClosingRecordAndTheStopFollows(t *testing.T) {
	f := newA198Fixture(t)
	f.p = f.s.entry("005930", "10", "70000", "68000")
	for _, px := range []string{"70100", "71000", "72000"} {
		f.broker.quote("005930", px)
		f.s.observe()
	}
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v, want the armed partial", f.sells())
	}
	partial := f.s.lastBrokerOrder()
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.broker.quote("005930", "67000")
	f.s.observe()
	f.broker.refuseCancels(0)
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateInDoubt)+"|"+journal.ReasonDispatchAmbiguous {
		t.Fatalf("cancel attempts = %v, want one IN_DOUBT on the armed sell", got)
	}
	f.fill(partial, "SELL", "2026-03-30", "CANCELLED", true, false, "4", "0")
	f.s.observe()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v before the sweep — the IN_DOUBT cancel still holds the symbol", f.sells())
	}
	f.sweep()
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateConfirmed)+"|"+journal.ReasonResolvedFound {
		t.Fatalf("cancel attempts after the sweep = %v, want CONFIRMED|%s", got, journal.ReasonResolvedFound)
	}
	f.s.clk.Advance(6 * time.Second)
	f.s.observe()
	if sells := f.sells(); len(sells) != 2 || fmt.Sprint(sells[1]["quantity"]) != "10" {
		t.Fatalf("broker sells = %v, want the stop of 10 after the sweep — attempts %v", sells, f.attemptRows())
	}
}

// §4-19 — 대상이 끝난 모양마다 그 취소의 종결 상태·사유(판정 표).
func TestA198TheClosingRecordDecidesTheSettledCancel(t *testing.T) {
	for _, tc := range []struct {
		state, filled string
		want          journal.AttemptState
		reason        string
	}{
		{"CANCELLED", "0", journal.StateConfirmed, journal.ReasonResolvedFound},
		{"CANCELLED_PARTIALLY_FILLED", "1", journal.StateConfirmed, journal.ReasonResolvedFound},
		{"FILLED", "3", journal.StateFailedConfirmed, journal.ReasonResolvedAbsent},
		{"REJECTED", "0", journal.StateFailedConfirmed, journal.ReasonResolvedAbsent},
		{"REJECTED_PARTIALLY_FILLED", "1", journal.StateFailedConfirmed, journal.ReasonResolvedAbsent},
	} {
		t.Run(tc.state, func(t *testing.T) {
			f := newA198Fixture(t)
			f.inDoubtBuyCancel()
			f.fill("W-BUY-1", "BUY", "2026-03-30", tc.state, true, false, "3", tc.filled)
			f.sweep()
			if got := f.cancelStates(); len(got) != 1 || got[0] != string(tc.want)+"|"+tc.reason {
				t.Fatalf("cancel attempts = %v, want %s|%s", got, tc.want, tc.reason)
			}
			f.next()
			if len(f.sells()) != 1 {
				t.Fatalf("broker sells = %v after the cancel settled, want the stop", f.sells())
			}
		})
	}
}

// §4-20·21 — 대상이 살아 있거나 판정할 수 없으면 무변(IN_DOUBT), 손절 보류 유지.
func TestA198AnInDoubtCancelStaysWhileItsTargetIsNotClosed(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(f *a198Fixture)
	}{
		{"no closing record", func(*a198Fixture) {}},
		{"open unfilled", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "OPEN_UNFILLED", false, false, "3", "0")
		}},
		{"open partially filled", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "OPEN_PARTIALLY_FILLED", false, false, "3", "1")
		}},
		{"cancel pending", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "CANCEL_PENDING", false, false, "3", "0")
		}},
		{"unknown (already-modified shape)", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "UNKNOWN_BROKER_STATE", false, true, "3", "0")
		}},
		{"replaced with a successor", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "REPLACED", true, false, "3", "1")
		}},
		{"terminal row that later failed closed", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-30", "CANCELLED_PARTIALLY_FILLED", true, false, "3", "2")
			// 누적 체결이 줄어든 관측 — 원장이 거절하고 그 행에 fail_closed 를 세움(terminal 은 그대로).
			f.fill("W-BUY-1", "BUY", "2026-03-30", "CANCELLED_PARTIALLY_FILLED", true, false, "3", "1")
		}},
		{"closing record of another trading day", func(f *a198Fixture) {
			f.fill("W-BUY-1", "BUY", "2026-03-31", "CANCELLED", true, false, "3", "0")
		}},
		// 다른 방향의 같은 번호 기록은 원장의 신원 충돌 가드가 계좌 전체를 막으므로 여기서 재지 않음 — 원장 단위 시험이 잼.
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA198Fixture(t)
			f.inDoubtBuyCancel()
			tc.setup(f)
			f.sweep()
			if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateInDoubt)+"|"+journal.ReasonDispatchAmbiguous {
				t.Fatalf("cancel attempts = %v, want the IN_DOUBT untouched", got)
			}
			f.next()
			if len(f.sells()) != 0 {
				t.Fatalf("broker sells = %v over an IN_DOUBT cancel whose target is not closed", f.sells())
			}
		})
	}
}

// §4-22 — 같은 번호의 소유 intent 가 거래일을 건너 둘이면 닫지 않음.
func TestA198AnInDoubtCancelWithTwoOwnersIsNotClosed(t *testing.T) {
	f := newA198Fixture(t)
	f.inDoubtBuyCancel()
	a198WorkingOrderOn(t, f.s, f.broker, "005930", "BUY", "W-BUY-1", "3", "69000", "2026-03-27", "dup")
	f.fill("W-BUY-1", "BUY", "2026-03-30", "CANCELLED", true, false, "3", "0")
	f.sweep()
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateInDoubt)+"|"+journal.ReasonDispatchAmbiguous {
		t.Fatalf("cancel attempts = %v, want the IN_DOUBT untouched — the order number has two owners", got)
	}
}

// §4-22 — 전일 주문: 소유 intent 의 거래일(전일)로 종결 기록을 찾음 — 취소 intent 의 거래일(오늘)을 쓰지 않음.
func TestA198AnInDoubtCancelOfAPreviousDayOrderClosesOnThatDaysRecord(t *testing.T) {
	f := newA198Fixture(t)
	f.p = f.s.entry("005930", "10", "70000", "68000")
	a198WorkingOrderOn(t, f.s, f.broker, "005930", "BUY", "W-BUY-PREV", "3", "69000", "2026-03-27", "")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.breach()
	f.broker.refuseCancels(0)
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateInDoubt)+"|"+journal.ReasonDispatchAmbiguous {
		t.Fatalf("cancel attempts = %v, want one IN_DOUBT", got)
	}
	f.fill("W-BUY-PREV", "BUY", "2026-03-27", "CANCELLED", true, false, "3", "0")
	f.sweep()
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateConfirmed)+"|"+journal.ReasonResolvedFound {
		t.Fatalf("cancel attempts = %v, want CONFIRMED — the owner's trading day is the previous day", got)
	}
}

// §4-31 (Manager Q6 논증) — D10 이 닫은 뒤에도 길 치우기가 같은 정의로 다시 판정함: (가) 매도 취소를 CONFIRMED 로 닫았는데
// 그 매도에 종결 증거가 없으면 손절 0, FAILED 로 닫혔는데 살아 있으면 다시 취소·손절 0. (나) 매수는 D8 로 치움 → 손절.
func TestA198AClosedCancelStillLeavesTheStopToTheSameClearingJudgement(t *testing.T) {
	armedSellInDoubt := func(t *testing.T) (*a198Fixture, string) {
		f := newA198Fixture(t)
		f.p = f.s.entry("005930", "10", "70000", "68000")
		for _, px := range []string{"70100", "71000", "72000"} {
			f.broker.quote("005930", px)
			f.s.observe()
		}
		partial := f.s.lastBrokerOrder()
		f.broker.refuseCancels(http.StatusInternalServerError)
		f.broker.quote("005930", "67000")
		f.s.observe()
		f.broker.refuseCancels(0)
		return f, partial
	}
	settle := func(f *a198Fixture, to journal.AttemptState, orderID string) {
		f.t.Helper()
		var id string
		if err := f.db().QueryRow(`SELECT id FROM mutation_attempts WHERE kind = 'CANCEL' AND state = 'IN_DOUBT'`).Scan(&id); err != nil {
			f.t.Fatalf("finding the IN_DOUBT cancel: %v", err)
		}
		a, err := f.s.engine.Journal.Resume(context.Background(), id)
		if err != nil {
			f.t.Fatal(err)
		}
		if to == journal.StateConfirmed {
			err = a.ResolveConfirmed(context.Background(), orderID, journal.ReasonResolvedFound, "a198 Q6 synthetic wrong close")
		} else {
			err = a.ResolveFailed(context.Background(), journal.ReasonResolvedAbsent, "a198 Q6 synthetic wrong close")
		}
		if err != nil {
			f.t.Fatalf("closing the cancel: %v", err)
		}
	}
	t.Run("sell closed CONFIRMED without closing evidence", func(t *testing.T) {
		f, partial := armedSellInDoubt(t)
		settle(f, journal.StateConfirmed, partial)
		cancels := len(f.broker.cancels())
		f.s.observe()
		f.s.observe()
		if len(f.sells()) != 1 || len(f.broker.cancels()) != cancels {
			t.Fatalf("sells %v cancels %v — a stop went out (or the cancel was re-sent) over a sell with no closing record",
				f.sells(), f.broker.cancels())
		}
	})
	t.Run("sell closed FAILED while it lives", func(t *testing.T) {
		f, _ := armedSellInDoubt(t)
		settle(f, journal.StateFailedConfirmed, "")
		cancels := len(f.broker.cancels())
		f.s.observe()
		if len(f.broker.cancels()) != cancels+1 {
			t.Fatalf("broker cancels = %v, want the living sell cancelled again", f.broker.cancels())
		}
		if len(f.sells()) != 1 {
			t.Fatalf("broker sells = %v, a stop went out over the living sell", f.sells())
		}
	})
	t.Run("buy closed CONFIRMED is cleared by D8", func(t *testing.T) {
		f := newA198Fixture(t)
		f.inDoubtBuyCancel()
		settle(f, journal.StateConfirmed, "W-BUY-1")
		f.next()
		if len(f.sells()) != 1 {
			t.Fatalf("broker sells = %v, want the stop — an accepted buy cancel clears the way (D8)", f.sells())
		}
	})
}

// a198BuyLimits 는 a198WorkingOrder 의 매수 한도와 같은 값임.
func a198BuyLimits() (string, error) {
	return execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(100), MaxNotional: execgw.Bound(1_000_000),
		MaxTotalExposure: execgw.Bound(10_000_000), MaxDailyLossAmount: execgw.Bound(100_000),
		MaxDailyLossRatio: execgw.Bound(0.01), Currency: "KRW",
	})
}

// a198WorkingOrderOn 은 a198WorkingOrder 와 같은 엔진 귀속 미체결 주문을 거래일·접미를 정해 남김(전일 주문·같은 번호의 둘째 소유).
func a198WorkingOrderOn(t *testing.T, s *e2eStack, b *a198Broker, symbol, side, orderID, quantity, price, day, suffix string) {
	t.Helper()
	ctx := context.Background()
	key := orderID + suffix
	decisionID := "d-a198-on-" + key
	limits, err := a198BuyLimits()
	if err != nil {
		t.Fatalf("EncodeLimits: %v", err)
	}
	if _, err := s.engine.Journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: s.engine.AccountRef, SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: s.engine.AccountRef, Market: "kr", Symbol: symbol, Side: side,
			Quantity: quantity, EntryPrice: price, StopPrice: "60000", TargetPrice: "90000", PolicyVersion: "a198-test",
		},
		LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: s.clk.Now(), ExpiresAt: s.clk.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("RecordDecision(%s): %v", key, err)
	}
	attempt, err := s.engine.Journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "i-a198-on-" + key, Market: "kr", TradingDay: day,
			AccountRef: s.engine.AccountRef, Symbol: symbol, Side: side,
			OrderType: "LIMIT", TimeInForce: "DAY", Quantity: quantity, Price: price,
			Currency: "KRW", Source: "engine", Fingerprint: "fp-a198-on-" + key,
		},
		Kind: journal.KindPlace, AttemptID: "a-a198-on-" + key,
		AccountRef: s.engine.AccountRef, DecisionID: decisionID, SafetyClass: journal.SafetyClassExposureRaising,
		ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		t.Fatalf("Prepare(%s): %v", key, err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		t.Fatalf("Settle: %v", err)
	}
	b.e2eBroker.mu.Lock()
	b.e2eBroker.symbols[orderID] = symbol
	b.e2eBroker.mu.Unlock()
}
