package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var errExecutionTermsInvalid = errors.New("reversal lane: execution terms invalid")

type PriceProvenance struct {
	PriceMinor, Source, Version, Digest, AsOf, Currency, UnitVersion string
	MinorScale                                                       int
}
type ExecutionTermsPreimage struct {
	entry, target                        PriceProvenance
	planDigest, evidenceDigest, identity string
	seal                                 [32]byte
}

func mintExecutionTermsPreimage(plan CampaignPlan, envelope CommonEnvelope, entryPriceMinor, targetPriceMinor string) (ExecutionTermsPreimage, error) {
	entry, entryOK := canonicalPositiveMinor(entryPriceMinor)
	target, targetOK := canonicalPositiveMinor(targetPriceMinor)
	if !plan.valid() || !entryOK || !targetOK || entry.Cmp(target) >= 0 || envelope.SourceDigest == "" || envelope.ConfigDigest != plan.ConfigDigest() ||
		envelope.Market != plan.Market() || envelope.AccountRef != plan.request.AccountRef || envelope.Symbol != plan.request.Symbol {
		return ExecutionTermsPreimage{}, errExecutionTermsInvalid
	}
	scale, ok := reversalCurrencyScale(plan.request.QuoteCurrency)
	if !ok {
		return ExecutionTermsPreimage{}, errExecutionTermsInvalid
	}
	v := ExecutionTermsPreimage{planDigest: plan.Digest(), evidenceDigest: envelope.SourceDigest}
	v.entry = PriceProvenance{entry.String(), envelope.SourceRecordID, envelope.SchemaVersion, envelope.SourceDigest, envelope.EffectiveAt.UTC().Format(time.RFC3339Nano), plan.request.QuoteCurrency, "minor-v1", scale}
	v.target = PriceProvenance{target.String(), "reversal-target-policy", plan.request.LaneVersion, plan.request.ConfigDigest, envelope.EffectiveAt.UTC().Format(time.RFC3339Nano), plan.request.QuoteCurrency, "minor-v1", scale}
	v.identity = reversalProvenanceIdentity("reversal-terms:v2", v.entry, v.target)
	v.seal = reversalExecutionTermsSeal(v)
	return v, nil
}

func (v ExecutionTermsPreimage) valid(plan CampaignPlan, envelope CommonEnvelope) bool {
	return plan.valid() && v.planDigest == plan.Digest() && v.evidenceDigest == envelope.SourceDigest && v.entry.Digest == envelope.SourceDigest && v.target.Digest == plan.ConfigDigest() && v.seal != ([32]byte{}) && v.seal == reversalExecutionTermsSeal(v) && v.identity == reversalProvenanceIdentity("reversal-terms:v2", v.entry, v.target)
}

func mintStopCandidate(plan CampaignPlan, envelope CommonEnvelope, input stopCandidateInput) (StopCandidate, error) {
	price, ok := canonicalPositiveMinor(input.PriceMinor)
	if !plan.valid() || !ok || input.Source == "" || input.Policy == "" || input.Version == "" || input.Digest == "" || input.ObservedAt.IsZero() || input.FreshUntil.IsZero() ||
		input.ObservedAt.After(input.FreshUntil) || input.ObservedAt.After(envelope.EvaluatedAt) || input.FreshUntil.Before(envelope.EvaluatedAt) {
		return StopCandidate{}, errExecutionTermsInvalid
	}
	c := StopCandidate{priceMinor: price.String(), source: input.Source, policy: input.Policy, version: input.Version, digest: input.Digest, observedAt: input.ObservedAt,
		freshUntil: input.FreshUntil, planDigest: plan.Digest(), evidenceDigest: envelope.SourceDigest}
	c.seal = reversalStopSeal(c)
	return c, nil
}

func (c StopCandidate) valid(plan CampaignPlan, envelope CommonEnvelope) bool {
	return plan.valid() && c.planDigest == plan.Digest() && c.evidenceDigest == envelope.SourceDigest && c.seal != ([32]byte{}) && c.seal == reversalStopSeal(c) && !envelope.EvaluatedAt.Before(c.observedAt) && !envelope.EvaluatedAt.After(c.freshUntil)
}
func reversalStopSeal(c StopCandidate) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"reversal-stop:v1", c.planDigest, c.evidenceDigest, c.priceMinor, c.source, c.policy, c.version, c.digest, c.observedAt.UTC().Format(time.RFC3339Nano), c.freshUntil.UTC().Format(time.RFC3339Nano)}, "\x00")))
}

func validatedExecutionTerms(plan CampaignPlan, envelope CommonEnvelope, value ExecutionTermsPreimage, stop StopCandidate) (PriceProvenance, PriceProvenance, PriceProvenance, string, bool) {
	if !value.valid(plan, envelope) || !stop.valid(plan, envelope) {
		return PriceProvenance{}, PriceProvenance{}, PriceProvenance{}, "", false
	}
	entry, _ := canonicalPositiveMinor(value.entry.PriceMinor)
	stopValue, ok := canonicalPositiveMinor(stop.priceMinor)
	target, _ := canonicalPositiveMinor(value.target.PriceMinor)
	if !ok || stopValue.Cmp(entry) >= 0 || entry.Cmp(target) >= 0 {
		return PriceProvenance{}, PriceProvenance{}, PriceProvenance{}, "", false
	}
	scale, _ := reversalCurrencyScale(plan.request.QuoteCurrency)
	sp := PriceProvenance{stop.priceMinor, stop.source, stop.version, stop.digest, stop.observedAt.UTC().Format(time.RFC3339Nano), plan.request.QuoteCurrency, "minor-v1", scale}
	return value.entry, sp, value.target, value.identity, true
}

func canonicalPositiveMinor(raw string) (*big.Int, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return nil, false
	}
	v, ok := parseMinor(raw)
	return v, ok && v.Sign() > 0 && v.String() == raw
}
func reversalCurrencyScale(c string) (int, bool) {
	if c == "KRW" {
		return 0, true
	}
	if c == "USD" {
		return 2, true
	}
	return 0, false
}
func reversalProvenanceIdentity(prefix string, values ...PriceProvenance) string {
	parts := []string{prefix}
	for _, p := range values {
		parts = append(parts, p.PriceMinor, p.Source, p.Version, p.Digest, p.AsOf, p.Currency, strconv.Itoa(p.MinorScale), p.UnitVersion)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
func reversalExecutionTermsSeal(v ExecutionTermsPreimage) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"reversal-execution-terms:v2", v.planDigest, v.evidenceDigest, v.identity}, "\x00")))
}
