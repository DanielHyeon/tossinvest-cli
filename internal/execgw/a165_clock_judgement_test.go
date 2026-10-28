package execgw

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA165RealTimeoutKeepsFirstAmbiguousAsCancelled(t *testing.T) {
	base := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	clk := clock.NewFake(base)
	g := &Gateway{clk: clk, cancelRefusal: FlattenCancelRefusal()}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	done := make(chan journal.DispatchOutcome, 1)
	go func() {
		done <- g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
			return a165AlreadyProcessingOutcome(1)
		})(ctx, nil)
	}()
	if !clk.WaitForSleepers(1, time.Second) {
		t.Fatal("retry sleep was not reached")
	}

	select {
	case out := <-done:
		if out.Class != journal.DispatchAmbiguous {
			t.Fatalf("class=%s, want first ambiguous outcome", out.Class)
		}
		if !strings.Contains(out.Detail, "broker refused cancel") ||
			!strings.Contains(out.Detail, "stop=cancelled outcome=today") {
			t.Fatalf("detail=%q, want first 409 ambiguous cancellation marker", out.Detail)
		}
	case <-time.After(time.Second):
		t.Fatal("real retry timeout did not cancel a fake-clock sleeper")
	}
	if got := clk.Since(base); got != 0 {
		t.Fatalf("fake clock advanced by %s, want real upper bound only", got)
	}
}

func TestA165CancelRefusalUsesOnlyInjectedClockForJudgement(t *testing.T) {
	body, err := os.ReadFile("cancel_refusal.go")
	if err != nil {
		t.Fatalf("read cancel_refusal.go: %v", err)
	}
	source := string(body)
	forbidden := []string{"time.Now(", "time.Since(", "context.WithDeadline("}
	for _, call := range forbidden {
		if strings.Contains(source, call) {
			t.Fatalf("cancel_refusal.go contains %q; D5 judgement must use g.clk and a timeout cap", call)
		}
	}
	if got := strings.Count(source, "context.WithTimeout(dctx, policy.Window)"); got != 1 {
		t.Fatalf("WithTimeout(dctx, policy.Window) count=%d, want 1", got)
	}
}
