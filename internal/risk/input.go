package risk

// input.go declares everything the chain reads. The split matters: Intent is
// what the caller proposes, AccountState is what the world is, Policy is what
// the operator configured, and the cost model is shared with the exit policy and
// the performance record.
//
// Nothing here is fetched. The issuer (task 4.1) fills these in from the
// journal, the broker snapshot and the configuration; keeping them as plain
// values is what stops a hidden read from turning a stale number fresh between
// two checks of the same decision.

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DefaultMinRewardRisk is the minimum (target − entry) / (entry − stop) an entry
// must offer.
//
// # Provenance
//
// This is a **new check** — StockOS's `evaluate_guardian` has no counterpart;
// the ratio lived in the strategy layer, and risk-management promotes it to
// intent arithmetic (신규 검사 — 세션 구조 기반 산출은 P3).
//
// The value is StockOS's, from two places:
//
//   - `strategies/parker_vwap/default_lock.py:35-38` — the §22 #2 lock
//     `min_expected_rr`, whose original locked value was 2.0. Plan 044 later
//     relaxed it to 1.3 (recorded at :83 of the same file) because KRX signals
//     were producing RRs of 1.0–1.3; TossOS does not follow that relaxation,
//     because §0.9 allows sizing- and stop-adjacent movement in the conservative
//     direction only.
//   - `live_entry_contract.py:53` — `_DEFAULT_US_MIN_RR = Decimal("2.0")`, the
//     floor the live buy gate still applies today.
//
// 1.5 was considered and rejected: it is the lowest tier of the live gate's
// per-setup range (2.0–4.0), not a floor anyone chose deliberately.
const DefaultMinRewardRisk = "2.0"

// Side is the direction of the proposed mutation.
//
// There are two, and the asymmetry between them is the whole shape of this
// package: a BUY raises exposure and passes every gate; a SELL reduces it and is
// bounded only by what the account holds.
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// OperatingMode is the account's mutation-safety mode (design D3).
//
// Three, not four: EXIT_ONLY was removed because it behaved identically to
// ENTRY_BLOCKED, and a rung of the approval ladder that changes nothing is a
// rung an operator learns to skip. Persistence and transitions belong to task
// 3.1; this package only reads the value.
type OperatingMode string

const (
	// ModeNormal permits entries.
	ModeNormal OperatingMode = "NORMAL"
	// ModeEntryBlocked refuses EXPOSURE_RAISING and permits RISK_REDUCING. It is
	// the target state of every automatic conservative escalation.
	ModeEntryBlocked OperatingMode = "ENTRY_BLOCKED"
	// ModeHaltAll refuses EXPOSURE_RAISING and PROTECTION_WEAKENING, and permits
	// RISK_REDUCING. Inside this change it behaves like ENTRY_BLOCKED for the
	// chain; the difference is that only an operator may enter it.
	ModeHaltAll OperatingMode = "HALT_ALL"
)

// Intent is the mutation the caller proposes, in the caller's own words.
//
// Prices and quantities are decimal strings, the journal's convention: these
// values are hashed into a decision preimage and stored, and a binary float
// cannot hold them exactly.
type Intent struct {
	AccountRef string
	Market     costs.Market
	Symbol     string
	Side       Side
	// Quantity is a whole number of units. Fractional orders are not a path
	// TossOS has, and accepting one here would produce an intent no broker call
	// can carry.
	Quantity string
	// LimitPrice is the entry price. Automated entries are LIMIT only
	// (riskcalc.CheckAutomatedEntry): a market entry has no exposure valuation,
	// so there is nothing to size or to reserve against.
	LimitPrice string
	// StopPrice is the protective stop. Required for an entry, and it does not
	// stop being required after the check: it becomes the exit policy's t0
	// baseline (exit-policy).
	StopPrice string
	// TargetPrice is the profit target. Required for an entry — the reward:risk
	// ratio and the break-even comparison are both computed from it.
	TargetPrice string
}

// AccountState is everything about the account the chain measures against.
//
// Every field is supplied by the caller. The aggregate valuations
// (OpenExposure, DailyRealizedLoss) are riskcalc's outputs, not recomputed here:
// riskcalc owns the calculation contract, including which inputs are too stale
// to use, and a second implementation would be a second answer.
type AccountState struct {
	// KillSwitchActive is the operator's block-only switch.
	KillSwitchActive bool
	// Mode is the current operating mode (journal `operating_modes`, task 3.1).
	Mode OperatingMode
	// EntryBlockedLatch is the entry gate's latch (task 2.4 wires it to
	// execgw's 401/403 · SLO · RECONCILE · recovery states).
	EntryBlockedLatch bool
	// EntryBlockedReason is the latch's own reason code, carried into the
	// refusal's detail so the two vocabularies stay traceable to each other.
	EntryBlockedReason string
	// AllowedSymbols is the entry allowlist. Empty admits nothing.
	AllowedSymbols []string
	// HeldQuantity is this symbol's holding in the broker snapshot. It bounds a
	// reduction; an empty value is unknown, not zero.
	HeldQuantity string
	// CashAvailable is the account's spendable cash.
	CashAvailable riskcalc.Money
	// OpenExposure is riskcalc.GrossLongExposure's result for the account,
	// excluding this intent.
	OpenExposure riskcalc.Money
	// DailyRealizedLoss is riskcalc.DailyLoss's result: a non-negative
	// magnitude, after costs.
	DailyRealizedLoss riskcalc.Money
	// AccountEquity is the capital the daily-loss ratio is measured against.
	AccountEquity riskcalc.Money
	// SameDayEntryCount is how many entries this symbol already had today.
	SameDayEntryCount int
	// LastEntryAt is when the most recent one was. A zero value with a non-zero
	// count means "we know it happened but not when", which is treated as
	// inside the cooldown rather than outside it.
	LastEntryAt time.Time
	// PendingBuy reports an unfilled buy on this symbol.
	PendingBuy bool
	// DuplicateOrder reports that this intent duplicates one already in flight.
	DuplicateOrder bool
}

// ReentryPolicy bounds how often one symbol may be entered in a day.
type ReentryPolicy struct {
	// MaxEntriesPerSymbol is the same-day ceiling. Default 2 `[미검증]`.
	MaxEntriesPerSymbol int
	// Cooldown is the minimum gap between two entries on one symbol. Default
	// 30 minutes `[미검증]`.
	Cooldown time.Duration
}

// Policy is the configured limit set.
//
// All money fields share one currency; the chain refuses to compare across
// currencies rather than converting, because conversion needs a rate with a
// staleness bound and that lives in riskcalc.
type Policy struct {
	// MaxOrderQuantity is the per-order unit ceiling, inclusive.
	MaxOrderQuantity string
	// MaxOrderNotional is the per-order quantity×price ceiling, inclusive.
	MaxOrderNotional riskcalc.Money
	// MaxOpenExposure is the account-wide ceiling. The boundary is ≥.
	MaxOpenExposure riskcalc.Money
	// MaxDailyLoss is the absolute daily realised-loss ceiling. Boundary ≥.
	MaxDailyLoss riskcalc.Money
	// MaxDailyLossRatio is the same ceiling as a fraction of equity, in (0, 1].
	MaxDailyLossRatio string
	// RiskBudget is what one trade may lose if its stop is hit: the numerator of
	// floor(위험예산 / (entry − stop)).
	RiskBudget riskcalc.Money
	// MinRewardRisk is the minimum reward:risk. Default DefaultMinRewardRisk.
	MinRewardRisk string
	// Reentry bounds same-day re-entry.
	Reentry ReentryPolicy
}

// DefaultPolicy is the conservative default set risk-management specifies for an
// operator who has not chosen numbers: 주문당 notional 1,000,000 KRW·주문당 수량
// 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW.
//
// Those six are transcribed. Two more are needed to evaluate a chain and are
// **derived** rather than measured — both are recorded here because a number
// with no provenance is a number nobody can review:
//
//   - RiskBudget = MaxDailyLoss. Not a sizing recommendation: it is the largest
//     per-trade risk the daily limit already implies, so using it as the chain's
//     ceiling adds no permission that was not there. An issuer that sizes
//     smaller is free to (§0.9 conservative direction); one that sizes larger is
//     refused.
//   - Reentry = 2 entries, 30-minute cooldown `[미검증]`, the values
//     risk-management names. StockOS's own defaults are stricter still
//     (guardian.py:24-29: 1 entry, 30 minutes).
func DefaultPolicy() Policy {
	return Policy{
		MaxOrderQuantity:  "100",
		MaxOrderNotional:  riskcalc.Money{Amount: "1000000", Currency: "KRW"},
		MaxOpenExposure:   riskcalc.Money{Amount: "10000000", Currency: "KRW"},
		MaxDailyLoss:      riskcalc.Money{Amount: "100000", Currency: "KRW"},
		MaxDailyLossRatio: "0.01",
		RiskBudget:        riskcalc.Money{Amount: "100000", Currency: "KRW"},
		MinRewardRisk:     DefaultMinRewardRisk,
		Reentry:           ReentryPolicy{MaxEntriesPerSymbol: 2, Cooldown: 30 * time.Minute},
	}
}

// Validate refuses a policy the chain cannot evaluate against.
//
// It is called before any check runs, so a misconfigured limit surfaces as one
// INPUT_UNAVAILABLE at the top rather than as an arbitrary refusal several rungs
// down. Every rule here is fail-closed: an unset ceiling is not an unlimited
// one, and a zero one is not "no limit".
func (p Policy) Validate() error {
	currency, err := requirePositiveMoney("maximum order notional", p.MaxOrderNotional, "")
	if err != nil {
		return err
	}
	for _, m := range []struct {
		label string
		value riskcalc.Money
	}{
		{"maximum open exposure", p.MaxOpenExposure},
		{"maximum daily loss", p.MaxDailyLoss},
		{"per-trade risk budget", p.RiskBudget},
	} {
		if _, err := requirePositiveMoney(m.label, m.value, currency); err != nil {
			return err
		}
	}

	quantity, err := parseWholeNumber("maximum order quantity", p.MaxOrderQuantity)
	if err != nil {
		return err
	}
	if quantity.Sign() <= 0 {
		return fmt.Errorf("maximum order quantity %q is not positive; an unset ceiling is not an unlimited one",
			p.MaxOrderQuantity)
	}

	ratio, err := parseDecimal("maximum daily loss ratio", p.MaxDailyLossRatio)
	if err != nil {
		return err
	}
	if ratio.Sign() <= 0 || ratio.Cmp(big.NewRat(1, 1)) > 0 {
		return fmt.Errorf("maximum daily loss ratio %q is outside (0, 1]", p.MaxDailyLossRatio)
	}

	minRR, err := parseDecimal("minimum reward:risk", p.MinRewardRisk)
	if err != nil {
		return err
	}
	if minRR.Sign() <= 0 {
		return fmt.Errorf("minimum reward:risk %q is not positive", p.MinRewardRisk)
	}

	if p.Reentry.MaxEntriesPerSymbol < 1 {
		return fmt.Errorf("same-day entry ceiling %d is below one; zero would refuse every entry silently",
			p.Reentry.MaxEntriesPerSymbol)
	}
	if p.Reentry.Cooldown < 0 {
		return fmt.Errorf("same-day cooldown %s is negative", p.Reentry.Cooldown)
	}
	return nil
}

// LimitCurrency returns the currency every money limit is expressed in.
func (p Policy) LimitCurrency() string {
	return strings.ToUpper(strings.TrimSpace(p.MaxOrderNotional.Currency))
}

// requirePositiveMoney checks one configured ceiling: present, readable,
// positive, and — when want is non-empty — in the expected currency.
func requirePositiveMoney(label string, m riskcalc.Money, want string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(m.Currency))
	if currency == "" {
		return "", fmt.Errorf("%s carries no currency; an amount nobody named a currency for cannot be compared", label)
	}
	if want != "" && currency != want {
		return "", fmt.Errorf("%s is in %s but the other limits are in %s; normalise before configuring, never inside a comparison",
			label, currency, want)
	}
	value, err := parseDecimal(label, m.Amount)
	if err != nil {
		return "", err
	}
	if value.Sign() <= 0 {
		return "", fmt.Errorf("%s %q is not positive; an unset ceiling is not an unlimited one", label, m.Amount)
	}
	return currency, nil
}

// currencyOf returns the currency a market's prices are quoted in.
//
// It is a two-row table rather than a configuration value: a KR order priced in
// anything but KRW, or a US order in anything but USD, is not a thing the broker
// can accept. Task 4.x still has to convert between them for a cross-currency
// limit, and riskcalc — with its FX staleness bound — is where that happens.
func currencyOf(market costs.Market) (string, error) {
	switch market {
	case costs.MarketKR:
		return "KRW", nil
	case costs.MarketUS:
		return "USD", nil
	default:
		return "", fmt.Errorf("market %q is not one this build prices; a mislabelled market is never defaulted", string(market))
	}
}
