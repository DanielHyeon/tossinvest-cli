// Package strategyflow composes approved candidates, the market/horizon
// router, and pure strategy-lane evaluators. It deliberately stops before
// Guardian, journal, broker, or runtime-state authority.
package strategyflow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"

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
	RefusalNone               RefusalCode = ""
	RefusalInvalidCandidate   RefusalCode = "INVALID_APPROVED_CANDIDATE"
	RefusalInvalidScope       RefusalCode = "INVALID_ROUTE_SCOPE"
	RefusalRouter             RefusalCode = "ROUTER_REFUSAL"
	RefusalUnsupportedBinding RefusalCode = "UNSUPPORTED_LANE_BINDING"
	RefusalLane               RefusalCode = "LANE_REFUSAL"
	RefusalLineageMismatch    RefusalCode = "LINEAGE_MISMATCH"
	RefusalLineageIncomplete  RefusalCode = "LINEAGE_INCOMPLETE"
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
	Complete                bool
	Identity                string
}

type Result struct {
	Code                    RefusalCode
	NativeCode              string
	Quantity                uint64
	Lineage                 Lineage
	CommonSafetyIndependent bool
	GuardianCalls           uint64
	BrokerCalls             uint64
	Mutations               uint64
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
