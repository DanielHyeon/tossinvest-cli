package reversallane

import (
	"crypto/sha256"
	"errors"
	"math/big"
	"strings"
)

var errExecutionTermsInvalid = errors.New("reversal lane: execution terms invalid")

// ExecutionTermsPreimage preserves explicit caller values and binds them to a
// single immutable campaign plan. It never calculates a missing target.
type ExecutionTermsPreimage struct {
	EntryPriceMinor  string
	TargetPriceMinor string
	planDigest       string
	seal             [32]byte
}

func NewExecutionTermsPreimage(plan CampaignPlan, entryPriceMinor, targetPriceMinor string) (ExecutionTermsPreimage, error) {
	entry, entryOK := canonicalPositiveMinor(entryPriceMinor)
	target, targetOK := canonicalPositiveMinor(targetPriceMinor)
	if !plan.valid() || !entryOK || !targetOK || entry.Cmp(target) >= 0 {
		return ExecutionTermsPreimage{}, errExecutionTermsInvalid
	}
	value := ExecutionTermsPreimage{EntryPriceMinor: entry.String(), TargetPriceMinor: target.String(), planDigest: plan.Digest()}
	value.seal = reversalExecutionTermsSeal(value)
	return value, nil
}

func (value ExecutionTermsPreimage) valid(plan CampaignPlan) bool {
	entry, entryOK := canonicalPositiveMinor(value.EntryPriceMinor)
	target, targetOK := canonicalPositiveMinor(value.TargetPriceMinor)
	return plan.valid() && value.planDigest == plan.Digest() && value.seal != ([32]byte{}) && value.seal == reversalExecutionTermsSeal(value) &&
		entryOK && targetOK && entry.Cmp(target) < 0
}

func validatedExecutionTerms(plan CampaignPlan, value ExecutionTermsPreimage, effectiveStopMinor string) (string, string, string, bool) {
	if !value.valid(plan) {
		return "", "", "", false
	}
	entry, entryOK := canonicalPositiveMinor(value.EntryPriceMinor)
	stop, stopOK := canonicalPositiveMinor(effectiveStopMinor)
	target, targetOK := canonicalPositiveMinor(value.TargetPriceMinor)
	if !entryOK || !stopOK || !targetOK || stop.Cmp(entry) >= 0 || entry.Cmp(target) >= 0 {
		return "", "", "", false
	}
	return entry.String(), stop.String(), target.String(), true
}

func canonicalPositiveMinor(raw string) (*big.Int, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return nil, false
	}
	value, ok := parseMinor(raw)
	return value, ok && value.Sign() > 0 && value.String() == raw
}

func reversalExecutionTermsSeal(value ExecutionTermsPreimage) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"reversal-execution-terms:v1", value.planDigest, value.EntryPriceMinor, value.TargetPriceMinor}, "\x00")))
}
