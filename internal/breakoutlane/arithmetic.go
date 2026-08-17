package breakoutlane

import (
	"crypto/sha256"
	"encoding/hex"
	"math/bits"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxUint64 = ^uint64(0)

func canonical(s string) bool {
	return s != "" && utf8.ValidString(s) && strings.TrimSpace(s) == s && !strings.ContainsRune(s, '\x00')
}
func u(v uint64) string        { return strconv.FormatUint(v, 10) }
func boolString(v bool) string { return strconv.FormatBool(v) }
func hashFields(fields ...string) string {
	h := sha256.New()
	for _, f := range fields {
		h.Write([]byte(strconv.Itoa(len(f))))
		h.Write([]byte{':'})
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
func hashNUL(fields ...string) string {
	h := sha256.Sum256([]byte(strings.Join(fields, "\x00")))
	return hex.EncodeToString(h[:])
}
func checkedAdd(values ...uint64) (uint64, bool) {
	var total uint64
	for _, v := range values {
		next, c := bits.Add64(total, v, 0)
		if c != 0 {
			return 0, true
		}
		total = next
	}
	return total, false
}
func mulDivCeil(a, b, d uint64) (uint64, bool) {
	if d == 0 {
		return 0, true
	}
	hi, lo := bits.Mul64(a, b)
	if hi >= d {
		return 0, true
	}
	q, r := bits.Div64(hi, lo, d)
	if r == 0 {
		return q, false
	}
	if q == maxUint64 {
		return 0, true
	}
	return q + 1, false
}
func mulDivFloor(a, b, d uint64) (uint64, RefusalCode) {
	if d == 0 {
		return 0, RefusalFXInvalidRate
	}
	hi, lo := bits.Mul64(a, b)
	if hi >= d {
		return 0, RefusalSizingOverflow
	}
	q, _ := bits.Div64(hi, lo, d)
	return q, RefusalNone
}
func absoluteDifference(a, b uint64) uint64 {
	if a >= b {
		return a - b
	}
	return b - a
}

func snapshotDigest(input EvidenceInput) string {
	parts := []string{"snapshot.v1", string(input.Market), input.Symbol, input.SessionID, input.CalendarVersion, input.LaneID, input.LaneVersion, input.Config.Digest(), u(input.ATRMinor), u(input.EvaluatedAtMS), input.Quote.value.Digest, input.FX.value.Digest,
		u(input.Sizing.ProposedEntryMinor), u(input.Sizing.StopMinor), u(input.Sizing.TargetMinor), u(input.Sizing.EntrySlippageMinor), u(input.Sizing.ExitSlippageMinor), u(input.Sizing.RoundTripCostAccountMinor), u(input.Sizing.RiskBudgetAccountMinor), u(input.Sizing.NotionalCapAccountMinor), u(input.Sizing.FinalCap), u(input.Sizing.MinRiskRewardPPM)}
	for _, bar := range input.Bars {
		b := bar.value
		parts = append(parts, u(b.Sequence), u(b.Revision), u(b.IntervalMS), b.ID, b.SessionID, u(b.HighMinor), u(b.LowMinor), u(b.CloseMinor), u(b.RVOLPPM), u(b.UpperWickRangePPM), strconv.FormatBool(b.RegularSession), strconv.FormatBool(b.Closed), strconv.FormatBool(b.VolumeExpanded))
	}
	return "sha256:" + hashFields(parts...)
}
