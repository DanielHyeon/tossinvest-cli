// Package strategyflow composes approved candidates, the market/horizon
// router, and pure strategy-lane evaluators. It deliberately stops before
// Guardian, journal, broker, or runtime-state authority.
package strategyflow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math/big"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

type State string

const StateOff State = "OFF"

type RuntimeState string

const RuntimeUnobserved RuntimeState = "UNOBSERVED"

type RefusalCode string

const (
	RefusalNone                  RefusalCode = ""
	RefusalInvalidCandidate      RefusalCode = "INVALID_APPROVED_CANDIDATE"
	RefusalInvalidScope          RefusalCode = "INVALID_ROUTE_SCOPE"
	RefusalRouter                RefusalCode = "ROUTER_REFUSAL"
	RefusalUnsupportedBinding    RefusalCode = "UNSUPPORTED_LANE_BINDING"
	RefusalLane                  RefusalCode = "LANE_REFUSAL"
	RefusalLineageMismatch       RefusalCode = "LINEAGE_MISMATCH"
	RefusalLineageIncomplete     RefusalCode = "LINEAGE_INCOMPLETE"
	RefusalExecutionTermsInvalid RefusalCode = "EXECUTION_TERMS_INVALID"
)

type Descriptor struct {
	Market      strategyrouter.Market
	Horizon     strategyrouter.Horizon
	LaneID      string
	LaneVersion string
	Release     string
	Desired     State
	Effective   State
	Runtime     RuntimeState
}

type Request struct {
	Approved strategy.ApprovedSnapshot
	Router   strategyrouter.RouteRequest
	Lane     LaneInput
}

// Lineage is a sealed value. Valid detects any post-evaluation field change.
// Complete distinguishes an accepted lane decision from a typed refusal.
type Lineage struct {
	AccountRef              string
	Market                  strategyrouter.Market
	Symbol                  string
	PositionGeneration      uint64
	CandidateState          string
	CandidateLifeID         string
	CandidateFirstSeenNS    int64
	CandidateLastSeenNS     int64
	CandidateValidUntilNS   int64
	CandidateApprovedAtNS   int64
	ThresholdVersion        string
	ThresholdSetDigest      string
	CandidateEvidenceDigest string
	RouterEvidenceDigest    string
	LaneEvidenceDigest      string
	RouterID                string
	RouterRelease           string
	Horizon                 strategyrouter.Horizon
	LaneID                  string
	LaneVersion             string
	LaneRelease             string
	ConfigDigest            string
	CampaignID              string
	LegOrdinal              int
	PlannedCeiling          uint64
	RiskBudgetDigest        string
	ExecutionPolicyDigest   string
	Complete                bool
	Identity                string
}

type Result struct {
	Code                    RefusalCode
	NativeCode              string
	Quantity                uint64
	ExecutionTerms          ExecutionTerms
	Lineage                 Lineage
	CommonSafetyIndependent bool
	GuardianCalls           uint64
	BrokerCalls             uint64
	Mutations               uint64
	proposalSeal            [32]byte
}

// ValidProposal proves that the result came from the cap-free production
// proposal composition and that none of its quantity, lineage, execution
// terms, or authority counters changed afterwards.
func (result Result) ValidProposal() bool {
	return result.proposalSeal != ([32]byte{}) && result.proposalSeal == proposalResultSeal(result) &&
		result.Code == RefusalNone && result.Quantity > 0 && result.CommonSafetyIndependent && result.GuardianCalls == 0 &&
		result.BrokerCalls == 0 && result.Mutations == 0 && result.Lineage.Complete && result.Lineage.Valid() &&
		result.ExecutionTerms.Valid() && result.ExecutionTerms.Quantity() == result.Quantity &&
		result.ExecutionTerms.LineageIdentity() == result.Lineage.Identity
}

func sealProposalResult(result Result) Result {
	result.proposalSeal = [32]byte{}
	if result.Code != RefusalNone || result.Quantity == 0 || !result.Lineage.Complete || !result.Lineage.Valid() ||
		!result.ExecutionTerms.Valid() || result.ExecutionTerms.Quantity() != result.Quantity ||
		result.ExecutionTerms.LineageIdentity() != result.Lineage.Identity || !result.CommonSafetyIndependent ||
		result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 {
		return result
	}
	result.proposalSeal = proposalResultSeal(result)
	return result
}

func proposalResultSeal(result Result) [32]byte {
	h := sha256.New()
	writeLineageString(h, "strategyflow-q-candidate-proposal:v1")
	writeLineageString(h, string(result.Code))
	writeLineageString(h, result.NativeCode)
	writeLineageUint64(h, result.Quantity)
	writeLineageString(h, result.Lineage.Identity)
	writeLineageString(h, result.ExecutionTerms.Identity())
	if result.CommonSafetyIndependent {
		writeLineageString(h, "1")
	} else {
		writeLineageString(h, "0")
	}
	writeLineageUint64(h, result.GuardianCalls)
	writeLineageUint64(h, result.BrokerCalls)
	writeLineageUint64(h, result.Mutations)
	var seal [32]byte
	copy(seal[:], h.Sum(nil))
	return seal
}

// FinalizeProposalQuantity creates the quantity projection used after a066 has
// independently issued q_final. It cannot increase q_candidate and preserves
// the exact proposal lineage and non-quantity execution terms.
func FinalizeProposalQuantity(proposal Result, qFinal uint64) (Result, bool) {
	if !proposal.ValidProposal() || qFinal == 0 || qFinal > proposal.Quantity {
		return Result{}, false
	}
	final := proposal
	final.Quantity = qFinal
	final.proposalSeal = [32]byte{}
	terms := final.ExecutionTerms
	terms.quantity = qFinal
	terms.identity = ""
	if !validExecutionTermsFields(terms) {
		return Result{}, false
	}
	terms.identity = executionTermsIdentity(terms)
	if !terms.Valid() {
		return Result{}, false
	}
	final.ExecutionTerms = terms
	return final, true
}

type PriceProvenance struct {
	priceMinor, source, version, digest, asOf, currency, unitVersion string
	minorScale                                                       int
}

func (p PriceProvenance) PriceMinor() string  { return p.priceMinor }
func (p PriceProvenance) Source() string      { return p.source }
func (p PriceProvenance) Version() string     { return p.version }
func (p PriceProvenance) Digest() string      { return p.digest }
func (p PriceProvenance) AsOf() string        { return p.asOf }
func (p PriceProvenance) Currency() string    { return p.currency }
func (p PriceProvenance) MinorScale() int     { return p.minorScale }
func (p PriceProvenance) UnitVersion() string { return p.unitVersion }

// MajorDecimal projects authenticated quote-minor evidence into the canonical
// decimal representation consumed by Guardian and official order intents. It
// never uses floating point and does not mutate or reseal the provenance.
func (p PriceProvenance) MajorDecimal() (string, bool) {
	minor, ok := canonicalExecutionMinor(p.priceMinor)
	if !ok || p.unitVersion != "minor-v1" || p.minorScale < 0 || p.minorScale > 18 {
		return "", false
	}
	if p.minorScale == 0 {
		return minor.String(), true
	}
	digits := minor.String()
	if len(digits) <= p.minorScale {
		digits = strings.Repeat("0", p.minorScale-len(digits)+1) + digits
	}
	point := len(digits) - p.minorScale
	integer, fractional := digits[:point], strings.TrimRight(digits[point:], "0")
	if fractional == "" {
		return integer, true
	}
	return integer + "." + fractional, true
}

type ExecutionPolicy struct {
	stagedTargetMinor, fairValueMinor, entryCostsMinor, exitCostsMinor string
	minimumRRPPM                                                       uint64
	decisionDigest, calendarDigest, capSnapshotID, identity            string
}

func (p ExecutionPolicy) StagedTargetMinor() string { return p.stagedTargetMinor }
func (p ExecutionPolicy) FairValueMinor() string    { return p.fairValueMinor }
func (p ExecutionPolicy) EntryCostsMinor() string   { return p.entryCostsMinor }
func (p ExecutionPolicy) ExitCostsMinor() string    { return p.exitCostsMinor }
func (p ExecutionPolicy) MinimumRRPPM() uint64      { return p.minimumRRPPM }
func (p ExecutionPolicy) DecisionDigest() string    { return p.decisionDigest }
func (p ExecutionPolicy) CalendarDigest() string    { return p.calendarDigest }
func (p ExecutionPolicy) CapSnapshotID() string     { return p.capSnapshotID }
func (p ExecutionPolicy) Identity() string          { return p.identity }

// ExecutionTerms is opaque outside this package. Callers can inspect copies
// through getters but cannot mutate or recompute its unexported seal.
type ExecutionTerms struct {
	accountRef, symbol, campaignID, lineageIdentity string
	market                                          strategyrouter.Market
	legOrdinal                                      int
	quantity                                        uint64
	entry, stop, target                             PriceProvenance
	policy                                          ExecutionPolicy
	identity                                        string
}

func (terms ExecutionTerms) AccountRef() string             { return terms.accountRef }
func (terms ExecutionTerms) Market() strategyrouter.Market  { return terms.market }
func (terms ExecutionTerms) Symbol() string                 { return terms.symbol }
func (terms ExecutionTerms) CampaignID() string             { return terms.campaignID }
func (terms ExecutionTerms) LegOrdinal() int                { return terms.legOrdinal }
func (terms ExecutionTerms) Quantity() uint64               { return terms.quantity }
func (terms ExecutionTerms) Entry() PriceProvenance         { return terms.entry }
func (terms ExecutionTerms) EffectiveStop() PriceProvenance { return terms.stop }
func (terms ExecutionTerms) Target() PriceProvenance        { return terms.target }
func (terms ExecutionTerms) Policy() ExecutionPolicy        { return terms.policy }
func (terms ExecutionTerms) LineageIdentity() string        { return terms.lineageIdentity }
func (terms ExecutionTerms) Identity() string               { return terms.identity }

func (terms ExecutionTerms) Valid() bool {
	if terms.identity == "" || !validExecutionTermsFields(terms) {
		return false
	}
	want := terms.identity
	terms.identity = ""
	return want == executionTermsIdentity(terms)
}

func sealExecutionTerms(lineage Lineage, evaluated laneEvaluation) (ExecutionTerms, bool) {
	terms := ExecutionTerms{accountRef: lineage.AccountRef, market: lineage.Market, symbol: lineage.Symbol, campaignID: lineage.CampaignID,
		legOrdinal: lineage.LegOrdinal, quantity: evaluated.quantity, entry: evaluated.entry, stop: evaluated.stop, target: evaluated.target,
		policy: evaluated.policy, lineageIdentity: lineage.Identity}
	if !lineage.Complete || !lineage.Valid() || !validExecutionTermsFields(terms) {
		return ExecutionTerms{}, false
	}
	terms.identity = executionTermsIdentity(terms)
	return terms, true
}

func validExecutionTermsFields(terms ExecutionTerms) bool {
	entry, entryOK := canonicalExecutionMinor(terms.entry.priceMinor)
	stop, stopOK := canonicalExecutionMinor(terms.stop.priceMinor)
	target, targetOK := canonicalExecutionMinor(terms.target.priceMinor)
	return terms.accountRef != "" && (terms.market == strategyrouter.MarketKR || terms.market == strategyrouter.MarketUS) &&
		terms.symbol != "" && terms.symbol == strings.ToUpper(strings.TrimSpace(terms.symbol)) && terms.campaignID != "" &&
		terms.legOrdinal > 0 && terms.quantity > 0 && terms.lineageIdentity != "" && terms.policy.identity != "" && validPriceProvenance(terms.entry) &&
		validPriceProvenance(terms.stop) && validPriceProvenance(terms.target) && entryOK && stopOK && targetOK &&
		stop.Cmp(entry) < 0 && entry.Cmp(target) < 0
}

func validPriceProvenance(p PriceProvenance) bool {
	return p.source != "" && p.version != "" && p.digest != "" && p.asOf != "" && p.currency != "" && p.minorScale >= 0 && p.unitVersion != ""
}

func canonicalExecutionMinor(raw string) (*big.Int, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return nil, false
	}
	value, ok := new(big.Int).SetString(raw, 10)
	return value, ok && value.Sign() > 0 && value.String() == raw
}

func executionTermsIdentity(terms ExecutionTerms) string {
	h := sha256.New()
	writeLineageString(h, terms.accountRef)
	writeLineageString(h, string(terms.market))
	writeLineageString(h, terms.symbol)
	writeLineageString(h, terms.campaignID)
	writeLineageUint64(h, uint64(terms.legOrdinal))
	writeLineageUint64(h, terms.quantity)
	for _, p := range []PriceProvenance{terms.entry, terms.stop, terms.target} {
		writeLineageString(h, p.priceMinor)
		writeLineageString(h, p.source)
		writeLineageString(h, p.version)
		writeLineageString(h, p.digest)
		writeLineageString(h, p.asOf)
		writeLineageString(h, p.currency)
		writeLineageUint64(h, uint64(p.minorScale))
		writeLineageString(h, p.unitVersion)
	}
	writeLineageString(h, terms.policy.identity)
	writeLineageString(h, terms.lineageIdentity)
	return "strategy-execution-terms:v1:sha256:" + hex.EncodeToString(h.Sum(nil))
}

type laneKind uint8

const (
	laneUnknown laneKind = iota
	laneContinuationKR
	laneContinuationUS
	laneReversalKR
	laneReversalUS
	laneWeeklyKR
	laneWeeklyUS
)

// LaneInput is an exact tagged union. Constructors bind a request to one
// market-specific evaluator without creating activation or execution authority.
type LaneInput struct {
	kind           laneKind
	continuationKR continuationlane.KREvaluationRequest
	continuationUS continuationlane.USEvaluationRequest
	reversalKR     reversallane.KREvaluationRequest
	reversalUS     reversallane.USEvaluationRequest
	weeklyKR       weeklyvaluelane.EvaluationRequest
	weeklyUS       weeklyvaluelane.EvaluationRequest
}

func ContinuationKR(request continuationlane.KREvaluationRequest) LaneInput {
	return LaneInput{kind: laneContinuationKR, continuationKR: request}
}

func ContinuationUS(request continuationlane.USEvaluationRequest) LaneInput {
	return LaneInput{kind: laneContinuationUS, continuationUS: request}
}

func ReversalKR(request reversallane.KREvaluationRequest) LaneInput {
	return LaneInput{kind: laneReversalKR, reversalKR: request}
}

func ReversalUS(request reversallane.USEvaluationRequest) LaneInput {
	return LaneInput{kind: laneReversalUS, reversalUS: request}
}

func WeeklyKR(request weeklyvaluelane.EvaluationRequest) LaneInput {
	return LaneInput{kind: laneWeeklyKR, weeklyKR: request}
}

func WeeklyUS(request weeklyvaluelane.EvaluationRequest) LaneInput {
	return LaneInput{kind: laneWeeklyUS, weeklyUS: request}
}

func (l Lineage) Valid() bool {
	if l.Identity == "" {
		return false
	}
	want := l.Identity
	l.Identity = ""
	return want == lineageIdentity(l)
}

func sealLineage(lineage Lineage) Lineage {
	lineage.Identity = ""
	lineage.Identity = lineageIdentity(lineage)
	return lineage
}

func lineageIdentity(lineage Lineage) string {
	h := sha256.New()
	writeLineageString(h, lineage.AccountRef)
	writeLineageString(h, string(lineage.Market))
	writeLineageString(h, lineage.Symbol)
	writeLineageUint64(h, lineage.PositionGeneration)
	writeLineageString(h, lineage.CandidateState)
	writeLineageString(h, lineage.CandidateLifeID)
	writeLineageUint64(h, uint64(lineage.CandidateFirstSeenNS))
	writeLineageUint64(h, uint64(lineage.CandidateLastSeenNS))
	writeLineageUint64(h, uint64(lineage.CandidateValidUntilNS))
	writeLineageUint64(h, uint64(lineage.CandidateApprovedAtNS))
	writeLineageString(h, lineage.ThresholdVersion)
	writeLineageString(h, lineage.ThresholdSetDigest)
	writeLineageString(h, lineage.CandidateEvidenceDigest)
	writeLineageString(h, lineage.RouterEvidenceDigest)
	writeLineageString(h, lineage.LaneEvidenceDigest)
	writeLineageString(h, lineage.RouterID)
	writeLineageString(h, lineage.RouterRelease)
	writeLineageString(h, string(lineage.Horizon))
	writeLineageString(h, lineage.LaneID)
	writeLineageString(h, lineage.LaneVersion)
	writeLineageString(h, lineage.LaneRelease)
	writeLineageString(h, lineage.ConfigDigest)
	writeLineageString(h, lineage.CampaignID)
	writeLineageUint64(h, uint64(lineage.LegOrdinal))
	writeLineageUint64(h, lineage.PlannedCeiling)
	writeLineageString(h, lineage.RiskBudgetDigest)
	writeLineageString(h, lineage.ExecutionPolicyDigest)
	if lineage.Complete {
		writeLineageString(h, "1")
	} else {
		writeLineageString(h, "0")
	}
	return "strategy-lineage:v1:sha256:" + hex.EncodeToString(h.Sum(nil))
}

func writeLineageString(h hash.Hash, value string) {
	writeLineageUint64(h, uint64(len(value)))
	_, _ = h.Write([]byte(value))
}

func writeLineageUint64(h hash.Hash, value uint64) {
	var buffer [8]byte
	binary.BigEndian.PutUint64(buffer[:], value)
	_, _ = h.Write(buffer[:])
}
