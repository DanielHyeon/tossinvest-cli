package weeklyvaluelane

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type PriceProvenance struct {
	PriceMinor, Source, Version, Digest, AsOf, Currency, UnitVersion string
	MinorScale                                                       int
}
type RRExecutionPolicy struct {
	StagedTargetMinor, FairValueMinor, EntryCostsMinor, ExitCostsMinor string
	MinimumRRPPM                                                       uint64
	DecisionDigest, CalendarDigest, CapSnapshotID, Identity            string
}
type ExecutionTermsPreimage struct {
	planDigest, evidenceDigest, entry, staged, fair, entryCosts, exitCosts string
	minimumRR                                                              uint64
	seal                                                                   [32]byte
}

type savedStopAuthority struct {
	provenance                 PriceProvenance
	planDigest, evidenceDigest string
	seal                       [32]byte
}

func mintExecutionTermsPreimage(plan CampaignPlan, evidence DisclosureEvidence, entry, staged, entryCosts, exitCosts string, minimumRR uint64) ExecutionTermsPreimage {
	if !plan.valid() || evidence.seal == ([32]byte{}) || evidence.seal != evidenceSnapshotSeal(evidence) {
		return ExecutionTermsPreimage{}
	}
	v := ExecutionTermsPreimage{planDigest: plan.digest, evidenceDigest: evidence.EvidenceDigest, entry: entry, staged: staged, fair: evidence.FairValueMinor, entryCosts: entryCosts, exitCosts: exitCosts, minimumRR: minimumRR}
	v.seal = weeklyExecutionPreimageSeal(v)
	return v
}
func (v ExecutionTermsPreimage) valid(plan CampaignPlan, evidence DisclosureEvidence, r EvaluationRequest) bool {
	return plan.valid() && v.planDigest == plan.digest && v.evidenceDigest == evidence.EvidenceDigest && v.entry == r.EntryPriceMinor && v.staged == r.StagedTargetMinor && v.fair == evidence.FairValueMinor && v.entryCosts == r.EntryCostsMinor && v.exitCosts == r.EstimatedExitCostsLeviesMinor && v.minimumRR == r.MinimumRRPPM && v.seal != ([32]byte{}) && v.seal == weeklyExecutionPreimageSeal(v)
}
func weeklyExecutionPreimageSeal(v ExecutionTermsPreimage) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"weekly-rr-preimage:v1", v.planDigest, v.evidenceDigest, v.entry, v.staged, v.fair, v.entryCosts, v.exitCosts, strconv.FormatUint(v.minimumRR, 10)}, "\x00")))
}

func mintSavedStopAuthority(plan CampaignPlan, evidence DisclosureEvidence, price string) savedStopAuthority {
	scale, ok := weeklyScale(plan.quoteCurrency)
	if !plan.valid() || !ok || price == "" || evidence.seal == ([32]byte{}) || evidence.seal != evidenceSnapshotSeal(evidence) {
		return savedStopAuthority{}
	}
	v := savedStopAuthority{planDigest: plan.digest, evidenceDigest: evidence.EvidenceDigest,
		provenance: PriceProvenance{price, "saved-effective-stop", "stop-state-v1", plan.digest, evidence.EvaluatedAt.UTC().Format(time.RFC3339Nano), plan.quoteCurrency, "minor-v1", scale}}
	v.seal = weeklySavedStopSeal(v)
	return v
}

func (v savedStopAuthority) valid(plan CampaignPlan, evidence DisclosureEvidence, price string) bool {
	return plan.valid() && v.planDigest == plan.digest && v.evidenceDigest == evidence.EvidenceDigest && v.provenance.PriceMinor == price &&
		v.seal != ([32]byte{}) && v.seal == weeklySavedStopSeal(v)
}

func (v savedStopAuthority) effectivePrice(plan CampaignPlan, evidence DisclosureEvidence, publicScalar string) (string, bool) {
	if v.seal == ([32]byte{}) {
		return "", publicScalar == ""
	}
	price := v.provenance.PriceMinor
	return price, v.valid(plan, evidence, price)
}

func weeklySavedStopSeal(v savedStopAuthority) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"weekly-saved-stop:v1", v.planDigest, v.evidenceDigest, v.provenance.PriceMinor,
		v.provenance.Source, v.provenance.Version, v.provenance.Digest, v.provenance.AsOf, v.provenance.Currency, strconv.Itoa(v.provenance.MinorScale), v.provenance.UnitVersion}, "\x00")))
}

func executionPolicy(v ExecutionTermsPreimage, lineage ResultLineage) RRExecutionPolicy {
	p := RRExecutionPolicy{StagedTargetMinor: v.staged, FairValueMinor: v.fair, EntryCostsMinor: v.entryCosts, ExitCostsMinor: v.exitCosts, MinimumRRPPM: v.minimumRR,
		DecisionDigest: lineage.DecisionDigest, CalendarDigest: lineage.CalendarDigest, CapSnapshotID: lineage.CapSnapshotID}
	sum := sha256.Sum256([]byte(strings.Join([]string{"weekly-rr-policy:v1", hex.EncodeToString(v.seal[:]), p.DecisionDigest, p.CalendarDigest, p.CapSnapshotID}, "\x00")))
	p.Identity = hex.EncodeToString(sum[:])
	return p
}
func weeklyScale(currency string) (int, bool) {
	if currency == "KRW" {
		return 0, true
	}
	if currency == "USD" {
		return 2, true
	}
	return 0, false
}
