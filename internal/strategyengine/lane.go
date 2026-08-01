package strategyengine

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
)

type ParkerConservativeLane struct{}

func (ParkerConservativeLane) Evaluate(in LaneInput) EntryDecision {
	approved := strategy.Value(in.Approved)
	base := EntryDecision{Reason: RefusalCandidate, LaneID: LaneID, LaneVersion: LaneVersion, SourceCommit: SourceCommit, SourceSetDigest: FrozenSourceSetDigest, ConstantsDigest: constantsDigest()}
	if !approved.Valid() {
		return base
	}
	key := approved.Key()
	base.CandidateLifeID = approved.CandidateLifeID().String()
	base.ThresholdVersion, base.ThresholdSetDigest, base.EvidenceDigest = approved.ThresholdVersion(), approved.SetDigest(), approved.EvidenceDigest()
	base.Market, base.Symbol = key.Market, key.Symbol
	if in.Market != "KR" || in.Session != "regular" || key.Market != "KR" {
		base.Reason = RefusalUnsupportedScope
		return base
	}
	if !in.SourceVerified {
		base.Reason = RefusalSource
		return base
	}
	if !in.RegularSession {
		base.Reason = RefusalSession
		return base
	}
	if !in.BarsClosedContiguous {
		base.Reason = RefusalBarIntegrity
		return base
	}
	if !in.SymbolStateNormal {
		base.Reason = RefusalSymbolState
		return base
	}
	if !in.NoExistingPosition {
		base.Reason = RefusalExistingPosition
		return base
	}
	volume, vok := decimal(in.Volume)
	if !vok || volume.Sign() <= 0 {
		base.Reason = RefusalZeroVolume
		return base
	}
	price, pok := decimal(in.Price)
	vwap, wok := decimal(in.VWAP)
	slope, sok := decimal(in.VWAPSlopePct)
	ema9, e9ok := decimal(in.EMA9)
	ema20, e20ok := decimal(in.EMA20)
	lvn, lok := decimal(in.LVNForwardSpacePct)
	expansion, exok := decimal(in.BandExpansionRate)
	rr, rrok := decimal(in.ExpectedRR)
	drift, dok := decimal(in.EntryPriceDriftPct)
	if !(pok && wok && sok && e9ok && e20ok && lok && exok && rrok && dok) {
		base.Reason = RefusalIndicator
		return base
	}
	if price.Cmp(vwap) <= 0 {
		base.Reason = RefusalVWAPAbove
		return base
	}
	if slope.Cmp(rat("0.08")) < 0 {
		base.Reason = RefusalVWAPSlope
		return base
	}
	if ema9.Cmp(ema20) <= 0 || absDiffPct(price, ema9).Cmp(rat("0.25")) > 0 {
		base.Reason = RefusalEMA9Pullback
		return base
	}
	if lvn.Cmp(rat("1.2")) < 0 {
		base.Reason = RefusalLVNSpace
		return base
	}
	if !in.Untangled {
		base.Reason = RefusalTangledBand
		return base
	}
	if expansion.Cmp(rat("1.8")) > 0 {
		base.Reason = RefusalBandExpansion
		return base
	}
	if rr.Cmp(rat("1.5")) < 0 {
		base.Reason = RefusalRR
		return base
	}
	if !in.HVNCeilingClear {
		base.Reason = RefusalHVNCeiling
		return base
	}
	if in.SignalAgeSeconds < 0 || in.SignalAgeSeconds > 15 {
		base.Reason = RefusalAge
		return base
	}
	if drift.Sign() < 0 || drift.Cmp(rat("0.20")) > 0 {
		base.Reason = RefusalDrift
		return base
	}
	stop := new(big.Rat).Mul(price, rat("0.993"))
	target := new(big.Rat).Mul(price, rat("1.021"))
	base.Accepted, base.Reason = true, RefusalNone
	base.EntryPrice, base.StopPrice, base.TargetPrice, base.ExpectedRR = decimalString(price), decimalString(stop), decimalString(target), decimalString(rr)
	payload := strings.Join([]string{"strategy-decision:v1", base.CandidateLifeID, LaneID, LaneVersion, base.EntryPrice, base.StopPrice, base.TargetPrice}, "\x00")
	sum := sha256.Sum256([]byte(payload))
	base.Identity = "strategy-decision:v1:sha256:" + hex.EncodeToString(sum[:])
	return base
}

func constantsDigest() string {
	payload := "min_vwap_slope_pct=0.08\nema_touch_tolerance_pct=0.25\nmin_forward_space_pct=1.2\nmin_expected_rr=1.5\ntangled_band_pct=0.35\nmax_band_expansion_rate=1.8\nhard_stop_pct=0.7\npartial_take_profit_at_r=3.0\nskip_open_minutes=10\nmax_signal_age_seconds=15\nmax_entry_price_drift_pct=0.20\nsymbol_state_stale_seconds=30\n"
	sum := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func decimal(raw string) (*big.Rat, bool) {
	if strings.TrimSpace(raw) != raw || raw == "" || strings.ContainsAny(raw, "eE/") {
		return nil, false
	}
	v, ok := new(big.Rat).SetString(raw)
	return v, ok
}
func rat(raw string) *big.Rat { v, _ := new(big.Rat).SetString(raw); return v }
func absDiffPct(a, b *big.Rat) *big.Rat {
	d := new(big.Rat).Sub(a, b)
	if d.Sign() < 0 {
		d.Neg(d)
	}
	return d.Mul(d, big.NewRat(100, 1)).Quo(d, b)
}
func decimalString(v *big.Rat) string {
	numerator := new(big.Int).Set(v.Num())
	denominator := new(big.Int).Set(v.Denom())
	two, five := big.NewInt(2), big.NewInt(5)
	zero, remainder := big.NewInt(0), new(big.Int)
	twos, fives := 0, 0
	for {
		quotient := new(big.Int)
		quotient.QuoRem(denominator, two, remainder)
		if remainder.Cmp(zero) != 0 {
			break
		}
		denominator = quotient
		twos++
	}
	for {
		quotient := new(big.Int)
		quotient.QuoRem(denominator, five, remainder)
		if remainder.Cmp(zero) != 0 {
			break
		}
		denominator = quotient
		fives++
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return v.RatString()
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
