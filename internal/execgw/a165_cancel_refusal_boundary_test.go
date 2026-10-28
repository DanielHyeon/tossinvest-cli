package execgw_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

type a165AdvancingClock struct{ *clock.Fake }

func (c a165AdvancingClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.Advance(d)
	return ctx.Err()
}

type a165BoundaryBroker struct {
	calls          int
	firstDelay     time.Duration
	retryAfter     string
	secondDeadline bool
}

func (b *a165BoundaryBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	return domain.MutationResult{}, fmt.Errorf("not used")
}

func (b *a165BoundaryBroker) CancelPendingOrder(ctx context.Context, _ orderintent.CancelIntent) (domain.MutationResult, error) {
	b.calls++
	if b.calls == 1 {
		if b.firstDelay > 0 {
			time.Sleep(b.firstDelay)
		}
		return domain.MutationResult{}, &official.APIError{Code: 409,
			Body: `{"code":"already-processing","data":{"retryAfterSeconds":` + b.retryAfter + `}}`}
	}
	_, b.secondDeadline = ctx.Deadline()
	return domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "C-a165"}, nil
}

func (b *a165BoundaryBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	return domain.MutationResult{}, fmt.Errorf("not used")
}

func (b *a165BoundaryBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func a165BoundaryGateway(t *testing.T, j *journal.Journal, clk clock.Clock, policy *execgw.CancelRefusalPolicy, broker *a165BoundaryBroker) *execgw.Gateway {
	t.Helper()
	gw, err := execgw.New(execgw.Options{
		Journal: j,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock: clk, AccountRef: "acct-7", Source: "a165-boundary-test", CancelRefusal: policy,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw
}

func a165BoundaryCancel(t *testing.T, gw *execgw.Gateway, j *journal.Journal, clk clock.Clock) (execgw.Outcome, error) {
	t.Helper()
	return gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	})
}

func TestA165ExpiredRetryWindowDoesNotCancelSettlement(t *testing.T) {
	clk := a165AdvancingClock{Fake: clock.NewFake(fixedNow)}
	j := openJournal(t, clk)
	broker := &a165BoundaryBroker{firstDelay: 25 * time.Millisecond, retryAfter: "0.001"}
	gw := a165BoundaryGateway(t, j, clk, &execgw.CancelRefusalPolicy{
		Retries: 2, MaxWait: 5 * time.Second, Window: 10 * time.Millisecond,
	}, broker)

	out, err := a165BoundaryCancel(t, gw, j, clk)
	if err == nil || out.State != journal.StateInDoubt {
		t.Fatalf("outcome=%+v err=%v, want settled IN_DOUBT", out, err)
	}
	if broker.calls != 1 {
		t.Fatalf("broker calls=%d, want one before the expired retry window", broker.calls)
	}
	record, lookupErr := j.LookupAttempt(context.Background(), out.AttemptID)
	if lookupErr != nil || record.State != journal.StateInDoubt {
		t.Fatalf("settled record=%+v err=%v, want IN_DOUBT", record, lookupErr)
	}
}

func TestA165RetryInvokesTheBrokerWithTheWindowContext(t *testing.T) {
	clk := a165AdvancingClock{Fake: clock.NewFake(fixedNow)}
	j := openJournal(t, clk)
	broker := &a165BoundaryBroker{retryAfter: "0.001"}
	gw := a165BoundaryGateway(t, j, clk, &execgw.CancelRefusalPolicy{
		Retries: 2, MaxWait: 5 * time.Second, Window: time.Second,
	}, broker)

	out, err := a165BoundaryCancel(t, gw, j, clk)
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("outcome=%+v err=%v, want confirmed retry", out, err)
	}
	if broker.calls != 2 || !broker.secondDeadline {
		t.Fatalf("broker calls=%d secondDeadline=%v, want retry with window deadline", broker.calls, broker.secondDeadline)
	}
}
