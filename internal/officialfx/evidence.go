// Package officialfx converts the read-only official exchange-rate endpoint
// into sealed, exact-decimal evidence. It owns no broker, journal, activation,
// configuration or mutation capability.
package officialfx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	OfficialSource         = "official-fx"
	OfficialVersion        = "toss-open-api/exchange-rate-v1"
	IdentitySource         = "same-currency"
	IdentityVersion        = "same-currency/v1"
	haircutPolicySource    = "fx-haircut-policy-authority"
	identitySnapshotSource = "risk-snapshot-identity-authority"
	maxDecimalBits         = 256
	maxIdentityBytes       = 256
	maxIdentityWindow      = 5 * time.Minute
)

var (
	ErrInvalidEvidence    = errors.New("officialfx: invalid exchange-rate evidence")
	ErrEvidenceNotCurrent = errors.New("officialfx: exchange-rate evidence is not current")
)

// HaircutPolicy is a package-owned, immutable policy capability. There is
// deliberately no public raw-value constructor: until a production policy
// loader exists, callers can only hold the zero value and authority minting is
// unavailable.
type HaircutPolicy struct {
	id, version, multiplier, digest string
	observedAt, freshUntil          time.Time
	seal                            [32]byte
}

func newHaircutPolicy(id, version, multiplier string, observedAt, freshUntil time.Time) (HaircutPolicy, error) {
	canonical, ok := canonicalPolicyDecimal(multiplier)
	if !boundedIdentity(id) || !boundedIdentity(version) || !ok || observedAt.IsZero() || freshUntil.IsZero() || observedAt.After(freshUntil) {
		return HaircutPolicy{}, fmt.Errorf("%w: haircut policy", ErrInvalidEvidence)
	}
	policy := HaircutPolicy{id: id, version: version, multiplier: canonical, observedAt: observedAt.UTC(), freshUntil: freshUntil.UTC()}
	policy.digest = sha256Identity(strings.Join([]string{haircutPolicySource, policy.id, policy.version, policy.multiplier,
		policy.observedAt.Format(time.RFC3339Nano), policy.freshUntil.Format(time.RFC3339Nano)}, "\x00"))
	policy.seal = haircutPolicySeal(policy)
	return policy, nil
}

func (p HaircutPolicy) validAt(at time.Time) bool {
	return p.seal != ([32]byte{}) && p.seal == haircutPolicySeal(p) && boundedIdentity(p.id) && boundedIdentity(p.version) &&
		canonicalDigest(p.digest) && !at.IsZero() && !at.Before(p.observedAt) && !at.After(p.freshUntil)
}

func (p HaircutPolicy) validWindow(from, until time.Time) bool {
	return p.seal != ([32]byte{}) && p.seal == haircutPolicySeal(p) && boundedIdentity(p.id) && boundedIdentity(p.version) &&
		canonicalDigest(p.digest) && !from.IsZero() && !until.IsZero() && !from.After(until) &&
		!p.observedAt.After(until) && !p.freshUntil.Before(from)
}

func haircutPolicySeal(p HaircutPolicy) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{haircutPolicySource, p.id, p.version, p.multiplier, p.digest,
		p.observedAt.Format(time.RFC3339Nano), p.freshUntil.Format(time.RFC3339Nano)}, "\x00")))
}

// IdentitySnapshot is a bounded package-owned snapshot capability used only for
// same-currency identity conversion. Callers cannot choose its freshness.
type IdentitySnapshot struct {
	currency, snapshotID, digest string
	observedAt, freshUntil       time.Time
	seal                         [32]byte
}

func newIdentitySnapshot(currency, snapshotID, digest string, observedAt, freshUntil time.Time) (IdentitySnapshot, error) {
	if !canonicalCurrency(currency) || !boundedIdentity(snapshotID) || !canonicalDigest(digest) ||
		observedAt.IsZero() || freshUntil.IsZero() || observedAt.After(freshUntil) || freshUntil.Sub(observedAt) > maxIdentityWindow {
		return IdentitySnapshot{}, fmt.Errorf("%w: identity snapshot", ErrInvalidEvidence)
	}
	snapshot := IdentitySnapshot{currency: currency, snapshotID: snapshotID, digest: digest,
		observedAt: observedAt.UTC(), freshUntil: freshUntil.UTC()}
	snapshot.seal = identitySnapshotSeal(snapshot)
	return snapshot, nil
}

func (s IdentitySnapshot) valid() bool {
	return s.seal != ([32]byte{}) && s.seal == identitySnapshotSeal(s) && canonicalCurrency(s.currency) &&
		boundedIdentity(s.snapshotID) && canonicalDigest(s.digest) && !s.observedAt.IsZero() && !s.freshUntil.IsZero() &&
		!s.observedAt.After(s.freshUntil) && s.freshUntil.Sub(s.observedAt) <= maxIdentityWindow
}

func identitySnapshotSeal(s IdentitySnapshot) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{identitySnapshotSource, s.currency, s.snapshotID, s.digest,
		s.observedAt.Format(time.RFC3339Nano), s.freshUntil.Format(time.RFC3339Nano)}, "\x00")))
}

// Evidence is an immutable quote-to-account conversion. Downstream code must
// retain this opaque value and revalidate it at the consuming clock boundary.
type Evidence struct {
	quoteCurrency, accountCurrency string
	rate, midRate                  string
	source, version, digest        string
	rawValidFrom, rawValidUntil    string
	observedAt, freshUntil         time.Time
	haircut                        string
	haircutPolicyID                string
	haircutPolicyVersion           string
	haircutPolicyDigest            string
	haircutObservedAt              time.Time
	haircutFreshUntil              time.Time
	identitySnapshotID             string
	identitySnapshotDigest         string
	seal                           [32]byte
}

// ReadOfficial is the only cross-currency mint. Configured endpoint or HTTP
// clients may perform ordinary reads but cannot mint official authority.
func ReadOfficial(ctx context.Context, client *official.Client, quoteCurrency, accountCurrency string, policy HaircutPolicy) (Evidence, error) {
	origin, originOK := client.AuthorityOrigin()
	if ctx == nil || client == nil || !originOK || !origin.Valid() || !canonicalCurrency(quoteCurrency) ||
		!canonicalCurrency(accountCurrency) || quoteCurrency == accountCurrency || policy.seal == ([32]byte{}) ||
		policy.seal != haircutPolicySeal(policy) {
		return Evidence{}, fmt.Errorf("%w: request authority", ErrInvalidEvidence)
	}
	rate, err := client.ExchangeRate(ctx, quoteCurrency, accountCurrency)
	if err != nil {
		return Evidence{}, fmt.Errorf("officialfx: reading official exchange rate: %w", err)
	}
	return sealOfficial(rate, quoteCurrency, accountCurrency, policy)
}

// Identity consumes a package-owned snapshot capability. Its rate and haircut
// are fixed to canonical one; the caller cannot supply a validity window.
func Identity(snapshot IdentitySnapshot) (Evidence, error) {
	if !snapshot.valid() {
		return Evidence{}, fmt.Errorf("%w: identity snapshot", ErrInvalidEvidence)
	}
	evidence := Evidence{
		quoteCurrency: snapshot.currency, accountCurrency: snapshot.currency,
		rate: "1", midRate: "1", haircut: "1",
		source: IdentitySource, version: IdentityVersion,
		rawValidFrom: snapshot.observedAt.Format(time.RFC3339Nano), rawValidUntil: snapshot.freshUntil.Format(time.RFC3339Nano),
		observedAt: snapshot.observedAt, freshUntil: snapshot.freshUntil,
		identitySnapshotID: snapshot.snapshotID, identitySnapshotDigest: snapshot.digest,
	}
	evidence.digest = digestFor(evidence)
	evidence.seal = sealFor(evidence)
	return evidence, nil
}

func sealOfficial(rate domain.ExchangeRate, quoteCurrency, accountCurrency string, policy HaircutPolicy) (Evidence, error) {
	observedAt, observedErr := time.Parse(time.RFC3339Nano, rate.ValidFromRaw)
	freshUntil, freshErr := time.Parse(time.RFC3339Nano, rate.ValidUntilRaw)
	_, rateOK := positiveDecimal(rate.RateRaw)
	_, midRateOK := positiveDecimal(rate.MidRateRaw)
	if rate.BaseCurrency != quoteCurrency || rate.QuoteCurrency != accountCurrency || rate.Code != quoteCurrency+"/"+accountCurrency ||
		!canonicalCurrency(rate.BaseCurrency) || !canonicalCurrency(rate.QuoteCurrency) || !rateOK || !midRateOK ||
		observedErr != nil || freshErr != nil || observedAt.IsZero() || freshUntil.IsZero() || observedAt.After(freshUntil) ||
		!policy.validWindow(observedAt, freshUntil) {
		return Evidence{}, fmt.Errorf("%w: official response", ErrInvalidEvidence)
	}
	from := observedAt.UTC()
	if policy.observedAt.After(from) {
		from = policy.observedAt
	}
	until := freshUntil.UTC()
	if policy.freshUntil.Before(until) {
		until = policy.freshUntil
	}
	if from.After(until) {
		return Evidence{}, fmt.Errorf("%w: no policy/rate window intersection", ErrInvalidEvidence)
	}
	evidence := Evidence{
		quoteCurrency: quoteCurrency, accountCurrency: accountCurrency,
		rate: rate.RateRaw, midRate: rate.MidRateRaw,
		source: OfficialSource, version: OfficialVersion,
		rawValidFrom: rate.ValidFromRaw, rawValidUntil: rate.ValidUntilRaw,
		observedAt: from, freshUntil: until, haircut: policy.multiplier,
		haircutPolicyID: policy.id, haircutPolicyVersion: policy.version, haircutPolicyDigest: policy.digest,
		haircutObservedAt: policy.observedAt, haircutFreshUntil: policy.freshUntil,
	}
	evidence.digest = digestFor(evidence)
	evidence.seal = sealFor(evidence)
	return evidence, nil
}

func (e Evidence) QuoteCurrency() string   { return e.quoteCurrency }
func (e Evidence) AccountCurrency() string { return e.accountCurrency }
func (e Evidence) Digest() string          { return e.digest }

// Reserve is an immutable view released only after Evidence validates. Its
// fields remain private so constructing look-alike strings cannot satisfy the
// q_final input contract.
type Reserve struct {
	rate, haircut, source, version, digest string
	observedAt, freshUntil                 time.Time
}

func (r Reserve) RateQuoteToBase() string { return r.rate }
func (r Reserve) Haircut() string         { return r.haircut }
func (r Reserve) Source() string          { return r.source }
func (r Reserve) Version() string         { return r.version }
func (r Reserve) Digest() string          { return r.digest }
func (r Reserve) ObservedAt() time.Time   { return r.observedAt }
func (r Reserve) FreshUntil() time.Time   { return r.freshUntil }

// EvidenceAt validates seal, exact currency pair and the intersection of source,
// haircut-policy or identity-snapshot freshness at the caller's clock.
func (e Evidence) EvidenceAt(at time.Time, quoteCurrency, accountCurrency string) (Reserve, error) {
	if e.seal == ([32]byte{}) || e.seal != sealFor(e) || e.quoteCurrency != quoteCurrency || e.accountCurrency != accountCurrency ||
		!canonicalCurrency(quoteCurrency) || !canonicalCurrency(accountCurrency) || !canonicalDigest(e.digest) {
		return Reserve{}, fmt.Errorf("%w: seal or scope", ErrInvalidEvidence)
	}
	if at.IsZero() || at.Before(e.observedAt) || at.After(e.freshUntil) {
		return Reserve{}, ErrEvidenceNotCurrent
	}
	if e.source == OfficialSource {
		if !boundedIdentity(e.haircutPolicyID) || !boundedIdentity(e.haircutPolicyVersion) || !canonicalDigest(e.haircutPolicyDigest) ||
			at.Before(e.haircutObservedAt) || at.After(e.haircutFreshUntil) {
			return Reserve{}, fmt.Errorf("%w: haircut policy", ErrInvalidEvidence)
		}
	} else if e.source == IdentitySource {
		if quoteCurrency != accountCurrency || e.rate != "1" || e.midRate != "1" || e.haircut != "1" ||
			!boundedIdentity(e.identitySnapshotID) || !canonicalDigest(e.identitySnapshotDigest) {
			return Reserve{}, fmt.Errorf("%w: identity snapshot", ErrInvalidEvidence)
		}
	} else {
		return Reserve{}, fmt.Errorf("%w: source", ErrInvalidEvidence)
	}
	return Reserve{rate: e.rate, haircut: e.haircut, source: e.source, version: e.version, digest: e.digest,
		observedAt: e.observedAt, freshUntil: e.freshUntil}, nil
}

func positiveDecimal(raw string) (*big.Rat, bool) {
	if raw == "" || len(raw) > 128 || strings.ContainsAny(raw, "/eE+-") || strings.Count(raw, ".") > 1 {
		return nil, false
	}
	parts := strings.Split(raw, ".")
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, false
			}
		}
	}
	value, ok := new(big.Rat).SetString(raw)
	return value, ok && value.Sign() > 0 && value.Num().BitLen() <= maxDecimalBits && value.Denom().BitLen() <= maxDecimalBits
}

func canonicalPolicyDecimal(raw string) (string, bool) {
	value, ok := positiveDecimal(raw)
	if !ok || value.Cmp(big.NewRat(1, 1)) < 0 {
		return "", false
	}
	parts := strings.Split(raw, ".")
	if len(parts[0]) > 1 && parts[0][0] == '0' {
		return "", false
	}
	if len(parts) == 1 {
		return parts[0], true
	}
	fraction := strings.TrimRight(parts[1], "0")
	if fraction == "" {
		return parts[0], true
	}
	return parts[0] + "." + fraction, true
}

func canonicalCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func boundedIdentity(value string) bool {
	if value == "" || len(value) > maxIdentityBytes || value != strings.TrimSpace(value) {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func canonicalDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, r := range value[len("sha256:"):] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func sha256Identity(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestFor(e Evidence) string {
	return sha256Identity(strings.Join([]string{
		e.source, e.version, e.quoteCurrency, e.accountCurrency, e.rate, e.midRate, e.haircut,
		e.rawValidFrom, e.rawValidUntil, e.haircutPolicyID, e.haircutPolicyVersion, e.haircutPolicyDigest,
		e.haircutObservedAt.Format(time.RFC3339Nano), e.haircutFreshUntil.Format(time.RFC3339Nano),
		e.identitySnapshotID, e.identitySnapshotDigest,
	}, "\x00"))
}

func sealFor(e Evidence) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		e.source, e.version, e.quoteCurrency, e.accountCurrency, e.rate, e.midRate, e.haircut, e.digest,
		e.rawValidFrom, e.rawValidUntil, e.observedAt.Format(time.RFC3339Nano), e.freshUntil.Format(time.RFC3339Nano),
		e.haircutPolicyID, e.haircutPolicyVersion, e.haircutPolicyDigest,
		e.haircutObservedAt.Format(time.RFC3339Nano), e.haircutFreshUntil.Format(time.RFC3339Nano),
		e.identitySnapshotID, e.identitySnapshotDigest,
	}, "\x00")))
}
