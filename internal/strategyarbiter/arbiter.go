// Package strategyarbiter 는 소유자 범위 하나 안에서 봉인된 제안 여러 개 중
// 최대 하나만 고르는 순수 중재자다.
//
// "소유자 범위"는 (계좌, 시장, 종목, 포지션 세대) 네 값이다. 같은 종목에
// 여러 전략군이 동시에 제안을 낼 수 있으므로, 누가 나갈지 정하는 규칙이
// 한 곳에 있어야 한다. 그 한 곳이 여기다.
//
// 이 패키지는 아무것도 바꾸지 않는다. 주문도 원장도 토글도 건드리지 않으며,
// 그 사실은 import 목록으로 증명된다(dependency_closure_test.go).
package strategyarbiter

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// Refusal 은 중재가 닫힌 이유다. 빈 값은 "닫히지 않았다"는 뜻이다.
//
// **이 여섯 개가 전부다.** 값은 동결 골든
// `openspec/changes/a112-.../analysis/goldens/four-family-runtime-v1.json` 의
// `refusal_enums.arbitration` 에서 그대로 읽었다. 그 파일은 "내용을 바꾸려면 Manager 가
// 쓴 OpenSpec amendment 와 새 manifest/receipt 가 필요하다"고 못박고 있으므로,
// 구현이 새 코드를 만들어 내보내는 것은 계약을 몰래 넓히는 일이다.
// 더 잘게 나눈 이유는 계약이 아니라 진단이며 Outcome.Detail 이 들고 간다.
type Refusal string

const (
	RefusalNone Refusal = ""
	// 승인된 공통 채점 기준이 없거나, 서로 다르거나, 점수를 견줄 수 없다.
	// 골든 매핑: incomparable_or_uncalibrated, singleton_without_calibration.
	RefusalUncalibrated Refusal = "ARBITRATION_UNCALIBRATED"
	// 최고점이 둘 이상이다. 골든 매핑: tie.
	RefusalTie Refusal = "ARBITRATION_TIE"
	// 활성 소유자가 둘 이상이거나, 소유자가 있는데 그 소유자 하나로 정리되지 않는다.
	// 골든 매핑: multiple_active_owners.
	RefusalMultipleOwner Refusal = "ARBITRATION_MULTIPLE_OWNER"
	// 소유자 스냅샷의 리비전이 기대와 다르거나 그 스냅샷이 이미 묵었다.
	// 골든 매핑: stale_owner_revision.
	RefusalStaleOwner Refusal = "ARBITRATION_STALE_OWNER"
	// 제안이 근거로 삼은 후보의 유효 기한이 지났다. 골든 매핑: stale_envelope.
	RefusalStaleEnvelope Refusal = "ARBITRATION_STALE_ENVELOPE"
	// 봉인으로 신원을 다시 세울 수 없다 — 제안 봉인 파손, 범위 불일치, 자격 집합
	// 재구성 실패, 같은 레인 중복, 자격 집합에 없는 레인, 증거·설정 다이제스트 어긋남.
	// 골든 매핑: seal_mismatch.
	RefusalSealMismatch Refusal = "ARBITRATION_SEAL_MISMATCH"
)

// 아래 상수들은 Outcome.Detail 에 들어가는 진단 문자열이다. 계약이 아니다 —
// 골든이 정한 여섯 코드만 계약이고, 이 값들은 사람이 원인을 좁히기 위한 것이다.
const (
	DetailInvalidRequest       = "invalid request identity"
	DetailNoProposal           = "no proposal to arbitrate"
	DetailProposalSeal         = "proposal seal does not match its contents"
	DetailScope                = "proposal or authority is outside the expected owner scope"
	DetailRouteSetUnusable     = "the sealed eligible set could not be reconstructed"
	DetailRouteSetDisagreement = "proposals are bound to different eligible sets"
	DetailDuplicateLane        = "the same lane appears twice"
	DetailIneligibleLane       = "the lane is not in the sealed eligible set"
	DetailEvidenceBinding      = "proposal evidence or config digest does not match its route decision"
	DetailUnknownFamily        = "the lane does not map to exactly one approved family score row"
	DetailScoreCeiling         = "the family score is above the approved ceiling"
	DetailCalibration          = "missing or disagreeing approved score version or calibration digest"
	DetailOwnerLane            = "an active owner exists but the proposals do not reduce to that owner"
	DetailMultipleOwners       = "the owner snapshot carries more than one active owner"
	DetailStaleOwnerRevision   = "the owner snapshot revision is not the expected one"
	DetailOwnerSnapshotStale   = "the owner snapshot is outside its freshness window"
	DetailStaleCandidate       = "the candidate validity deadline has passed"
	DetailTie                  = "two or more proposals share the highest score"
)

// Proposal 은 중재에 들어가는 제안 하나다.
//
// 채점 기준을 호출자가 말로 넘기지 않고 봉인된 권한 객체째로 받는다.
// "봉인을 확인했다"를 호출자가 스스로 주장할 수 있으면 그 확인은 없는 것과 같다.
type Proposal struct {
	Result    strategyflow.Result
	Authority strategyrouter.ProductionRouteAuthority
}

// Request 는 한 소유자 범위의 중재 입력이다. Market/Symbol 등은 호출자가
// *기대하는* 범위이며, 권한과 제안이 그 기대와 어긋나면 닫는다.
type Request struct {
	AccountRef         string
	Market             strategyrouter.Market
	Symbol             string
	PositionGeneration uint64
	ObservedAt         time.Time
	Proposals          []Proposal
}

// Outcome 은 중재 결과다. 거절이면 Selected 는 -1이다 — 거절을 무시하고
// 색인을 쓰는 호출자는 조용히 0번을 고르는 대신 그 자리에서 터진다.
//
// Refusal 은 계약(골든의 여섯 코드)이고 Detail 은 진단이다. 소비자가 판정에 쓰는 값은
// Refusal 하나뿐이며, Detail 은 사람이 읽으려고 있는 것이다.
type Outcome struct {
	Refusal         Refusal
	Detail          string
	Selected        int
	ExistingOwner   bool
	Family          strategyrouter.Family
	ScorePPM        uint32
	LineageIdentity string
}

func refuse(code Refusal, detail string) Outcome {
	return Outcome{Refusal: code, Detail: detail, Selected: -1}
}

// Arbitrate 는 한 소유자 범위에서 제안 최대 하나를 고른다.
//
// 고를 수 없으면 그 범위 전체를 닫는다. 문제 있는 제안 하나만 빼고 나머지를
// 계속 견주지 않는다 — 하나를 빼면 남은 것들의 비교 결과가 달라지고,
// 그러면 막으려던 것과 상관없는 다른 제안이 대신 통과한다.
func Arbitrate(request Request) Outcome {
	expected := strategyrouter.OwnerKey{AccountRef: request.AccountRef, Market: request.Market,
		Symbol: request.Symbol, PositionGeneration: request.PositionGeneration}
	if expected.AccountRef == "" || expected.Symbol == "" || expected.PositionGeneration == 0 ||
		(expected.Market != strategyrouter.MarketKR && expected.Market != strategyrouter.MarketUS) || request.ObservedAt.IsZero() {
		return refuse(RefusalSealMismatch, DetailInvalidRequest)
	}
	if len(request.Proposals) == 0 {
		return refuse(RefusalSealMismatch, DetailNoProposal)
	}

	// 1. 모든 제안이 같은 범위, 같은 자격 집합에 묶여 있는가.
	//
	//    자격 집합이 못 서면 그 이유를 소유자 스냅샷에서 먼저 좁힌다. RouteSet 은
	//    묵은 리비전과 소유자 중복을 같은 코드 하나로 돌려주는데, 골든은 그 둘에
	//    서로 다른 거절 코드를 요구한다. 코드를 되짚어 추측하지 않고 스냅샷을 직접 본다.
	primary := request.Proposals[0].Authority.Request()
	routed := strategyrouter.RouteSet(primary)
	if routed.Code != strategyrouter.RefusalNone || !routed.Valid() || len(routed.Decisions) == 0 {
		if code, detail := ownerSnapshotFault(primary); code != RefusalNone {
			return refuse(code, detail)
		}
		return refuse(RefusalSealMismatch, DetailRouteSetUnusable)
	}
	for _, proposal := range request.Proposals {
		if proposal.Authority.Request().Key != expected {
			return refuse(RefusalSealMismatch, DetailScope)
		}
		if strategyrouter.RouteSet(proposal.Authority.Request()).SetDigest() != routed.SetDigest() {
			return refuse(RefusalSealMismatch, DetailRouteSetDisagreement)
		}
	}

	// 2. 제안 하나하나가 스스로 성립하는가.
	seen := make(map[laneKey]bool, len(request.Proposals))
	for _, proposal := range request.Proposals {
		if !proposal.Result.ValidProposal() {
			return refuse(RefusalSealMismatch, DetailProposalSeal)
		}
		lineage := proposal.Result.Lineage
		if lineage.AccountRef != expected.AccountRef || lineage.Market != expected.Market ||
			lineage.Symbol != expected.Symbol || lineage.PositionGeneration != expected.PositionGeneration {
			return refuse(RefusalSealMismatch, DetailScope)
		}
		if lineage.CandidateValidUntilNS <= request.ObservedAt.UnixNano() {
			return refuse(RefusalStaleEnvelope, DetailStaleCandidate)
		}
		lane := laneKey{horizon: lineage.Horizon, laneID: lineage.LaneID, laneVersion: lineage.LaneVersion}
		if seen[lane] {
			return refuse(RefusalSealMismatch, DetailDuplicateLane)
		}
		seen[lane] = true
	}

	// 3. 승인된 공통 채점 기준이 있는가. 제안이 하나뿐이어도 요구한다.
	//    보정 값이 비어 있는 권한은 SealsValid 가 이미 거절한다. 그래서 빈 값을
	//    여기서 또 검사하지 않는다 — 같은 판단을 두 곳에 적으면 한쪽은 영영
	//    실행되지 않는 죽은 코드가 되고, 죽은 코드는 지켜 주는 것이 없다.
	calibration := request.Proposals[0].Authority.Calibration()
	for _, proposal := range request.Proposals {
		if !proposal.Authority.SealsValid() || proposal.Authority.Calibration() != calibration {
			return refuse(RefusalUncalibrated, DetailCalibration)
		}
	}

	// 4. 활성 소유자가 있으면 그 소유자만 이어 간다. 점수는 소유자를 못 바꾼다.
	if routed.ExistingOwner {
		return continueExistingOwner(request, routed)
	}

	// 5. 소유자가 없을 때만 점수를 견준다.
	return selectHighestScore(request, routed)
}

// ownerSnapshotFault 는 자격 집합이 서지 못한 이유를 소유자 스냅샷에서 좁힌다.
// 골든이 묵은 리비전과 소유자 중복에 서로 다른 코드를 요구하는데, RouteSet 은
// 둘을 같은 코드로 뭉치기 때문이다. 여기서 아무것도 못 찾으면 봉인 문제로 본다.
func ownerSnapshotFault(request strategyrouter.RouteRequest) (Refusal, string) {
	active := 0
	for _, owner := range request.Snapshot.Owners {
		if owner.Active {
			active++
		}
	}
	if active > 1 {
		return RefusalMultipleOwner, DetailMultipleOwners
	}
	if request.ExpectedOwnerRevision != 0 && request.Snapshot.Revision != request.ExpectedOwnerRevision {
		return RefusalStaleOwner, DetailStaleOwnerRevision
	}
	if !request.EvaluatedAt.IsZero() && !request.Snapshot.ObservedAt.IsZero() &&
		(request.EvaluatedAt.Before(request.Snapshot.ObservedAt) || !request.EvaluatedAt.Before(request.Snapshot.FreshUntil)) {
		return RefusalStaleOwner, DetailOwnerSnapshotStale
	}
	return RefusalNone, ""
}

type laneKey struct {
	horizon     strategyrouter.Horizon
	laneID      string
	laneVersion string
}

// continueExistingOwner 는 이미 자리를 잡은 소유자의 제안 하나만 통과시킨다.
func continueExistingOwner(request Request, routed strategyrouter.RouteSetResult) Outcome {
	if len(routed.Decisions) != 1 || len(request.Proposals) != 1 {
		return refuse(RefusalMultipleOwner, DetailOwnerLane)
	}
	owner := routed.Decisions[0]
	lineage := request.Proposals[0].Result.Lineage
	if lineage.Horizon != owner.Horizon || lineage.LaneID != owner.LaneID ||
		lineage.LaneVersion != owner.LaneVersion || lineage.CampaignID != owner.CampaignID {
		return refuse(RefusalMultipleOwner, DetailOwnerLane)
	}
	score, code, detail := familyScore(request.Proposals[0])
	if code != RefusalNone {
		return refuse(code, detail)
	}
	return Outcome{Selected: 0, ExistingOwner: true, Family: score.Family,
		ScorePPM: score.ScorePPM, LineageIdentity: lineage.Identity}
}

// selectHighestScore 는 자격 있는 제안 중 점수가 가장 높은 하나를 고른다.
func selectHighestScore(request Request, routed strategyrouter.RouteSetResult) Outcome {
	eligible := make(map[laneKey]strategyrouter.RouteDecision, len(routed.Decisions))
	for _, decision := range routed.Decisions {
		eligible[laneKey{horizon: decision.Horizon, laneID: decision.LaneID, laneVersion: decision.LaneVersion}] = decision
	}
	best, ties := -1, 0
	var winner strategyrouter.ProductionRouteFamilyScore
	for index, proposal := range request.Proposals {
		lineage := proposal.Result.Lineage
		decision, ok := eligible[laneKey{horizon: lineage.Horizon, laneID: lineage.LaneID, laneVersion: lineage.LaneVersion}]
		if !ok {
			return refuse(RefusalSealMismatch, DetailIneligibleLane)
		}
		// 레인 이름이 같다고 같은 제안이 아니다. Propose 는 그때의 경로 결정에서
		// 증거·설정 다이제스트를 계보에 박아 넣는다(flow.go:67-68). 지금 자격 집합의
		// 결정과 그 두 값이 다르면, 이 제안은 *지금이 아닌 어떤 증거* 로 만들어진 것이다.
		if lineage.RouterEvidenceDigest != decision.EvidenceDigest || lineage.ConfigDigest != decision.ConfigDigest {
			return refuse(RefusalSealMismatch, DetailEvidenceBinding)
		}
		score, code, detail := familyScore(proposal)
		if code != RefusalNone {
			return refuse(code, detail)
		}
		switch {
		case best < 0 || score.ScorePPM > winner.ScorePPM:
			best, ties, winner = index, 1, score
		case score.ScorePPM == winner.ScorePPM:
			ties++
		}
	}
	// best 가 -1 로 남는 경우는 제안이 하나도 없을 때뿐인데, Arbitrate 가
	// 그 요청을 이미 거절하고 여기까지 오지 않는다. 그래서 따로 검사하지
	// 않는다 — 실행되지 않는 검사는 지켜 주는 것이 없다. 설령 빈 목록이
	// 들어와도 ties 가 0이라 아래에서 닫히고, 색인은 쓰이지 않는다.
	if ties != 1 {
		return refuse(RefusalTie, DetailTie)
	}
	return Outcome{Selected: best, Family: winner.Family, ScorePPM: winner.ScorePPM,
		LineageIdentity: request.Proposals[best].Result.Lineage.Identity}
}

// ProposalFamily 는 제안이 묶인 전략군을 알려준다. 봉인된 가족 점수 표에서
// 정확히 한 행에 붙지 않으면 빈 값이다.
//
// 조정자가 중복 제거 열쇠를 만들 때 가족이 필요해서 내보낸다. 같은 판단을
// 조정자가 따로 구현하면 두 곳이 언젠가 서로 다른 가족을 말하게 된다.
//
// **붙지 않는 이유는 여기서 판정하지 않는다.** 채점 기준이 아예 없는 권한은
// Arbitrate 가 보정 문제로 먼저 거절하는데, 열쇠를 만드는 자리에서 같은 입력을
// "모르는 가족"으로 먼저 판정해 버리면 운영자가 보는 진단이 원인에서 멀어진다.
func ProposalFamily(proposal Proposal) strategyrouter.Family {
	score, code, _ := familyScore(proposal)
	if code != RefusalNone {
		return ""
	}
	return score.Family
}

// familyScore 는 제안의 레인을 봉인된 가족 점수 행 정확히 하나에 붙인다.
// 두 행에 걸치거나 한 행에도 안 붙으면 그 제안은 견줄 수 없는 것이다.
func familyScore(proposal Proposal) (strategyrouter.ProductionRouteFamilyScore, Refusal, string) {
	lineage := proposal.Result.Lineage
	matches := 0
	var found strategyrouter.ProductionRouteFamilyScore
	for _, score := range proposal.Authority.FamilyScores() {
		if score.Horizon == lineage.Horizon && score.LaneID == lineage.LaneID && score.LaneVersion == lineage.LaneVersion {
			matches++
			found = score
		}
	}
	if matches != 1 || !found.Family.Known() {
		return strategyrouter.ProductionRouteFamilyScore{}, RefusalUncalibrated, DetailUnknownFamily
	}
	if found.ScorePPM > strategyrouter.ScorePPMMax {
		return strategyrouter.ProductionRouteFamilyScore{}, RefusalUncalibrated, DetailScoreCeiling
	}
	return found, RefusalNone, ""
}
