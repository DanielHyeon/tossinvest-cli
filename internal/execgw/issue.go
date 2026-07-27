package execgw

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// issue.go is the issuer side of the decision contract for the callers that
// have no Guardian yet: the manual flatten saga, and the tests.
//
// It exists so "persist the decision, then hand the gateway a reference" is
// written once. The alternative — every caller assembling a
// journal.DecisionRequest — would spread the ordering that matters (write
// first, submit second) across every call site, and the ordering is the point:
// a crash between deciding and submitting must leave a decision on disk with
// nothing sent, never the reverse.
//
// The real Guardian (sizing rules, limit sources, kill switch) belongs to a
// later change. This type deliberately computes nothing: the caller supplies the
// numbers, and the value added is durability and the derived key, not judgement.

// DefaultDecisionTTL is how long an issued decision stays valid when the caller
// names no other bound. Short, because a decision is spent immediately after it
// is issued; long enough that a slow broker call cannot expire one mid-flight.
const DefaultDecisionTTL = 60 * time.Second

// Issuer records decisions in the journal.
type Issuer struct {
	// Journal is where decisions are persisted. Required.
	Journal *journal.Journal
	// Clock stamps issue and expiry. It must be the same clock the gateway
	// verifies against — two time sources would make the expiry meaningless.
	Clock clock.Clock
	// AccountRef is the account every decision is issued for. Required.
	AccountRef string
	// TTL bounds the decisions. Zero uses DefaultDecisionTTL.
	TTL time.Duration
	// NewID mints decision ids and nonces. Defaults to 128 bits of randomness.
	NewID func() string
}

// ReductionRequest describes an exit: a cancel, or a reduce-only sell.
type ReductionRequest struct {
	// Kind is the mutation the decision authorises. A PLACE gets an idempotency
	// key; a CANCEL or an AMEND does not.
	Kind journal.MutationKind
	// Market, Symbol and Side identify the order being reduced. Side is the side
	// of the mutation itself — cancelling a resting BUY is a BUY-side reduction.
	Market, Symbol, Side string
	// MaxQuantity is the ceiling the exit may not exceed. A smaller exit is
	// allowed: sizing from a fresher sellable-quantity read must never be
	// refused for being more conservative (§0.3).
	MaxQuantity float64
	// Reason is why the exit is happening, for the operator reading it later.
	Reason string
}

// EntryRequest describes new exposure.
type EntryRequest struct {
	// Kind is the mutation the decision authorises. Empty means PLACE, which is
	// what new exposure normally is; an amend that raises quantity or price also
	// adds exposure and is authorised the same way, without a key (the modify
	// endpoint takes none).
	Kind                 journal.MutationKind
	Market, Symbol, Side string
	// Quantity and EntryPrice are exact: an entry is authorised for one size at
	// one price, because the stop distance the size came from is only true there.
	Quantity, EntryPrice float64
	// StopPrice is required. A decision that does not say where the risk ends is
	// not a risk decision.
	StopPrice float64
	// TargetPrice is optional (0 means no take-profit).
	TargetPrice float64
	// PolicyVersion names the rules that produced the size.
	PolicyVersion string
	// Limits is the snapshot the gateway re-measures the order against.
	Limits Limits
}

// IssueReduction persists a RISK_REDUCING decision and returns the reference.
func (i *Issuer) IssueReduction(ctx context.Context, req ReductionRequest) (GuardianDecision, error) {
	if err := i.validate(); err != nil {
		return GuardianDecision{}, err
	}
	account := strings.TrimSpace(i.AccountRef)
	return i.record(ctx, journal.DecisionRequest{
		SafetyClass: journal.SafetyClassRiskReducing,
		Kind:        req.Kind,
		Preimage: journal.ReductionIntent{
			AccountRef:  account,
			Market:      req.Market,
			Symbol:      req.Symbol,
			Side:        req.Side,
			MaxQuantity: decimalString(req.MaxQuantity),
			Reason:      req.Reason,
		},
	})
}

// IssueEntry persists an EXPOSURE_RAISING decision and returns the reference.
func (i *Issuer) IssueEntry(ctx context.Context, req EntryRequest) (GuardianDecision, error) {
	if err := i.validate(); err != nil {
		return GuardianDecision{}, err
	}
	limits, err := EncodeLimits(req.Limits)
	if err != nil {
		return GuardianDecision{}, fmt.Errorf("execgw: encoding the limit snapshot: %w", err)
	}
	account := strings.TrimSpace(i.AccountRef)
	kind := req.Kind
	if kind == "" {
		kind = journal.KindPlace
	}
	return i.record(ctx, journal.DecisionRequest{
		SafetyClass: journal.SafetyClassExposureRaising,
		Kind:        kind,
		Preimage: journal.RiskIntent{
			AccountRef:    account,
			Market:        req.Market,
			Symbol:        req.Symbol,
			Side:          req.Side,
			Quantity:      decimalString(req.Quantity),
			EntryPrice:    priceString(req.EntryPrice),
			StopPrice:     priceString(req.StopPrice),
			TargetPrice:   priceString(req.TargetPrice),
			PolicyVersion: req.PolicyVersion,
		},
		LimitsJSON: limits,
	})
}

func (i *Issuer) record(ctx context.Context, req journal.DecisionRequest) (GuardianDecision, error) {
	now := i.clock().Now()
	ttl := i.TTL
	if ttl <= 0 {
		ttl = DefaultDecisionTTL
	}
	req.ID = i.newID()
	req.AccountRef = strings.TrimSpace(i.AccountRef)
	// Generation 0 is the only one this change issues; the mechanism that
	// advances it belongs to the change that introduces reissue.
	req.Generation = 0
	req.Nonce = i.newID()
	req.IssuedAt = now
	req.ExpiresAt = now.Add(ttl)

	dec, err := i.Journal.RecordDecision(ctx, req)
	if err != nil {
		return GuardianDecision{}, err
	}
	return GuardianDecision{ID: dec.ID, Generation: dec.Generation}, nil
}

func (i *Issuer) validate() error {
	switch {
	case i == nil || i.Journal == nil:
		return fmt.Errorf("execgw: an issuer needs a journal — a decision that is not on disk authorises nothing")
	case strings.TrimSpace(i.AccountRef) == "":
		return fmt.Errorf("execgw: an issuer needs an account reference")
	}
	return nil
}

func (i *Issuer) clock() clock.Clock {
	if i.Clock == nil {
		return clock.System()
	}
	return i.Clock
}

func (i *Issuer) newID() string {
	if i.NewID != nil {
		return i.NewID()
	}
	return randomID()
}
