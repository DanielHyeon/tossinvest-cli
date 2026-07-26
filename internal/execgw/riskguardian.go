package execgw

// riskguardian.go is the Guardian: the authority that judges an entry with the
// decision chain (internal/risk) and, when the chain allows, issues the decision
// and takes its reservation in one journal transaction (task 4.1/4.2, design D1;
// risk-management "발급 절차와 예약의 권위" · "GuardianDecision 발급자").
//
// # The procedure, and why it is this order
//
//	chain ALLOW → RecordDecisionAndReserve (one transaction) → Gateway 제출
//
// The reservation's decision_id is a NOT NULL foreign key, so "reserve, then
// decide" cannot be expressed; "decide, then reserve" leaves a submittable
// decision on disk that nobody paid limit headroom for. The atomic API is what
// removes the third state, and this type is the only production caller of it.
//
// # Why the issuer lives in execgw and the chain does not
//
// internal/risk is a pure function over values: it imports no journal, no
// gateway and no clock, which is what lets the whole chain be tested row by row
// and what keeps the entry latch a *projection* of execgw's gate rather than a
// second copy of its rules (issues.md, task 2.4). Putting the issuer there would
// drag the journal in behind it. The dependency already points this way —
// execgw → journal, execgw → risk — so the issuer sits at the end of it, next to
// the decision contract it produces (GuardianDecision, Limits, DefaultDecisionTTL).
//
// # What the re-collection loop may and may not change
//
// The chain runs **once**, before the loop. A re-collection replaces the
// reservation snapshot — the account's open exposure and the ledger version —
// and nothing else (design D1: 재수집 시 체인은 재실행하지 않는다). The freshness
// of everything the chain read is bounded by the decision's own 60-second TTL,
// and mode and latch changes are caught by the EntryGate re-check the Gateway
// performs at submission. The loop's budget (10s) is under the TTL for exactly
// that reason, and a loop that outlives the decision surfaces as DECISION_EXPIRED
// rather than as a successful issuance of stale reasoning.
//
// # Broker-behaviour claims
//
// None. Every broker fact — the holding, the cash, the open exposure, the
// snapshot's as-of — arrives as a value the caller collected.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// evaluateChain is risk.Evaluate, reached through a variable.
//
// The indirection buys one thing and it is worth the variable: "the chain is
// evaluated exactly once per issuance, and never again inside the re-collection
// loop" (design D1) becomes a property a test can *count* rather than one a
// reviewer has to re-derive by reading the loop. Production never reassigns it —
// the only writer is export_test.go, which is not in the built binary.
var evaluateChain = risk.Evaluate

// IssueStage names which authority refused an issuance.
//
// The two vocabularies stay apart on purpose: a chain refusal is a judgement
// about the intent (the operator changes the intent or the policy), an issuance
// refusal is the ledger's answer about the account (the operator waits, or the
// account is full). Collapsing them into one code space would make
// "LIMIT_REACHED" and "MAX_ORDER_EXCEEDED" look like the same kind of fact.
type IssueStage string

const (
	// StageChain: the Guardian chain refused before anything was written.
	StageChain IssueStage = "chain"
	// StageIssuance: the chain allowed and the atomic issuance transaction
	// refused. Nothing was written either — the decision rolls back with the
	// reservation.
	StageIssuance IssueStage = "issuance"
)

// IssueRefusal is a refusal to issue, carrying the stable reason-code that goes
// into the record of why an entry did not happen.
//
// Reason is a risk.ReasonCode string at StageChain and one of journal's four
// issuance reasons (LIMIT_REACHED / SNAPSHOT_RECOLLECTION_EXHAUSTED /
// VERSION_CONFLICT / DECISION_EXPIRED) at StageIssuance. The underlying sentinel
// is wrapped, so errors.Is still answers questions about the cause.
type IssueRefusal struct {
	Stage  IssueStage
	Reason string
	Detail string
	cause  error
}

func (e *IssueRefusal) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("execgw: the %s refused the issuance: %s", e.Stage, e.Reason)
	}
	return fmt.Sprintf("execgw: the %s refused the issuance (%s): %s", e.Stage, e.Reason, e.Detail)
}

func (e *IssueRefusal) Unwrap() error { return e.cause }

// Issued is what one successful issuance produced.
type Issued struct {
	// Decision is the reference a submission carries. It is the whole
	// authorisation surface: the Gateway re-reads the row, never these values.
	Decision GuardianDecision
	// ExpiresAt is when the decision stops authorising anything.
	ExpiresAt time.Time
	// Reservations are the rows the same transaction inserted. Empty for a
	// RISK_REDUCING issuance, which takes none.
	Reservations []journal.Reservation
	// Version is the reservation ledger version after the commit, usable as the
	// next observed version without a re-read.
	Version int64
}

// ExposureSnapshot is the reservation side of one collection: what the broker
// says is already open, and the ledger version that was true before the
// collection started.
//
// It is deliberately narrow. The re-collection loop may refresh *this* and
// nothing else, so the type itself is what makes "the chain is not re-run"
// structural rather than remembered: there is no field here a chain input could
// arrive through.
type ExposureSnapshot struct {
	// AsOf is when the broker data was current. The transaction judges its age.
	AsOf time.Time
	// Version is journal.ReservationVersion read *before* the collection.
	Version int64
	// OpenExposure is the account's gross long exposure excluding this intent —
	// a magnitude, in the policy's currency.
	OpenExposure riskcalc.Money
}

// CollectExposure produces one reservation snapshot. Attempt counts from 1.
//
// It is the caller's function and it is where the broker round trip happens —
// outside the journal transaction, by construction, because this signature is
// the only way a snapshot reaches the issuance.
type CollectExposure func(ctx context.Context, attempt int) (ExposureSnapshot, error)

// EntryIssuance is one entry the Guardian judges.
type EntryIssuance struct {
	// Intent is the proposed entry, in the caller's own words. Side must be BUY:
	// TossOS is long-only and an entry is a buy.
	Intent risk.Intent
	// Account is the state the chain measures the intent against, collected by
	// the caller. It is read once — see the file header.
	Account risk.AccountState
	// Collect refreshes the reservation snapshot on every attempt, including the
	// first. Required.
	Collect CollectExposure
}

// ReductionIssuance is one exit the Guardian judges: a reduce-only sell.
//
// Cancels are not here. A cancel is not sized against a holding, so the chain
// has nothing to say about it, and the plain Issuer (issue.go) is what the
// flatten saga already uses for one. Adding a holdings check that a cancel would
// have to be exempted from would be a rule with an exception rather than a rule.
type ReductionIssuance struct {
	// Intent describes the exit. Side must be SELL and Quantity is the exact
	// ceiling the decision authorises.
	Intent risk.Intent
	// Account carries the holding the reduction is bounded by.
	Account risk.AccountState
	// Reason is why the exit is happening, for the operator reading it later.
	Reason string
}

// RiskGuardianOptions constructs a Guardian.
type RiskGuardianOptions struct {
	// Journal is where decisions and reservations are written. Required.
	Journal *journal.Journal
	// Clock stamps the evaluation instant, the issue time and the expiry. It has
	// to be the gateway's clock: two time sources make the expiry meaningless.
	Clock clock.Clock
	// AccountRef is the one account this Guardian authorises for. Required.
	AccountRef string
	// Policy is the configured limit set. It is the single source: the limit
	// snapshot stamped on every EXPOSURE_RAISING decision is derived from it, so
	// the numbers the chain sizes against and the numbers the gateway re-measures
	// against cannot be two different sets.
	Policy risk.Policy
	// Costs is the shared cost model. An unconfigured model prices every trade as
	// free, so it is refused at construction rather than one entry at a time.
	Costs costs.Model
	// PolicyVersion names the rules that produced the size. Required — a decision
	// that cannot name its rules cannot be audited against them.
	PolicyVersion string
	// TTL bounds the decisions. Zero uses DefaultDecisionTTL (60s).
	TTL time.Duration
	// Recollect bounds the snapshot re-collection loop. Zero uses the journal's
	// defaults (3 attempts, 10s), which are under the TTL by design.
	Recollect journal.RecollectPolicy
	// Announcer is told about an operating-mode transition this Guardian's
	// escalation caused. Optional.
	Announcer journal.ModeAnnouncer
	// NewID mints decision ids and nonces. Defaults to 128 bits of randomness.
	NewID func() string
}

// RiskGuardian is the risk authority. Its fields are unexported because two of
// them are only correct together: the limit snapshot is *derived* from the
// policy at construction, so there is no way to configure a Guardian whose
// stamped limits differ from the limits it sized against.
type RiskGuardian struct {
	journal       *journal.Journal
	clk           clock.Clock
	accountRef    string
	policy        risk.Policy
	costs         costs.Model
	limits        Limits
	limitsJSON    string
	policyVersion string
	ttl           time.Duration
	recollect     journal.RecollectPolicy
	announcer     journal.ModeAnnouncer
	newID         func() string
}

// NewRiskGuardian validates the configuration and derives the limit snapshot.
//
// Every refusal here is a refusal to exist. A Guardian with an unusable policy
// would refuse every entry at evaluation time with INPUT_UNAVAILABLE, which is
// the same outcome reported at the least useful moment.
func NewRiskGuardian(opts RiskGuardianOptions) (*RiskGuardian, error) {
	account := strings.TrimSpace(opts.AccountRef)
	switch {
	case opts.Journal == nil:
		return nil, errors.New("execgw: a Guardian needs a journal — a decision that is not on disk authorises nothing")
	case account == "":
		return nil, errors.New("execgw: a Guardian authorises one account; none was named")
	case strings.TrimSpace(opts.PolicyVersion) == "":
		return nil, errors.New("execgw: a Guardian needs a policy version — a decision that cannot name " +
			"the rules that produced it cannot be audited against them")
	}
	if err := opts.Policy.Validate(); err != nil {
		return nil, fmt.Errorf("execgw: the Guardian's policy is not one a chain can be evaluated against: %w", err)
	}
	if !opts.Costs.Configured() {
		return nil, errors.New("execgw: the Guardian has no cost model; an absent model prices every trade " +
			"as free, which understates the cash an entry needs and drops break-even onto the entry price")
	}
	limits, err := ExposureLimitsFor(opts.Policy)
	if err != nil {
		return nil, fmt.Errorf("execgw: deriving the Guardian's limit snapshot: %w", err)
	}
	encoded, err := EncodeLimits(limits)
	if err != nil {
		return nil, fmt.Errorf("execgw: encoding the Guardian's limit snapshot: %w", err)
	}

	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = DefaultDecisionTTL
	}
	newID := opts.NewID
	if newID == nil {
		newID = randomID
	}
	return &RiskGuardian{
		journal:       opts.Journal,
		clk:           clk,
		accountRef:    account,
		policy:        opts.Policy,
		costs:         opts.Costs,
		limits:        limits,
		limitsJSON:    encoded,
		policyVersion: strings.TrimSpace(opts.PolicyVersion),
		ttl:           ttl,
		recollect:     opts.Recollect,
		announcer:     opts.Announcer,
		newID:         newID,
	}, nil
}

// ExposureLimits reports the snapshot this Guardian stamps on EXPOSURE_RAISING
// decisions. It is what the startup interlock compares against the audited
// configuration (engine-safety clause 4; engine.ExposureLimiter).
func (g *RiskGuardian) ExposureLimits() Limits { return g.limits }

// AccountRef is the account this Guardian authorises for.
func (g *RiskGuardian) AccountRef() string { return g.accountRef }

// Authorize satisfies the Guardian interface and refuses both classes.
//
// AuthorizationRequest is the draft shape from before an issuer existed: it
// carries no stop price, so an entry authorised through it would be a decision
// with no place where the risk ends (No Stop = No Trade), and it carries no
// holding, so a reduction through it could not be told from a short. Neither is
// something to guess a default for, and issuing on a guess is exactly what a
// checkpoint exists not to do. The real entry points are IssueEntry and
// IssueReduction.
func (g *RiskGuardian) Authorize(context.Context, AuthorizationRequest) (GuardianDecision, error) {
	return GuardianDecision{}, errors.New(
		"execgw: this Guardian does not authorise through the draft AuthorizationRequest — " +
			"it carries neither a stop price nor a holding, so use IssueEntry (entries) or " +
			"IssueReduction (reduce-only exits)")
}

// IssueEntry judges an entry and, if the chain allows, issues it.
//
// On a refusal at either stage nothing is on disk: the chain writes nothing at
// all, and the issuance transaction rolls the decision back with the reservation
// it could not take (risk-management: 롤백으로 제출 가능한 결정이 존재하지 않는다).
func (g *RiskGuardian) IssueEntry(ctx context.Context, req EntryIssuance) (Issued, error) {
	if req.Collect == nil {
		return Issued{}, errors.New("execgw: issuing an entry needs a snapshot collector")
	}
	intent, err := g.scopedIntent(req.Intent)
	if err != nil {
		return Issued{}, err
	}
	if intent.Side != risk.SideBuy {
		return Issued{}, fmt.Errorf("execgw: IssueEntry authorises entries, which are buys; the intent is a %s",
			string(intent.Side))
	}

	in := risk.Input{
		Now:     g.clk.Now().UTC(),
		Intent:  intent,
		Account: req.Account,
		Policy:  g.policy,
		Costs:   g.costs,
	}
	// The chain, once. Everything after this point is the ledger's answer, not a
	// second judgement of the intent.
	if verdict := evaluateChain(in); !verdict.Allowed {
		return Issued{}, errors.Join(chainRefusal(verdict), g.escalateFor(ctx, verdict))
	}
	exposure, verdict := risk.EntryExposureValue(in)
	if !verdict.Allowed {
		return Issued{}, chainRefusal(verdict)
	}

	// One decision id and one nonce for the whole loop. A refused attempt rolls
	// its row back, so the next attempt is not competing with the corpse of the
	// last one and the reference this call returns is stable across retries.
	decision := journal.DecisionRequest{
		ID:          g.newID(),
		AccountRef:  g.accountRef,
		Generation:  0,
		SafetyClass: journal.SafetyClassExposureRaising,
		Kind:        journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef:    g.accountRef,
			Market:        string(intent.Market),
			Symbol:        intent.Symbol,
			Side:          string(intent.Side),
			Quantity:      intent.Quantity,
			EntryPrice:    intent.LimitPrice,
			StopPrice:     intent.StopPrice,
			TargetPrice:   intent.TargetPrice,
			PolicyVersion: g.policyVersion,
		},
		LimitsJSON: g.limitsJSON,
		Nonce:      g.newID(),
		IssuedAt:   in.Now,
		ExpiresAt:  in.Now.Add(g.ttl),
	}
	reservationID := "res-" + decision.ID
	currency := exposure.Currency

	out, err := g.journal.RecordDecisionAndReserveWithRecollection(ctx,
		func(ctx context.Context, attempt int) (journal.IssueRequest, error) {
			snapshot, err := req.Collect(ctx, attempt)
			if err != nil {
				return journal.IssueRequest{}, err
			}
			usage, err := exposureUsage(snapshot.OpenExposure, currency)
			if err != nil {
				return journal.IssueRequest{}, err
			}
			return journal.IssueRequest{
				Decision: decision,
				Reserve: journal.ReserveRequest{
					SnapshotAsOf:    snapshot.AsOf,
					ObservedVersion: snapshot.Version,
					SnapshotUsage: []journal.AggregateAmount{
						{Kind: journal.ReservationKindOpenExposure, Amount: usage, Currency: currency},
					},
					Limits: []journal.AggregateAmount{
						{
							Kind:     journal.ReservationKindOpenExposure,
							Amount:   g.policy.MaxOpenExposure.Amount,
							Currency: currency,
						},
					},
					Reservations: []journal.ReservationRequest{
						{
							ID:       reservationID,
							Kind:     journal.ReservationKindOpenExposure,
							Amount:   exposure.Amount,
							Currency: currency,
						},
					},
				},
			}, nil
		}, g.recollect)
	if err != nil {
		return Issued{}, issuanceRefusal(err)
	}
	return Issued{
		Decision:     GuardianDecision{ID: out.Decision.ID, Generation: out.Decision.Generation},
		ExpiresAt:    out.Decision.ExpiresAt,
		Reservations: out.Reservations,
		Version:      out.Version,
	}, nil
}

// IssueReduction judges a reduce-only sell and issues it.
//
// No limits and no reservation, and both omissions are rules: an exit carries no
// limit snapshot because a limit that could refuse a liquidation is a trap
// (§0.3), and it takes no reservation because it lowers the aggregate it would
// otherwise be reserving against. The chain's reduction branch checks the one
// thing an exit can get wrong — selling more than the account holds.
func (g *RiskGuardian) IssueReduction(ctx context.Context, req ReductionIssuance) (Issued, error) {
	intent, err := g.scopedIntent(req.Intent)
	if err != nil {
		return Issued{}, err
	}
	if intent.Side != risk.SideSell {
		return Issued{}, fmt.Errorf("execgw: IssueReduction authorises reduce-only exits, which are sells; "+
			"the intent is a %s", string(intent.Side))
	}

	in := risk.Input{
		Now:     g.clk.Now().UTC(),
		Intent:  intent,
		Account: req.Account,
		Policy:  g.policy,
		Costs:   g.costs,
	}
	if verdict := evaluateChain(in); !verdict.Allowed {
		return Issued{}, chainRefusal(verdict)
	}

	now := in.Now
	dec, err := g.journal.RecordDecision(ctx, journal.DecisionRequest{
		ID:          g.newID(),
		AccountRef:  g.accountRef,
		Generation:  0,
		SafetyClass: journal.SafetyClassRiskReducing,
		Kind:        journal.KindPlace,
		Preimage: journal.ReductionIntent{
			AccountRef:  g.accountRef,
			Market:      string(intent.Market),
			Symbol:      intent.Symbol,
			Side:        string(intent.Side),
			MaxQuantity: intent.Quantity,
			Reason:      req.Reason,
		},
		Nonce:     g.newID(),
		IssuedAt:  now,
		ExpiresAt: now.Add(g.ttl),
	})
	if err != nil {
		return Issued{}, err
	}
	return Issued{
		Decision:  GuardianDecision{ID: dec.ID, Generation: dec.Generation},
		ExpiresAt: dec.ExpiresAt,
	}, nil
}

// escalateFor persists the automatic operating-mode tightening a chain refusal
// triggers, when it is one of the enumerated ones.
//
// # Why the issuer is the daily-loss producer
//
// The trigger is "일일 손실 한도 도달", and the place that judges whether the
// limit is reached is the chain's daily-loss rung — which runs here, with the
// account's realised loss as an input. The issuer does not compute a P&L and
// must not: the value arrives on AccountState.DailyRealizedLoss from
// riskcalc.DailyLoss, and the loop that keeps it current is task 8.1/7.x's. What
// this seam supplies is the transition; what those tasks supply is a caller that
// passes a realised loss derived from `trade_outcomes` rather than a zero.
//
// Until then this fires exactly when a caller hands over a loss at or past the
// limit — which is the correct behaviour on the day the input exists, and no
// behaviour at all before it. The equity-at-or-below-zero branch reports the
// same reason and escalates for the same purpose: both mean the account has no
// loss budget left to measure against.
//
// The enumeration is the journal's, and it is closed: a refusal that is not a
// trigger returns before reaching it, which is how "분석·성과 작업 실패는
// 트리거가 아니다" stays true without anybody remembering it here.
//
// The error is joined onto the refusal rather than replacing it. The entry is
// refused either way — that part is fail-closed and already decided — but a
// tightening that did not persist is a block a restart lifts, and it must not
// disappear because the thing it accompanies was already an error.
func (g *RiskGuardian) escalateFor(ctx context.Context, verdict risk.Decision) error {
	if verdict.Reason != risk.ReasonDailyLossLimitReached {
		return nil
	}
	if _, _, err := g.journal.EscalateOperatingMode(ctx, g.accountRef,
		journal.ModeTriggerDailyLossLimit, g.announcer); err != nil {
		return fmt.Errorf("execgw: the daily-loss limit did not reach the operating mode, "+
			"so a restart would lift the block: %w", err)
	}
	return nil
}

// scopedIntent binds the intent to this Guardian's account.
//
// An empty account reference is filled; a different one is refused rather than
// overwritten. A caller who names another account has made a mistake worth
// stopping on — silently re-pointing the intent would issue a decision for an
// account nobody asked about.
func (g *RiskGuardian) scopedIntent(intent risk.Intent) (risk.Intent, error) {
	named := strings.TrimSpace(intent.AccountRef)
	if named == "" {
		intent.AccountRef = g.accountRef
		return intent, nil
	}
	if named != g.accountRef {
		return risk.Intent{}, fmt.Errorf("execgw: the intent names account %q and this Guardian authorises %q",
			named, g.accountRef)
	}
	intent.AccountRef = named
	return intent, nil
}

// chainRefusal wraps a chain verdict as an issuance refusal.
func chainRefusal(verdict risk.Decision) *IssueRefusal {
	return &IssueRefusal{Stage: StageChain, Reason: string(verdict.Reason), Detail: verdict.Detail}
}

// issuanceRefusal classifies an atomic-issuance failure.
//
// journal.IssueRefusalReason is the single mapping from the sentinels to the
// four stable reason-codes, so two call sites cannot classify the same failure
// differently. Anything it does not recognise is not a risk refusal — a
// malformed request or a database error is a bug or an outage, and dressing one
// up as LIMIT_REACHED would hide it.
func issuanceRefusal(err error) error {
	reason, ok := journal.IssueRefusalReason(err)
	if !ok {
		return err
	}
	return &IssueRefusal{Stage: StageIssuance, Reason: reason, Detail: err.Error(), cause: err}
}

// exposureUsage validates the collected open exposure and returns its amount.
//
// The negative check is the one input error here that opens the gate instead of
// closing it: riskcalc.GrossLongExposure is a magnitude, and a caller who passed
// a signed valuation instead would hand "−1,000,000" to a comparison that would
// read it as an account with more headroom than it has. It is refused by name,
// exactly as the chain refuses it (internal/risk magnitudeIn).
func exposureUsage(m riskcalc.Money, want string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(m.Currency))
	if currency == "" {
		return "", errors.New("execgw: the collected open exposure carries no currency, " +
			"and an amount nobody named a currency for cannot be compared")
	}
	if currency != want {
		return "", fmt.Errorf("execgw: the collected open exposure is in %s but the policy limits are in %s; "+
			"normalise before comparing, never inside a comparison", currency, want)
	}
	amount := strings.TrimSpace(m.Amount)
	sign, err := riskcalc.CompareDecimal(amount, "0")
	if err != nil {
		return "", fmt.Errorf("execgw: the collected open exposure %q is not a decimal: %w", m.Amount, err)
	}
	if sign < 0 {
		return "", fmt.Errorf("execgw: the collected open exposure is %s %s; this aggregate is a magnitude, "+
			"and a negative one would be read as unused headroom", amount, currency)
	}
	return amount, nil
}

// ExposureLimitsFor renders a risk.Policy as the limit snapshot an
// EXPOSURE_RAISING decision carries.
//
// It is the join between the two vocabularies — the chain's exact decimal
// strings and the decision row's JSON floats — and it exists so the join happens
// once. The startup interlock compares the result field by field, Set bit
// included, against the limits it audited from the configuration (engine-safety
// clause 4), so a Guardian built from a policy and a gate built from the same
// numbers must produce byte-identical snapshots.
func ExposureLimitsFor(p risk.Policy) (Limits, error) {
	quantity, err := limitValue("maximum order quantity", p.MaxOrderQuantity)
	if err != nil {
		return Limits{}, err
	}
	notional, err := limitValue("maximum order notional", p.MaxOrderNotional.Amount)
	if err != nil {
		return Limits{}, err
	}
	exposure, err := limitValue("maximum open exposure", p.MaxOpenExposure.Amount)
	if err != nil {
		return Limits{}, err
	}
	dailyLoss, err := limitValue("maximum daily loss", p.MaxDailyLoss.Amount)
	if err != nil {
		return Limits{}, err
	}
	ratio, err := limitValue("maximum daily loss ratio", p.MaxDailyLossRatio)
	if err != nil {
		return Limits{}, err
	}
	limits := Limits{
		MaxQuantity:        Bound(quantity),
		MaxNotional:        Bound(notional),
		MaxTotalExposure:   Bound(exposure),
		MaxDailyLossAmount: Bound(dailyLoss),
		MaxDailyLossRatio:  Bound(ratio),
		Currency:           p.LimitCurrency(),
	}
	// The same predicate the gateway applies to a decision's snapshot. A policy
	// that cannot produce a valid snapshot is a policy no entry can be authorised
	// by, and finding that out at construction beats finding it out per order.
	if err := limits.Validate(); err != nil {
		return Limits{}, err
	}
	return limits, nil
}

// limitValue reads one configured ceiling as the float64 the decision row
// carries. Non-finite values are refused here rather than in Validate, so the
// message names the field that could not be read.
func limitValue(label, decimal string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(decimal), 64)
	if err != nil {
		return 0, fmt.Errorf("%s %q is not a number", label, decimal)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%s %q is not a finite number", label, decimal)
	}
	return v, nil
}
