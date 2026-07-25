// Package flatten is the emergency exit: cancel everything outstanding, then
// sell down every position, as a durable saga that survives a crash.
//
// # Why this is its own package
//
// It cannot live in internal/execgw, because it needs internal/reconcile (for
// the account snapshot the liquidation phase measures against) and reconcile
// already imports execgw. It should not live in cmd/, because a flatten has to be
// resumable by a process that is not the one that started it, and logic in a
// command is logic only that command has.
//
// # The one rule that outranks everything else here
//
// WORKFLOW §0.3: nothing may weaken or delay the exit. Concretely, in this
// package:
//
//   - No confirmation, wait, stabilisation or freshness check sits in front of a
//     cancel. The entry block is raised first because it is free and instant;
//     everything else that could delay is on the *entry* side of the account.
//   - No error path returns early with orders still live. A cancel that fails is
//     recorded and the saga keeps going to the next one, because seven orders
//     with one failure is a better place to be than seven orders with one
//     failure and six untried.
//   - The entry gate is latched, not consulted. This package never asks whether
//     it is allowed to reduce exposure.
//
// # And the rule that constrains it
//
// A cancel whose outcome is unknown means the order may still be live. Selling
// the full holding while a live sell order also exists is how an emergency exit
// becomes a short position — so the symbol is *held* out of the liquidation phase
// until its cancel is settled (engine-safety: "해당 심볼 청산은 취소 확정 시까지
// 보류되고 oversell이 방지된다"). Holding a symbol back from the *sell* is the one
// delay this package permits, and it exists to stop the exit overshooting.
package flatten

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// DefaultMaxPages bounds the open-order walk (cursor-loop defence).
const DefaultMaxPages = 50

// DecisionTTL is how long a Guardian decision the saga issues stays valid.
//
// Short, because a decision is spent immediately after it is issued; long enough
// that a slow broker call cannot expire one mid-flight.
const DecisionTTL = 60 * time.Second

// Saga runs the flatten sequence.
type Saga struct {
	// Journal holds the durable plan. Required.
	Journal *journal.Journal
	// Gateway is the only way this package mutates. Required — a flatten that
	// bypassed the gateway would bypass the journal and the IN_DOUBT rules.
	Gateway *execgw.Gateway
	// Gate is latched for the duration and afterwards. Optional but expected.
	Gate *execgw.EntryGate
	// Orders walks the broker's open-order list. Required.
	Orders execgw.OrderPager
	// Resolver settles IN_DOUBT cancels. Optional; without it an ambiguous
	// cancel parks the symbol rather than being resolved.
	Resolver *execgw.Resolver
	// Notifier raises the operator alerts. Optional.
	Notifier *obs.Notifier
	// Log is the structured log. Optional.
	Log *obs.Logger
	// Clock stamps the saga. Defaults to clock.System().
	Clock clock.Clock
	// AccountRef identifies the account being flattened. Required.
	AccountRef string
	// Operator names who asked for it, for the record.
	Operator string
	// Reason is why, for the record and the alert.
	Reason string
	// MaxPages bounds the order-list walk. Zero uses DefaultMaxPages.
	MaxPages int
	// DryRun reports what would happen and mutates nothing.
	DryRun bool
	// NewID generates the saga id. Defaults to random hex.
	NewID func() string

	// --- liquidation phase (task 4.5, liquidate.go) -------------------------

	// Positions sweeps the account's holdings. Required by Liquidate.
	Positions PositionReader
	// Balance reads buying power, for the account snapshot. Required by Liquidate.
	Balance reconcile.BuyingPowerReader
	// Sellable reports how much of a symbol can be sold now. Required by
	// Liquidate: sizing a sell from the holding alone can oversell.
	Sellable SellableReader
	// Prices supplies the aggressive limit price. Required by Liquidate.
	Prices PriceReader
	// Currencies are the balances the snapshot reads. Empty uses KRW and USD.
	Currencies []string
	// Rounds is how many sell-and-recheck cycles to run. Zero uses
	// DefaultLiquidationRounds.
	Rounds int
	// StabiliseInterval is the minimum gap between corroborating snapshots. Zero
	// uses reconcile.DefaultStabilisationInterval.
	StabiliseInterval time.Duration
	// StabiliseAttempts bounds the wait for a still account. Zero uses 4.
	StabiliseAttempts int
	// AggressiveDiscount is how far below the last trade a liquidation limit sits
	// when the broker reports no price band. Zero uses DefaultAggressiveDiscount.
	AggressiveDiscount float64
}

// CancelOutcome is what happened to one cancel target.
type CancelOutcome struct {
	OrderID string
	Symbol  string
	Market  string
	Side    string
	// Quantity is the target order's quantity, as the broker reported it.
	Quantity string
	State    string
	Reason   string
	Detail   string
}

// CancelReport is the result of the cancel-all phase.
type CancelReport struct {
	SagaID string
	// Resumed reports that this run continued an earlier saga rather than
	// starting one.
	Resumed bool
	DryRun  bool
	// Found is how many outstanding orders the walk saw.
	Found int
	// Cancelled, Failed, InDoubt and Held partition Found by outcome.
	Cancelled int
	Failed    int
	InDoubt   int
	Held      int
	Outcomes  []CancelOutcome
	// Phase is the saga's phase after this run.
	Phase string
}

// Settled reports whether every cancel reached a terminal state, which is the
// precondition for the liquidation phase.
func (r CancelReport) Settled() bool { return r.InDoubt == 0 && r.Held == 0 }

// ErrNotConfigured means a required dependency is missing.
var ErrNotConfigured = errors.New("flatten: the saga is not fully wired")

// CancelAll runs phase 1: block new entries, then cancel every outstanding order
// and settle each result.
//
// It returns a report rather than an error for per-order failures. An error means
// the saga could not proceed at all (the journal is unwritable, the order list is
// unreadable); a refused or ambiguous individual cancel is data, and stopping the
// loop on one would leave the remaining orders live.
func (s *Saga) CancelAll(ctx context.Context) (CancelReport, error) {
	if err := s.validate(); err != nil {
		return CancelReport{}, err
	}
	clk := s.clock()

	// 1. Block new entries. First, because it is instant and free, and because
	//    every later step is slower than a new order arriving.
	s.blockEntries()

	saga, resumed, err := s.openSaga(ctx)
	if err != nil {
		return CancelReport{}, err
	}
	report := CancelReport{SagaID: saga.ID, Resumed: resumed, DryRun: s.DryRun, Phase: saga.Phase}

	s.event(obs.EventFlattenStarted, saga.ID, map[string]any{
		"saga_id":       saga.ID,
		"resumed":       resumed,
		"dry_run":       s.DryRun,
		obs.FieldReason: s.Reason,
		obs.FieldDetail: "flatten-all: cancelling every outstanding order",
	})

	if err := s.Journal.SetFlattenPhase(ctx, saga.ID, journal.FlattenPhaseCancelling, ""); err != nil {
		return report, err
	}
	report.Phase = journal.FlattenPhaseCancelling

	// 2. Enumerate. The walk must complete: an order on page two that we never
	//    saw is an order this flatten leaves live.
	targets, err := s.enumerate(ctx)
	if err != nil {
		return report, err
	}
	report.Found = len(targets)

	// 3. Record every target before acting on any of them. A crash between the
	//    plan and the first cancel then resumes with the full plan rather than
	//    re-deriving it from a list that has since changed.
	steps := make([]journal.FlattenStep, 0, len(targets))
	for _, target := range targets {
		step, err := s.Journal.AddFlattenStep(ctx, journal.FlattenStep{
			SagaID:        saga.ID,
			Kind:          journal.FlattenStepCancel,
			Market:        target.Market,
			Symbol:        target.Symbol,
			TargetOrderID: target.OrderID,
			Side:          target.Side,
			Quantity:      decimalString(target.Quantity),
			Price:         decimalString(target.Price),
			Currency:      target.Currency,
		})
		if err != nil {
			return report, err
		}
		steps = append(steps, step)
	}

	// A resumed saga may carry steps this walk did not produce — an order that
	// has since been filled, for instance. They still have to be settled, or the
	// saga can never leave the cancelling phase.
	recorded, err := s.Journal.FlattenSteps(ctx, saga.ID)
	if err != nil {
		return report, err
	}
	steps = mergeSteps(steps, recorded)

	// 4. Cancel each, one at a time. Serial rather than concurrent: the gateway
	//    holds one in-flight mutation per symbol, and two cancels racing on the
	//    same symbol would refuse each other rather than go faster.
	for _, step := range steps {
		if step.Kind != journal.FlattenStepCancel {
			continue
		}
		outcome := s.settleStep(ctx, saga.ID, step, targetsByOrder(targets))
		report.Outcomes = append(report.Outcomes, outcome)
		switch outcome.State {
		case journal.FlattenStepDone:
			report.Cancelled++
		case journal.FlattenStepFailed:
			report.Failed++
		case journal.FlattenStepInDoubt:
			report.InDoubt++
		case journal.FlattenStepHeld:
			report.Held++
		}
	}
	if report.Found == 0 {
		report.Found = len(report.Outcomes)
	}

	// 5. Advance, or stop and say why.
	phase := journal.FlattenPhaseCancelled
	detail := fmt.Sprintf("cancelled %d, already gone %d, unresolved %d",
		report.Cancelled, report.Failed, report.InDoubt+report.Held)
	if s.DryRun {
		// A dry run must not move the saga forward: the next real run has to see
		// the same work still to do.
		phase = journal.FlattenPhaseBlocking
		detail = "dry run: nothing was submitted"
	}
	if err := s.Journal.SetFlattenPhase(ctx, saga.ID, phase, detail); err != nil {
		return report, err
	}
	report.Phase = phase

	if !report.Settled() && !s.DryRun {
		s.event(obs.EventFlattenStalled, saga.ID+"|cancel", map[string]any{
			"saga_id":       saga.ID,
			"in_doubt":      report.InDoubt,
			"held":          report.Held,
			obs.FieldDetail: "some cancels did not settle; those symbols are held out of liquidation",
		})
	}
	s.logf(obs.EventFlattenProgress,
		"saga_id", saga.ID,
		"found", report.Found,
		"cancelled", report.Cancelled,
		"failed", report.Failed,
		"in_doubt", report.InDoubt,
		"held", report.Held,
		"dry_run", s.DryRun)

	_ = clk
	return report, nil
}

// --- steps ------------------------------------------------------------------

// settleStep cancels one order and records what happened.
func (s *Saga) settleStep(ctx context.Context, sagaID string, step journal.FlattenStep,
	byOrder map[string]cancelTarget,
) CancelOutcome {
	outcome := CancelOutcome{
		OrderID:  step.TargetOrderID,
		Symbol:   step.Symbol,
		Market:   step.Market,
		Side:     step.Side,
		Quantity: step.Quantity,
		State:    step.State,
	}

	// Already settled by an earlier run. Resuming must not re-submit: the
	// official API has no idempotency key, so a second cancel of an order that
	// was replaced is a cancel of somebody else's order number.
	switch step.State {
	case journal.FlattenStepDone, journal.FlattenStepFailed:
		outcome.Detail = "already settled by an earlier run"
		return outcome

	case journal.FlattenStepInDoubt:
		// An earlier run left this cancel ambiguous. Re-submitting is the one
		// thing that must not happen: the first cancel may well have worked, and
		// a second one aimed at the same order number after a replacement would
		// cancel a different order. So the only move available is to look again.
		outcome.State, outcome.Reason, outcome.Detail = s.resolveStep(ctx, step)
		if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State,
			step.IntentID, step.AttemptID, outcome.Reason, outcome.Detail); err != nil {
			outcome.Detail = "the outcome could not be recorded: " + err.Error()
		}
		return outcome
	}

	if s.DryRun {
		outcome.State = journal.FlattenStepPending
		outcome.Detail = "dry run: would cancel"
		return outcome
	}

	intent := orderintent.CancelIntent{OrderID: step.TargetOrderID, Symbol: step.Symbol}
	ref := execgw.OrderRef{
		Market:   step.Market,
		Side:     step.Side,
		Quantity: parseDecimal(step.Quantity),
		Price:    parseDecimal(step.Price),
		Currency: step.Currency,
	}
	if target, ok := byOrder[step.TargetOrderID]; ok {
		ref = execgw.OrderRef{
			Market: target.Market, Side: target.Side,
			Quantity: target.Quantity, Price: target.Price, Currency: target.Currency,
		}
	}

	out, err := s.Gateway.Cancel(ctx, execgw.CancelRequest{
		Intent:   intent,
		Order:    ref,
		Decision: s.decisionFor(execgw.CancelHash(intent)),
	})

	state, detail := classifyCancel(out, err)
	outcome.State = state
	outcome.Reason = string(out.Reason)
	outcome.Detail = detail

	// An ambiguous cancel is the case the resolver exists for. Resolving it is
	// observation only — no re-submission, ever.
	if state == journal.FlattenStepInDoubt {
		step.AttemptID = out.AttemptID
		outcome.State, outcome.Reason, outcome.Detail = s.resolveStep(ctx, step)
	}

	if outcome.State == journal.FlattenStepInDoubt {
		// The symbol is now dangerous to liquidate. Say so where the liquidation
		// phase will look, and where an operator will.
		s.event(obs.EventOrderInDoubt, sagaID+"|"+step.Symbol, map[string]any{
			obs.FieldSymbol:  step.Symbol,
			obs.FieldOrderID: step.TargetOrderID,
			"saga_id":        sagaID,
			obs.FieldDetail:  "a flatten cancel is unresolved; this symbol is held out of liquidation",
		})
	}

	if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State,
		out.IntentID, out.AttemptID, outcome.Reason, outcome.Detail); err != nil {
		outcome.Detail = "the outcome could not be recorded: " + err.Error()
	}
	return outcome
}

// resolveStep settles an ambiguous cancel by observation.
//
// There is no branch here that submits anything. The IN_DOUBT resolver's whole
// contract is that it can find an order, prove one is absent, or say it cannot
// tell — and the third answer keeps the symbol held rather than guessing.
func (s *Saga) resolveStep(ctx context.Context, step journal.FlattenStep) (state string, reason string, detail string) {
	if s.Resolver == nil || step.AttemptID == "" {
		return journal.FlattenStepInDoubt, string(execgw.ReasonBrokerOutcomeUnknown),
			"the cancel outcome is unresolved and no resolver is wired; this symbol stays held"
	}
	res, err := s.Resolver.Resolve(ctx, step.AttemptID)
	if err != nil {
		return journal.FlattenStepInDoubt, string(execgw.ReasonBrokerOutcomeUnknown),
			"resolution did not settle the cancel: " + err.Error()
	}
	switch res.State {
	case journal.StateConfirmed:
		return journal.FlattenStepDone, string(res.Reason),
			"resolved by observation: the cancel took effect"
	case journal.StateFailedConfirmed:
		// Proven absent: the cancel never happened. The order may still be live,
		// so this is not "done" — but it *is* settled, and the next run may
		// legitimately submit a fresh cancel for it.
		return journal.FlattenStepPending, string(res.Reason),
			"resolved by observation: the cancel never reached the broker; it will be retried"
	default:
		return journal.FlattenStepInDoubt, string(res.Reason), res.Detail
	}
}

// classifyCancel maps a gateway outcome onto a step state.
//
// The mapping is deliberately generous about what counts as success: the saga's
// goal is that the order is not live, and an order the broker refuses to cancel
// because it is already filled or already cancelled satisfies that goal. Only a
// genuinely unknown outcome blocks.
func classifyCancel(out execgw.Outcome, err error) (string, string) {
	switch out.State {
	case journal.StateConfirmed:
		return journal.FlattenStepDone, "cancel confirmed"
	case journal.StateFailedConfirmed, journal.StateNotDispatched:
		detail := out.Detail
		if detail == "" && err != nil {
			detail = err.Error()
		}
		return journal.FlattenStepFailed, detail
	case journal.StateInDoubt, journal.StateUnresolvedInDoubt:
		return journal.FlattenStepInDoubt, out.Detail
	}
	if err != nil {
		// No journal state at all: the gateway refused before recording, or the
		// journal itself failed. Nothing was sent, but we cannot say the order is
		// gone, so the symbol stays held.
		return journal.FlattenStepHeld, err.Error()
	}
	return journal.FlattenStepHeld, "the cancel produced no settled state"
}

// --- enumeration ------------------------------------------------------------

// cancelTarget is one outstanding order to cancel.
type cancelTarget struct {
	OrderID  string
	Symbol   string
	Market   string
	Side     string
	Quantity float64
	Price    float64
	Currency string
}

// enumerate walks the open-order list to its last page.
func (s *Saga) enumerate(ctx context.Context) ([]cancelTarget, error) {
	maxPages := s.MaxPages
	if maxPages <= 0 {
		maxPages = DefaultMaxPages
	}
	raws, err := execgw.ScanOrders(ctx, s.Orders, execgw.OrderQuery{Status: "OPEN"}, maxPages)
	if err != nil {
		return nil, fmt.Errorf("flatten: walking the open-order list: %w", err)
	}

	targets := make([]cancelTarget, 0, len(raws))
	for _, raw := range raws {
		target, perr := parseTarget(raw)
		if perr != nil {
			// A payload we cannot read is an order we cannot cancel by id. That
			// is a stop condition for the whole saga, not a skip: continuing
			// would report "everything cancelled" while an order we never
			// understood is still live.
			return nil, fmt.Errorf("flatten: reading an open order: %w", perr)
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// openOrderPayload is the subset of the broker's order the saga needs. Decimals
// arrive as strings and are parsed once, here.
type openOrderPayload struct {
	OrderID  string  `json:"orderId"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity *string `json:"quantity"`
	Price    *string `json:"price"`
	Currency string  `json:"currency"`
}

func parseTarget(raw json.RawMessage) (cancelTarget, error) {
	var payload openOrderPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return cancelTarget{}, err
	}
	if strings.TrimSpace(payload.OrderID) == "" {
		return cancelTarget{}, errors.New("the order has no id, so it cannot be cancelled")
	}
	quantity, err := parseOptional(payload.Quantity)
	if err != nil {
		return cancelTarget{}, fmt.Errorf("quantity: %w", err)
	}
	price, err := parseOptional(payload.Price)
	if err != nil {
		return cancelTarget{}, fmt.Errorf("price: %w", err)
	}
	symbol := strings.ToUpper(strings.TrimSpace(payload.Symbol))
	return cancelTarget{
		OrderID:  strings.TrimSpace(payload.OrderID),
		Symbol:   symbol,
		Market:   MarketOf(symbol, payload.Currency),
		Side:     strings.ToUpper(strings.TrimSpace(payload.Side)),
		Quantity: quantity,
		Price:    price,
		Currency: strings.ToUpper(strings.TrimSpace(payload.Currency)),
	}, nil
}

// MarketOf infers the market from what the broker's order payload actually
// carries.
//
// The payload has no market field. Currency is the stronger signal — a KRW order
// is a KR order — and the six-digit symbol pattern is the fallback for a payload
// that omits currency too. Getting this wrong would produce a cancel the gateway
// journals against the wrong trading calendar, which is why it is a named
// function with its own test rather than an expression inline.
func MarketOf(symbol, currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "KRW":
		return "kr"
	case "USD":
		return "us"
	}
	trimmed := strings.TrimSpace(symbol)
	if len(trimmed) == 6 && isAllDigits(trimmed) {
		return "kr"
	}
	return "us"
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// --- saga plumbing ----------------------------------------------------------

func (s *Saga) validate() error {
	switch {
	case s.Journal == nil:
		return fmt.Errorf("%w: a journal is required — an unrecorded flatten cannot be resumed", ErrNotConfigured)
	case s.Gateway == nil && !s.DryRun:
		return fmt.Errorf("%w: a gateway is required to cancel anything", ErrNotConfigured)
	case s.Orders == nil:
		return fmt.Errorf("%w: an order pager is required to find what to cancel", ErrNotConfigured)
	case strings.TrimSpace(s.AccountRef) == "":
		return fmt.Errorf("%w: an account reference is required", ErrNotConfigured)
	}
	return nil
}

// openSaga starts a saga or picks up the unfinished one.
func (s *Saga) openSaga(ctx context.Context) (journal.FlattenSaga, bool, error) {
	existing, err := s.Journal.ActiveFlatten(ctx)
	switch {
	case err == nil:
		return existing, true, nil
	case !errors.Is(err, journal.ErrFlattenNotFound):
		return journal.FlattenSaga{}, false, err
	}

	newID := s.NewID
	if newID == nil {
		newID = randomID
	}
	saga, err := s.Journal.StartFlatten(ctx, journal.FlattenSaga{
		ID:         newID(),
		AccountRef: s.AccountRef,
		Phase:      journal.FlattenPhaseBlocking,
		Reason:     s.Reason,
		Operator:   s.Operator,
		DryRun:     s.DryRun,
	})
	if err != nil {
		return journal.FlattenSaga{}, false, err
	}
	return saga, false, nil
}

// blockEntries latches the gate.
//
// A dry run latches it too. That looks surprising and is deliberate: somebody has
// just decided to find out what flattening this account would do, and letting the
// engine open a new position between that decision and the real run would make
// the dry run's answer wrong.
func (s *Saga) blockEntries() {
	if s.Gate == nil {
		return
	}
	detail := "a flatten-all is in progress; new entries stay blocked until an operator clears this"
	if s.Reason != "" {
		detail = detail + " (" + s.Reason + ")"
	}
	s.Gate.Block(execgw.ReasonFlattenInProgress, detail)
}

// decisionFor issues the Guardian decision for one cancel.
//
// A cancel carries no limit snapshot, and that is enforced rather than assumed:
// execgw.verifyLimits exempts KindCancel entirely, because a limit is not a
// reason to refuse an exit (§0.3). What the decision still provides is the
// binding — this nonce authorises this exact intent, once.
func (s *Saga) decisionFor(hash string) execgw.GuardianDecision {
	now := s.clock().Now()
	return execgw.GuardianDecision{
		Nonce:      "flatten-" + randomID(),
		IntentHash: hash,
		IssuedAt:   now,
		ExpiresAt:  now.Add(DecisionTTL),
		Authority:  "flatten-saga",
	}
}

func (s *Saga) clock() clock.Clock {
	if s.Clock == nil {
		return clock.System()
	}
	return s.Clock
}

// event raises an operator alert and logs it.
func (s *Saga) event(t obs.EventType, key string, fields map[string]any) {
	if s.Notifier == nil {
		s.logf(t, flatten(fields)...)
		return
	}
	_ = s.Notifier.Notify(context.Background(), obs.Event{
		Type:   t,
		Key:    key,
		Title:  string(t),
		Fields: fields,
	})
}

func (s *Saga) logf(t obs.EventType, args ...any) {
	if s.Log == nil {
		return
	}
	s.Log.Event(t, args...)
}

func flatten(fields map[string]any) []any {
	out := make([]any, 0, 2*len(fields))
	for k, v := range fields {
		out = append(out, k, v)
	}
	return out
}

// mergeSteps appends recorded steps that the fresh enumeration did not produce,
// preserving order and without duplicating.
func mergeSteps(fresh, recorded []journal.FlattenStep) []journal.FlattenStep {
	seen := make(map[int64]bool, len(fresh))
	for _, s := range fresh {
		seen[s.ID] = true
	}
	out := append([]journal.FlattenStep(nil), fresh...)
	for _, s := range recorded {
		if !seen[s.ID] {
			out = append(out, s)
		}
	}
	return out
}

func targetsByOrder(targets []cancelTarget) map[string]cancelTarget {
	out := make(map[string]cancelTarget, len(targets))
	for _, t := range targets {
		out[t.OrderID] = t
	}
	return out
}

func parseOptional(s *string) (float64, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(*s), 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a decimal", *s)
	}
	return v, nil
}

func parseDecimal(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func randomID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("flatten: crypto/rand is unavailable: " + err.Error())
	}
	return hex.EncodeToString(buf[:])
}
