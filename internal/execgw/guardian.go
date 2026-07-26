package execgw

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// Guardian is the risk authority that authorises a mutation before the gateway
// will submit it.
//
// DRAFT INTERFACE — the authority that *issues* decisions (the sizing rules, the
// limit sources, the kill switch) belongs to a later change. What is fixed here
// is the contract an issuer has to satisfy: persist the decision first
// (journal.RecordDecision), then hand the gateway a reference to it. Returning a
// decision that was never written is not a usable answer — the gateway reads the
// row, not the value.
type Guardian interface {
	// Authorize records a decision for the described mutation and returns the
	// reference, or an error when the mutation must not happen.
	Authorize(ctx context.Context, req AuthorizationRequest) (GuardianDecision, error)
}

// AuthorizationRequest describes a mutation to a Guardian. DRAFT (see Guardian).
type AuthorizationRequest struct {
	Kind       journal.MutationKind
	AccountRef string
	Market     string
	Symbol     string
	Side       string
	Quantity   float64
	Price      float64
	Currency   string
}

// Limits is the risk snapshot taken when the decision was made.
//
// It is a snapshot, not a live lookup, on purpose: the gateway re-verifies the
// mutation against the numbers that were true when the decision was taken, so a
// limit that changed in between cannot silently widen an in-flight
// authorisation. It travels as JSON in the decision row's limits_json, which is
// why the fields carry tags.
type Limits struct {
	// MaxQuantity is the largest quantity this decision authorises.
	MaxQuantity float64 `json:"max_quantity"`
	// MaxNotional is the largest quantity×price this decision authorises, in
	// Currency.
	MaxNotional float64 `json:"max_notional"`
	// Currency is the currency MaxNotional is expressed in.
	Currency string `json:"currency"`
}

// IsZero reports whether the snapshot carries no usable limit at all.
func (l Limits) IsZero() bool { return l.MaxQuantity <= 0 && l.MaxNotional <= 0 }

// EncodeLimits renders a snapshot for the decision row.
func EncodeLimits(l Limits) (string, error) {
	encoded, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// DecodeLimits reads a snapshot back from the decision row. Unknown fields are
// refused: a limits blob this build does not fully understand is not a snapshot
// it may measure an order against.
func DecodeLimits(s string) (Limits, error) {
	var l Limits
	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&l); err != nil {
		return Limits{}, err
	}
	return l, nil
}

// GuardianDecision is a reference to a decision that is already in the journal.
//
// It deliberately carries no risk data. The gateway re-verifies against the
// persisted row — preimage, class, key, expiry, nonce — and a submitter that
// could hand over those values could also hand over matching forgeries
// (engine-safety: "제출 호출자가 공급한 위험 데이터로 재검증해서는 안 된다 —
// 검증이 순환한다"). The reference is the whole authorisation surface.
type GuardianDecision struct {
	// ID is the decisions row id the issuer minted and persisted.
	ID string
	// Generation is the reissue counter the reference expects the row to carry.
	// Always 0 in this change; a row that disagrees is refused rather than
	// followed.
	Generation int
}

// IsZero reports whether the reference names nothing.
func (d GuardianDecision) IsZero() bool { return strings.TrimSpace(d.ID) == "" }

// ErrNonceReused means a one-shot nonce was presented twice.
var ErrNonceReused = errors.New("execgw: guardian nonce has already been used")

// NonceStore spends one-shot nonces. Implementations must be safe for concurrent
// use and must make Consume atomic: two goroutines presenting the same nonce is
// exactly the race that would duplicate a live order.
type NonceStore interface {
	// Consume marks the nonce as spent, returning ErrNonceReused if it already was.
	Consume(nonce string, now time.Time) error
}

// NewMemoryNonceStore returns the in-memory store.
//
// It is not durable, and durability is required (engine-safety: "one-shot nonce
// 저장소는 journal 기반이어야 한다"). It survives here as the default for the
// narrower gateway tests until the journal-backed store folds consumption into
// the dispatch transaction (task 1.7).
func NewMemoryNonceStore() NonceStore {
	return &memoryNonceStore{seen: make(map[string]time.Time)}
}

type memoryNonceStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func (s *memoryNonceStore) Consume(nonce string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, used := s.seen[nonce]; used {
		return ErrNonceReused
	}
	s.seen[nonce] = now
	return nil
}

// loadDecision reads the decision a submission refers to.
//
// A reference to nothing and a reference to a row that does not exist are the
// same refusal on purpose: both mean no issuer ever committed to this mutation,
// and the difference between them is not something the caller should be able to
// probe for.
func (g *Gateway) loadDecision(ctx context.Context, ref GuardianDecision) (journal.Decision, *RejectedError) {
	if ref.IsZero() {
		return journal.Decision{}, reject(ReasonGuardianMissing,
			"no GuardianDecision was supplied")
	}
	dec, err := g.journal.LookupDecision(ctx, strings.TrimSpace(ref.ID))
	if errors.Is(err, journal.ErrDecisionNotFound) {
		return journal.Decision{}, reject(ReasonGuardianMissing,
			"decision %s was never persisted; there is nothing that authorises this mutation",
			short(ref.ID))
	}
	if err != nil {
		return journal.Decision{}, reject(ReasonGuardianMissing,
			"the decision could not be read: %v", err)
	}
	return dec, nil
}

// checkDecision re-derives the authorisation from the persisted row and measures
// the mutation against it.
//
// It is called twice per mutation — once before the dispatch is recorded and
// once, on a freshly re-read row, immediately before the broker call. Only the
// second one can honestly be called "직전 재검증": an expiry checked a
// transaction earlier is an expiry that has not been checked.
//
// The order of the checks is deliberate. Integrity first (is this row internally
// consistent), then identity (is it about this account, this order), then
// authority (does the class match what the order actually does), then time, then
// the key, then the size. Reporting "expired" for a decision that was also about
// a different symbol would send an operator after the wrong problem.
func (g *Gateway) checkDecision(dec journal.Decision, ref GuardianDecision,
	plan mutationPlan, now time.Time,
) *RejectedError {
	// --- integrity ----------------------------------------------------------
	if journal.HashPreimage(dec.RiskPreimage) != dec.RiskHash {
		return reject(ReasonGuardianDecisionTampered,
			"decision %s: the stored risk hash does not match the stored preimage",
			short(dec.ID))
	}
	preimage, err := journal.ParsePreimage(dec.PreimageKind, dec.RiskPreimage)
	if err != nil {
		return reject(ReasonGuardianDecisionTampered,
			"decision %s: the stored preimage is not readable as %s: %v",
			short(dec.ID), dec.PreimageKind, err)
	}

	// --- identity -----------------------------------------------------------
	switch {
	case dec.Generation != ref.Generation:
		return reject(ReasonGuardianDecisionTampered,
			"decision %s is generation %d, the submission expects %d",
			short(dec.ID), dec.Generation, ref.Generation)
	case dec.AccountRef != g.accountRef:
		return reject(ReasonGuardianIntentMismatch,
			"decision %s authorises account %s, this gateway trades %s",
			short(dec.ID), dec.AccountRef, g.accountRef)
	}

	// --- class ⇔ shape ------------------------------------------------------
	//
	// The class is a caller-visible label, so it is worth exactly nothing on its
	// own. The gateway computes whether the mutation raises exposure from the
	// mutation itself and requires the two to agree, which is what stops a BUY
	// wearing a RISK_REDUCING badge to skip the limits (engine-safety: "class
	// 선언은 그것만으로 효력이 없다").
	switch dec.SafetyClass {
	case journal.SafetyClassExposureRaising:
		if !plan.raisesExposure {
			return reject(ReasonGuardianClassMismatch,
				"decision %s is EXPOSURE_RAISING but the %s of %s does not raise exposure",
				short(dec.ID), plan.kind, plan.symbol)
		}
	case journal.SafetyClassRiskReducing:
		if plan.raisesExposure {
			return reject(ReasonGuardianClassMismatch,
				"decision %s is RISK_REDUCING but the %s of %s raises exposure",
				short(dec.ID), plan.kind, plan.symbol)
		}
	default:
		return reject(ReasonGuardianClassMismatch,
			"decision %s carries safety class %s, which this build does not issue",
			short(dec.ID), dec.SafetyClass)
	}

	// --- the preimage against the order -------------------------------------
	if rejected := checkPreimage(dec, preimage, plan); rejected != nil {
		return rejected
	}

	// --- time ---------------------------------------------------------------
	if dec.ExpiresAt.IsZero() {
		return reject(ReasonGuardianMissing,
			"decision %s carries no expiry", short(dec.ID))
	}
	if !now.Before(dec.ExpiresAt) {
		return reject(ReasonGuardianExpired,
			"the decision expired at %s and it is now %s",
			dec.ExpiresAt.UTC().Format(time.RFC3339), now.UTC().Format(time.RFC3339))
	}

	// --- the idempotency key ------------------------------------------------
	if rejected := checkIdempotencyKey(dec, plan); rejected != nil {
		return rejected
	}

	// --- the limit snapshot -------------------------------------------------
	return checkLimits(dec, plan)
}

// checkPreimage compares the persisted risk data with the order about to be sent.
//
// This is the substance of "결정 → 제출 사이의 변조" being closed: the numbers the
// decision was justified by live in the journal, and the order has to still be
// the one they justify.
func checkPreimage(dec journal.Decision, preimage journal.Preimage, plan mutationPlan) *RejectedError {
	mismatch := func(field, want, got string) *RejectedError {
		return reject(ReasonGuardianIntentMismatch,
			"decision %s authorises %s %s but the mutation carries %s",
			short(dec.ID), field, want, got)
	}
	venue := func(account, market, symbol, side string) *RejectedError {
		switch {
		case account != dec.AccountRef:
			return mismatch("account", account, dec.AccountRef)
		case market != plan.market:
			return mismatch("market", market, plan.market)
		case symbol != plan.symbol:
			return mismatch("symbol", symbol, plan.symbol)
		case side != plan.side:
			return mismatch("side", side, plan.side)
		}
		return nil
	}

	switch p := preimage.(type) {
	case journal.RiskIntent:
		if dec.SafetyClass != journal.SafetyClassExposureRaising {
			return reject(ReasonGuardianDecisionTampered,
				"decision %s pairs %s with a risk intent", short(dec.ID), dec.SafetyClass)
		}
		if rejected := venue(p.AccountRef, strings.ToLower(p.Market),
			strings.ToUpper(p.Symbol), strings.ToUpper(p.Side)); rejected != nil {
			return rejected
		}
		// An entry is authorised for one size at one price. Not "at most": the
		// stop distance the size was derived from is only true at that entry.
		if want, got := normalized(p.Quantity), decimalString(plan.quantity); want != got {
			return mismatch("quantity", want, got)
		}
		if want, got := normalized(p.EntryPrice), priceString(plan.price); want != got {
			return mismatch("entry price", want, got)
		}

	case journal.ReductionIntent:
		if dec.SafetyClass != journal.SafetyClassRiskReducing {
			return reject(ReasonGuardianDecisionTampered,
				"decision %s pairs %s with a reduction intent", short(dec.ID), dec.SafetyClass)
		}
		if rejected := venue(p.AccountRef, strings.ToLower(p.Market),
			strings.ToUpper(p.Symbol), strings.ToUpper(p.Side)); rejected != nil {
			return rejected
		}
		// A ceiling, not an equality: an exit sized from a fresher sellable
		// quantity may be smaller than the decision and must not be refused for
		// it (§0.3). Larger is refused — that is the direction that oversells.
		ok, err := decimalAtMost(decimalString(plan.quantity), p.MaxQuantity)
		if err != nil {
			return reject(ReasonGuardianDecisionTampered,
				"decision %s: max quantity %q is not a decimal", short(dec.ID), p.MaxQuantity)
		}
		if !ok {
			return mismatch("at most quantity", normalized(p.MaxQuantity), decimalString(plan.quantity))
		}

	default:
		return reject(ReasonGuardianDecisionTampered,
			"decision %s carries a preimage kind this build does not verify", short(dec.ID))
	}
	return nil
}

// checkIdempotencyKey ties the key that will be sent to the decision that
// derived it.
//
// A place must carry one, because the replay procedure is built on the
// assumption that every place was sent under a key (design D2/D3). A cancel or
// an amend must not: the broker's cancel and modify endpoints take no
// clientOrderId (openapi — the field exists on OrderCreateRequest and on the
// conditional-order request only), so a key on one of those describes a request
// that cannot exist.
func checkIdempotencyKey(dec journal.Decision, plan mutationPlan) *RejectedError {
	want := journal.DeriveClientOrderID(dec.ID, dec.Generation)
	if plan.kind != journal.KindPlace {
		if dec.ClientOrderID != "" {
			return reject(ReasonGuardianKeyMismatch,
				"decision %s carries an idempotency key but a %s cannot send one",
				short(dec.ID), plan.kind)
		}
		return nil
	}
	switch {
	case dec.ClientOrderID == "":
		return reject(ReasonGuardianKeyMismatch,
			"decision %s authorises a place without an idempotency key; a place that cannot be "+
				"replayed is not submittable", short(dec.ID))
	case dec.ClientOrderID != want:
		return reject(ReasonGuardianKeyMismatch,
			"decision %s carries key %q, which is not the key derived from it",
			short(dec.ID), dec.ClientOrderID)
	case !journal.ValidClientOrderID(dec.ClientOrderID):
		return reject(ReasonGuardianKeyMismatch,
			"decision %s carries a key the broker would refuse", short(dec.ID))
	}
	return nil
}

// checkLimits measures the mutation against the snapshot the decision carries.
//
// The exemption is by verified class, not by mutation kind. That is the whole
// point of the class check above: "a cancel is exempt" was a literal on the
// mutation verb, and it could say nothing about a reduce-only sell — which is
// also an exit, is also not something a limit may refuse (§0.3), and is a place.
func checkLimits(dec journal.Decision, plan mutationPlan) *RejectedError {
	if dec.SafetyClass == journal.SafetyClassRiskReducing {
		// An exit carries no snapshot at all. One that does is a row nobody in
		// this build could have written.
		if dec.LimitsJSON != "" {
			return reject(ReasonGuardianDecisionTampered,
				"decision %s is RISK_REDUCING but carries a limit snapshot", short(dec.ID))
		}
		return nil
	}

	if strings.TrimSpace(dec.LimitsJSON) == "" {
		return reject(ReasonGuardianLimitsUnset,
			"the decision for %s carries no limit snapshot", plan.symbol)
	}
	limits, err := DecodeLimits(dec.LimitsJSON)
	if err != nil {
		return reject(ReasonGuardianLimitsUnset,
			"the limit snapshot on decision %s is unreadable: %v", short(dec.ID), err)
	}
	if limits.IsZero() {
		return reject(ReasonGuardianLimitsUnset,
			"the decision for %s carries an empty limit snapshot", plan.symbol)
	}
	if limits.MaxQuantity > 0 && plan.quantity > limits.MaxQuantity {
		return reject(ReasonGuardianLimitExceeded,
			"quantity %s exceeds the authorised maximum %s",
			decimalString(plan.quantity), decimalString(limits.MaxQuantity))
	}
	if limits.MaxNotional > 0 {
		if notional := plan.notional(); notional > limits.MaxNotional {
			return reject(ReasonGuardianLimitExceeded,
				"notional %s exceeds the authorised maximum %s",
				decimalString(notional), decimalString(limits.MaxNotional))
		}
	}
	if limits.Currency != "" && plan.currency != "" && limits.Currency != plan.currency {
		return reject(ReasonGuardianLimitExceeded,
			"the limit snapshot is in %s but the order is in %s", limits.Currency, plan.currency)
	}
	return nil
}

// notional is the exposure a mutation adds, in the order's currency. An
// amount-based (fractional) order carries the notional directly.
func (p mutationPlan) notional() float64 {
	if p.amount > 0 {
		return p.amount
	}
	if p.price <= 0 || math.IsNaN(p.price) {
		return 0
	}
	return p.quantity * p.price
}

// normalized is the canonical spelling of a stored decimal, so a comparison
// against a freshly formatted number is about the value and not the spelling.
func normalized(decimal string) string {
	if out, ok := journal.NormalizeDecimal(decimal); ok {
		return out
	}
	return decimal
}

// decimalAtMost reports whether value ≤ bound, exactly. big.Rat rather than
// float64 because the bound comes out of the journal as text and the journal
// stores decimals as text for exactly this reason.
func decimalAtMost(value, bound string) (bool, error) {
	v, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	if !ok {
		return false, errors.New("execgw: value is not a decimal")
	}
	b, ok := new(big.Rat).SetString(strings.TrimSpace(bound))
	if !ok {
		return false, errors.New("execgw: bound is not a decimal")
	}
	return v.Cmp(b) <= 0, nil
}

func short(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
