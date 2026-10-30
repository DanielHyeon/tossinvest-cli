package journal

import (
	"context"
	"testing"
)

func TestA087HasProtectiveExitAttemptScopesAndFiltersActions(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertPosition(t, j, "pos-a087", nil)

	// 익절 intent는 같은 계좌·종목이어도 첫 보호 매도 금지 근거가 아님.
	a087ConfirmSell(t, j, "intent-take-profit", "order-take-profit", "2026-03-30")
	a087ExitEvent(t, j, "pos-a087", "70000", "LADDER_TAKE_PROFIT", "intent-take-profit")
	actions := []string{"BASELINE_BREACH_EXIT", "STOP_LOSS_LADDER"}
	got, err := j.HasProtectiveExitAttempt(ctx, "acct-1", "KR", "2026-03-30", " 005930 ", actions)
	if err != nil {
		t.Fatalf("HasProtectiveExitAttempt: %v", err)
	}
	if got {
		t.Fatal("take-profit intent prohibited the first protective lower-limit attempt")
	}

	// 시도 결과 상태·가격과 무관하게 보호 action을 가리키는 SELL intent 하나면 금지함.
	a087ConfirmSell(t, j, "intent-protective", "order-protective", "2026-03-30")
	a087ExitEvent(t, j, "pos-a087", "67900", "BASELINE_BREACH_EXIT", "intent-protective")
	got, err = j.HasProtectiveExitAttempt(ctx, "acct-1", "kr", "2026-03-30", "005930", actions)
	if err != nil || !got {
		t.Fatalf("protective attempt = %v, %v; want true, nil", got, err)
	}

	for _, scope := range []struct {
		name, account, market, day, symbol string
	}{
		{"other account", "acct-2", "kr", "2026-03-30", "005930"},
		{"other market", "acct-1", "us", "2026-03-30", "005930"},
		{"other trading day", "acct-1", "kr", "2026-03-31", "005930"},
		{"other symbol", "acct-1", "kr", "2026-03-30", "000660"},
	} {
		t.Run(scope.name, func(t *testing.T) {
			got, err := j.HasProtectiveExitAttempt(ctx, scope.account, scope.market, scope.day, scope.symbol, actions)
			if err != nil {
				t.Fatal(err)
			}
			if got {
				t.Fatalf("scope %+v found an unrelated protective attempt", scope)
			}
		})
	}
}

func TestA087HasProtectiveExitAttemptRejectsEmptyActionSet(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.HasProtectiveExitAttempt(t.Context(), "acct-1", "kr", "2026-03-30", "005930", nil); err == nil {
		t.Fatal("empty protective action set was accepted")
	}
}

func TestA087HasProtectiveExitAttemptIgnoresOutcomeState(t *testing.T) {
	actions := []string{"BASELINE_BREACH_EXIT", "STOP_LOSS_LADDER"}
	for _, state := range []AttemptState{StateConfirmed, StateFailedConfirmed, StateInDoubt} {
		t.Run(string(state), func(t *testing.T) {
			j := openTestJournal(t)
			insertPosition(t, j, "pos-a087-state", nil)
			intentID := "intent-a087-" + string(state)
			a087ProtectiveSellState(t, j, intentID, state)
			a087ExitEvent(t, j, "pos-a087-state", "67900", "BASELINE_BREACH_EXIT", intentID)
			got, err := j.HasProtectiveExitAttempt(t.Context(), "acct-1", "kr", "2026-03-30", "005930", actions)
			if err != nil || !got {
				t.Fatalf("state %s attempt = %v, %v; want true, nil", state, got, err)
			}
		})
	}
}

// a087ProtectiveSellState 는 결과 상태와 무관한 append-only intent 존재를 시험용으로 만듦.
func a087ProtectiveSellState(t *testing.T, j *Journal, intentID string, state AttemptState) {
	t.Helper()
	ctx := t.Context()
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-1",
			Symbol: "005930", Side: "SELL", OrderType: "LIMIT", Quantity: "10", Price: "67900",
			Currency: "KRW", Source: "engine/a087-test", Fingerprint: "fp-" + intentID,
		},
		Kind: KindPlace, AttemptID: "attempt-" + intentID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	switch state {
	case StateConfirmed:
		if err := attempt.MarkInDoubt(ctx, "timeout", "a087 test"); err != nil {
			t.Fatal(err)
		}
		if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "a087 test"); err != nil {
			t.Fatal(err)
		}
	case StateFailedConfirmed:
		if err := attempt.Settle(ctx, StateFailedConfirmed, "broker_rejected", "a087 test"); err != nil {
			t.Fatal(err)
		}
	case StateInDoubt:
		if err := attempt.MarkInDoubt(ctx, "timeout", "a087 test"); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported state %s", state)
	}
}
