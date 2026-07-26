// Package costs is the trade-cost model: what a round trip costs, and what a
// sale has to fetch before it stops being a loss.
//
// It is the artifact trade-analytics asks for — "KRW·USD 거래의 수수료·거래세·
// 환전 비용은 순수 함수 비용 모델로 계산되어야 한다(SHALL)" — and it is the only
// one: Guardian's cash check, the exit policy's 실질 본전 and the performance
// record's cost deduction all read this model (SHALL — 이중 정의 금지).
//
// # What was ported from StockOS, and what was not
//
// The shape comes from
// `/mnt/D/project/axipient/stockos/packages/trading/stockos_trading/costs.py`:
// per-market commission and tax rates (:72-116), the break-even sell price
// (:240-266), and the validation gate every configured rate passes through
// (:179-206) under the 5% cap (:69). Four things deliberately did not come
// across:
//
//  1. **KIS's numbers and `KIS_*` naming** (costs.py:11-21, :74-90). SHALL NOT —
//     they describe another broker's rate sheet, and a number copied from it
//     would read as measured when nobody measured it here. The rates below are
//     TossOS placeholders tagged `[미검증]`.
//  2. **Env-variable overrides** (costs.py:118-176). Overrides arrive as
//     injected configuration instead; see Overrides below.
//  3. **The KRX board split** (KOSPI/KOSDAQ/KONEX, costs.py:76-78). TossOS's
//     market vocabulary is internal/clock.Market — "kr" and "us", with no board
//     dimension anywhere in the journal, the intents or the broker client. The
//     single KR rate is therefore set to the conservative maximum over the
//     boards rather than to any one board's rate. A board dimension can be added
//     when something in TossOS can tell the boards apart.
//  4. **FX conversion of a profit floor** (costs.py:269-284). Currency
//     normalisation, and the staleness bound that makes it safe, belong to
//     internal/riskcalc (FXRateStaleness). Re-implementing it here with an
//     unbounded rate would put a second, weaker conversion in the tree. The
//     floor is taken in the trade's own currency.
//
// # The numbers here are placeholders, and the cap is what makes that safe
//
// Every rate is an over-estimate awaiting 2b 실측 (trade-analytics: 수치는 2b
// 실측으로 채우며, 실측 전에는 상한 이내의 과대 추정 보수값을 "미검증"
// provenance와 함께 사용한다). Over-estimating is the conservative direction on
// the *entry* side — it makes a trade look more expensive, so fewer trades pass
// the cash check. It is not conservative on the exit side, which is why two
// rules bound it:
//
//   - MaxRate caps every configured rate (exit-policy: 상한 없는 과대 추정은
//     본전 기준선을 무한히 끌어올려 승자 포지션을 즉시 청산시킨다), and
//   - nothing in TossOS may use this model as a liquidation gate (§0.3 — see
//     sellgate_test.go, which pins that structurally).
//
// # Decimals
//
// Amounts arrive and leave as decimal strings, the convention internal/riskcalc
// and internal/journal already use for every broker decimal. Arithmetic is done
// in float64 after parsing and results are rendered with
// `strconv.FormatFloat(v, 'f', -1, 64)`, exactly as riskcalc's valuation does —
// a cost is a valuation (a rate times an amount), not a ledger entry, and its
// precision claim is the same bounded-relative-error one riskcalc documents.
// Rates themselves are float64: they are configuration fractions, never money.
//
// Everything here is a pure function of its inputs. There is no clock, no
// configuration lookup and no network client in this package, and no journal
// import — a cost that changed depending on when it was asked would not be a
// cost model.
package costs

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// MaxRate is the inclusive ceiling on any configured rate, ported from
// StockOS costs.py:69.
//
// The bound is not a tuning knob. It exists because the two realistic
// misconfigurations are catastrophic in opposite directions: an off-by-100 typo
// (0.20 where 0.0020 was meant) makes every entry look unaffordable, and — worse
// — it drags the 실질 본전 baseline so far above the entry that the exit policy
// closes a winning position the moment it opens. 5% is far above any real
// commission or tax and far below the values a typo produces, so it separates
// the two cleanly.
//
// Raising it is a §0.9 decision with an audit record, and it must be raised in
// code: configuration cannot bypass it.
const MaxRate = 0.05

// Market is the market a cost is computed for.
//
// It is internal/clock's vocabulary rather than a second one. Two market enums
// in one program is two chances to disagree about what "us" means, and the
// disagreement would be silent: a cost computed under the wrong market is still
// a number.
type Market = clock.Market

const (
	// MarketKR is the Korean market. Commission plus the sell-side
	// transaction tax.
	MarketKR = clock.MarketKR
	// MarketUS is the US market. Commission plus sell-side regulatory fees
	// plus the FX conversion cost on both legs.
	MarketUS = clock.MarketUS
)

// Side is the leg being costed. The transaction tax and the regulatory fee are
// sell-side only; commission and the FX conversion cost apply to both.
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// Sentinel errors. Every failure here is fail-closed: the caller gets no
// number at all rather than a plausible-looking one.
var (
	// ErrInvalidRate: a configured rate did not pass the validation gate.
	// Startup refuses rather than trading on it.
	ErrInvalidRate = errors.New("costs: invalid rate")
	// ErrUnknownMarket: a market this model has no rates for. It is never
	// defaulted to KR — a mislabelled market must stop the caller.
	ErrUnknownMarket = errors.New("costs: unknown market")
	// ErrUnknownSide: a leg that is neither BUY nor SELL.
	ErrUnknownSide = errors.New("costs: unknown side")
	// ErrInvalidAmount: an amount that is absent, unreadable, non-finite or
	// outside its allowed sign.
	ErrInvalidAmount = errors.New("costs: invalid amount")
)

// Override keys. These are the configuration surface: a config block injects a
// map of these keys to decimal strings, and NewModel validates each one.
//
// They are stable strings because they end up in config files and in audit
// records of limit changes (§0.5). Renaming one silently drops an operator's
// override, so add a key rather than rename one.
const (
	KeyKRBuyCommissionRate     = "kr.buy_commission_rate"
	KeyKRSellCommissionRate    = "kr.sell_commission_rate"
	KeyKRSellTaxRate           = "kr.sell_tax_rate"
	KeyUSBuyCommissionRate     = "us.buy_commission_rate"
	KeyUSSellCommissionRate    = "us.sell_commission_rate"
	KeyUSSellRegulatoryFeeRate = "us.sell_regulatory_fee_rate"
	KeyUSFXConversionFeeRate   = "us.fx_conversion_fee_rate"
)

// Model holds one validated set of rates.
//
// The fields are unexported and the type is comparable: a Model can only be
// produced by DefaultModel or by NewModel, so a rate that never passed the gate
// cannot exist inside one, and two models can be compared for equality in a
// test without reflection.
type Model struct {
	// configured distinguishes "every rate is zero because that is what the
	// operator configured" from "this is the zero value and nobody configured
	// anything". The two are numerically identical and operationally opposite:
	// an unconfigured model prices every trade as free, which understates the
	// cash a trade needs and drops the break-even price onto the entry. Callers
	// that judge anything on cost check Configured first.
	configured bool

	krBuyCommission  float64
	krSellCommission float64
	krSellTax        float64
	usBuyCommission  float64
	usSellCommission float64
	usSellRegulatory float64
	usFXConversion   float64
}

// DefaultModel returns the conservative placeholder rates.
//
// # Provenance — every number below is `[미검증]`
//
// None of these is a measured Toss rate. They are deliberate over-estimates,
// chosen an order of magnitude above the published retail figures they stand in
// for, so that the entry-side cash check errs toward refusing a trade. They are
// replaced by 2b 실측 (proposal: 수치는 2b 실측으로 채우며), and until then
// nothing in TossOS may describe them as verified (risk-management: "Toss
// 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다 — SHALL NOT).
//
// The *structure* they fill is StockOS's (costs.py:72-116): a per-market buy and
// sell commission, a sell-side tax or regulatory fee, and an FX conversion cost
// that applies to both legs of a foreign trade. The KIS values that structure
// carried (costs.py:74-90) are not reproduced.
func DefaultModel() Model {
	return Model{
		configured: true,
		// KR commission: over-estimate of a domestic retail online rate.
		// `[미검증 — 2b 실측 대체 대상]`
		krBuyCommission:  0.0010,
		krSellCommission: 0.0010,
		// KR sell-side transaction tax. Over-estimate of the statutory rate,
		// set above the highest KRX board rather than per board (see the
		// package comment on why there is no board dimension).
		// `[미검증 — 2b 실측 대체 대상]`
		krSellTax: 0.0030,
		// US commission: over-estimate of an overseas retail online rate.
		// `[미검증 — 2b 실측 대체 대상]`
		usBuyCommission:  0.0050,
		usSellCommission: 0.0050,
		// US sell-side regulatory fees (the SEC/TAF-shaped charges). Small and
		// broker-specific; over-estimated rather than assumed zero, because a
		// zero here would understate the break-even price.
		// `[미검증 — 2b 실측 대체 대상]`
		usSellRegulatory: 0.0010,
		// FX conversion spread on a USD trade, charged on both legs.
		// `[미검증 — 2b 실측 대체 대상]`
		usFXConversion: 0.0050,
	}
}

// overrideKeys is the registry, in a stable order. It is the single source of
// truth for "which rates can be configured": NewModel iterates it, and the test
// sweep iterates it, so a rate added to Model without being wired here shows up
// as a test failure rather than as a silently unconfigurable rate.
var overrideKeys = []string{
	KeyKRBuyCommissionRate,
	KeyKRSellCommissionRate,
	KeyKRSellTaxRate,
	KeyUSBuyCommissionRate,
	KeyUSSellCommissionRate,
	KeyUSSellRegulatoryFeeRate,
	KeyUSFXConversionFeeRate,
}

// OverrideKeys returns the configurable rate keys, in a stable order.
func OverrideKeys() []string {
	out := make([]string, len(overrideKeys))
	copy(out, overrideKeys)
	return out
}

// NewModel applies configured overrides to the defaults, validating every one.
//
// This is the injection seam StockOS spells `FeeTaxModel.from_env`
// (costs.py:118-167), with the env map replaced by injected configuration. Two
// rules differ from the original, both deliberately stricter:
//
//   - An **unrecognised key is refused**. StockOS reads process env, where
//     unrelated keys are the norm; a TossOS override map is a dedicated config
//     block, so an unknown key is a typo — and a typo that silently keeps the
//     default is a rate the operator believes they set and did not.
//   - An **empty value is refused** (as in StockOS): a key present with no value
//     means the deployment artifact is broken, not that the default was wanted.
//
// A refusal is fatal at startup by design: the caller propagates it rather than
// trading on a rate nobody can read (costs.py module docstring, "Failure mode").
func NewModel(overrides map[string]string) (Model, error) {
	m := DefaultModel()
	if len(overrides) == 0 {
		return m, nil
	}

	// Refuse unknown keys first, in a deterministic order, so an operator with
	// several typos sees the same one named on every run.
	unknown := make([]string, 0, len(overrides))
	for key := range overrides {
		if !isOverrideKey(key) {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return Model{}, fmt.Errorf("%w: %q is not a cost override key (known keys: %s)",
			ErrInvalidRate, unknown[0], strings.Join(overrideKeys, ", "))
	}

	for _, key := range overrideKeys {
		raw, ok := overrides[key]
		if !ok {
			continue
		}
		rate, err := parseRate(key, raw)
		if err != nil {
			return Model{}, err
		}
		if err := m.set(key, rate); err != nil {
			return Model{}, err
		}
	}
	return m, nil
}

func isOverrideKey(key string) bool {
	for _, k := range overrideKeys {
		if k == key {
			return true
		}
	}
	return false
}

// set writes one validated rate. It switches on the same constants Rate reads,
// which is what keeps the two directions from drifting apart.
func (m *Model) set(key string, rate float64) error {
	switch key {
	case KeyKRBuyCommissionRate:
		m.krBuyCommission = rate
	case KeyKRSellCommissionRate:
		m.krSellCommission = rate
	case KeyKRSellTaxRate:
		m.krSellTax = rate
	case KeyUSBuyCommissionRate:
		m.usBuyCommission = rate
	case KeyUSSellCommissionRate:
		m.usSellCommission = rate
	case KeyUSSellRegulatoryFeeRate:
		m.usSellRegulatory = rate
	case KeyUSFXConversionFeeRate:
		m.usFXConversion = rate
	default:
		return fmt.Errorf("%w: %q has no field in the model", ErrInvalidRate, key)
	}
	return nil
}

// Configured reports whether this model came from DefaultModel or NewModel.
//
// The zero Model is not a free-trade model, it is an absent one, and a caller
// that would refuse a trade on cost grounds must refuse an absent model instead
// of pricing everything at zero. (Callers that must never refuse on cost —
// liquidations, §0.3 — do not ask this question at all.)
func (m Model) Configured() bool { return m.configured }

// Rate returns the configured rate for an override key.
//
// An unknown key returns NaN rather than zero: a rate nobody defined is not a
// free trade, and NaN propagates into any arithmetic that ignored this comment.
func (m Model) Rate(key string) float64 {
	switch key {
	case KeyKRBuyCommissionRate:
		return m.krBuyCommission
	case KeyKRSellCommissionRate:
		return m.krSellCommission
	case KeyKRSellTaxRate:
		return m.krSellTax
	case KeyUSBuyCommissionRate:
		return m.usBuyCommission
	case KeyUSSellCommissionRate:
		return m.usSellCommission
	case KeyUSSellRegulatoryFeeRate:
		return m.usSellRegulatory
	case KeyUSFXConversionFeeRate:
		return m.usFXConversion
	default:
		return math.NaN()
	}
}

// parseRate is the validation gate, ported from costs.py:179-206.
//
// The five rules, in the order a bad value is most likely to hit them:
//
//  1. Empty or blank — the operator set the key and forgot the value.
//  2. Not a number.
//  3. Not finite (NaN, ±Inf). Go's ParseFloat accepts "nan" and "inf" happily,
//     exactly as Python's float() does, so this check is not redundant.
//  4. Negative — it would flip the sign of P&L and of the break-even price,
//     making a misconfiguration read as a trade that pays you to place it.
//  5. Above MaxRate — the off-by-100 typo guard.
func parseRate(key, value string) (float64, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("%w: %s must be a numeric rate (got an empty value)", ErrInvalidRate, key)
	}
	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be a numeric rate (got %q)", ErrInvalidRate, key, value)
	}
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return 0, fmt.Errorf("%w: %s must be finite (got %q)", ErrInvalidRate, key, value)
	}
	if rate < 0 {
		return 0, fmt.Errorf("%w: %s must be non-negative (got %q)", ErrInvalidRate, key, value)
	}
	if rate > MaxRate {
		return 0, fmt.Errorf(
			"%w: %s must be at most %v (got %q) — a rate above the cap is treated as a typo; "+
				"check it is a decimal fraction and not a percentage",
			ErrInvalidRate, key, MaxRate, value)
	}
	return rate, nil
}

// --- per-market rate accessors (costs.py:92-116) -----------------------------

// BuyCommissionRate returns the buy-leg commission rate for a market.
func (m Model) BuyCommissionRate(market Market) (float64, error) {
	switch market {
	case MarketKR:
		return m.krBuyCommission, nil
	case MarketUS:
		return m.usBuyCommission, nil
	default:
		return 0, unknownMarket(market)
	}
}

// SellCommissionRate returns the sell-leg commission rate for a market.
func (m Model) SellCommissionRate(market Market) (float64, error) {
	switch market {
	case MarketKR:
		return m.krSellCommission, nil
	case MarketUS:
		return m.usSellCommission, nil
	default:
		return 0, unknownMarket(market)
	}
}

// SellTaxRate returns the sell-side statutory charge: the transaction tax in KR,
// the regulatory fees in US. They are one field because they occupy the same
// place in the arithmetic — a sell-only levy on notional.
func (m Model) SellTaxRate(market Market) (float64, error) {
	switch market {
	case MarketKR:
		return m.krSellTax, nil
	case MarketUS:
		return m.usSellRegulatory, nil
	default:
		return 0, unknownMarket(market)
	}
}

// FXConversionFeeRate returns the currency-conversion cost charged on each leg
// of a trade in a foreign currency. Zero for a domestic market.
func (m Model) FXConversionFeeRate(market Market) (float64, error) {
	switch market {
	case MarketKR:
		return 0, nil
	case MarketUS:
		return m.usFXConversion, nil
	default:
		return 0, unknownMarket(market)
	}
}

// SellSideRate is the fraction of the sale proceeds that never reaches the
// account: commission plus the statutory charge plus the FX conversion cost.
//
// It is the denominator term of the break-even price, which is why it is one
// named function rather than three additions repeated at each call site.
func (m Model) SellSideRate(market Market) (float64, error) {
	commission, err := m.SellCommissionRate(market)
	if err != nil {
		return 0, err
	}
	tax, err := m.SellTaxRate(market)
	if err != nil {
		return 0, err
	}
	fx, err := m.FXConversionFeeRate(market)
	if err != nil {
		return 0, err
	}
	return commission + tax + fx, nil
}

func unknownMarket(market Market) error {
	return fmt.Errorf("%w: market %q has no rates; a cost is never defaulted to another market's",
		ErrUnknownMarket, string(market))
}

// --- costing -----------------------------------------------------------------

// TradeCost is one leg's cost, broken out so an audit record can show which part
// of a refusal was tax and which was commission. All four are decimal strings in
// the trade's own currency.
type TradeCost struct {
	Commission string
	Tax        string
	FXFee      string
	Total      string
}

// EstimateTradeCost costs one leg of a trade (costs.py:220-237).
//
// The buy leg carries commission and the FX conversion cost; the sell leg adds
// the statutory charge. A negative notional is refused rather than costed: it
// would produce a negative cost, which is a discount nobody offers.
func (m Model) EstimateTradeCost(notional string, side Side, market Market) (TradeCost, error) {
	value, err := parseNonNegative("notional", notional)
	if err != nil {
		return TradeCost{}, err
	}
	commissionRate, err := m.commissionRateFor(side, market)
	if err != nil {
		return TradeCost{}, err
	}
	fxRate, err := m.FXConversionFeeRate(market)
	if err != nil {
		return TradeCost{}, err
	}
	tax := 0.0
	if side == SideSell {
		taxRate, err := m.SellTaxRate(market)
		if err != nil {
			return TradeCost{}, err
		}
		tax = value * taxRate
	}
	commission := value * commissionRate
	fx := value * fxRate
	return TradeCost{
		Commission: formatAmount(commission),
		Tax:        formatAmount(tax),
		FXFee:      formatAmount(fx),
		Total:      formatAmount(commission + tax + fx),
	}, nil
}

func (m Model) commissionRateFor(side Side, market Market) (float64, error) {
	switch side {
	case SideBuy:
		return m.BuyCommissionRate(market)
	case SideSell:
		return m.SellCommissionRate(market)
	default:
		return 0, fmt.Errorf("%w: side %q is neither %s nor %s", ErrUnknownSide, string(side), SideBuy, SideSell)
	}
}

// BreakEvenSellPrice is the price a holding must fetch for the round trip to net
// zero: 실질 본전 (costs.py:240-266, with min_profit_krw = 0).
//
// This is the strict fee-and-tax floor, and it is what Guardian's
// TARGET_BELOW_BREAK_EVEN check and the exit policy's BREAKEVEN ratchet level
// both read. A target below it describes a trade that cannot make money even
// when it works.
func (m Model) BreakEvenSellPrice(buyPrice, quantity string, market Market) (string, error) {
	return m.BreakEvenSellPriceWithFloor(buyPrice, quantity, market, "0")
}

// BreakEvenSellPriceWithFloor adds a required profit, in the trade's own
// currency, on top of the costs.
//
// The formula is StockOS's, and the division is the whole point of it: the sale
// itself is taxed, so recovering `buy + costs` requires grossing up by
// `1 / (1 − sell-side rate)`. Adding the sell-side cost instead of dividing by
// its complement understates the break-even price, and understating it is how a
// "break-even" exit books a loss.
//
//	break_even = (buy_notional + buy_cost + required_profit) / (quantity × (1 − sell_rate))
//
// The min-profit floor is native-currency here; StockOS converts a KRW floor
// into USD inside this function (costs.py:269-284) and TossOS does not — see the
// package comment.
func (m Model) BreakEvenSellPriceWithFloor(buyPrice, quantity string, market Market, minProfit string) (string, error) {
	price, err := parsePositive("buy price", buyPrice)
	if err != nil {
		return "", err
	}
	qty, err := parsePositive("quantity", quantity)
	if err != nil {
		return "", err
	}
	profit, err := parseNonNegative("minimum profit", minProfit)
	if err != nil {
		return "", err
	}
	buyNotional := price * qty
	buyCost, err := m.EstimateTradeCost(formatAmount(buyNotional), SideBuy, market)
	if err != nil {
		return "", err
	}
	buyTotal, err := parseNonNegative("buy cost", buyCost.Total)
	if err != nil {
		return "", err
	}
	sellRate, err := m.SellSideRate(market)
	if err != nil {
		return "", err
	}
	// Structurally impossible while MaxRate < 1/3 caps each of the three
	// sell-side components, and checked anyway: at a sell-side rate of 1 the
	// division below is by zero, and above 1 it silently changes sign.
	if sellRate >= 1 {
		return "", fmt.Errorf("%w: sell-side rate for %q is %v; a break-even price is undefined at or above 1",
			ErrInvalidRate, string(market), sellRate)
	}
	return formatAmount((buyNotional + buyTotal + profit) / (qty * (1 - sellRate))), nil
}

// NetPnL is the realised result of a round trip after both legs' costs
// (costs.py:287-301). It is signed: a loss is negative.
func (m Model) NetPnL(buyPrice, sellPrice, quantity string, market Market) (string, error) {
	buy, err := parseNonNegative("buy price", buyPrice)
	if err != nil {
		return "", err
	}
	sell, err := parseNonNegative("sell price", sellPrice)
	if err != nil {
		return "", err
	}
	qty, err := parsePositive("quantity", quantity)
	if err != nil {
		return "", err
	}
	buyNotional := buy * qty
	sellNotional := sell * qty
	buyCost, err := m.EstimateTradeCost(formatAmount(buyNotional), SideBuy, market)
	if err != nil {
		return "", err
	}
	sellCost, err := m.EstimateTradeCost(formatAmount(sellNotional), SideSell, market)
	if err != nil {
		return "", err
	}
	buyTotal, err := parseNonNegative("buy cost", buyCost.Total)
	if err != nil {
		return "", err
	}
	sellTotal, err := parseNonNegative("sell cost", sellCost.Total)
	if err != nil {
		return "", err
	}
	return formatAmount(sellNotional - buyNotional - buyTotal - sellTotal), nil
}

// --- decimal-string plumbing --------------------------------------------------
//
// These mirror internal/riskcalc's parseAmount / parseNonNegative /
// formatAmount, which are unexported there. The duplication is a handful of
// lines and buys this package a stdlib-only dependency set; what matters is that
// the *rules* match, because the two packages' numbers are added together at the
// call site: an empty decimal is unknown rather than zero, a non-finite one is
// unusable, and results are rendered as the shortest string that round-trips.

func parseAmount(field, value string) (float64, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("%w: %s is empty; an empty decimal is unknown, not zero", ErrInvalidAmount, field)
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s %q is not a decimal", ErrInvalidAmount, field, value)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%w: %s %q is not finite", ErrInvalidAmount, field, value)
	}
	return v, nil
}

func parseNonNegative(field, value string) (float64, error) {
	v, err := parseAmount(field, value)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("%w: %s %q is negative", ErrInvalidAmount, field, strings.TrimSpace(value))
	}
	return v, nil
}

func parsePositive(field, value string) (float64, error) {
	v, err := parseAmount(field, value)
	if err != nil {
		return 0, err
	}
	if v <= 0 {
		return 0, fmt.Errorf("%w: %s %q must be positive", ErrInvalidAmount, field, strings.TrimSpace(value))
	}
	return v, nil
}

func formatAmount(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
