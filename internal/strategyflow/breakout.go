package strategyflow

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// BreakoutRelease 는 이 가족의 레인 릴리스 이름이다.
// 다른 세 가족은 자기 lane 패키지에서 릴리스 상수를 가져오지만, breakoutlane 은
// L2 소유라 편집할 수 없고 동결된 골든에도 release 항목이 없다(결정 48).
const BreakoutRelease = "breakout-retest-v1"

// breakoutPolicyDomain 은 실행 정책 신원의 도메인 분리 문자열이다.
// 다른 도메인의 해시와 절대 같은 값이 나오지 않게 앞에 붙인다.
const breakoutPolicyDomain = "strategyflow-breakout-execution-policy:v1"

// BreakoutPriceEnvelope 는 순수 코어가 만들지 않는(그리고 만들어서도 안 되는)
// 가격 출처 증명이다. 매니페스트가 준 값을 그대로 옮기기만 한다.
type BreakoutPriceEnvelope struct {
	Source      string
	Version     string
	Digest      string
	AsOf        string
	Currency    string
	UnitVersion string
	MinorScale  int
}

// BreakoutPolicyInput 은 실행 정책 항목이다. 순수 코어의 관심사가 아니다.
type BreakoutPolicyInput struct {
	StagedTargetMinor string
	FairValueMinor    string
	EntryCostsMinor   string
	ExitCostsMinor    string
	MinimumRRPPM      uint64
	DecisionDigest    string
	CalendarDigest    string
	CapSnapshotID     string
}

// BreakoutRequest 는 봉인된 순수 코어 증거와, 합성에 필요한 계보 봉투를 함께 든다.
// 활성화나 실행 권한을 만들지 않는다.
type BreakoutRequest struct {
	Evidence           breakoutlane.EvidenceInput
	Prior              *breakoutlane.Decision
	AccountRef         string
	CandidateID        string
	CampaignID         string
	PositionGeneration uint64
	LegOrdinal         int
	PlannedCeiling     uint64
	RiskBudgetDigest   string
	Price              BreakoutPriceEnvelope
	Policy             BreakoutPolicyInput
}

// BreakoutKR 은 KR 돌파-되돌림 평가자에 요청을 묶는다.
func BreakoutKR(request BreakoutRequest) LaneInput {
	return LaneInput{kind: laneBreakoutKR, breakoutKR: request}
}

// BreakoutUS 는 US 돌파-되돌림 평가자에 요청을 묶는다.
func BreakoutUS(request BreakoutRequest) LaneInput {
	return LaneInput{kind: laneBreakoutUS, breakoutUS: request}
}

func evaluateBreakoutKR(input LaneInput) laneEvaluation {
	return runBreakout(input.breakoutKR, breakoutlane.MarketKR, false)
}

func evaluateBreakoutUS(input LaneInput) laneEvaluation {
	return runBreakout(input.breakoutUS, breakoutlane.MarketUS, false)
}

func proposeBreakoutKR(input LaneInput) laneEvaluation {
	return runBreakout(input.breakoutKR, breakoutlane.MarketKR, true)
}

func proposeBreakoutUS(input LaneInput) laneEvaluation {
	return runBreakout(input.breakoutUS, breakoutlane.MarketUS, true)
}

// runBreakout 은 순수 코어를 한 번 돌리고 그 결정을 합성 계층의 모양으로 옮긴다.
// 어떤 값도 다시 계산하지 않는다 — 옮기기만 한다.
// capFree 가 참이면 상한 없는 q_candidate 를, 거짓이면 FinalCap 이 걸린 값을 읽는다(결정 48).
func runBreakout(request BreakoutRequest, market breakoutlane.Market, capFree bool) laneEvaluation {
	if request.Evidence.Market != market {
		return laneEvaluation{nativeCode: breakoutMarketMismatch}
	}
	snapshot, err := breakoutlane.NewEvidenceSnapshot(request.Evidence)
	if err != nil {
		return laneEvaluation{nativeCode: breakoutSnapshotInvalid}
	}
	decision := breakoutlane.Evaluate(snapshot, request.Prior)
	accepted := decision.Phase() == breakoutProposedPhase && decision.Refusal() == breakoutlane.RefusalNone
	quantity := decision.FinalQuantity()
	if capFree {
		quantity = decision.CandidateQuantity()
	}
	if !accepted {
		quantity = 0
	}
	return laneEvaluation{
		accepted:   accepted,
		nativeCode: breakoutNativeCode(decision),
		quantity:   quantity,
		entry:      breakoutPrice(request.Price, request.Evidence.Sizing.ProposedEntryMinor),
		stop:       breakoutPrice(request.Price, request.Evidence.Sizing.StopMinor),
		target:     breakoutPrice(request.Price, request.Evidence.Sizing.TargetMinor),
		policy:     breakoutPolicy(request.Policy),
		lineage: laneLineage{
			AccountRef:         request.AccountRef,
			Market:             strategyrouter.Market(market),
			Symbol:             request.Evidence.Symbol,
			PositionGeneration: request.PositionGeneration,
			LaneID:             request.Evidence.LaneID,
			LaneVersion:        request.Evidence.LaneVersion,
			CandidateID:        request.CandidateID,
			EvidenceDigest:     decision.SnapshotDigest(),
			ConfigDigest:       decision.ConfigDigest(),
			CampaignID:         request.CampaignID,
			LegOrdinal:         request.LegOrdinal,
			PlannedCeiling:     request.PlannedCeiling,
			RiskBudgetDigest:   request.RiskBudgetDigest,
		},
	}
}

const (
	breakoutProposedPhase   = "PROPOSED"
	breakoutMarketMismatch  = "BREAKOUT_MARKET_MISMATCH"
	breakoutSnapshotInvalid = "BREAKOUT_SNAPSHOT_INVALID"
	breakoutPhasePrefix     = "BREAKOUT_PHASE_"
)

// breakoutNativeCode 는 거절 사유를 항상 읽을 수 있는 문자열로 만든다.
// 순수 코어는 "아직 조건이 안 됐다"를 거절 코드 없이 중간 상태로만 알리므로,
// 그 경우에는 상태 이름을 코드로 쓴다 — 빈 문자열은 침묵이라서 금지다.
func breakoutNativeCode(decision breakoutlane.Decision) string {
	if refusal := decision.Refusal(); refusal != breakoutlane.RefusalNone {
		return string(refusal)
	}
	if decision.Phase() == breakoutProposedPhase {
		return ""
	}
	return breakoutPhasePrefix + decision.Phase()
}

func breakoutPrice(envelope BreakoutPriceEnvelope, minor uint64) PriceProvenance {
	if minor == 0 {
		return PriceProvenance{}
	}
	return PriceProvenance{
		priceMinor:  strconv.FormatUint(minor, 10),
		source:      envelope.Source,
		version:     envelope.Version,
		digest:      envelope.Digest,
		asOf:        envelope.AsOf,
		currency:    envelope.Currency,
		minorScale:  envelope.MinorScale,
		unitVersion: envelope.UnitVersion,
	}
}

func breakoutPolicy(input BreakoutPolicyInput) ExecutionPolicy {
	policy := ExecutionPolicy{
		stagedTargetMinor: input.StagedTargetMinor,
		fairValueMinor:    input.FairValueMinor,
		entryCostsMinor:   input.EntryCostsMinor,
		exitCostsMinor:    input.ExitCostsMinor,
		minimumRRPPM:      input.MinimumRRPPM,
		decisionDigest:    input.DecisionDigest,
		calendarDigest:    input.CalendarDigest,
		capSnapshotID:     input.CapSnapshotID,
	}
	policy.identity = breakoutPolicyIdentity(policy)
	return policy
}

func breakoutPolicyIdentity(policy ExecutionPolicy) string {
	if policy.decisionDigest == "" || policy.calendarDigest == "" || policy.capSnapshotID == "" {
		return ""
	}
	h := sha256.New()
	writeLineageString(h, breakoutPolicyDomain)
	writeLineageString(h, policy.stagedTargetMinor)
	writeLineageString(h, policy.fairValueMinor)
	writeLineageString(h, policy.entryCostsMinor)
	writeLineageString(h, policy.exitCostsMinor)
	writeLineageUint64(h, policy.minimumRRPPM)
	writeLineageString(h, policy.decisionDigest)
	writeLineageString(h, policy.calendarDigest)
	writeLineageString(h, policy.capSnapshotID)
	return breakoutPolicyDomain + ":sha256:" + hex.EncodeToString(h.Sum(nil))
}
