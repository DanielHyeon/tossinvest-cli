package risk

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var ErrAccountBaseFXUnavailable = errors.New("risk: account-base FX authority unavailable")

// AccountBaseFX is one opaque, request-scoped conversion shared by sizing,
// Guardian aggregate checks and the later reservation/Gateway boundary. Its
// arithmetic fields are private: production callers can only obtain a value by
// presenting sealed officialfx evidence to BindAccountBaseFX.
type AccountBaseFX struct {
	quoteCurrency, accountCurrency string
	rate, haircut                  string
	source, version, digest        string
	observedAt, freshUntil         time.Time
	evaluatedAt                    time.Time
	seal                           [32]byte
}

func (f AccountBaseFX) QuoteCurrency() string   { return f.quoteCurrency }
func (f AccountBaseFX) AccountCurrency() string { return f.accountCurrency }
func (f AccountBaseFX) Digest() string          { return f.digest }
func (f AccountBaseFX) Source() string          { return f.source }
func (f AccountBaseFX) Version() string         { return f.version }
func (f AccountBaseFX) ObservedAt() time.Time   { return f.observedAt }
func (f AccountBaseFX) FreshUntil() time.Time   { return f.freshUntil }
func (f AccountBaseFX) EvaluatedAt() time.Time  { return f.evaluatedAt }

// BindAccountBaseFX consumes opaque official/identity evidence at the exact
// Guardian evaluation instant. A returned value cannot be moved to another
// market, policy currency or instant without invalidating its seal/scope.
func BindAccountBaseFX(at time.Time, market costs.Market, policy Policy, authority officialfx.Evidence) (AccountBaseFX, error) {
	if at.IsZero() {
		return AccountBaseFX{}, fmt.Errorf("%w: evaluation instant", ErrAccountBaseFXUnavailable)
	}
	if err := policy.Validate(); err != nil {
		return AccountBaseFX{}, fmt.Errorf("%w: policy: %v", ErrAccountBaseFXUnavailable, err)
	}
	quote, err := currencyOf(market)
	if err != nil {
		return AccountBaseFX{}, fmt.Errorf("%w: %v", ErrAccountBaseFXUnavailable, err)
	}
	base := policy.LimitCurrency()
	return bindAccountBaseFXPair(at, quote, base, authority)
}

// BindAccountBaseFXPair is the Gateway-side binding boundary. The account
// currency comes from the persisted decision envelope, the quote currency from
// the market, and the rate/haircut only from opaque official evidence.
func BindAccountBaseFXPair(at time.Time, market costs.Market, accountCurrency string, authority officialfx.Evidence) (AccountBaseFX, error) {
	if at.IsZero() {
		return AccountBaseFX{}, fmt.Errorf("%w: evaluation instant", ErrAccountBaseFXUnavailable)
	}
	quote, err := currencyOf(market)
	if err != nil {
		return AccountBaseFX{}, fmt.Errorf("%w: %v", ErrAccountBaseFXUnavailable, err)
	}
	return bindAccountBaseFXPair(at, quote, strings.ToUpper(strings.TrimSpace(accountCurrency)), authority)
}

func bindAccountBaseFXPair(at time.Time, quote, base string, authority officialfx.Evidence) (AccountBaseFX, error) {
	reserve, err := authority.EvidenceAt(at.UTC(), quote, base)
	if err != nil {
		return AccountBaseFX{}, fmt.Errorf("%w: %v", ErrAccountBaseFXUnavailable, err)
	}
	fx := AccountBaseFX{
		quoteCurrency: quote, accountCurrency: base,
		rate: reserve.RateQuoteToBase(), haircut: reserve.Haircut(),
		source: reserve.Source(), version: reserve.Version(), digest: reserve.Digest(),
		observedAt: reserve.ObservedAt().UTC(), freshUntil: reserve.FreshUntil().UTC(), evaluatedAt: at.UTC(),
	}
	fx.seal = sealAccountBaseFX(fx)
	if _, err := fx.multiplierAt(at.UTC(), quote, base); err != nil {
		return AccountBaseFX{}, err
	}
	return fx, nil
}

func sealAccountBaseFX(f AccountBaseFX) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		f.quoteCurrency, f.accountCurrency, f.rate, f.haircut, f.source, f.version, f.digest,
		f.observedAt.Format(time.RFC3339Nano), f.freshUntil.Format(time.RFC3339Nano), f.evaluatedAt.Format(time.RFC3339Nano),
	}, "\x00")))
}

func (f AccountBaseFX) multiplierAt(at time.Time, quote, base string) (*big.Rat, error) {
	if f == (AccountBaseFX{}) || f.seal == ([32]byte{}) || f.seal != sealAccountBaseFX(f) ||
		f.quoteCurrency != quote || f.accountCurrency != base || f.source == "" || f.version == "" || f.digest == "" ||
		at.IsZero() || !f.evaluatedAt.Equal(at.UTC()) || f.observedAt.IsZero() || f.freshUntil.IsZero() ||
		at.Before(f.observedAt) || at.After(f.freshUntil) {
		return nil, fmt.Errorf("%w: seal, scope or freshness", ErrAccountBaseFXUnavailable)
	}
	rate, err := parseDecimal("quote-to-account FX rate", f.rate)
	if err != nil || rate.Sign() <= 0 {
		return nil, fmt.Errorf("%w: rate", ErrAccountBaseFXUnavailable)
	}
	haircut, err := parseDecimal("FX haircut", f.haircut)
	if err != nil || haircut.Cmp(big.NewRat(1, 1)) < 0 {
		return nil, fmt.Errorf("%w: haircut", ErrAccountBaseFXUnavailable)
	}
	if quote == base {
		if rate.Cmp(big.NewRat(1, 1)) != 0 || haircut.Cmp(big.NewRat(1, 1)) != 0 || f.source != officialfx.IdentitySource {
			return nil, fmt.Errorf("%w: same-currency identity", ErrAccountBaseFXUnavailable)
		}
	} else if f.source != officialfx.OfficialSource {
		return nil, fmt.Errorf("%w: cross-currency source", ErrAccountBaseFXUnavailable)
	}
	return new(big.Rat).Mul(rate, haircut), nil
}

// accountBaseMultiplier preserves the released same-currency Guardian API for
// legacy callers. Cross-currency use never receives that compatibility path,
// and any non-zero opaque value is always validated even for same currency.
func accountBaseMultiplier(at time.Time, market costs.Market, policy Policy, fx AccountBaseFX) (*big.Rat, string, string, bool, error) {
	quote, err := currencyOf(market)
	if err != nil {
		return nil, "", "", false, err
	}
	base := policy.LimitCurrency()
	if fx == (AccountBaseFX{}) && quote == base {
		return big.NewRat(1, 1), quote, base, false, nil
	}
	multiplier, err := fx.multiplierAt(at.UTC(), quote, base)
	if err != nil {
		return nil, quote, base, true, err
	}
	return multiplier, quote, base, true, nil
}

// accountBaseMoney converts quote money with exact rational arithmetic and
// rounds upward once to the base accounting unit. The same value is returned by
// EntryExposureValue and compared by checkOpenExposure, preventing a caller-side
// second conversion from producing less reservation than Guardian admitted.
func accountBaseMoney(at time.Time, market costs.Market, policy Policy, fx AccountBaseFX, quoteMoney riskcalc.Money) (riskcalc.Money, error) {
	multiplier, quote, base, bound, err := accountBaseMultiplier(at, market, policy, fx)
	if err != nil {
		return riskcalc.Money{}, err
	}
	amount, err := moneyIn("quote-currency entry value", quoteMoney, quote)
	if err != nil {
		return riskcalc.Money{}, err
	}
	if !bound {
		return riskcalc.Money{Amount: amount, Currency: base}, nil
	}
	value, err := parseDecimal("quote-currency entry value", amount)
	if err != nil || value.Sign() < 0 {
		return riskcalc.Money{}, fmt.Errorf("%w: entry value", ErrAccountBaseFXUnavailable)
	}
	value.Mul(value, multiplier)
	return riskcalc.Money{Amount: ceilNonNegative(value).String(), Currency: base}, nil
}

// AccountBaseOrderNotional converts one exact quote-currency order notional to
// the account-base accounting unit and rounds upward once. It accepts only an
// already sealed binding, so a Gateway cannot reconstruct authority from the
// auditable decision envelope.
func AccountBaseOrderNotional(at time.Time, market costs.Market, quoteNotional string, fx AccountBaseFX) (riskcalc.Money, error) {
	value, err := parseDecimal("quote-currency order notional", quoteNotional)
	if err != nil || value.Sign() < 0 {
		return riskcalc.Money{}, fmt.Errorf("%w: order notional", ErrAccountBaseFXUnavailable)
	}
	return accountBaseOrderValue(at, market, value, fx)
}

// AccountBaseOrderValue is the quantity-order counterpart used by Gateway.
// Quantity and quote price remain exact decimals until the single conservative
// base-unit ceiling.
func AccountBaseOrderValue(at time.Time, market costs.Market, quantity, quotePrice string, fx AccountBaseFX) (riskcalc.Money, error) {
	q, err := parseDecimal("order quantity", quantity)
	if err != nil || q.Sign() < 0 {
		return riskcalc.Money{}, fmt.Errorf("%w: order quantity", ErrAccountBaseFXUnavailable)
	}
	price, err := parseDecimal("quote-currency order price", quotePrice)
	if err != nil || price.Sign() < 0 {
		return riskcalc.Money{}, fmt.Errorf("%w: order price", ErrAccountBaseFXUnavailable)
	}
	return accountBaseOrderValue(at, market, new(big.Rat).Mul(q, price), fx)
}

func accountBaseOrderValue(at time.Time, market costs.Market, value *big.Rat, fx AccountBaseFX) (riskcalc.Money, error) {
	quote, err := currencyOf(market)
	if err != nil {
		return riskcalc.Money{}, err
	}
	multiplier, err := fx.multiplierAt(at.UTC(), quote, fx.accountCurrency)
	if err != nil {
		return riskcalc.Money{}, err
	}
	value.Mul(value, multiplier)
	return riskcalc.Money{Amount: ceilNonNegative(value).String(), Currency: fx.accountCurrency}, nil
}

func ceilNonNegative(value *big.Rat) *big.Int {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func accountBaseRiskQuantity(policy Policy, market costs.Market, entryPrice, stopPrice string, fx AccountBaseFX, at time.Time) (string, error) {
	multiplier, _, _, _, err := accountBaseMultiplier(at, market, policy, fx)
	if err != nil {
		return "", err
	}
	budget, err := parseDecimal("risk budget", policy.RiskBudget.Amount)
	if err != nil || budget.Sign() < 0 {
		return "", fmt.Errorf("risk budget %q is invalid", policy.RiskBudget.Amount)
	}
	entry, err := parseDecimal("entry price", entryPrice)
	if err != nil {
		return "", err
	}
	stop, err := parseDecimal("stop price", stopPrice)
	if err != nil {
		return "", err
	}
	width := new(big.Rat).Sub(entry, stop)
	if width.Sign() <= 0 {
		return "0", nil
	}
	width.Mul(width, multiplier)
	quantity := new(big.Rat).Quo(budget, width)
	return new(big.Int).Quo(quantity.Num(), quantity.Denom()).String(), nil
}

// AccountBaseStrategyEntryQuantity is the paired KR/US Guardian sizing rule.
// Risk and notional money are account-base values; the market price/stop remain
// quote values. Every monetary conversion uses the same opaque FX and each
// quantity cap is floored exactly once.
func AccountBaseStrategyEntryQuantity(policy Policy, market costs.Market, entryPrice, stopPrice string, fx AccountBaseFX) (string, error) {
	// This API is the production multi-market boundary, not the released
	// same-currency Evaluate compatibility path. Requiring an explicit sealed
	// value for both markets keeps KR identity evidence and US official evidence
	// on the same authority contract.
	if fx == (AccountBaseFX{}) {
		return "", fmt.Errorf("%w: explicit account-base FX required", ErrAccountBaseFXUnavailable)
	}
	if err := policy.Validate(); err != nil {
		return "", err
	}
	riskRaw, err := accountBaseRiskQuantity(policy, market, entryPrice, stopPrice, fx, fx.evaluatedAt)
	if err != nil {
		return "", err
	}
	riskCap, err := parseWholeNumber("risk-based quantity", riskRaw)
	if err != nil {
		return "", err
	}
	quantityCap, err := parseWholeNumber("maximum order quantity", policy.MaxOrderQuantity)
	if err != nil {
		return "", err
	}
	multiplier, _, _, _, err := accountBaseMultiplier(fx.evaluatedAt, market, policy, fx)
	if err != nil {
		return "", err
	}
	entry, err := parseDecimal("entry price", entryPrice)
	if err != nil {
		return "", err
	}
	notional, err := parseDecimal("maximum order notional", policy.MaxOrderNotional.Amount)
	if err != nil {
		return "", err
	}
	entry.Mul(entry, multiplier)
	if entry.Sign() <= 0 || notional.Sign() <= 0 {
		return "", ErrStrategyQuantityZero
	}
	notionalRatio := new(big.Rat).Quo(notional, entry)
	notionalCap := new(big.Int).Quo(notionalRatio.Num(), notionalRatio.Denom())
	capacity := new(big.Int).Set(riskCap)
	if quantityCap.Cmp(capacity) < 0 {
		capacity.Set(quantityCap)
	}
	if notionalCap.Cmp(capacity) < 0 {
		capacity.Set(notionalCap)
	}
	if capacity.Sign() <= 0 {
		return "", ErrStrategyQuantityZero
	}
	return capacity.String(), nil
}
