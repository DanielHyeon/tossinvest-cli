package engine_test

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

type a166PriorityPrices struct {
	underlying engine.PriceReader
	got        accountbudget.Priority
}

func (p *a166PriorityPrices) Prices(ctx context.Context, symbols []string) ([]domain.Quote, error) {
	p.got = accountbudget.PriorityFromContext(ctx)
	return p.underlying.Prices(ctx, symbols)
}
func TestA166ExitObservePricesCritical(t *testing.T) {
	var capture *a166PriorityPrices
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		capture = &a166PriorityPrices{underlying: opts.Prices}
		opts.Prices = capture
	})
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	if result := h.observe(); result.Err != nil {
		t.Fatal(result.Err)
	}
	if capture.got != accountbudget.Critical {
		t.Fatalf("observation priority = %v", capture.got)
	}
}

type a166SubmitSpy struct {
	underlying      engine.ExitSubmitter
	places, cancels []accountbudget.Priority
}

func (s *a166SubmitSpy) Place(ctx context.Context, r execgw.PlaceRequest) (execgw.Outcome, error) {
	s.places = append(s.places, accountbudget.PriorityFromContext(ctx))
	return s.underlying.Place(ctx, r)
}
func (s *a166SubmitSpy) Cancel(ctx context.Context, r execgw.CancelRequest) (execgw.Outcome, error) {
	s.cancels = append(s.cancels, accountbudget.PriorityFromContext(ctx))
	return s.underlying.Cancel(ctx, r)
}
func (s *a166SubmitSpy) UnsettledOnSymbol(ctx context.Context, m, symbol string) ([]journal.AttemptRecord, error) {
	return s.underlying.UnsettledOnSymbol(ctx, m, symbol)
}
func TestA166ExitProposalPriority(t *testing.T) {
	for _, tc := range []struct {
		name  string
		price float64
		want  accountbudget.Priority
	}{{"stop", 67900, accountbudget.Critical}, {"take-profit", 72000, accountbudget.Normal}} {
		t.Run(tc.name, func(t *testing.T) {
			var spy *a166SubmitSpy
			h := newExitHarness(t, func(o *engine.ExitObserverOptions) { spy = &a166SubmitSpy{underlying: o.Submit}; o.Submit = spy })
			h.entry("005930", "10", "70000", "68000", "70000")
			h.workingEntry("005930", "5", "69500")
			h.quote("005930", tc.price)
			if result := h.observe(); result.Err != nil {
				t.Fatal(result.Err)
			}
			if (tc.name == "stop" && (len(spy.cancels) != 1 || spy.cancels[0] != tc.want)) || (tc.name == "take-profit" && len(spy.cancels) != 0) || len(spy.places) != 1 || spy.places[0] != tc.want {
				t.Fatalf("cancel=%v place=%v", spy.cancels, spy.places)
			}
		})
	}
}

func TestA166FullTakeProfitCancelIsNormal(t *testing.T) {
	var spy *a166SubmitSpy
	ladder := exitpolicy.DefaultLadderPolicy()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		spy = &a166SubmitSpy{underlying: o.Submit}
		o.Submit = spy
		o.Ladder = &ladder
	})
	p := h.entry("005930", "10", "10000", "9800", "10000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: "default_v1", EntryPrice: "10000", InitialStop: "9800"}); err != nil {
		t.Fatal(err)
	}
	h.workingEntry("005930", "5", "9900")
	h.quote("005930", 10600)
	result := h.observer.ObserveOnce(accountbudget.WithPriority(context.Background(), accountbudget.Critical))
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(spy.cancels) != 1 || spy.cancels[0] != accountbudget.Normal || len(spy.places) != 1 || spy.places[0] != accountbudget.Normal {
		t.Fatalf("cancel=%v place=%v result=%+v alerts=%+v", spy.cancels, spy.places, result, h.alerts.events)
	}
}
