package execgw

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// Guardian is the risk authority that authorises a mutation before the gateway
// will submit it.
//
// DRAFT INTERFACE — Phase 1 only declares the shape. The engine-safety spec fixes
// the four properties a decision must carry (intent hash, limit snapshot, expiry,
// one-shot nonce) and the gateway enforces all four today, but the authority that
// *issues* decisions — the sizing rules, the limit sources, the kill switch — is
// Phase 2 work ("Guardian 수치 구현(P2 — 인터페이스 초안만)", design Non-Goals).
// Expect this interface to change when Phase 2 lands; nothing in Phase 1 depends
// on its method set, because callers supply a decision directly.
type Guardian interface {
	// Authorize returns a decision for the described mutation, or an error when
	// the mutation must not happen.
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
	// IntentHash is the value a returned decision must carry back.
	IntentHash string
}

// Limits is the risk snapshot taken when the decision was made.
//
// It is a snapshot, not a live lookup, on purpose: the gateway re-verifies the
// mutation against the numbers that were true when the decision was taken, so a
// limit that changed in between cannot silently widen an in-flight authorisation.
type Limits struct {
	// MaxQuantity is the largest quantity this decision authorises.
	MaxQuantity float64
	// MaxNotional is the largest quantity×price this decision authorises, in
	// Currency.
	MaxNotional float64
	// Currency is the currency MaxNotional is expressed in.
	Currency string
}

// IsZero reports whether the snapshot carries no usable limit at all.
func (l Limits) IsZero() bool { return l.MaxQuantity <= 0 && l.MaxNotional <= 0 }

// GuardianDecision is a one-shot authorisation for exactly one mutation.
//
// DRAFT (see Guardian) — the fields are the spec's four requirements and are
// expected to survive Phase 2; the issuing side is not.
type GuardianDecision struct {
	// Nonce is spent the moment the gateway submits. Reusing it is refused.
	Nonce string
	// IntentHash binds the decision to one exact intent (see PlaceHash).
	IntentHash string
	// Limits is the risk snapshot the mutation is re-measured against.
	Limits Limits
	// IssuedAt / ExpiresAt bound the decision in time. An expired decision is
	// refused at the last possible moment, immediately before the broker call.
	IssuedAt  time.Time
	ExpiresAt time.Time
	// Authority names who issued it, for the audit trail.
	Authority string
}

// IntentHash hashes a canonical intent string. The canonical strings come from
// internal/orderintent, which is also what the CLI's confirm token is derived
// from — but this is a full SHA-256 rather than the short confirm token, because
// it is a binding identity and not something a human retypes.
func IntentHash(canonical string) string {
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// PlaceHash returns the hash a GuardianDecision must carry to authorise this place.
func PlaceHash(i orderintent.PlaceIntent) string {
	return IntentHash(orderintent.CanonicalPlace(i))
}

// CancelHash returns the hash a GuardianDecision must carry to authorise this cancel.
func CancelHash(i orderintent.CancelIntent) string {
	return IntentHash(orderintent.CanonicalCancel(i))
}

// AmendHash returns the hash a GuardianDecision must carry to authorise this amend.
func AmendHash(i orderintent.AmendIntent) string {
	return IntentHash(orderintent.CanonicalAmend(i))
}

// ErrNonceReused means a one-shot nonce was presented twice.
var ErrNonceReused = errors.New("execgw: guardian nonce has already been used")

// NonceStore spends one-shot nonces. Implementations must be safe for concurrent
// use and must make Consume atomic: two goroutines presenting the same nonce is
// exactly the race that would duplicate a live order.
type NonceStore interface {
	// Consume marks the nonce as spent, returning ErrNonceReused if it already was.
	Consume(nonce string, now time.Time) error
}

// NewMemoryNonceStore returns the Phase 1 in-memory store.
//
// In-memory is sufficient while decisions are short-lived and issued in-process:
// a restart invalidates every outstanding decision anyway, because the issuing
// Guardian is gone with it. Phase 2, where decisions may outlive a process, is
// expected to back this with the journal database.
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

// verifyDecision checks everything about a decision except the nonce, which is
// spent separately at the last moment. It is called twice per mutation — once
// before the dispatch is recorded and once immediately before the broker call —
// because "재검증 직전" is what makes an expiry meaningful.
func verifyDecision(d GuardianDecision, plan mutationPlan, now time.Time) *RejectedError {
	switch {
	case d.Nonce == "":
		return reject(ReasonGuardianMissing,
			"no GuardianDecision nonce was supplied for the %s of %s", plan.kind, plan.symbol)
	case d.IntentHash == "":
		return reject(ReasonGuardianMissing,
			"the GuardianDecision for %s carries no intent hash", plan.symbol)
	case d.ExpiresAt.IsZero():
		return reject(ReasonGuardianMissing,
			"the GuardianDecision for %s carries no expiry", plan.symbol)
	case d.IntentHash != plan.intentHash:
		return reject(ReasonGuardianIntentMismatch,
			"the decision authorises intent %s but the mutation is intent %s",
			short(d.IntentHash), short(plan.intentHash))
	case !now.Before(d.ExpiresAt):
		return reject(ReasonGuardianExpired,
			"the decision expired at %s and it is now %s",
			d.ExpiresAt.UTC().Format(time.RFC3339), now.UTC().Format(time.RFC3339))
	}
	return verifyLimits(d.Limits, plan)
}

// verifyLimits measures a mutation against the snapshot.
//
// Two asymmetries are deliberate:
//
//   - A cancel is never limit-checked. It removes exposure, and a limit snapshot
//     is not a reason to refuse an exit (§0.3: changes to stop/exit behaviour may
//     only move in the conservative direction).
//   - An empty snapshot refuses only the mutations that *raise* exposure. "No
//     limits" is not an authorisation to buy, but it must not close the exits
//     either.
func verifyLimits(l Limits, plan mutationPlan) *RejectedError {
	if plan.kind == journal.KindCancel {
		return nil
	}
	if l.IsZero() {
		if !plan.raisesExposure {
			return nil
		}
		return reject(ReasonGuardianLimitsUnset,
			"the decision for %s carries no limit snapshot", plan.symbol)
	}
	if l.MaxQuantity > 0 && plan.quantity > l.MaxQuantity {
		return reject(ReasonGuardianLimitExceeded,
			"quantity %s exceeds the authorised maximum %s",
			decimalString(plan.quantity), decimalString(l.MaxQuantity))
	}
	if l.MaxNotional > 0 {
		notional := plan.notional()
		if notional > l.MaxNotional {
			return reject(ReasonGuardianLimitExceeded,
				"notional %s exceeds the authorised maximum %s",
				decimalString(notional), decimalString(l.MaxNotional))
		}
	}
	if l.Currency != "" && plan.currency != "" && l.Currency != plan.currency {
		return reject(ReasonGuardianLimitExceeded,
			"the limit snapshot is in %s but the order is in %s", l.Currency, plan.currency)
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

func short(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
