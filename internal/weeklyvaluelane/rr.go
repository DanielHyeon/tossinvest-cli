package weeklyvaluelane

import (
	"math/big"
	"time"
)

type RRInput struct {
	EntryPriceMinor, StagedTargetMinor, FairValueMinor, EffectiveStopMinor string
	Quantity                                                               uint64
	EntryCostsMinor, EstimatedExitCostsLeviesMinor                         string
	MinimumRRPPM                                                           uint64
	AccountCurrency, QuoteCurrency                                         string
	FX                                                                     *FrozenFX
	EvaluatedAt                                                            time.Time
}

type RRResult struct {
	Accepted                            bool
	Code                                RefusalCode
	TargetMinor, RewardMinor, RiskMinor string
	RRPPM                               uint64
}

func CalculateRR(input RRInput) RRResult {
	refuse := func(code RefusalCode) RRResult { return RRResult{Code: code} }
	entry, entryOK := parseUnsigned(input.EntryPriceMinor)
	staged, stagedOK := parseUnsigned(input.StagedTargetMinor)
	fair, fairOK := parseUnsigned(input.FairValueMinor)
	stop, stopOK := parseUnsigned(input.EffectiveStopMinor)
	entryCosts, entryCostsOK := parseUnsigned(input.EntryCostsMinor)
	exitCosts, exitCostsOK := parseUnsigned(input.EstimatedExitCostsLeviesMinor)
	if !entryOK || !stagedOK || !fairOK || !stopOK || !entryCostsOK || !exitCostsOK || input.Quantity == 0 || input.MinimumRRPPM == 0 ||
		input.AccountCurrency == "" || input.QuoteCurrency == "" || entry.Cmp(stop) <= 0 {
		return refuse(RefusalRRInvalid)
	}
	target := new(big.Int).Set(staged)
	if fair.Cmp(target) < 0 {
		target.Set(fair)
	}
	if target.Cmp(entry) <= 0 {
		return refuse(RefusalRRThreshold)
	}
	rewardPerShare := new(big.Int).Sub(target, entry)
	riskPerShare := new(big.Int).Sub(entry, stop)
	quantity := new(big.Int).SetUint64(input.Quantity)
	grossReward, rewardOK := checkedMul(rewardPerShare, quantity)
	grossRisk, riskOK := checkedMul(riskPerShare, quantity)
	if !rewardOK || !riskOK {
		return refuse(RefusalRRInvalid)
	}
	convertedReward, convertRewardOK := convertQuoteAmount(grossReward, input)
	convertedRisk, convertRiskOK := convertQuoteAmount(grossRisk, input)
	convertedEntryCosts, convertEntryOK := convertQuoteAmount(entryCosts, input)
	convertedExitCosts, convertExitOK := convertQuoteAmount(exitCosts, input)
	if !convertRewardOK || !convertRiskOK || !convertEntryOK || !convertExitOK {
		return refuse(RefusalRRInvalid)
	}
	totalCosts, costsOK := checkedAdd(convertedEntryCosts, convertedExitCosts)
	if !costsOK || convertedReward.Cmp(totalCosts) <= 0 {
		return refuse(RefusalRRThreshold)
	}
	reward := new(big.Int).Sub(convertedReward, totalCosts)
	risk, totalRiskOK := checkedAdd(convertedRisk, totalCosts)
	if !totalRiskOK || risk.Sign() <= 0 {
		return refuse(RefusalRRInvalid)
	}
	ppmNumerator, ppmOK := checkedMul(reward, big.NewInt(1_000_000))
	if !ppmOK {
		return refuse(RefusalRRInvalid)
	}
	ppmBig := new(big.Int).Quo(ppmNumerator, risk)
	if !ppmBig.IsUint64() {
		return refuse(RefusalRRInvalid)
	}
	ppm := ppmBig.Uint64()
	result := RRResult{TargetMinor: target.String(), RewardMinor: reward.String(), RiskMinor: risk.String(), RRPPM: ppm}
	if ppm < input.MinimumRRPPM {
		result.Code = RefusalRRThreshold
		return result
	}
	result.Accepted = true
	return result
}

func convertQuoteAmount(value *big.Int, input RRInput) (*big.Int, bool) {
	if input.AccountCurrency == input.QuoteCurrency {
		if input.FX != nil {
			return nil, false
		}
		return new(big.Int).Set(value), true
	}
	if input.FX == nil || !input.FX.validAt(input.EvaluatedAt) {
		return nil, false
	}
	rate, rateOK := parsePositiveDecimal(input.FX.rateQuoteToAccount)
	haircut, haircutOK := parsePositiveDecimal(input.FX.haircut)
	if !rateOK || !haircutOK {
		return nil, false
	}
	converted := new(big.Rat).SetInt(value)
	converted.Mul(converted, rate)
	converted.Mul(converted, haircut)
	return ceilRat(converted)
}
