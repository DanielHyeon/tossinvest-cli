package breakoutlane

func BreakoutBufferMinor(tick, atr uint64) uint64 {
	return breakoutBufferMinor(tick, atr, v1BreakoutBufferPPM)
}
func breakoutBufferMinor(tick, atr, bufferPPM uint64) uint64 {
	q, overflow := mulDivCeil(atr, bufferPPM, v1PPMScale)
	if overflow {
		return maxUint64
	}
	if q > tick {
		return q
	}
	return tick
}
func BreakoutCloseQualifies(close, resistance, atr uint64, c V1Config) bool {
	if !c.Valid() {
		return false
	}
	b := breakoutBufferMinor(c.value.TickMinor, atr, c.value.BreakoutBufferPPM)
	return b != maxUint64 && resistance <= maxUint64-b && close >= resistance+b
}
func RetestQualifies(close, resistance, atr uint64, c V1Config) bool {
	if !c.Valid() {
		return false
	}
	d := absoluteDifference(close, resistance)
	left, ov := mulDivCeil(d, v1PPMScale, c.value.RetestTolerancePPM)
	return !ov && left <= atr
}
func quoteValid(q QuoteSeal) bool { _, err := NewQuoteSeal(q.value); return err == nil }
func validateQuote(q QuoteSeal, evaluated, entry uint64, c V1Config) RefusalCode {
	if !quoteValid(q) || entry == 0 {
		return RefusalEvidenceInvalid
	}
	v := q.value
	if v.SourceObservedAtMS > v.ReceivedAtMS || v.ReceivedAtMS > evaluated {
		return RefusalQuoteStale
	}
	if evaluated-v.SourceObservedAtMS > c.value.MaxQuoteAgeMS {
		return RefusalQuoteStale
	}
	mid := v.BidMinor/2 + v.AskMinor/2 + (v.BidMinor%2+v.AskMinor%2)/2
	spread, ov := mulDivCeil(v.AskMinor-v.BidMinor, v1PPMScale, mid)
	if ov {
		return RefusalSizingOverflow
	}
	if spread > c.value.MaxSpreadPPM {
		return RefusalSpreadTooWide
	}
	drift, ov := mulDivCeil(absoluteDifference(v.AskMinor, entry), v1PPMScale, entry)
	if ov {
		return RefusalSizingOverflow
	}
	if drift > c.value.MaxEntryDriftPPM {
		return RefusalEntryDriftExceeded
	}
	return RefusalNone
}
