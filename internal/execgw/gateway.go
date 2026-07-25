// Package execgw is the engine's ExecutionGateway: the single door every order
// mutation the automated engine performs has to go through.
//
// # Why a wrapper and not a change to internal/trading
//
// The CLI's order path (internal/trading.Service) is interactive: it enforces the
// user's config policy and a confirm token a human retypes. That path must keep
// behaving exactly as upstream does (WORKFLOW §0.2), so the engine's extra
// obligations — durable intent journalling, Guardian authorisation, one dispatch
// per attempt, IN_DOUBT resolution — are layered *around* it here instead of
// inside it (design D1).
//
// # The ordering this package exists to guarantee
//
//	journal.Prepare  → RECORDED, committed and fsynced (nothing sent yet)
//	verify decision  → intent hash, expiry, limit snapshot
//	MarkDispatchStarted → committed, so a crash from here on is discoverable
//	spend the nonce  → one-shot, immediately before the call
//	broker call      → exactly once, never retried
//	settle           → CONFIRMED | FAILED_CONFIRMED | NOT_DISPATCHED | IN_DOUBT
//
// Each step is a precondition for the next, and the two that protect a live
// account are the first and the last: an intent that is not on disk is never
// submitted, and an outcome we cannot classify becomes IN_DOUBT rather than
// "probably fine".
//
// # What is deliberately absent
//
// There is no exported way to reach the wrapped broker. The gateway takes a
// *trading.Service and never hands it back, so "submit without a GuardianDecision"
// is not an API a caller can spell (engine-safety: "raw mutator 미노출").
package execgw

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Options are the gateway's dependencies.
type Options struct {
	// Journal is the durable intent journal. Required.
	Journal *journal.Journal
	// Trading is the policy-enforcing order service the gateway wraps. Required.
	// It is stored unexported and never returned.
	Trading *trading.Service
	// Clock is the injected time source. Defaults to clock.System().
	Clock clock.Clock
	// AccountRef identifies the account every intent is recorded against. Required.
	AccountRef string
	// Source names what produces the intents (strategy id, "operator", …).
	// Defaults to "engine".
	Source string
	// Nonces spends one-shot Guardian nonces. Defaults to NewMemoryNonceStore().
	Nonces NonceStore
	// Entry gates new exposure on required-query freshness and on latched
	// failures (retry matrix, task 2.6). Optional: nil means no gate, which is
	// what the pure gateway tests use. Engine wiring always supplies one.
	Entry *EntryGate
	// NewID generates intent and attempt ids. Defaults to a 128-bit random hex
	// string; tests inject a deterministic sequence.
	NewID func() string
}

// Gateway is the engine's only order mutation surface.
//
// Every field is unexported on purpose: the wrapped trading service is a mutator,
// and a caller that could read it back could bypass the journal and the Guardian.
type Gateway struct {
	journal    *journal.Journal
	trading    *trading.Service
	clk        clock.Clock
	accountRef string
	source     string
	nonces     NonceStore
	entry      *EntryGate
	newID      func() string
}

// New validates the wiring and returns a gateway.
func New(opts Options) (*Gateway, error) {
	switch {
	case opts.Journal == nil:
		return nil, errors.New("execgw: a journal is required — an unjournalled mutation is not submittable")
	case opts.Trading == nil:
		return nil, errors.New("execgw: a trading service is required")
	case strings.TrimSpace(opts.AccountRef) == "":
		return nil, errors.New("execgw: an account reference is required")
	}
	g := &Gateway{
		journal:    opts.Journal,
		trading:    opts.Trading,
		clk:        opts.Clock,
		accountRef: strings.TrimSpace(opts.AccountRef),
		source:     strings.TrimSpace(opts.Source),
		nonces:     opts.Nonces,
		entry:      opts.Entry,
		newID:      opts.NewID,
	}
	if g.clk == nil {
		g.clk = clock.System()
	}
	if g.source == "" {
		g.source = "engine"
	}
	if g.nonces == nil {
		g.nonces = NewMemoryNonceStore()
	}
	if g.newID == nil {
		g.newID = randomID
	}
	return g, nil
}

// Now reports the gateway clock's instant. Exposed because whoever issues a
// GuardianDecision must stamp its expiry from the same clock the gateway verifies
// it against — two time sources would make the expiry meaningless.
func (g *Gateway) Now() time.Time { return g.clk.Now() }

// OrderRef describes the broker order a cancel or an amend acts on. The gateway
// cannot read these off a CancelIntent/AmendIntent — upstream's intents carry only
// an order id — and the journal needs them to record what was attempted and to
// build the fingerprint an IN_DOUBT resolution matches on.
type OrderRef struct {
	// Market is "kr" or "us".
	Market string
	// Side is the side of the order being acted on ("BUY"/"SELL", any case).
	Side string
	// Quantity is the target order's quantity.
	Quantity float64
	// Price is the target order's limit price, 0 for a market order.
	Price float64
	// Currency is the target order's currency ("KRW"/"USD").
	Currency string
}

// PlaceRequest is a new order.
type PlaceRequest struct {
	Intent   orderintent.PlaceIntent
	Decision GuardianDecision
}

// CancelRequest cancels an existing order.
type CancelRequest struct {
	Intent   orderintent.CancelIntent
	Order    OrderRef
	Decision GuardianDecision
}

// AmendRequest amends an existing order. Symbol is required because
// orderintent.AmendIntent carries only the order id.
type AmendRequest struct {
	Intent   orderintent.AmendIntent
	Symbol   string
	Order    OrderRef
	Decision GuardianDecision
}

// Outcome is what a mutation ended as. It is returned even for refusals, because
// the attempt id is how an operator finds the record of what was refused.
type Outcome struct {
	IntentID  string
	AttemptID string
	// State is the attempt's terminal (or, for IN_DOUBT, current) journal state.
	State journal.AttemptState
	// Class is how the dispatch was classified.
	Class journal.DispatchClass
	// BrokerOrderID is the order number the broker named, when it named one.
	BrokerOrderID string
	Reason        ReasonCode
	Detail        string
	// Result is the broker's own report, empty unless the call completed.
	Result domain.MutationResult
}

// Blocking reports whether this outcome leaves the symbol blocked (not settled).
func (o Outcome) Blocking() bool { return o.State != "" && !o.State.IsTerminal() }

// mutationPlan is the gateway's internal description of one mutation: everything
// the journal, the Guardian check and the broker call each need, resolved once.
type mutationPlan struct {
	kind          journal.MutationKind
	intentHash    string
	market        string
	symbol        string
	side          string // "BUY" | "SELL"
	orderType     string // "LIMIT" | "MARKET"
	quantity      float64
	price         float64
	amount        float64 // amount-based (fractional buy) orders
	currency      string
	targetOrderID string
	// raisesExposure marks a mutation that can increase risk, which is what makes
	// a limit snapshot mandatory rather than merely checked.
	raisesExposure bool
	// call performs the broker mutation exactly once.
	call func(ctx context.Context) (domain.MutationResult, error)
}

// Place submits a new order through the full gateway sequence.
func (g *Gateway) Place(ctx context.Context, req PlaceRequest) (Outcome, error) {
	intent := req.Intent
	plan := mutationPlan{
		kind:           journal.KindPlace,
		intentHash:     PlaceHash(intent),
		market:         strings.ToLower(strings.TrimSpace(intent.Market)),
		symbol:         strings.ToUpper(strings.TrimSpace(intent.Symbol)),
		side:           strings.ToUpper(strings.TrimSpace(intent.Side)),
		orderType:      strings.ToUpper(strings.TrimSpace(intent.OrderType)),
		quantity:       intent.Quantity,
		price:          intent.Price,
		amount:         intent.Amount,
		currency:       strings.ToUpper(strings.TrimSpace(intent.CurrencyMode)),
		raisesExposure: strings.EqualFold(intent.Side, "buy"),
	}
	plan.call = func(ctx context.Context) (domain.MutationResult, error) {
		return g.trading.Place(ctx, intent, g.executeOpts(g.trading.PreviewPlace(intent).ConfirmToken))
	}
	return g.submit(ctx, plan, req.Decision)
}

// Cancel cancels an existing order through the full gateway sequence.
func (g *Gateway) Cancel(ctx context.Context, req CancelRequest) (Outcome, error) {
	intent := req.Intent
	plan := mutationPlan{
		kind:          journal.KindCancel,
		intentHash:    CancelHash(intent),
		market:        strings.ToLower(strings.TrimSpace(req.Order.Market)),
		symbol:        strings.ToUpper(strings.TrimSpace(intent.Symbol)),
		side:          strings.ToUpper(strings.TrimSpace(req.Order.Side)),
		orderType:     orderTypeFor(req.Order.Price),
		quantity:      req.Order.Quantity,
		price:         req.Order.Price,
		currency:      strings.ToUpper(strings.TrimSpace(req.Order.Currency)),
		targetOrderID: strings.TrimSpace(intent.OrderID),
		// A cancel can only remove exposure.
		raisesExposure: false,
	}
	plan.call = func(ctx context.Context) (domain.MutationResult, error) {
		return g.trading.Cancel(ctx, intent, g.executeOpts(g.trading.PreviewCancel(intent).ConfirmToken))
	}
	return g.submit(ctx, plan, req.Decision)
}

// Amend amends an existing order through the full gateway sequence.
func (g *Gateway) Amend(ctx context.Context, req AmendRequest) (Outcome, error) {
	intent := req.Intent
	quantity := req.Order.Quantity
	if intent.Quantity != nil {
		quantity = *intent.Quantity
	}
	price := req.Order.Price
	if intent.Price != nil {
		price = *intent.Price
	}
	plan := mutationPlan{
		kind:          journal.KindAmend,
		intentHash:    AmendHash(intent),
		market:        strings.ToLower(strings.TrimSpace(req.Order.Market)),
		symbol:        strings.ToUpper(strings.TrimSpace(req.Symbol)),
		side:          strings.ToUpper(strings.TrimSpace(req.Order.Side)),
		orderType:     orderTypeFor(price),
		quantity:      quantity,
		price:         price,
		currency:      strings.ToUpper(strings.TrimSpace(req.Order.Currency)),
		targetOrderID: strings.TrimSpace(intent.OrderID),
		// An amend that raises quantity or price adds exposure to a live order,
		// so it is measured against the snapshot exactly like a place.
		raisesExposure: quantity > req.Order.Quantity || price > req.Order.Price,
	}
	plan.call = func(ctx context.Context) (domain.MutationResult, error) {
		return g.trading.Amend(ctx, intent, g.executeOpts(g.trading.PreviewAmend(intent).ConfirmToken))
	}
	return g.submit(ctx, plan, req.Decision)
}

// executeOpts satisfies upstream's confirm-token gate.
//
// The token is not the engine's authorisation — the GuardianDecision is. It exists
// so a human cannot fat-finger a CLI order, and the gateway computes it from the
// same canonical intent trading.Service will compare against, which keeps
// internal/trading unmodified (design D1) while the real gate stays the Guardian.
func (g *Gateway) executeOpts(token string) trading.ExecuteOptions {
	return trading.ExecuteOptions{Execute: true, Confirm: token}
}

// submit runs the ordering documented at the top of this file.
func (g *Gateway) submit(ctx context.Context, plan mutationPlan, decision GuardianDecision) (Outcome, error) {
	prep, err := g.prepareRequest(plan)
	if err != nil {
		return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err
	}

	// 1. Durable record first. Prepare only returns a handle after its
	//    BEGIN IMMEDIATE transaction committed on a synchronous=FULL connection,
	//    so there is nothing to dispatch with until the intent is on disk.
	attempt, err := g.journal.Prepare(ctx, prep)
	if err != nil {
		return Outcome{IntentID: prep.Intent.ID, Reason: ReasonInvalidRequest, Detail: err.Error()},
			fmt.Errorf("execgw: recording the intent (nothing was submitted): %w", err)
	}

	out := Outcome{IntentID: prep.Intent.ID, AttemptID: attempt.ID()}

	// 2. Refusals, in order of cost. Both are recorded against the attempt rather
	//    than raised before it, so "why did the engine not trade" is answerable
	//    from the journal alone.
	//
	//    2a. The entry gate, for mutations that add exposure. Exits are never
	//        gated (§0.3).
	if rejected := g.checkEntry(plan); rejected != nil {
		return g.refuse(ctx, attempt, out, rejected)
	}
	//    2b. The Guardian decision, before the dispatch is even recorded, so an
	//        unauthorised mutation never reaches DISPATCH_STARTED and never
	//        blocks the symbol.
	if rejected := verifyDecision(decision, plan, g.clk.Now()); rejected != nil {
		return g.refuse(ctx, attempt, out, rejected)
	}

	// 3-5. Dispatch exactly once and settle from the classification.
	var result domain.MutationResult
	res, err := attempt.Dispatch(ctx, func(dctx context.Context, _ *journal.Attempt) journal.DispatchOutcome {
		// Re-verified immediately before the call: the expiry only means
		// something if it is checked at the last possible moment.
		now := g.clk.Now()
		if rejected := verifyDecision(decision, plan, now); rejected != nil {
			return notSent(rejected)
		}
		if err := g.nonces.Consume(decision.Nonce, now); err != nil {
			return notSent(reject(ReasonGuardianNonceReused,
				"the decision for %s was already spent", plan.symbol))
		}

		// The tracker is what separates "provably never sent" from "may have
		// executed"; without it a connection error would have to be treated as
		// ambiguous every time.
		tctx, tracker := journal.WithSendTracker(dctx)
		var callErr error
		result, callErr = plan.call(tctx)
		return classifyMutation(tracker.State(), result, callErr)
	})
	if err != nil {
		return out, fmt.Errorf("execgw: settling the %s of %s: %w", plan.kind, plan.symbol, err)
	}

	out.State = res.Final
	out.Class = res.Class
	out.BrokerOrderID = res.BrokerOrderID
	out.Reason = ReasonCode(res.ReasonCode)
	out.Detail = res.Detail
	out.Result = result

	if res.Final == journal.StateConfirmed {
		return out, nil
	}
	if res.Err != nil {
		return out, res.Err
	}
	return out, &RejectedError{Reason: out.Reason, Detail: out.Detail}
}

// checkEntry asks the gate whether new exposure is allowed. Mutations that do not
// raise exposure — cancels, and amends that only lower quantity or price — are
// never gated: the whole point of blocking entries is to keep the exits open.
func (g *Gateway) checkEntry(plan mutationPlan) *RejectedError {
	if g.entry == nil || !plan.raisesExposure {
		return nil
	}
	return g.entry.CheckEntry()
}

// refuse closes a journalled attempt that never reached the broker.
func (g *Gateway) refuse(ctx context.Context, attempt *journal.Attempt, out Outcome, rejected *RejectedError) (Outcome, error) {
	out.Reason = rejected.Reason
	out.Detail = rejected.Detail
	out.State = journal.StateNotDispatched
	if err := attempt.Settle(ctx, journal.StateNotDispatched,
		string(rejected.Reason), rejected.Detail); err != nil {
		return out, fmt.Errorf("execgw: closing a refused attempt: %w", err)
	}
	return out, rejected
}

// notSent turns a gateway refusal into a dispatch outcome that settles the attempt
// as NOT_DISPATCHED. It is only used on paths where no byte was written.
func notSent(rejected *RejectedError) journal.DispatchOutcome {
	return journal.DispatchOutcome{
		Class:      journal.DispatchNotSent,
		ReasonCode: string(rejected.Reason),
		Detail:     rejected.Detail,
		Err:        rejected,
	}
}

// prepareRequest turns a plan into the journal's record of it.
func (g *Gateway) prepareRequest(plan mutationPlan) (journal.PrepareRequest, error) {
	if plan.symbol == "" {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest, "the mutation has no symbol")
	}
	market, err := clock.ParseMarket(plan.market)
	if err != nil {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest,
			"market %q is not one this build trades", plan.market)
	}
	day, err := market.TradingDay(g.clk.Now())
	if err != nil {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest, "resolving the trading day: %v", err)
	}
	side := plan.side
	if side != "BUY" && side != "SELL" {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest,
			"side %q is neither BUY nor SELL", plan.side)
	}

	fingerprint := Fingerprint(FingerprintInput{
		AccountRef: g.accountRef,
		Market:     string(market),
		Symbol:     plan.symbol,
		Side:       side,
		Quantity:   plan.quantity,
		Price:      plan.price,
		Currency:   plan.currency,
		TradingDay: day,
	})

	intent := journal.Intent{
		ID:          g.newID(),
		Market:      string(market),
		TradingDay:  day,
		AccountRef:  g.accountRef,
		Symbol:      plan.symbol,
		Side:        side,
		OrderType:   plan.orderType,
		Quantity:    decimalString(plan.quantity),
		Price:       priceString(plan.price),
		Currency:    plan.currency,
		Source:      g.source,
		Fingerprint: fingerprint,
	}
	if intent.OrderType == "" {
		intent.OrderType = "LIMIT"
	}
	if intent.Currency == "" {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest, "the mutation has no currency")
	}

	return journal.PrepareRequest{
		Intent:        intent,
		Kind:          plan.kind,
		AttemptID:     g.newID(),
		TargetOrderID: plan.targetOrderID,
	}, nil
}

func orderTypeFor(price float64) string {
	if price > 0 {
		return "LIMIT"
	}
	return "MARKET"
}

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// priceString renders a price for the journal, where "" means "no price" and is
// stored as NULL (market orders).
func priceString(v float64) string {
	if v <= 0 {
		return ""
	}
	return decimalString(v)
}

// randomID is the default id generator: 128 bits of randomness, which is enough
// that an id collision is not a failure mode worth designing against.
func randomID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand failing is not recoverable in a way that keeps the journal
		// meaningful, and returning a duplicate id silently would be worse.
		panic("execgw: crypto/rand is unavailable: " + err.Error())
	}
	return hex.EncodeToString(buf[:])
}
