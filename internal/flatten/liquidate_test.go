package flatten_test

// liquidate_test.go covers the reduce-only liquidation phase and the human
// confirmation (task 4.5).
//
// The cases that carry the requirement: the order is sized from the *sellable*
// quantity and not the holding, a symbol whose cancel is unresolved is not sold
// at all, the limit is aggressive but valid, --dry-run submits nothing, and the
// confirmation cannot be answered by anything but a person at a terminal.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/flatten"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// --- fakes ------------------------------------------------------------------

// accountFake serves holdings, sellable quantities, balances and prices, and can
// be mutated between rounds to simulate fills landing.
type accountFake struct {
	mu        sync.Mutex
	positions []domain.Position
	sellable  map[string]float64
	lower     map[string]float64
	last      map[string]float64
	priceErr  error
}

func (a *accountFake) Positions(context.Context) ([]domain.Position, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]domain.Position(nil), a.positions...), nil
}

func (a *accountFake) BuyingPower(context.Context, string) (float64, error) { return 1000, nil }

func (a *accountFake) SellableQuantity(_ context.Context, symbol string) (domain.SellableQuantity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return domain.SellableQuantity{Symbol: symbol, Quantity: a.sellable[symbol]}, nil
}

func (a *accountFake) PriceLimits(_ context.Context, symbol string) (domain.PriceLimits, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.priceErr != nil {
		return domain.PriceLimits{}, a.priceErr
	}
	lower, ok := a.lower[symbol]
	if !ok {
		return domain.PriceLimits{}, errors.New("no price band")
	}
	return domain.PriceLimits{Symbol: symbol, LowerLimit: lower}, nil
}

func (a *accountFake) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.priceErr != nil {
		return nil, a.priceErr
	}
	var out []domain.Quote
	for _, s := range symbols {
		if last, ok := a.last[s]; ok {
			out = append(out, domain.Quote{Symbol: s, Last: last})
		}
	}
	return out, nil
}

func (a *accountFake) setFlat() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.positions = nil
	a.sellable = map[string]float64{}
}

// sellBroker records the place intents the liquidation submits.
type sellBroker struct {
	*cancelBroker
	mu     sync.Mutex
	placed []orderintent.PlaceIntent
	fail   error
	after  func()
}

func newSellBroker() *sellBroker { return &sellBroker{cancelBroker: newCancelBroker()} }

func (b *sellBroker) PlacePendingOrder(_ context.Context, intent orderintent.PlaceIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	b.placed = append(b.placed, intent)
	fail := b.fail
	after := b.after
	b.mu.Unlock()

	if after != nil {
		after()
	}
	if fail != nil {
		return domain.MutationResult{}, fail
	}
	return domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "S-" + intent.Symbol}, nil
}

func (b *sellBroker) orders() []orderintent.PlaceIntent {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]orderintent.PlaceIntent(nil), b.placed...)
}

// --- harness ----------------------------------------------------------------

type liqHarness struct {
	*harness
	account *accountFake
	sell    *sellBroker
}

func newLiqHarness(t *testing.T, pages [][]json.RawMessage, account *accountFake) *liqHarness {
	t.Helper()
	h := newHarness(t, pages)

	// Rebuild the gateway over a broker that can also place, since liquidation
	// sells through the same gateway the cancels used.
	sell := &sellBroker{cancelBroker: h.broker}
	gw, err := execgw.New(execgw.Options{
		Journal:    h.journal,
		Trading:    newTradingService(sell),
		Clock:      h.clock,
		AccountRef: "acct-7",
		Source:     "flatten-test",
		Entry:      h.gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	h.gateway = gw
	return &liqHarness{harness: h, account: account, sell: sell}
}

// autoClock is the fake clock with a Sleep that advances instead of blocking.
//
// The stabilisation loop legitimately sleeps between snapshots, and a fake clock
// that only moves when a test advances it would deadlock inside a call the test
// is already inside. Advancing on sleep keeps every timestamp deterministic —
// including the ones the gateway checks a Guardian decision's expiry against,
// which is why the test cannot simply switch to the system clock here.
type autoClock struct{ *clock.Fake }

func (c autoClock) Sleep(_ context.Context, d time.Duration) error {
	c.Advance(d)
	return nil
}

func (h *liqHarness) saga(dryRun bool) *flatten.Saga {
	s := h.harness.saga(dryRun)
	s.Gateway = h.gateway
	s.Clock = autoClock{h.clock}
	s.Positions = h.account
	s.Balance = h.account
	s.Sellable = h.account
	s.Prices = h.account
	s.Currencies = []string{"KRW"}
	s.Rounds = 2
	s.StabiliseInterval = time.Second
	s.StabiliseAttempts = 3
	return s
}

func position(symbol, market string, quantity float64) domain.Position {
	return domain.Position{Symbol: symbol, MarketType: market, Quantity: quantity}
}

// --- tests ------------------------------------------------------------------

// TestLiquidationSizesFromTheSellableQuantity is the reduce-only rule: the order
// carries what the account says can be sold, never what we think we hold.
func TestLiquidationSizesFromTheSellableQuantity(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		// Only 6 are settled and sellable; 4 are not.
		sellable: map[string]float64{"005930": 6},
		lower:    map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)

	// The sell fills as soon as it is submitted, so the verification round sees
	// a flat account.
	h.sell.after = account.setFlat

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	orders := h.sell.orders()
	if len(orders) != 1 {
		t.Fatalf("orders = %+v, want one", orders)
	}
	if orders[0].Quantity != 6 {
		t.Errorf("quantity = %v, want 6 (the sellable quantity, not the 10 held)", orders[0].Quantity)
	}
	if orders[0].Side != "sell" {
		t.Errorf("side = %q, want sell — nothing here may produce a buy", orders[0].Side)
	}
	if orders[0].OrderType != "limit" {
		t.Errorf("order type = %q, want limit", orders[0].OrderType)
	}
	if orders[0].Price != 63000 {
		t.Errorf("price = %v, want the exchange lower limit 63000", orders[0].Price)
	}
	if report.Submitted != 1 {
		t.Errorf("submitted = %d, want 1", report.Submitted)
	}
}

// TestHeldSymbolIsNotSold is engine-safety's oversell scenario, end to end: a
// cancel that did not settle means the symbol is not sold at all.
func TestHeldSymbolIsNotSold(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10), position("000660", "kr", 4)},
		sellable:  map[string]float64{"005930": 10, "000660": 4},
		lower:     map[string]float64{"005930": 63000, "000660": 100000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{
		order("O-1", "000660", "SELL", "4", "120000", "KRW"),
	}}, account)
	h.broker.ambiguou["O-1"] = true
	ctx := context.Background()
	saga := h.saga(false)

	cancelReport, err := saga.CancelAll(ctx)
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if cancelReport.InDoubt != 1 {
		t.Fatalf("cancel report = %+v, want one in doubt", cancelReport)
	}

	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	for _, o := range h.sell.orders() {
		if o.Symbol == "000660" {
			t.Fatalf("000660's cancel is unresolved and it was sold anyway: %+v", o)
		}
	}
	var held *flatten.LiquidationTarget
	for i := range report.Targets {
		if report.Targets[i].Symbol == "000660" {
			held = &report.Targets[i]
		}
	}
	if held == nil || held.State != journal.FlattenStepHeld {
		t.Fatalf("000660 target = %+v, want held", held)
	}
	if !strings.Contains(held.Detail, "oversell") {
		t.Errorf("held detail = %q, want it to explain the oversell risk", held.Detail)
	}
	if report.Flat() {
		t.Error("a held symbol means the account is not flat")
	}
	if report.Phase != journal.FlattenPhaseStalled {
		t.Errorf("phase = %q, want STALLED so an operator finishes it", report.Phase)
	}
}

// TestDryRunLiquidationSubmitsNothing.
func TestDryRunLiquidationSubmitsNothing(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		sellable:  map[string]float64{"005930": 10},
		lower:     map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(true)

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	if got := h.sell.orders(); len(got) != 0 {
		t.Fatalf("a dry run submitted %d order(s): %+v", len(got), got)
	}
	if len(report.Targets) != 1 {
		t.Fatalf("targets = %+v, want the plan listed", report.Targets)
	}
	if !strings.Contains(report.Targets[0].Detail, "would sell") {
		t.Errorf("detail = %q, want it to state what would happen", report.Targets[0].Detail)
	}
	pending, err := h.journal.PendingAttempts(ctx)
	if err != nil {
		t.Fatalf("PendingAttempts: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("a dry run left %d attempt(s) in the journal", len(pending))
	}
}

// TestRepeatedRoundsCatchWhatTheFirstDidNot: a fill that lands between rounds, a
// quantity that settles late. The loop is bounded, and a position that survives
// it stalls the saga rather than looping forever.
func TestRepeatedRoundsCatchWhatTheFirstDidNot(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		sellable:  map[string]float64{"005930": 4}, // only part is sellable at first
		lower:     map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()

	saga := h.saga(false)
	saga.Rounds = 2
	// After the first sell, the rest settles and becomes sellable.
	h.sell.after = func() {
		account.mu.Lock()
		account.sellable["005930"] = 6
		account.positions = []domain.Position{position("005930", "kr", 6)}
		account.mu.Unlock()
	}

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}
	if report.Rounds != 2 {
		t.Errorf("rounds = %d, want the loop to have run twice", report.Rounds)
	}
	// The second round must not be blocked by the first round's step being
	// already recorded for the same symbol: the plan is per symbol, and a symbol
	// with a settled sell is not re-sold.
	if len(h.sell.orders()) == 0 {
		t.Fatal("no sell was submitted at all")
	}
	if report.Remaining == 0 {
		t.Error("the fake account still shows a position; the report must say so")
	}
	if report.Phase != journal.FlattenPhaseStalled {
		t.Errorf("phase = %q; a position that survives the rounds must stall the saga", report.Phase)
	}
}

// TestAggressivePriceFallsBackToTheLastTrade when the broker reports no band, and
// rounds down to a valid tick.
func TestAggressivePriceFallsBackToTheLastTrade(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 1)},
		sellable:  map[string]float64{"005930": 1},
		// No price band; last trade 71,234 with the default 5% discount is
		// 67,672.3, which rounds down to the 100-won tick.
		last: map[string]float64{"005930": 71234},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)

	h.sell.after = account.setFlat

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if _, err := saga.Liquidate(ctx); err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	orders := h.sell.orders()
	if len(orders) != 1 {
		t.Fatalf("orders = %+v, want one", orders)
	}
	if orders[0].Price != 67600 {
		t.Errorf("price = %v, want 67600 (71234 × 0.95 rounded down to the 100-won tick)", orders[0].Price)
	}
}

// TestUnpriceableSymbolIsHeldRatherThanGuessed: an order with no price is an
// order that cannot be sent, and inventing one on a live account is worse than
// reporting it.
func TestUnpriceableSymbolIsHeldRatherThanGuessed(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 5)},
		sellable:  map[string]float64{"005930": 5},
		priceErr:  errors.New("the quote endpoint is down"),
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}
	if got := h.sell.orders(); len(got) != 0 {
		t.Fatalf("an unpriceable symbol was sold anyway: %+v", got)
	}
	if report.Held != 1 {
		t.Errorf("held = %d, want 1 (%+v)", report.Held, report.Targets)
	}
}

// TestFractionalOnlyPositionIsReported: a limit order cannot express it, so it is
// surfaced rather than silently left behind.
func TestFractionalOnlyPositionIsReported(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("AAPL", "us", 0.4)},
		sellable:  map[string]float64{"AAPL": 0.4},
		last:      map[string]float64{"AAPL": 180},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	report, err := saga.Liquidate(ctx)
	if err != nil {
		t.Fatalf("Liquidate: %v", err)
	}
	if got := h.sell.orders(); len(got) != 0 {
		t.Fatalf("a fractional-only position produced an order: %+v", got)
	}
	if report.Held != 1 {
		t.Fatalf("held = %d, want 1 (%+v)", report.Held, report.Targets)
	}
	if !strings.Contains(report.Targets[0].Detail, "fractional") {
		t.Errorf("detail = %q, want it to name the reason", report.Targets[0].Detail)
	}
}

// TestLiquidateRefusesWithoutASellableReader: sizing a sell from the holding
// alone can oversell, so the dependency is mandatory rather than optional.
func TestLiquidateRefusesWithoutASellableReader(t *testing.T) {
	account := &accountFake{}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	saga := h.saga(false)
	saga.Sellable = nil

	if _, err := saga.Liquidate(context.Background()); !errors.Is(err, flatten.ErrNotConfigured) {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}

// --- confirmation -----------------------------------------------------------

func sampleConfirmation(now time.Time) flatten.Confirmation {
	return flatten.NewConfirmation("123-45678", 3, []flatten.LiquidationTarget{
		{Symbol: "005930", Held: 10},
		{Symbol: "AAPL", Held: 2.5},
	}, now)
}

// TestConfirmationPromptCarriesEveryRequiredFact.
func TestConfirmationPromptCarriesEveryRequiredFact(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	c := sampleConfirmation(now)
	prompt := c.Prompt()

	// Masked account, never the full number.
	if strings.Contains(prompt, "123-45678") || strings.Contains(prompt, "12345") {
		t.Errorf("the prompt leaks the account number:\n%s", prompt)
	}
	if !strings.Contains(prompt, "*****5678") {
		t.Errorf("the prompt does not show the masked account:\n%s", prompt)
	}
	// Position count, expected quantity, open orders, nonce, expiry.
	for _, want := range []string{"positions        2", "12.5", "open orders      3", c.Nonce, "expires"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt is missing %q:\n%s", want, prompt)
		}
	}
}

// TestConfirmationRequiresATerminal. There is no flag, env var or parameter that
// answers this for an automation.
func TestConfirmationRequiresATerminal(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	c := sampleConfirmation(now)

	var out strings.Builder
	err := flatten.Confirm(strings.NewReader(c.Nonce+"\n"), &out, c, false, now)
	if !errors.Is(err, flatten.ErrNotATerminal) {
		t.Fatalf("err = %v, want ErrNotATerminal — a pipe must not be able to confirm a flatten", err)
	}
	if out.Len() != 0 {
		t.Error("nothing should be printed to a non-terminal")
	}
}

func TestConfirmationAcceptsTheExactNonce(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	c := sampleConfirmation(now)

	var out strings.Builder
	if err := flatten.Confirm(strings.NewReader(c.Nonce+"\n"), &out, c, true, now); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !strings.Contains(out.String(), "FLATTEN-ALL") {
		t.Errorf("the prompt was not shown:\n%s", out.String())
	}
}

func TestConfirmationRejectsAnythingElse(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	c := sampleConfirmation(now)

	for _, answer := range []string{"y\n", "yes\n", "\n", "FLATTEN\n", strings.ToLower(c.Nonce) + "\n"} {
		var out strings.Builder
		err := flatten.Confirm(strings.NewReader(answer), &out, c, true, now)
		if !errors.Is(err, flatten.ErrConfirmationMismatch) {
			t.Errorf("answer %q: err = %v, want ErrConfirmationMismatch", strings.TrimSpace(answer), err)
		}
	}
}

// TestConfirmationExpires: a string copied into a runbook must not work months
// later against an account that has changed.
func TestConfirmationExpires(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	c := sampleConfirmation(now)

	if err := c.Verify(c.Nonce, now.Add(flatten.ConfirmationTTL-time.Second)); err != nil {
		t.Fatalf("inside the window: %v", err)
	}
	if err := c.Verify(c.Nonce, now.Add(flatten.ConfirmationTTL)); !errors.Is(err, flatten.ErrConfirmationExpired) {
		t.Fatalf("at the expiry: err = %v, want ErrConfirmationExpired", err)
	}
}

// TestNoncesAreNotReusableAcrossPlans.
func TestNoncesAreNotReusableAcrossPlans(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	first := sampleConfirmation(now)
	second := sampleConfirmation(now)

	if first.Nonce == second.Nonce {
		t.Fatal("two plans produced the same confirmation string")
	}
	if err := second.Verify(first.Nonce, now); !errors.Is(err, flatten.ErrConfirmationMismatch) {
		t.Errorf("a nonce from another plan was accepted: %v", err)
	}
}

// --- helper -----------------------------------------------------------------

func newTradingService(broker *sellBroker) *trading.Service {
	return trading.NewService(openPolicy(), broker)
}
