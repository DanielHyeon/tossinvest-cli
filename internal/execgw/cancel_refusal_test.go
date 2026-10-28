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

func TestA165CancelRefusalPoliciesAreExact(t *testing.T) {
	if got, want := EngineCancelRefusal(), (CancelRefusalPolicy{Definitive: true}); got == nil || *got != want {
		t.Fatalf("engine policy: got %+v, want %+v", got, want)
	}
	if got, want := FlattenCancelRefusal(), (CancelRefusalPolicy{
		Retries: 2, MaxWait: 5 * time.Second, Window: 8 * time.Second,
	}); got == nil || *got != want {
		t.Fatalf("flatten policy: got %+v, want %+v", got, want)
	}
}

func TestA165AlreadyProcessingRefusalRequiresAnUnambiguousCode(t *testing.T) {
	cases := []struct {
		name string
		body string
		want time.Duration
		ok   bool
	}{
		{"top-level code", `{"code":"ALREADY-PROCESSING","data":{"retryAfterSeconds":3}}`, 3 * time.Second, true},
		{"nested code and retry after", `{"error":{"code":"already-processing","data":{"retryAfterSeconds":2}}}`, 2 * time.Second, true},
		{"matching codes", `{"code":"already-processing","error":{"code":"ALREADY-PROCESSING"},"data":{"retryAfterSeconds":0}}`, time.Second, true},
		{"contradictory codes", `{"code":"already-processing","error":{"code":"other"}}`, 0, false},
		{"missing code", `{"data":{"retryAfterSeconds":3}}`, 0, false},
		{"status only", `not-json`, 0, false},
		{"non-numeric retry", `{"code":"already-processing","data":{"retryAfterSeconds":"later"}}`, time.Second, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := alreadyProcessingRefusal(&official.APIError{Code: 409, Body: tc.body})
			if ok != tc.ok || got != tc.want {
				t.Fatalf("alreadyProcessingRefusal() = (%s, %v), want (%s, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestA165EngineAlreadyProcessingIsOneConfirmedRefusalWithoutDelay(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC))
	g := &Gateway{clk: clk, cancelRefusal: EngineCancelRefusal()}
	calls := 0
	out := g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		calls++
		return a165AlreadyProcessingOutcome(1)
	})(context.Background(), nil)
	if calls != 1 {
		t.Fatalf("broker calls: got %d, want 1", calls)
	}
	if out.Class != journal.DispatchRejected || out.ReasonCode != string(ReasonBrokerRejected) {
		t.Fatalf("engine outcome: %+v", out)
	}
	if !strings.Contains(out.Detail, "[a165 sends=1 branch=cancel_already_processing stop=definitive outcome=definitive]") {
		t.Fatalf("engine detail: %q", out.Detail)
	}
	if elapsed := clk.Since(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)); elapsed != 0 {
		t.Fatalf("fake clock advanced by %s", elapsed)
	}
}

func TestA165FlattenRetriesWithinCumulativeBudget(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC))
	g := &Gateway{clk: clk, cancelRefusal: FlattenCancelRefusal()}
	calls := 0
	done := make(chan journal.DispatchOutcome, 1)
	go func() {
		done <- g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
			calls++
			if calls < 3 {
				return a165AlreadyProcessingOutcome(4)
			}
			return journal.DispatchOutcome{Class: journal.DispatchAcked, BrokerOrderID: "O-a165"}
		})(context.Background(), nil)
	}()
	if !clk.WaitForSleepers(1, time.Second) {
		t.Fatal("first retry sleep was not reached")
	}
	clk.Advance(4 * time.Second)
	if !clk.WaitForSleepers(1, time.Second) {
		t.Fatal("capped retry sleep was not reached")
	}
	clk.Advance(time.Second)
	out := <-done
	if calls != 3 || out.Class != journal.DispatchAcked {
		t.Fatalf("calls=%d outcome=%+v", calls, out)
	}
	if got := clk.Since(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)); got != 5*time.Second {
		t.Fatalf("cumulative wait: got %s, want 5s", got)
	}
	if !strings.Contains(out.Detail, "[a165 sends=3 branch=cancel_already_processing stop=ack outcome=retried]") {
		t.Fatalf("retry detail: %q", out.Detail)
	}
}

func TestA165WindowStartsBeforeFirstDispatch(t *testing.T) {
	base := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	clk := clock.NewFake(base)
	g := &Gateway{clk: clk, cancelRefusal: FlattenCancelRefusal()}
	calls := 0
	out := g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		calls++
		clk.Advance(7 * time.Second)
		return a165AlreadyProcessingOutcome(1)
	})(context.Background(), nil)
	if calls != 1 || out.Class != journal.DispatchAmbiguous {
		t.Fatalf("calls=%d outcome=%+v", calls, out)
	}
	if !strings.Contains(out.Detail, "[a165 sends=1 branch=cancel_already_processing stop=window outcome=today]") {
		t.Fatalf("window detail: %q", out.Detail)
	}
}

func TestA165CallerCancellationStopsRetry(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC))
	g := &Gateway{clk: clk, cancelRefusal: FlattenCancelRefusal()}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	out := g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		calls++
		cancel()
		return a165AlreadyProcessingOutcome(1)
	})(ctx, nil)
	if calls != 1 || out.Class != journal.DispatchAmbiguous {
		t.Fatalf("calls=%d outcome=%+v", calls, out)
	}
}

func TestA165FlattenRetryNotSentKeepsTheFirstAmbiguousOutcome(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC))
	g := &Gateway{clk: clk, cancelRefusal: FlattenCancelRefusal()}
	calls := 0
	done := make(chan journal.DispatchOutcome, 1)
	go func() {
		done <- g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
			calls++
			if calls == 1 {
				return a165AlreadyProcessingOutcome(1)
			}
			return journal.DispatchOutcome{Class: journal.DispatchNotSent, Detail: "pre-send boundary refused"}
		})(context.Background(), nil)
	}()
	if !clk.WaitForSleepers(1, time.Second) {
		t.Fatal("retry sleep was not reached")
	}
	clk.Advance(time.Second)
	out := <-done
	if calls != 2 || out.Class != journal.DispatchAmbiguous || !strings.Contains(out.Detail, "stop=not_sent outcome=today") {
		t.Fatalf("calls=%d outcome=%+v", calls, out)
	}
}

func TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome(t *testing.T) {
	original := a165AlreadyProcessingOutcome(1)
	dispatch := func(context.Context, *journal.Attempt) journal.DispatchOutcome { return original }
	for _, tc := range []struct {
		name string
		g    *Gateway
		plan mutationPlan
	}{
		{"nil policy cancel", &Gateway{clk: clock.NewFake(time.Now())}, mutationPlan{kind: journal.KindCancel}},
		{"flatten policy place", &Gateway{clk: clock.NewFake(time.Now()), cancelRefusal: FlattenCancelRefusal()}, mutationPlan{kind: journal.KindPlace}},
		{"flatten policy amend", &Gateway{clk: clock.NewFake(time.Now()), cancelRefusal: FlattenCancelRefusal()}, mutationPlan{kind: journal.KindAmend}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.g.withCancelRefusal(tc.plan, dispatch)(context.Background(), nil); got != original {
				t.Fatalf("outcome changed: got %+v want %+v", got, original)
			}
		})
	}
}

func a165AlreadyProcessingOutcome(retryAfter int) journal.DispatchOutcome {
	return journal.DispatchOutcome{
		Class: journal.DispatchAmbiguous,
		Err: &official.APIError{Code: 409, Body: `{"code":"already-processing","data":{"retryAfterSeconds":` +
			string(rune('0'+retryAfter)) + `}}`},
		Detail: "broker refused cancel",
	}
}
