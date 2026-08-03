package protectionreadiness

import (
	"context"
	"time"
)

// DispatchScope is the immutable broker scope an exposure-raising mutation must
// match immediately before transport dispatch.
type DispatchScope struct {
	AccountID string
	ProfileID string
	Market    Market
}

// DispatchDecision is a read-only, market-scoped view of one sealed snapshot.
// SnapshotID is derived from one market verdict, so KR churn cannot invalidate
// an unchanged US decision (and vice versa).
type DispatchDecision struct {
	Allowed    bool
	Market     Market
	Code       RefusalCode
	Generation uint64
	Provenance Provenance
	SnapshotID string
}

// Dispatch validates the complete sealed snapshot and then the exact requested
// market scope and current expiry. It never changes a lane, approval or order.
func (snapshot ReadinessSnapshot) Dispatch(scope DispatchScope, now time.Time) DispatchDecision {
	decision := DispatchDecision{Market: scope.Market}
	if snapshot.release != ReadinessRelease {
		decision.Code = RefusalStateCorrupt
		return decision
	}
	if scope.AccountID == "" || scope.ProfileID == "" || !validMarket(scope.Market) || now.IsZero() {
		decision.Code = RefusalInvalid
		return decision
	}
	var verdict Verdict
	switch scope.Market {
	case MarketKR:
		verdict = snapshot.kr
		if snapshot.krSeal != marketVerdictSeal(snapshot.release, verdict) {
			decision.Code = RefusalStateCorrupt
			return decision
		}
	case MarketUS:
		verdict = snapshot.us
		if snapshot.usSeal != marketVerdictSeal(snapshot.release, verdict) {
			decision.Code = RefusalStateCorrupt
			return decision
		}
	}
	decision.Code, decision.Provenance, decision.Generation = verdict.Code, verdict.Provenance, verdict.Provenance.Serial
	if verdict.State != Wired || verdict.Code != RefusalNone {
		return decision
	}
	provenance := verdict.Provenance
	if provenance.AccountID != scope.AccountID || provenance.ProfileID != scope.ProfileID || provenance.Serial == 0 ||
		!validDigest(provenance.BodyDigest) || !validDigest(provenance.BuildDigest) || !validDigest(provenance.EvidenceDigest) || !validDigest(provenance.SupervisorDigest) {
		decision.Code = RefusalScopeMismatch
		return decision
	}
	now = now.UTC()
	if provenance.IssuedAt.IsZero() || provenance.ExpiresAt.IsZero() || now.Before(provenance.IssuedAt) {
		decision.Code = RefusalInvalid
		return decision
	}
	if !now.Before(provenance.ExpiresAt) {
		decision.Code = RefusalExpired
		return decision
	}
	decision.SnapshotID = marketSnapshotID(snapshot.release, verdict)
	decision.Allowed, decision.Code = true, RefusalNone
	return decision
}

func marketSnapshotID(release string, verdict Verdict) string {
	digest := marketVerdictSeal(release, verdict)
	return hexBytes(digest[:])
}

func marketVerdictSeal(release string, verdict Verdict) [32]byte {
	return hashStrings(release, string(verdict.Market), string(verdict.State), string(verdict.Code),
		verdict.Provenance.AccountID, verdict.Provenance.ProfileID, verdict.Provenance.KeyID,
		stringUint(verdict.Provenance.Serial), verdict.Provenance.BodyDigest, verdict.Provenance.BuildDigest,
		verdict.Provenance.EvidenceDigest, verdict.Provenance.SupervisorDigest,
		formatTime(verdict.Provenance.IssuedAt), formatTime(verdict.Provenance.ExpiresAt))
}

func stringUint(value uint64) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = digits[value%10]
		value /= 10
	}
	return string(buffer[index:])
}

// SnapshotProvider returns the latest immutable signed/supervisor-bound
// snapshot. Implementing the interface cannot mint WIRED: snapshot fields and
// its seal remain private to this package.
type SnapshotProvider interface {
	Current(context.Context) (ReadinessSnapshot, error)
}

type defaultProvider struct{}

func (defaultProvider) Current(context.Context) (ReadinessSnapshot, error) {
	return DefaultSnapshot(), nil
}

// DefaultProvider is the only shipped provider assembly in this wave. Both
// markets are independently UNWIRED.
func DefaultProvider() SnapshotProvider { return defaultProvider{} }
