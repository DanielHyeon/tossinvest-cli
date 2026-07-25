package execgw

// failclosed.go holds the refusals that happen before, or instead of, a broker
// call (harden-execution-base task 2.10).
//
// order-execution names three: an interactive auth challenge, a currency balance
// too small for the order, and an order shape the official path does not support.
// All three end the same way — no submission, a stable reason code, a journal
// record — and one thing is explicitly forbidden in all three: answering them
// automatically. "자동 환전·자동 승인은 금지된다".
//
// The reason each is a *refusal* rather than an error is that the engine has no
// way to satisfy them without a human: currency cannot be conjured, an app-push
// challenge cannot be tapped, and an unsupported order shape will not become
// supported by being sent again.

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Preflight performs the fail-closed checks that can be made before submitting.
type Preflight struct {
	// Account reads the currency balance. A nil reader does not mean "skip the
	// check": a buy with no way to verify funding is refused.
	Account AccountReader
	// Clock is unused today; it is here so a future freshness rule on the balance
	// read does not change the type's shape.
	Clock clock.Clock
}

// CheckPlace returns the refusal for an order that must not be submitted, or nil.
func (p *Preflight) CheckPlace(ctx context.Context, intent orderintent.PlaceIntent) *RejectedError {
	if rejected := checkOrderShape(intent); rejected != nil {
		return rejected
	}
	return p.checkFunding(ctx, intent)
}

// checkOrderShape refuses what the official order path cannot express.
//
// This mirrors internal/trading's capability check rather than calling it (that
// predicate is unexported, and internal/trading is not ours to change — design
// D1). Mirroring is safe in the one direction that matters: this check is a strict
// subset filter, so anything it accepts still faces the real check at the service.
func checkOrderShape(intent orderintent.PlaceIntent) *RejectedError {
	market := strings.ToLower(strings.TrimSpace(intent.Market))
	side := strings.ToLower(strings.TrimSpace(intent.Side))
	orderType := strings.ToLower(strings.TrimSpace(intent.OrderType))
	currency := strings.ToUpper(strings.TrimSpace(intent.CurrencyMode))

	if market != "kr" && market != "us" {
		return reject(ReasonUnsupportedOrderType,
			"market %q is not one the official order path trades", intent.Market)
	}
	if side != "buy" && side != "sell" {
		return reject(ReasonInvalidRequest, "side %q is neither buy nor sell", intent.Side)
	}
	if currency != "KRW" && currency != "USD" {
		return reject(ReasonUnsupportedOrderType, "currency mode %q is not supported", intent.CurrencyMode)
	}

	if intent.Fractional {
		switch {
		case market != "us":
			return reject(ReasonUnsupportedOrderType, "fractional orders exist only on the US market")
		case orderType != "market":
			return reject(ReasonUnsupportedOrderType, "a fractional order is always a market order")
		case side == "buy" && intent.Amount <= 0:
			return reject(ReasonInvalidRequest, "a fractional buy needs a positive amount")
		case side == "sell" && intent.Quantity <= 0:
			return reject(ReasonInvalidRequest, "a fractional sell needs a positive quantity")
		}
		return nil
	}

	if orderType != "limit" {
		return reject(ReasonUnsupportedOrderType,
			"only limit orders (and US fractional market orders) are supported, got %q", intent.OrderType)
	}
	if market == "kr" && currency != "KRW" {
		return reject(ReasonUnsupportedOrderType, "a KR order is priced in KRW, not %s", currency)
	}
	if intent.Quantity <= 0 {
		return reject(ReasonInvalidRequest, "quantity %s is not positive", decimalString(intent.Quantity))
	}
	if intent.Price <= 0 {
		return reject(ReasonInvalidRequest, "a limit order needs a positive price")
	}
	return nil
}

// checkFunding refuses a buy the account cannot pay for, and a buy that could only
// be paid for by converting currency.
//
// Sells are exempt: they produce currency rather than spending it, and gating an
// exit on buying power would close the way out (§0.3).
func (p *Preflight) checkFunding(ctx context.Context, intent orderintent.PlaceIntent) *RejectedError {
	if !strings.EqualFold(strings.TrimSpace(intent.Side), "buy") {
		return nil
	}
	market := strings.ToLower(strings.TrimSpace(intent.Market))
	currency := strings.ToUpper(strings.TrimSpace(intent.CurrencyMode))

	// A US buy priced in KRW settles only by converting KRW into USD. Upstream's
	// CLI lets a human take that decision; the engine must not, because the
	// conversion is exactly the "자동 환전" the spec forbids.
	if market == "us" && currency == "KRW" {
		return reject(ReasonFXConsentRequired,
			"a US buy priced in KRW would require a currency conversion, which the engine never triggers automatically; "+
				"price the order in USD and fund it from the USD balance")
	}

	needed := requiredFunds(intent)
	if needed <= 0 {
		return reject(ReasonInvalidRequest, "the order's cost could not be computed")
	}
	if p.Account == nil {
		return reject(ReasonBalanceUnavailable,
			"no account reader is configured, so the %s balance behind this buy cannot be verified", currency)
	}
	available, err := p.Account.BuyingPower(ctx, currency)
	if err != nil {
		return reject(ReasonBalanceUnavailable,
			"the %s buying power could not be read: %v", currency, err)
	}
	if available < needed {
		return reject(ReasonBalanceInsufficient,
			"the order needs %s %s but only %s is available; the engine never converts currency to cover the difference",
			decimalString(needed), currency, decimalString(available))
	}
	return nil
}

// requiredFunds is what the buy will cost in its own currency.
func requiredFunds(intent orderintent.PlaceIntent) float64 {
	if intent.Fractional {
		return intent.Amount
	}
	return intent.Quantity * intent.Price
}

// ClassifyBrokerRefusal recognises a broker answer that needs a human, and names
// it with a stable reason code.
//
// The classification is by error value first and by response body second. The body
// matching exists because the official client collapses a 4xx into an *APIError
// carrying the raw body, and the difference between "needs an app confirmation"
// and "bad request" is the difference between an operator alert and a bug report.
func ClassifyBrokerRefusal(err error) (ReasonCode, bool) {
	if err == nil {
		return "", false
	}
	if errors.Is(err, trading.ErrInteractiveAuthRequired) {
		return ReasonInteractiveAuthRequired, true
	}
	var branch *trading.BranchRequiredError
	if errors.As(err, &branch) {
		switch branch.Branch {
		case trading.BranchFXConsentRequired:
			return ReasonFXConsentRequired, true
		case trading.BranchFundingRequired:
			return ReasonFundingRequired, true
		default:
			return ReasonInteractiveAuthRequired, true
		}
	}

	var prepare *trading.PrepareRejectedError
	if errors.As(err, &prepare) {
		if reason, ok := classifyRefusalBody(prepare.BrokerMessage); ok {
			return reason, true
		}
	}
	if reason, ok := classifyRefusalBody(refusalBody(err)); ok {
		return reason, true
	}
	if isAuthRefusal(err) {
		return ReasonBrokerAuthRejected, true
	}
	return "", false
}

// refusalBody returns the broker's own message, and only that.
//
// It never falls back to err.Error(): a Go error string can carry anything, and
// matching operator-branch markers against text we produced ourselves is how a
// wrapped message turns into a wrong classification.
func refusalBody(err error) string {
	var apiErr *official.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Body
	}
	var branch *trading.BranchRequiredError
	if errors.As(err, &branch) {
		return branch.BrokerMessage
	}
	return ""
}

// isAuthRefusal reports a credential-level refusal (401/403).
func isAuthRefusal(err error) bool {
	if errors.Is(err, official.ErrAuth) || errors.Is(err, official.ErrIPNotAllowed) {
		return true
	}
	var apiErr *official.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 401 || apiErr.Code == 403
	}
	return false
}

// classifyRefusalBody looks for the broker's own vocabulary in a response body.
// The markers are matched case-insensitively and are deliberately narrow: a body
// this does not recognise stays unclassified rather than being guessed at.
func classifyRefusalBody(body string) (ReasonCode, bool) {
	lowered := strings.ToLower(body)
	if lowered == "" {
		return "", false
	}
	switch {
	case containsAny(lowered, "trade_auth_required", "interactive", "거래 인증"):
		return ReasonInteractiveAuthRequired, true
	case containsAny(lowered, "fx_consent", "exchange_consent", "환전 동의"):
		return ReasonFXConsentRequired, true
	case containsAny(lowered, "funding_required", "insufficient_deposit", "입금"):
		return ReasonFundingRequired, true
	default:
		return "", false
	}
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

// AllReasonCodes returns every reason code this build can emit, sorted.
//
// It exists so the vocabulary can be pinned by a test: these strings land in the
// journal and in operator alerts, and Phase 2's ledger reads them, so a rename is
// a data migration and not a refactor.
func AllReasonCodes() []ReasonCode {
	codes := []ReasonCode{
		ReasonGuardianMissing,
		ReasonGuardianExpired,
		ReasonGuardianIntentMismatch,
		ReasonGuardianNonceReused,
		ReasonGuardianLimitsUnset,
		ReasonGuardianLimitExceeded,

		ReasonInvalidRequest,
		ReasonPolicyDisabled,
		ReasonUnsupportedOrderType,
		ReasonSymbolInFlight,

		ReasonBrokerAccepted,
		ReasonBrokerRejected,
		ReasonBrokerOutcomeUnknown,

		ReasonBrokerAuthRejected,
		ReasonQueryStale,
		ReasonUnresolvedInDoubt,
		ReasonAlertUndelivered,

		ReasonBalanceInsufficient,
		ReasonBalanceUnavailable,
		ReasonInteractiveAuthRequired,
		ReasonFXConsentRequired,
		ReasonFundingRequired,

		// Fill detection and reconciliation (tasks 3.1-3.6). They were declared
		// with the rest of the enum but never registered here, which meant the
		// golden fixture — the thing that makes a rename a visible diff — did not
		// cover them.
		ReasonRecoveryIncomplete,
		ReasonFillDetectionSLO,
		ReasonBrokerStateUnknown,
		ReasonReconcileMismatch,
		ReasonReconcilePermanent,

		// Flatten (task 4.4).
		ReasonFlattenInProgress,
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
	return codes
}
