package continuationlane

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
)

var errExecutionTermsInvalid = errors.New("continuation lanes: execution terms invalid")

type PriceProvenance struct {
	PriceMinor, Source, Version, Digest, AsOf, Currency, UnitVersion string
	MinorScale                                                       int
}

type ExecutionTermsPreimage struct {
	entry, target                        PriceProvenance
	planDigest, evidenceDigest, identity string
	seal                                 [32]byte
}

type savedStopAuthority struct {
	provenance                 PriceProvenance
	planDigest, evidenceDigest string
	seal                       [32]byte
}

func mintExecutionTermsPreimage(plan CampaignPlan, envelope EvidenceEnvelope, entryPriceMinor, targetPriceMinor string) (ExecutionTermsPreimage, error) {
	entry, entryOK := canonicalPositiveMinor(entryPriceMinor)
	target, targetOK := canonicalPositiveMinor(targetPriceMinor)
	if !validatePlan(plan) || !entryOK || !targetOK || entry.Cmp(target) >= 0 || envelope.SourceDigest == "" || envelope.ConfigDigest != plan.ConfigDigest ||
		envelope.Market != plan.Market || envelope.Symbol != plan.Symbol {
		return ExecutionTermsPreimage{}, errExecutionTermsInvalid
	}
	scale, scaleOK := currencyMinorScale(plan.QuoteCurrency)
	if !scaleOK {
		return ExecutionTermsPreimage{}, errExecutionTermsInvalid
	}
	value := ExecutionTermsPreimage{planDigest: plan.RiskBudgetDigest, evidenceDigest: envelope.SourceDigest}
	value.entry = PriceProvenance{PriceMinor: entry.String(), Source: envelope.SourceRecordID, Version: envelope.SchemaVersion, Digest: envelope.SourceDigest,
		AsOf: envelope.EffectiveAt, Currency: plan.QuoteCurrency, MinorScale: scale, UnitVersion: "minor-v1"}
	value.target = PriceProvenance{PriceMinor: target.String(), Source: "continuation-target-policy", Version: plan.LaneVersion, Digest: plan.ConfigDigest,
		AsOf: envelope.EffectiveAt, Currency: plan.QuoteCurrency, MinorScale: scale, UnitVersion: "minor-v1"}
	value.identity = provenanceIdentity("continuation-terms:v2", value.entry, value.target)
	value.seal = continuationExecutionTermsSeal(value)
	return value, nil
}

func (value ExecutionTermsPreimage) valid(plan CampaignPlan, envelope EvidenceEnvelope) bool {
	return validatePlan(plan) && value.planDigest == plan.RiskBudgetDigest && value.evidenceDigest == envelope.SourceDigest &&
		value.entry.Digest == envelope.SourceDigest && value.entry.Version == envelope.SchemaVersion && value.target.Digest == plan.ConfigDigest &&
		value.seal != ([32]byte{}) && value.seal == continuationExecutionTermsSeal(value) && value.identity == provenanceIdentity("continuation-terms:v2", value.entry, value.target)
}

func validatedExecutionTerms(plan CampaignPlan, envelope EvidenceEnvelope, value ExecutionTermsPreimage, effectiveStopMinor string, stop PriceProvenance) (PriceProvenance, PriceProvenance, PriceProvenance, string, bool) {
	if !value.valid(plan, envelope) || stop.PriceMinor != effectiveStopMinor || stop.Digest == "" || stop.Source == "" || stop.Version == "" || stop.AsOf == "" {
		return PriceProvenance{}, PriceProvenance{}, PriceProvenance{}, "", false
	}
	entry, entryOK := canonicalPositiveMinor(value.entry.PriceMinor)
	stopValue, stopOK := canonicalPositiveMinor(stop.PriceMinor)
	target, targetOK := canonicalPositiveMinor(value.target.PriceMinor)
	if !entryOK || !stopOK || !targetOK || stopValue.Cmp(entry) >= 0 || entry.Cmp(target) >= 0 {
		return PriceProvenance{}, PriceProvenance{}, PriceProvenance{}, "", false
	}
	return value.entry, stop, value.target, value.identity, true
}

func stopProvenance(plan CampaignPlan, envelope EvidenceEnvelope, candidate StopCandidate, effective string, saved savedStopAuthority) (PriceProvenance, bool) {
	if effective == candidate.PriceMinor {
		scale, ok := currencyMinorScale(plan.QuoteCurrency)
		if !ok {
			return PriceProvenance{}, false
		}
		return PriceProvenance{PriceMinor: effective, Source: candidate.Source, Version: candidate.Version, Digest: candidate.Digest,
			AsOf: candidate.ObservedAt, Currency: plan.QuoteCurrency, MinorScale: scale, UnitVersion: "minor-v1"}, true
	}
	return saved.provenance, saved.valid(plan, envelope, effective)
}

func mintSavedStopProvenance(plan CampaignPlan, envelope EvidenceEnvelope, price string) savedStopAuthority {
	scale, ok := currencyMinorScale(plan.QuoteCurrency)
	if !ok {
		return savedStopAuthority{}
	}
	value := savedStopAuthority{planDigest: plan.RiskBudgetDigest, evidenceDigest: envelope.SourceDigest,
		provenance: PriceProvenance{PriceMinor: price, Source: "saved-effective-stop", Version: "stop-state-v1", Digest: plan.RiskBudgetDigest,
			AsOf: envelope.EffectiveAt, Currency: plan.QuoteCurrency, MinorScale: scale, UnitVersion: "minor-v1"}}
	value.seal = savedStopAuthoritySeal(value)
	return value
}

func (value savedStopAuthority) valid(plan CampaignPlan, envelope EvidenceEnvelope, price string) bool {
	return validatePlan(plan) && value.planDigest == plan.RiskBudgetDigest && value.evidenceDigest == envelope.SourceDigest &&
		value.provenance.PriceMinor == price && value.seal != ([32]byte{}) && value.seal == savedStopAuthoritySeal(value)
}

func (value savedStopAuthority) effectivePrice(plan CampaignPlan, envelope EvidenceEnvelope, publicScalar string) (string, bool) {
	if value.seal == ([32]byte{}) {
		return "", publicScalar == ""
	}
	price := value.provenance.PriceMinor
	return price, value.valid(plan, envelope, price)
}

func savedStopAuthoritySeal(value savedStopAuthority) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"continuation-saved-stop:v1", value.planDigest, value.evidenceDigest,
		provenanceIdentity("saved-stop", value.provenance)}, "\x00")))
}

func currencyMinorScale(currency string) (int, bool) {
	if currency == "KRW" {
		return 0, true
	}
	if currency == "USD" {
		return 2, true
	}
	return 0, false
}
func canonicalPositiveMinor(raw string) (*big.Int, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return nil, false
	}
	v, err := parseUnsigned(raw)
	return v, err == nil && v.Sign() > 0 && v.String() == raw
}
func provenanceIdentity(prefix string, values ...PriceProvenance) string {
	parts := []string{prefix}
	for _, p := range values {
		parts = append(parts, p.PriceMinor, p.Source, p.Version, p.Digest, p.AsOf, p.Currency, strconv.Itoa(p.MinorScale), p.UnitVersion)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
func continuationExecutionTermsSeal(value ExecutionTermsPreimage) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"continuation-execution-terms:v2", value.planDigest, value.evidenceDigest, value.identity}, "\x00")))
}
