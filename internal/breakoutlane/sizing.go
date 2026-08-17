package breakoutlane

type sizingResult struct {
	RiskPerShareMinor, WorstEntryMinor, CandidateQuantity, FinalQuantity uint64
	Refusal                                                              RefusalCode
}

func fxValid(f FXSeal) bool { _, err := NewFXSeal(f.value); return err == nil }
func convertCapacity(amount uint64, f FXSeal, evaluated uint64) (uint64, RefusalCode) {
	if !fxValid(f) {
		return 0, RefusalFXInvalidRate
	}
	v := f.value
	if evaluated < v.AsOfMS || evaluated > v.FreshUntilMS {
		return 0, RefusalFXStale
	}
	return mulDivFloor(amount, v.RateNum, v.RateDen)
}
func convertCost(amount uint64, f FXSeal, evaluated uint64) (uint64, RefusalCode) {
	if !fxValid(f) {
		return 0, RefusalFXInvalidRate
	}
	v := f.value
	if evaluated < v.AsOfMS || evaluated > v.FreshUntilMS {
		return 0, RefusalFXStale
	}
	q, ov := mulDivCeil(amount, v.RateNum, v.RateDen)
	if ov {
		return 0, RefusalSizingOverflow
	}
	return q, RefusalNone
}
func size(in SizingInput, q QuoteSeal, f FXSeal, evaluated uint64) sizingResult {
	if !fxValid(f) {
		return sizingResult{Refusal: RefusalFXMissing}
	}
	if q.value.Currency != f.value.InstrumentCurrency {
		return sizingResult{Refusal: RefusalFXCurrencyMismatch}
	}
	if in.ProposedEntryMinor == 0 || q.value.AskMinor == 0 || in.StopMinor == 0 || in.StopMinor >= in.ProposedEntryMinor {
		return sizingResult{Refusal: RefusalNonProtectiveStop}
	}
	if in.TargetMinor <= in.ProposedEntryMinor {
		return sizingResult{Refusal: RefusalNonProtectiveTarget}
	}
	roundTrip, r := convertCost(in.RoundTripCostAccountMinor, f, evaluated)
	if r != RefusalNone {
		return sizingResult{Refusal: r}
	}
	base := in.ProposedEntryMinor
	if q.value.AskMinor > base {
		base = q.value.AskMinor
	}
	worst, ov := checkedAdd(base, in.EntrySlippageMinor)
	if ov {
		return sizingResult{Refusal: RefusalSizingOverflow}
	}
	risk, ov := checkedAdd(worst-in.StopMinor, in.ExitSlippageMinor, roundTrip)
	if ov {
		return sizingResult{Refusal: RefusalSizingOverflow}
	}
	costExit, ov := checkedAdd(worst, in.ExitSlippageMinor, roundTrip)
	if ov {
		return sizingResult{Refusal: RefusalSizingOverflow}
	}
	if in.TargetMinor <= costExit {
		return sizingResult{Refusal: RefusalNonProtectiveTarget}
	}
	netReward := in.TargetMinor - costExit
	if in.MinRiskRewardPPM > 0 {
		need, ov := mulDivCeil(risk, in.MinRiskRewardPPM, v1PPMScale)
		if ov || netReward < need {
			return sizingResult{Refusal: RefusalNonProtectiveTarget}
		}
	}
	budget, r := convertCapacity(in.RiskBudgetAccountMinor, f, evaluated)
	if r != RefusalNone {
		return sizingResult{Refusal: r}
	}
	notional, r := convertCapacity(in.NotionalCapAccountMinor, f, evaluated)
	if r != RefusalNone {
		return sizingResult{Refusal: r}
	}
	candidate := budget / risk
	if n := notional / worst; n < candidate {
		candidate = n
	}
	if candidate == 0 {
		return sizingResult{Refusal: RefusalZeroQuantity}
	}
	final := candidate
	if in.FinalCap < final {
		final = in.FinalCap
	}
	if final == 0 {
		return sizingResult{Refusal: RefusalZeroQuantity}
	}
	return sizingResult{risk, worst, candidate, final, RefusalNone}
}
