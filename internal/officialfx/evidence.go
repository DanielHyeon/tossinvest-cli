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
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

const (
	OfficialSource  = "official-fx"
	OfficialVersion = "toss-open-api/exchange-rate-v1"
	IdentitySource  = "same-currency"
	IdentityVersion = "same-currency/v1"
	maxDecimalBits  = 256
)

var (
	ErrInvalidEvidence    = errors.New("officialfx: invalid exchange-rate evidence")
	ErrEvidenceNotCurrent = errors.New("officialfx: exchange-rate evidence is not current")
)

// Evidence is an immutable quote-to-account conversion. Its fields are private
// so downstream code can obtain riskbucket evidence only after seal and
// freshness validation.
type Evidence struct {
	quoteCurrency, accountCurrency string
	rate, midRate, haircut         string
	source, version, digest        string
	rawValidFrom, rawValidUntil    string
	observedAt, freshUntil         time.Time
	seal                           [32]byte
}

// ReadOfficial is the only cross-currency mint. It performs the official GET
// itself rather than accepting a caller-constructed domain value.
func ReadOfficial(ctx context.Context, client *official.Client, quoteCurrency, accountCurrency, haircut string) (Evidence, error) {
	if ctx == nil || client == nil || !canonicalCurrency(quoteCurrency) || !canonicalCurrency(accountCurrency) || quoteCurrency == accountCurrency {
		return Evidence{}, fmt.Errorf("%w: request scope", ErrInvalidEvidence)
	}
	rate, err := client.ExchangeRate(ctx, quoteCurrency, accountCurrency)
	if err != nil {
		return Evidence{}, fmt.Errorf("officialfx: reading official exchange rate: %w", err)
	}
	return sealOfficial(rate, quoteCurrency, accountCurrency, haircut)
}

// Identity seals the exact same-currency conversion. Its caller supplies the
// validity window that already bounds the surrounding read snapshot; the
// conversion itself is fixed to rate=mid-rate=haircut=1.
func Identity(currency string, observedAt, freshUntil time.Time) (Evidence, error) {
	if !canonicalCurrency(currency) || observedAt.IsZero() || freshUntil.IsZero() || observedAt.After(freshUntil) {
		return Evidence{}, fmt.Errorf("%w: identity scope", ErrInvalidEvidence)
	}
	evidence := Evidence{
		quoteCurrency: currency, accountCurrency: currency,
		rate: "1", midRate: "1", haircut: "1",
		source: IdentitySource, version: IdentityVersion,
		rawValidFrom: observedAt.Format(time.RFC3339Nano), rawValidUntil: freshUntil.Format(time.RFC3339Nano),
		observedAt: observedAt.UTC(), freshUntil: freshUntil.UTC(),
	}
	evidence.digest = digestFor(evidence)
	evidence.seal = sealFor(evidence)
	return evidence, nil
}

func sealOfficial(rate domain.ExchangeRate, quoteCurrency, accountCurrency, haircut string) (Evidence, error) {
	observedAt, observedErr := time.Parse(time.RFC3339Nano, rate.ValidFromRaw)
	freshUntil, freshErr := time.Parse(time.RFC3339Nano, rate.ValidUntilRaw)
	_, rateOK := positiveDecimal(rate.RateRaw)
	_, midRateOK := positiveDecimal(rate.MidRateRaw)
	haircutValue, haircutOK := positiveDecimal(haircut)
	if rate.BaseCurrency != quoteCurrency || rate.QuoteCurrency != accountCurrency || rate.Code != quoteCurrency+"/"+accountCurrency ||
		!canonicalCurrency(rate.BaseCurrency) || !canonicalCurrency(rate.QuoteCurrency) || !rateOK || !midRateOK || !haircutOK ||
		haircutValue.Cmp(big.NewRat(1, 1)) < 0 || observedErr != nil || freshErr != nil || observedAt.IsZero() || freshUntil.IsZero() || observedAt.After(freshUntil) {
		return Evidence{}, fmt.Errorf("%w: official response", ErrInvalidEvidence)
	}
	evidence := Evidence{
		quoteCurrency: quoteCurrency, accountCurrency: accountCurrency,
		rate: rate.RateRaw, midRate: rate.MidRateRaw, haircut: haircut,
		source: OfficialSource, version: OfficialVersion,
		rawValidFrom: rate.ValidFromRaw, rawValidUntil: rate.ValidUntilRaw,
		observedAt: observedAt.UTC(), freshUntil: freshUntil.UTC(),
	}
	evidence.digest = digestFor(evidence)
	evidence.seal = sealFor(evidence)
	return evidence, nil
}

func (e Evidence) QuoteCurrency() string   { return e.quoteCurrency }
func (e Evidence) AccountCurrency() string { return e.accountCurrency }
func (e Evidence) Digest() string          { return e.digest }

// EvidenceAt releases a copy only while the sealed source window is current.
func (e Evidence) EvidenceAt(at time.Time) (riskbucket.FXEvidence, error) {
	if e.seal == ([32]byte{}) || e.seal != sealFor(e) {
		return riskbucket.FXEvidence{}, fmt.Errorf("%w: seal", ErrInvalidEvidence)
	}
	if at.IsZero() || at.Before(e.observedAt) || at.After(e.freshUntil) {
		return riskbucket.FXEvidence{}, ErrEvidenceNotCurrent
	}
	return riskbucket.FXEvidence{
		RateQuoteToBase: e.rate,
		Haircut:         e.haircut,
		Evidence: riskbucket.Evidence{
			Source: e.source, Version: e.version, Digest: e.digest,
			Official: true, Frozen: true, ObservedAt: e.observedAt, FreshUntil: e.freshUntil,
		},
	}, nil
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

func digestFor(e Evidence) string {
	preimage := strings.Join([]string{
		e.source, e.version, e.quoteCurrency, e.accountCurrency, e.rate, e.midRate, e.haircut,
		e.rawValidFrom, e.rawValidUntil,
	}, "\x00")
	sum := sha256.Sum256([]byte(preimage))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sealFor(e Evidence) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		e.source, e.version, e.quoteCurrency, e.accountCurrency, e.rate, e.midRate, e.haircut, e.digest,
		e.rawValidFrom, e.rawValidUntil, e.observedAt.Format(time.RFC3339Nano), e.freshUntil.Format(time.RFC3339Nano),
	}, "\x00")))
}
