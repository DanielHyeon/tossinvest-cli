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

func TestA165AlreadyProcessingRequiresHTTP409(t *testing.T) {
	if _, ok := alreadyProcessingRefusal(&official.APIError{Code: 400, Body: `{"code":"already-processing"}`}); ok {
		t.Fatal("non-409 already-processing was accepted")
	}
}

func TestA165CancelRefusalDoesNotOverrideNonAmbiguousDispatch(t *testing.T) {
	g := &Gateway{clk: clock.NewFake(time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)), cancelRefusal: EngineCancelRefusal()}
	original := journal.DispatchOutcome{
		Class:  journal.DispatchRejected,
		Err:    &official.APIError{Code: 409, Body: `{"code":"already-processing"}`},
		Detail: "the dispatch already classified this result",
	}
	got := g.withCancelRefusal(mutationPlan{kind: journal.KindCancel}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		return original
	})(context.Background(), nil)
	if got.Class != original.Class || got.Detail != original.Detail || strings.Contains(got.Detail, cancelAlreadyProcessingMarker) {
		t.Fatalf("non-ambiguous outcome changed: got %+v want %+v", got, original)
	}
}
