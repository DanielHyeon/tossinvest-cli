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
//	read the decision   → from the journal; the caller supplies only its id
//	journal.Prepare     → RECORDED with the decision binding, committed and fsynced
//	verify the decision → preimage, class, expiry, key, limit snapshot
//	MarkDispatchStarted → committed *with the nonce spent in the same transaction*
//	re-read and re-verify → the row again, immediately before the call
//	broker call         → exactly once, never retried
//	settle              → CONFIRMED | FAILED_CONFIRMED | NOT_DISPATCHED | IN_DOUBT
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
	"sync"
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
	// Entry gates new exposure on required-query freshness and on latched
	// failures (retry matrix, task 2.6). Optional: nil means no gate, which is
	// what the pure gateway tests use. Engine wiring always supplies one.
	Entry *EntryGate
	// Preflight runs the fail-closed checks a place can be refused by before it
	// is sent (task 2.10). Optional; nil skips them, which only the narrower
	// gateway tests do.
	Preflight *Preflight
	// Orders reads one order back by id, which is how a created order's
	// identifier is confirmed to name something real before the attempt settles
	// (task 5.2, roundtrip.go). execgw.OfficialOrders satisfies it.
	//
	// Optional only in the sense that the type allows nil: without it a place is
	// confirmed on the broker's ack alone, which is what this build did before
	// the round trip existed. Engine wiring supplies it.
	Orders OrderReader
	// NewID generates intent and attempt ids. Defaults to a 128-bit random hex
	// string; tests inject a deterministic sequence.
	NewID func() string

	// Replay resends the request body an IN_DOUBT attempt stored, so its
	// identity can be recovered from the broker's idempotent answer (replay.go).
	// Optional: without it, an IN_DOUBT attempt is resolved by observation
	// alone, which is what this build does today.
	Replay ReplayTransport
	// ReplayConfig tunes the replay bounds. Zero fields take the defaults.
	ReplayConfig ReplayConfig
	// Attested reports whether the replay capability has been verified against
	// the real broker. Nil — the default — means "not attested", and the replay
	// entry point resends nothing: the capability is measured by
	// verify-execution-capability and the replay path stays off until then
	// [미측정 — 2b 전 비활성].
	Attested func(ctx context.Context) bool
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
	entry      *EntryGate
	preflight  *Preflight
	orders     OrderReader
	newID      func() string
	replay     ReplayTransport
	replayCfg  ReplayConfig
	attested   func(ctx context.Context) bool

	// inflight holds the in-process claim on a symbol for the duration of a
	// mutation, so two goroutines cannot both pass the journal's in-flight check.
	inflightMu sync.Mutex
	inflight   map[string]bool
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
		entry:      opts.Entry,
		preflight:  opts.Preflight,
		orders:     opts.Orders,
		newID:      opts.NewID,
		replay:     opts.Replay,
		replayCfg:  opts.ReplayConfig,
		attested:   opts.Attested,
	}
	if g.clk == nil {
		g.clk = clock.System()
	}
	if g.source == "" {
		g.source = "engine"
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

// Wiring reports which optional dependencies a gateway was constructed with.
//
// Every one of these is nil-able in Options, and every one of them changes what
// the gateway guarantees rather than merely how fast it is: without Orders a
// created order's identifier is never checked against the broker, without Entry
// nothing stops a new entry during a RECONCILE. "Optional" is a statement about
// the narrower unit tests, not about a production gateway — so the startup
// interlock asks this question instead of assuming the answer
// (engine-safety: "엔진 프로필에 ExecutionGateway가 구성됨").
type Wiring struct {
	// Orders reports that the round-trip reader is wired, so a created order's
	// identifier is confirmed against the broker before the attempt settles.
	Orders bool
	// Entry reports that the entry gate is wired.
	Entry bool
	// Preflight reports that the fail-closed pre-submission checks are wired.
	Preflight bool
	// Replay reports that a replay transport exists. It does not mean replay is
	// enabled — see Attested.
	Replay bool
	// Attested reports that an attestation predicate is wired. Without one the
	// replay entry point resends nothing [미측정 — 2b 전 비활성].
	Attested bool
}

// Wiring reports this gateway's optional dependencies.
func (g *Gateway) Wiring() Wiring {
	return Wiring{
		Orders:    g.orders != nil,
		Entry:     g.entry != nil,
		Preflight: g.preflight != nil,
		Replay:    g.replay != nil,
		Attested:  g.attested != nil,
	}
}

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
	// Baseline is the account snapshot taken before this order was decided. It
	// is recorded on the intent and is what lets IN_DOUBT resolution prove
	// absence later. Optional — see Baseline for what omitting it costs.
	Baseline *Baseline
	// IntentID pre-names the intent this place will record. Empty is the normal
	// case and the gateway mints one.
	//
	// It exists for exactly one caller: the exit observation loop, whose crash
	// contract is "arm the proposal, mint the intent, attach it, *then* submit"
	// (apply_hook.go). A proposal armed with no intent id is answered by no fill,
	// because the exit applier resolves on the intent the proposal named — so a
	// crash between a submit and a later attach would leave a live sell order
	// above a proposal nothing can clear, and every later proposal on that
	// position suppressed by it. Naming the intent before the order exists closes
	// that window: the id is on disk in `exit_states.pending_intent_id` before
	// anything is sent, and the fill that arrives carries the same one.
	//
	// It authorises nothing. The authority is the GuardianDecision, which is
	// re-read from the ledger; this is a primary key, and a collision is refused
	// by Prepare rather than silently reusing a record.
	IntentID string
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
	kind      journal.MutationKind
	market    string
	symbol    string
	side      string // "BUY" | "SELL"
	orderType string // "LIMIT" | "MARKET"
	quantity  float64
	price     float64
	amount    float64 // amount-based (fractional buy) orders
	currency  string
	// targetOrderID is the broker order a cancel or an amend acts on. It is an
	// opaque token (openapi contracts no shape for `orderId`), so it is carried
	// and stored exactly as received — trimming is only ever an emptiness test.
	targetOrderID string
	// raisesExposure marks a mutation that can increase risk, which is what makes
	// a limit snapshot mandatory rather than merely checked.
	raisesExposure bool
	// baseline is the pre-dispatch account snapshot, recorded on the intent for
	// IN_DOUBT resolution.
	baseline *Baseline
	// intentID is the caller's pre-minted intent id, empty for the usual case
	// where the gateway mints one (see PlaceRequest.IntentID).
	intentID string
	// preflight runs the fail-closed checks that apply to this mutation kind.
	// Only a place has any today; nil means there is nothing to check.
	preflight func(ctx context.Context) *RejectedError
	// call performs the broker mutation exactly once. key is the idempotency key
	// the verified decision authorised, empty for mutations that carry none —
	// it is threaded through the call rather than captured, because it is only
	// known after the persisted decision has been read.
	call func(ctx context.Context, key string) (domain.MutationResult, error)
	// wireBody renders the exact request body the call will send under key, so
	// the replay procedure can resend those bytes instead of rebuilding them
	// (design D2). nil for mutations with no key and therefore no replay.
	wireBody func(key string) (string, error)
}

// Place submits a new order through the full gateway sequence.
func (g *Gateway) Place(ctx context.Context, req PlaceRequest) (Outcome, error) {
	intent := req.Intent
	plan := mutationPlan{
		kind:      journal.KindPlace,
		market:    strings.ToLower(strings.TrimSpace(intent.Market)),
		symbol:    strings.ToUpper(strings.TrimSpace(intent.Symbol)),
		side:      strings.ToUpper(strings.TrimSpace(intent.Side)),
		orderType: strings.ToUpper(strings.TrimSpace(intent.OrderType)),
		quantity:  intent.Quantity,
		price:     intent.Price,
		amount:    intent.Amount,
		currency:  strings.ToUpper(strings.TrimSpace(intent.CurrencyMode)),
		// Computed here, from the mutation itself, and never taken from the
		// decision: a class the caller declares is a label, and this is the fact
		// it has to agree with (engine-safety "결정의 Safety Class와 형태 일치").
		raisesExposure: strings.EqualFold(intent.Side, "buy"),
		baseline:       req.Baseline,
		intentID:       strings.TrimSpace(req.IntentID),
	}
	if g.preflight != nil {
		plan.preflight = func(ctx context.Context) *RejectedError {
			return g.preflight.CheckPlace(ctx, intent)
		}
	}
	// The key is never taken from the caller's intent: it is derived from the
	// persisted decision and written over whatever the caller supplied.
	keyed := func(key string) orderintent.PlaceIntent {
		i := intent
		i.ClientOrderID = key
		return i
	}
	plan.call = func(ctx context.Context, key string) (domain.MutationResult, error) {
		i := keyed(key)
		return g.trading.Place(ctx, i, g.executeOpts(g.trading.PreviewPlace(i).ConfirmToken))
	}
	plan.wireBody = func(key string) (string, error) { return PlaceWireBody(keyed(key)) }
	return g.submit(ctx, plan, req.Decision)
}

// Cancel cancels an existing order through the full gateway sequence.
func (g *Gateway) Cancel(ctx context.Context, req CancelRequest) (Outcome, error) {
	intent := req.Intent
	plan := mutationPlan{
		kind:          journal.KindCancel,
		market:        strings.ToLower(strings.TrimSpace(req.Order.Market)),
		symbol:        strings.ToUpper(strings.TrimSpace(intent.Symbol)),
		side:          strings.ToUpper(strings.TrimSpace(req.Order.Side)),
		orderType:     orderTypeFor(req.Order.Price),
		quantity:      req.Order.Quantity,
		price:         req.Order.Price,
		currency:      strings.ToUpper(strings.TrimSpace(req.Order.Currency)),
		targetOrderID: intent.OrderID,
		// A cancel can only remove exposure.
		raisesExposure: false,
	}
	plan.call = func(ctx context.Context, _ string) (domain.MutationResult, error) {
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
		market:        strings.ToLower(strings.TrimSpace(req.Order.Market)),
		symbol:        strings.ToUpper(strings.TrimSpace(req.Symbol)),
		side:          strings.ToUpper(strings.TrimSpace(req.Order.Side)),
		orderType:     orderTypeFor(price),
		quantity:      quantity,
		price:         price,
		currency:      strings.ToUpper(strings.TrimSpace(req.Order.Currency)),
		targetOrderID: intent.OrderID,
		// An amend that raises quantity or price adds exposure to a live order,
		// so it is measured against the snapshot exactly like a place.
		raisesExposure: quantity > req.Order.Quantity || price > req.Order.Price,
	}
	plan.call = func(ctx context.Context, _ string) (domain.MutationResult, error) {
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
func (g *Gateway) submit(ctx context.Context, plan mutationPlan, ref GuardianDecision) (Outcome, error) {
	// The decision is *read* before the attempt is recorded, because the record
	// has to carry the binding (which decision, which class, which key, which
	// bytes) at RECORDED. It is *verified* after, so a refusal is journalled
	// rather than invisible — the two are separate steps on purpose.
	decision, loadRejected := g.loadDecision(ctx, ref)

	prep, err := g.prepareRequest(plan, decision)
	if err != nil {
		return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err
	}

	// 0. One in-flight mutation per symbol. This is not throttling: it is what
	//    makes an IN_DOUBT fingerprint match unique, because two identical
	//    outstanding orders on one symbol would be indistinguishable in the
	//    broker's list. The in-process claim closes the check-then-act race; the
	//    journal query catches whatever an earlier process left behind.
	//
	//    Refusals here are not journalled: the blocking attempt is already in the
	//    journal and is itself the record of why.
	symbolKey := plan.market + "|" + plan.symbol
	if !g.claimSymbol(symbolKey) {
		return Outcome{Reason: ReasonSymbolInFlight, Detail: "another mutation on " + plan.symbol + " is in flight"},
			reject(ReasonSymbolInFlight, "another mutation on %s is already in flight", plan.symbol)
	}
	defer g.releaseSymbol(symbolKey)

	if rejected, err := g.checkSymbolFree(ctx, plan); err != nil {
		return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err
	} else if rejected != nil {
		return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected
	}

	//    0b. A decision is spent once. Re-submitting one is caught here rather
	//        than by the UNIQUE index or the dispatch transaction, so it reports
	//        "already spent" instead of a write error — and, like the check
	//        above, it is not journalled: the attempt that spent it is the record.
	if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil {
		return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err
	} else if rejected != nil {
		return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected
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
	//    2b. The fail-closed branches: an order shape the path cannot express, a
	//        balance that cannot cover the buy, a conversion the engine must not
	//        trigger. None of these become true by being sent.
	if plan.preflight != nil {
		if rejected := plan.preflight(ctx); rejected != nil {
			return g.refuse(ctx, attempt, out, rejected)
		}
	}
	//    2c. The Guardian decision, before the dispatch is even recorded, so an
	//        unauthorised mutation never reaches DISPATCH_STARTED and never
	//        blocks the symbol. A decision that could not be read at all is
	//        refused here for the same reason: the refusal belongs in the
	//        journal next to the attempt it refused.
	if loadRejected != nil {
		return g.refuse(ctx, attempt, out, loadRejected)
	}
	if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil {
		return g.refuse(ctx, attempt, out, rejected)
	}
	//    2d. The reservation, for the entries that need one. It comes after the
	//        decision check because the *verified* class is what decides whether
	//        one is required — a class the caller merely asserted could exempt
	//        itself from the aggregate limits by claiming to be an exit.
	if rejected := g.checkReservation(ctx, decision); rejected != nil {
		return g.refuse(ctx, attempt, out, rejected)
	}

	// 3-5. Dispatch exactly once, confirm a created order id exists, and settle
	//      from the classification. The round trip runs inside the journal's
	//      dispatch (after MarkAcked, before Settle) because that is the only
	//      window where an unconfirmable id can still become IN_DOUBT instead of
	//      CONFIRMED.
	var result domain.MutationResult
	res, err := attempt.DispatchVerified(ctx, func(dctx context.Context, _ *journal.Attempt) journal.DispatchOutcome {
		// Re-read and re-verified immediately before the call. Re-*read*,
		// not merely re-checked: "제출 직전 journal에서 읽은 preimage로 재검증"
		// means the row itself is the thing consulted at the last moment, so a
		// decision revoked or rewritten since Prepare cannot be submitted from a
		// copy this goroutine happens to still hold.
		now := g.clk.Now()
		fresh, rejected := g.loadDecision(dctx, ref)
		if rejected != nil {
			return notSent(rejected)
		}
		if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil {
			return notSent(rejected)
		}
		if fresh.ClientOrderID != prep.ClientOrderID {
			return notSent(reject(ReasonGuardianKeyMismatch,
				"the idempotency key recorded for this attempt is not the one the decision now carries"))
		}
		// The tracker is what separates "provably never sent" from "may have
		// executed"; without it a connection error would have to be treated as
		// ambiguous every time.
		tctx, tracker := journal.WithSendTracker(dctx)
		var callErr error
		result, callErr = plan.call(tctx, fresh.ClientOrderID)
		return classifyMutation(tracker.State(), result, callErr)
	}, g.roundTripFor(plan))
	if err != nil {
		// The nonce is spent inside the transaction that records the dispatch, so
		// a reuse surfaces here rather than as a refusal from the closure. The
		// attempt is still RECORDED and nothing was sent: close it as such.
		if errors.Is(err, journal.ErrNonceSpent) {
			return g.refuse(ctx, attempt, out, reject(ReasonGuardianNonceReused,
				"the decision for %s was already spent", plan.symbol))
		}
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

// claimSymbol takes the in-process slot for a symbol, reporting false when
// another goroutine already holds it.
func (g *Gateway) claimSymbol(key string) bool {
	g.inflightMu.Lock()
	defer g.inflightMu.Unlock()
	if g.inflight == nil {
		g.inflight = make(map[string]bool)
	}
	if g.inflight[key] {
		return false
	}
	g.inflight[key] = true
	return true
}

func (g *Gateway) releaseSymbol(key string) {
	g.inflightMu.Lock()
	defer g.inflightMu.Unlock()
	delete(g.inflight, key)
}

// checkSymbolFree asks the journal whether this symbol already has an outstanding
// mutation, or an attempt parked as UNRESOLVED_IN_DOUBT.
//
// The two cases differ in scope on purpose:
//
//   - An unsettled attempt blocks *every* mutation on the symbol, including a
//     cancel: with two outstanding mutations we could not tell which broker order
//     belongs to which attempt, and a cancel aimed at the wrong one is worse than
//     a delayed cancel.
//   - An UNRESOLVED_IN_DOUBT attempt blocks only new exposure. The symbol's
//     history contains an order we could not account for, so adding more is out
//     of the question — but the operator must still be able to cancel and
//     liquidate through the engine (§0.3).
func (g *Gateway) checkSymbolFree(ctx context.Context, plan mutationPlan) (*RejectedError, error) {
	pending, err := g.journal.PendingAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("execgw: checking for in-flight mutations on %s: %w", plan.symbol, err)
	}
	for _, rec := range pending {
		same, err := g.attemptTargets(ctx, rec, plan)
		if err != nil {
			return nil, err
		}
		if same {
			return reject(ReasonSymbolInFlight,
				"attempt %s (%s, %s) on %s has not settled yet",
				rec.ID, rec.Kind, rec.State, plan.symbol), nil
		}
	}
	if !plan.raisesExposure {
		return nil, nil
	}
	unresolved, err := g.journal.UnresolvedAttempts(ctx)
	if err != nil {
		return nil, fmt.Errorf("execgw: checking unresolved attempts on %s: %w", plan.symbol, err)
	}
	for _, rec := range unresolved {
		same, err := g.attemptTargets(ctx, rec, plan)
		if err != nil {
			return nil, err
		}
		if same {
			return reject(ReasonUnresolvedInDoubt,
				"attempt %s on %s is unresolved; an operator has to close it before new exposure is added",
				rec.ID, plan.symbol), nil
		}
	}
	return nil, nil
}

// attemptTargets reports whether a recorded attempt concerns the same market and
// symbol as the plan.
func (g *Gateway) attemptTargets(ctx context.Context, rec journal.AttemptRecord, plan mutationPlan) (bool, error) {
	intent, err := g.journal.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return false, fmt.Errorf("execgw: reading the intent of attempt %s: %w", rec.ID, err)
	}
	return strings.EqualFold(intent.Symbol, plan.symbol) &&
		strings.EqualFold(intent.Market, plan.market), nil
}

// checkEntry asks the gate whether new exposure is allowed. Mutations that do not
// raise exposure — cancels, and amends that only lower quantity or price — are
// never gated: the whole point of blocking entries is to keep the exits open.
//
// The question is asked *for this market and symbol* (task 4.2). Account-wide
// blocks still apply to everything; a block scoped to one symbol now stops only
// that symbol, which is what the reconciliation state table always said and what
// the gate could not express until it had a symbol dimension.
func (g *Gateway) checkEntry(plan mutationPlan) *RejectedError {
	if g.entry == nil || !plan.raisesExposure {
		return nil
	}
	return g.entry.CheckEntryFor(plan.market, plan.symbol)
}

// checkReservation is the gateway's half of "예약이 총계 한도의 권위" (task 5.1,
// engine-safety: "Gateway는 EXPOSURE_RAISING 결정의 제출 시 HELD 상태의 예약
// 존재를 검증한다 — 예약 없는 진입 결정은 거부된다").
//
// # Why this exists when the issuance is already atomic
//
// The atomic API makes "a submittable decision with no reservation" unreachable
// *from the issuer* — a refused reservation rolls the decision back with it. But
// unreachable-by-construction only covers the issuers this build has. A decision
// row can also arrive from an older build, from a hand-edited journal, or from a
// future issuer that takes the reservation in a second transaction, and in every
// one of those cases the row looks exactly like a valid authorisation. The two
// checks together are what make the statement true rather than intended: one
// removes the state, the other refuses it.
//
// # Why it is not repeated immediately before the broker call
//
// The last-moment re-verification exists for the things that can change under a
// held decision: the row (tamper) and the clock (expiry). A HELD reservation is
// released by the attempt settling — this attempt, which has not dispatched — or
// by the lapse sweep, which releases holds whose decision has expired; and an
// expired decision is refused by the re-read `checkDecision` in that same
// window. So there is no interleaving that turns a HELD reservation into a
// released one while the decision it belongs to is still submittable.
//
// # Failing closed on a read error
//
// A reservation that cannot be read is treated as one that is not there. "The
// database did not answer" is not evidence that the headroom was taken, and this
// is the entry path — the direction where the conservative answer is to refuse.
func (g *Gateway) checkReservation(ctx context.Context, dec journal.Decision) *RejectedError {
	if dec.SafetyClass != journal.SafetyClassExposureRaising {
		return nil
	}
	reservations, err := g.journal.ReservationsForDecision(ctx, dec.ID)
	if err != nil {
		return reject(ReasonGuardianReservationMissing,
			"the reservations of decision %s could not be read, so the headroom it consumes "+
				"cannot be shown to have been taken: %v", short(dec.ID), err)
	}
	for _, reservation := range reservations {
		if reservation.Held() {
			return nil
		}
	}
	return reject(ReasonGuardianReservationMissing,
		"decision %s holds no risk reservation (%d recorded, none HELD); the aggregate limits "+
			"are enforced by the reservation ledger, so an entry that never took its headroom "+
			"is not an authorised one", short(dec.ID), len(reservations))
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

// checkDecisionUnspent reports whether this decision has already been submitted
// once.
//
// Both halves are durable reads, not in-memory state: the attempt that spent the
// decision may belong to an earlier process, and a restarted engine holding a
// persisted decision snapshot is exactly the reuse the spec names
// ("재시작 후 결정 재사용 시도"). The authoritative refusal is still the insert
// inside the dispatch transaction — this one exists so the common case is
// refused before an attempt is recorded for it, and so the reason code says
// "already spent" rather than surfacing as a write error.
func (g *Gateway) checkDecisionUnspent(ctx context.Context, prep journal.PrepareRequest,
	decision journal.Decision,
) (*RejectedError, error) {
	if prep.ClientOrderID != "" {
		rec, err := g.journal.AttemptForClientOrderID(ctx, prep.Intent.AccountRef, prep.ClientOrderID)
		switch {
		case err == nil:
			return reject(ReasonGuardianNonceReused,
				"attempt %s already submitted this decision", rec.ID), nil
		case !errors.Is(err, journal.ErrAttemptNotFound):
			return nil, fmt.Errorf("execgw: checking whether the idempotency key is claimed: %w", err)
		}
	}
	if decision.Nonce == "" {
		return nil, nil
	}
	spent, err := g.journal.NonceSpent(ctx, decision.Nonce)
	if err != nil {
		return nil, fmt.Errorf("execgw: checking whether the decision was spent: %w", err)
	}
	if spent {
		return reject(ReasonGuardianNonceReused,
			"decision %s was already spent", short(decision.ID)), nil
	}
	return nil, nil
}

// prepareRequest turns a plan and its persisted decision into the journal's
// record of both.
//
// A zero decision produces a request with no binding, which Prepare accepts and
// the verification then refuses: the refusal is worth recording, and it can only
// be recorded against an attempt that exists.
func (g *Gateway) prepareRequest(plan mutationPlan, decision journal.Decision) (journal.PrepareRequest, error) {
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

	intentID := plan.intentID
	if intentID == "" {
		intentID = g.newID()
	}
	intent := journal.Intent{
		ID:          intentID,
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
		Notes:       EncodeBaseline(plan.baseline),
	}
	if intent.OrderType == "" {
		intent.OrderType = "LIMIT"
	}
	if intent.Currency == "" {
		return journal.PrepareRequest{}, reject(ReasonInvalidRequest, "the mutation has no currency")
	}

	req := journal.PrepareRequest{
		Intent:        intent,
		Kind:          plan.kind,
		AttemptID:     g.newID(),
		TargetOrderID: plan.targetOrderID,
	}
	if decision.ID == "" {
		return req, nil
	}

	req.AccountRef = decision.AccountRef
	req.DecisionID = decision.ID
	req.SafetyClass = decision.SafetyClass
	req.Generation = decision.Generation
	req.ClientOrderID = decision.ClientOrderID
	// The bytes are rendered under the key the decision authorises and stored
	// before anything is sent, because that is the only moment at which "what
	// will be sent" and "what may be re-sent" are the same thing (design D2).
	if plan.wireBody != nil && decision.ClientOrderID != "" {
		body, err := plan.wireBody(decision.ClientOrderID)
		if err != nil {
			return journal.PrepareRequest{}, reject(ReasonUnsupportedOrderType,
				"the mutation cannot be rendered as a request body: %v", err)
		}
		req.WireBody = body
		req.SerializerVersion = WireSerializerVersion
	}
	return req, nil
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
