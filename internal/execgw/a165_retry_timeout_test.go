package execgw

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA165RetryDeadlineKeepsFirstAmbiguousAsCancelled(t *testing.T) {
	base := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	clk := clock.NewFake(base)
	g := &Gateway{clk: clk, cancelRefusal: &CancelRefusalPolicy{
		Retries: 1, MaxWait: time.Second, Window: 25 * time.Millisecond,
	}}
	first := journal.DispatchOutcome{
		Class:  journal.DispatchAmbiguous,
		Err:    &official.APIError{Code: 409, Body: `{"code":"already-processing","data":{"retryAfterSeconds":0.001}}`},
		Detail: "broker refused cancel",
	}
	retryStarted := make(chan struct{})
	done := make(chan journal.DispatchOutcome, 1)
	calls := 0
	go func() {
		done <- g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(ctx context.Context, _ *journal.Attempt) journal.DispatchOutcome {
			calls++
			if calls == 1 {
				return first
			}
			close(retryStarted)
			<-ctx.Done()
			return journal.DispatchOutcome{Class: journal.DispatchAmbiguous, Detail: "retry transport timeout", Err: ctx.Err()}
		})(context.Background(), nil)
	}()
	if !clk.WaitForSleepers(1, time.Second) {
		t.Fatal("retry sleep was not reached")
	}
	clk.Advance(time.Millisecond)
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("retry dispatch was not reached")
	}

	select {
	case out := <-done:
		if calls != 2 {
			t.Fatalf("broker calls=%d, want retry dispatch", calls)
		}
		if out.Class != journal.DispatchAmbiguous || !strings.Contains(out.Detail, "broker refused cancel") ||
			!strings.Contains(out.Detail, "stop=cancelled outcome=today") {
			t.Fatalf("outcome=%+v, want first 409 AMBIGUOUS preserved as cancelled", out)
		}
	case <-time.After(time.Second):
		t.Fatal("retry deadline did not finish")
	}
}
