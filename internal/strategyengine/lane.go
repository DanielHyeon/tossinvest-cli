package strategyengine

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
)

type ParkerConservativeLane struct{}

func (ParkerConservativeLane) Evaluate(in LaneInput) Evaluation {
	refuse := func(reason Refusal) Evaluation {
		return Evaluation{Reason: reason, SourceReason: sourceRefusalReason(reason)}
	}
	refuseAs := func(reason Refusal, sourceReason string) Evaluation {
		return Evaluation{Reason: reason, SourceReason: sourceReason}
	}
	if !in.Approved.Valid() {
		return refuse(RefusalCandidate)
	}
	if in.Approved.Market() != "KR" || in.Approved.Symbol() == "" {
		return refuse(RefusalUnsupportedScope)
	}
	if !in.Source.Valid() {
		return refuse(RefusalSource)
	}
	market := in.Market
	if !market.valid || market.market != in.Approved.Market() || market.symbol != in.Approved.Symbol() {
		return refuse(RefusalIndicator)
	}
	evaluatedAt := market.evaluatedAt.UTC()
	approvedAt := time.Unix(0, in.Approved.ApprovedAtUnixNano()).UTC()
	lastSeenAt := time.Unix(0, in.Approved.LastSeenUnixNano()).UTC()
	validUntil := time.Unix(0, in.Approved.ValidUntilUnixNano()).UTC()
	if in.Approved.State() != "active" || approvedAt.After(evaluatedAt) ||
		evaluatedAt.Before(lastSeenAt) || !evaluatedAt.Before(validUntil) {
		return refuse(RefusalCandidate)
	}
	if !market.tradingDay {
		return refuseAs(RefusalSession, SourceRejectNonTradingDay)
	}
	if evaluatedAt.Before(market.sessionOpenAt) || evaluatedAt.After(market.sessionCloseAt) {
		return refuseAs(RefusalSession, SourceRejectAfterHours)
	}
	if evaluatedAt.Before(market.sessionOpenAt.Add(10 * time.Minute)) {
		return refuseAs(RefusalSession, SourceRejectOpeningWindow)
	}
	if evaluatedAt.After(market.noEntryAfter) {
		return refuseAs(RefusalSession, SourceRejectAfterEntryCutoff)
	}
	if !in.Bar.Valid() || in.Bar.Market() != market.market || in.Bar.Symbol() != market.symbol ||
		in.Bar.Source() != string(strategymarket.SourceOfficialOpenAPI) || in.Bar.Adjusted() {
		return refuse(RefusalBarIntegrity)
	}
	if in.Bar.ClosedAt().After(evaluatedAt) || in.Bar.ClosedAt().After(market.sessionCloseAt) {
		return refuse(RefusalSession)
	}
	if !in.State.Valid() || in.State.Market() != "KR" || in.State.Symbol() != in.Approved.Symbol() {
		return refuse(RefusalSymbolState)
	}
	if !in.Position.Valid() || in.Position.Market() != "KR" || in.Position.Symbol() != in.Approved.Symbol() {
		return refuse(RefusalExistingPosition)
	}

	open, openOK := positive(in.Bar.Open())
	high, highOK := positive(in.Bar.High())
	low, lowOK := positive(in.Bar.Low())
	closePrice, closeOK := positive(in.Bar.Close())
	volume, volumeOK := nonnegative(in.Bar.Volume())
	if !(openOK && highOK && lowOK && closeOK && volumeOK) ||
		high.Cmp(low) < 0 || high.Cmp(open) < 0 || high.Cmp(closePrice) < 0 ||
		low.Cmp(open) > 0 || low.Cmp(closePrice) > 0 {
		return refuse(RefusalInvalidBar)
	}
	if volume.Sign() == 0 {
		return refuse(RefusalIlliquidBar)
	}

	vwap, vwapOK := positive(market.vwap)
	slope, slopeOK := decimal(market.vwapSlopePct)
	ema9, ema9OK := positive(market.ema9)
	lvn, lvnOK := decimal(market.lvnForwardSpacePct)
	tangled, tangledOK := nonnegative(market.tangledScorePct)
	if !(vwapOK && slopeOK && ema9OK && lvnOK && tangledOK) {
		return refuse(RefusalIndicator)
	}
	if closePrice.Cmp(vwap) <= 0 {
		return refuse(RefusalVWAPAbove)
	}
	if slope.Cmp(rat("0.08")) < 0 {
		return refuse(RefusalVWAPSlope)
	}
	touchCeiling := new(big.Rat).Mul(ema9, rat("1.0025"))
	if closePrice.Cmp(ema9) <= 0 || low.Cmp(touchCeiling) > 0 {
		return refuse(RefusalEMA9Pullback)
	}
	if closePrice.Cmp(open) <= 0 {
		return refuse(RefusalFakeBreakout)
	}
	if lvn.Cmp(rat("1.2")) < 0 {
		return refuse(RefusalLVNSpace)
	}
	if tangled.Cmp(rat("0.35")) < 0 {
		return refuse(RefusalTangledBand)
	}
	if market.bandExpansionRate != "" {
		expansion, ok := decimal(market.bandExpansionRate)
		if !ok {
			return refuse(RefusalIndicator)
		}
		if expansion.Cmp(rat("1.8")) > 0 {
			return refuse(RefusalBandExpansion)
		}
	}

	stop := new(big.Rat).Mul(closePrice, rat("0.993"))
	risk := new(big.Rat).Sub(closePrice, stop)
	forwardDistance := new(big.Rat).Quo(new(big.Rat).Mul(closePrice, lvn), rat("100"))
	expectedRR := new(big.Rat).Quo(forwardDistance, risk)
	if expectedRR.Cmp(rat("1.5")) < 0 {
		return refuse(RefusalRR)
	}
	if market.hvnAboveDistancePct != "" {
		hvnDistance, ok := decimal(market.hvnAboveDistancePct)
		if !ok {
			return refuse(RefusalIndicator)
		}
		if hvnDistance.Cmp(lvn) < 0 {
			return refuse(RefusalHVNCeiling)
		}
	}
	age := evaluatedAt.Sub(in.Bar.ClosedAt().UTC())
	if age < 0 || age > 15*time.Second {
		return refuse(RefusalAge)
	}
	currentPrice := new(big.Rat).Set(closePrice)
	if market.currentPrice != "" {
		live, ok := decimal(market.currentPrice)
		if !ok || live.Sign() <= 0 {
			return refuse(RefusalDrift)
		}
		currentPrice = live
	}
	drift := new(big.Rat).Sub(currentPrice, closePrice)
	if drift.Sign() < 0 {
		drift.Neg(drift)
	}
	drift.Quo(drift, closePrice)
	drift.Mul(drift, rat("100"))
	if drift.Cmp(rat("0.20")) > 0 {
		return refuse(RefusalDrift)
	}

	target := new(big.Rat).Add(closePrice, new(big.Rat).Mul(risk, rat("3.0")))
	expansion := market.bandExpansionRate
	hvnDistance := market.hvnAboveDistancePct
	record := DecisionRecord{
		CandidateLifeID:     in.Approved.CandidateLifeID(),
		CandidateState:      in.Approved.State(),
		CandidateFirstSeen:  in.Approved.FirstSeenUnixNano(),
		CandidateLastSeen:   in.Approved.LastSeenUnixNano(),
		CandidateValidUntil: in.Approved.ValidUntilUnixNano(),
		CandidateApprovedAt: in.Approved.ApprovedAtUnixNano(),
		Market:              "KR",
		Symbol:              in.Approved.Symbol(),
		LaneID:              LaneID,
		LaneVersion:         LaneVersion,
		SourceCommit:        SourceCommit,
		SourceDigest:        in.Source.Digest(),
		ConstantsDigest:     constantsDigest(),
		ThresholdVersion:    in.Approved.ThresholdVersion(),
		ThresholdSetDigest:  in.Approved.SetDigest(),
		EvidenceDigest:      in.Approved.EvidenceDigest(),
		MarketInputVersion:  market.version,
		CalendarSource:      market.calendarSource,
		CalendarVersion:     market.calendarVersion,
		ConfigSource:        market.configSource,
		ConfigVersion:       market.configVersion,
		IndicatorSource:     market.indicatorSource,
		IndicatorVersion:    market.indicatorVersion,
		IndicatorComputedAt: market.indicatorComputedAt.UnixNano(),
		TradingDay:          market.tradingDay,
		SessionOpenAt:       market.sessionOpenAt.UnixNano(),
		SessionCloseAt:      market.sessionCloseAt.UnixNano(),
		NoEntryAfter:        market.noEntryAfter.UnixNano(),
		BarSource:           in.Bar.Source(),
		BarAdjusted:         in.Bar.Adjusted(),
		BarOpenAt:           in.Bar.OpenAt().UTC().UnixNano(),
		BarClosedAt:         in.Bar.ClosedAt().UTC().UnixNano(),
		EvaluatedAt:         evaluatedAt.UnixNano(),
		ExpiresAt:           in.Bar.ClosedAt().UTC().Add(15*time.Second + time.Nanosecond).UnixNano(),
		Open:                decimalString(open),
		High:                decimalString(high),
		Low:                 decimalString(low),
		Close:               decimalString(closePrice),
		Volume:              in.Bar.Volume(),
		Currency:            in.Bar.Currency(),
		VWAP:                decimalString(vwap),
		VWAPSlopePct:        decimalString(slope),
		EMA9:                decimalString(ema9),
		LVNSpacePct:         decimalString(lvn),
		TangledPct:          decimalString(tangled),
		Expansion:           expansion,
		HVNAboveDistancePct: hvnDistance,
		StateSource:         in.State.Authority(),
		StateAt:             in.State.ObservedAt().UTC().UnixNano(),
		PositionSource:      in.Position.Authority(),
		PositionAt:          in.Position.ObservedAt().UTC().UnixNano(),
		EntryPrice:          decimalString(closePrice),
		LivePrice:           decimalString(currentPrice),
		LivePriceObserved:   market.currentPrice != "",
		EntryPriceDriftPct:  roundedDecimalString(drift),
		StopPrice:           decimalString(stop),
		TargetPrice:         decimalString(target),
		ExpectedRR:          roundedDecimalString(expectedRR),
		AcceptReasons: [7]string{
			"VWAP_ABOVE", "VWAP_SLOPE_UP", "EMA9_PULLBACK_CONFIRMED",
			"VOLUME_PROFILE_SPACE_OK", "RR_GE_2", "NOT_TANGLED", "NOT_AFTER_ENTRY_CUTOFF",
		},
	}
	identity, err := decisionIdentity(record)
	if err != nil {
		return refuse(RefusalDecision)
	}
	record.Identity = identity
	decision, err := mintDecision(record)
	if err != nil {
		return refuse(RefusalDecision)
	}
	return Evaluation{Decision: decision, Reason: RefusalNone}
}

func sourceRefusalReason(reason Refusal) string {
	switch reason {
	case RefusalCandidate, RefusalSource:
		return SourceRejectProfileDisabled
	case RefusalUnsupportedScope:
		return SourceRejectScopeFrozen
	case RefusalBarIntegrity, RefusalSession:
		return SourceRejectBarNotClosed
	case RefusalSymbolState:
		return SourceRejectSymbolStateStale
	case RefusalExistingPosition:
		return SourceRejectPositionAlreadyOpen
	case RefusalIlliquidBar:
		return SourceRejectIlliquidBar
	case RefusalInvalidBar, RefusalIndicator, RefusalDecision:
		return SourceRejectIndicatorUnavailable
	case RefusalVWAPAbove:
		return SourceRejectVWAPBelow
	case RefusalVWAPSlope:
		return SourceRejectVWAPSlopeDown
	case RefusalEMA9Pullback:
		return SourceRejectEMA9PullbackMissing
	case RefusalFakeBreakout:
		return SourceRejectFakeBreakout
	case RefusalLVNSpace, RefusalHVNCeiling:
		return SourceRejectHVNBlock
	case RefusalTangledBand:
		return SourceRejectTangled
	case RefusalBandExpansion:
		return SourceRejectVolatilityExpansion
	case RefusalRR:
		return SourceRejectRRTooLow
	case RefusalAge:
		return SourceRejectStaleSignal
	case RefusalDrift:
		return SourceRejectPriceDrift
	default:
		return ""
	}
}

func constantsDigest() string {
	payload := "min_vwap_slope_pct=0.08\n" +
		"ema_touch_tolerance_pct=0.25\n" +
		"min_forward_space_pct=1.2\n" +
		"min_expected_rr=1.5\n" +
		"tangled_band_pct=0.35\n" +
		"max_band_expansion_rate=1.8\n" +
		"hard_stop_pct=0.7\n" +
		"partial_take_profit_at_r=3.0\n" +
		"skip_open_minutes=10\n" +
		"max_signal_age_seconds=15\n" +
		"max_entry_price_drift_pct=0.20\n" +
		"symbol_state_stale_seconds=30\n"
	sum := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func decimal(raw string) (*big.Rat, bool) {
	if strings.TrimSpace(raw) != raw || raw == "" || strings.ContainsAny(raw, "eE/") {
		return nil, false
	}
	value, ok := new(big.Rat).SetString(raw)
	return value, ok
}

func positive(raw string) (*big.Rat, bool) {
	value, ok := decimal(raw)
	return value, ok && value.Sign() > 0
}

func nonnegative(raw string) (*big.Rat, bool) {
	value, ok := decimal(raw)
	return value, ok && value.Sign() >= 0
}

func rat(raw string) *big.Rat {
	value, _ := new(big.Rat).SetString(raw)
	return value
}

func decimalString(value *big.Rat) string {
	numerator := new(big.Int).Set(value.Num())
	denominator := new(big.Int).Set(value.Denom())
	two, five := big.NewInt(2), big.NewInt(5)
	remainder := new(big.Int)
	twos, fives := 0, 0
	for {
		quotient := new(big.Int)
		quotient.QuoRem(denominator, two, remainder)
		if remainder.Sign() != 0 {
			break
		}
		denominator = quotient
		twos++
	}
	for {
		quotient := new(big.Int)
		quotient.QuoRem(denominator, five, remainder)
		if remainder.Sign() != 0 {
			break
		}
		denominator = quotient
		fives++
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return value.RatString()
	}
	scale := max(twos, fives)
	numerator.Mul(numerator, new(big.Int).Exp(two, big.NewInt(int64(scale-twos)), nil))
	numerator.Mul(numerator, new(big.Int).Exp(five, big.NewInt(int64(scale-fives)), nil))
	negative := numerator.Sign() < 0
	numerator.Abs(numerator)
	digits := numerator.String()
	if scale == 0 {
		if negative {
			return "-" + digits
		}
		return digits
	}
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	point := len(digits) - scale
	out := strings.TrimRight(digits[:point]+"."+digits[point:], "0")
	out = strings.TrimSuffix(out, ".")
	if negative && out != "0" {
		out = "-" + out
	}
	return out
}

func roundedDecimalString(value *big.Rat) string {
	if value == nil {
		return ""
	}
	if exact := decimalString(value); !strings.Contains(exact, "/") {
		return exact
	}
	out := value.FloatString(28)
	out = strings.TrimRight(out, "0")
	out = strings.TrimSuffix(out, ".")
	return out
}
