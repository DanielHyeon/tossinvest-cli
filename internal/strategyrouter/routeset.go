package strategyrouter

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// routeSetDomain 은 집합 다이제스트의 도메인 분리 문자열이다.
// 다른 해시와 같은 값이 나올 수 없게 항상 맨 앞에 들어간다.
const routeSetDomain = "strategyrouter-route-set:v1"

// RouteSetResult 는 한 소유자 범위에서 *자격 있는 모든* 가족 후보다.
// Route 와 달리 여기서는 아무것도 고르지 않는다 — 고르는 일은 순수 평가가
// 끝난 뒤 조정자의 보정 점수 중재가 한다.
type RouteSetResult struct {
	Code                    RefusalCode
	Decisions               []RouteDecision
	ExistingOwner           bool
	CommonSafetyIndependent bool
	Mutations               uint64
	seal                    [32]byte
}

// SetDigest 는 결정 집합의 재현 가능한 신원이다. 거절이면 빈 문자열이다.
func (result RouteSetResult) SetDigest() string {
	if result.Code != RefusalNone || len(result.Decisions) == 0 {
		return ""
	}
	return routeSetDomain + ":sha256:" + hex.EncodeToString(routeSetSum(result.Decisions, result.ExistingOwner))
}

// Valid 는 집합이 만들어진 뒤 한 글자도 바뀌지 않았음을 증명한다.
func (result RouteSetResult) Valid() bool {
	if result.Code != RefusalNone || len(result.Decisions) == 0 || !result.CommonSafetyIndependent || result.Mutations != 0 {
		return false
	}
	var want [32]byte
	copy(want[:], routeSetSum(result.Decisions, result.ExistingOwner))
	return result.seal != ([32]byte{}) && result.seal == want
}

// RouteSet 은 Route 와 똑같은 전단 검증을 통과시킨 뒤, 승자를 하나 고르는 대신
// 자격 있는 후보를 전부 결정론적 순서로 내보낸다.
// Route 를 부르지 않고 같은 검증을 다시 유도한다 — Route 를 부르면 이 로트가
// 없애려는 바로 그 사전 선택이 다시 끼어들기 때문이다(결정 47).
func RouteSet(request RouteRequest) RouteSetResult {
	result := RouteSetResult{CommonSafetyIndependent: true}
	if !validOwnerKey(request.Key) || request.ExpectedOwnerRevision == 0 || request.EvaluatedAt.IsZero() {
		result.Code = RefusalInvalid
		return result
	}
	snapshot := request.Snapshot
	if snapshot.seal != ownerSnapshotSeal(snapshot) || !validOwnerKey(snapshot.Key) {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if snapshot.Key != request.Key {
		result.Code = RefusalScopeMismatch
		return result
	}
	if snapshot.Revision != request.ExpectedOwnerRevision {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if request.EvaluatedAt.Before(snapshot.ObservedAt) || !request.EvaluatedAt.Before(snapshot.FreshUntil) {
		result.Code = RefusalOwnerSnapshotStale
		return result
	}
	active := make([]Owner, 0, 1)
	for _, owner := range snapshot.Owners {
		if owner.Key != request.Key {
			result.Code = RefusalScopeMismatch
			return result
		}
		if !validOwner(owner) {
			result.Code = RefusalReconstructionMismatch
			return result
		}
		if owner.Active {
			active = append(active, owner)
		}
	}
	if len(active) > 1 {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if !validMarketRecord(request.MarketRecord) {
		result.Code = RefusalReconstructionMismatch
		return result
	}
	if request.MarketRecord.Market != request.Key.Market {
		result.Code = RefusalScopeMismatch
		return result
	}
	if request.ExpectedMarketRevision == 0 || request.MarketRecord.Revision != request.ExpectedMarketRevision {
		result.Code = RefusalVersionConflict
		return result
	}
	if EvaluateMarketLifecycle(request.MarketRecord, request.EvaluatedAt) != LifecycleReady {
		result.Code = RefusalDisabled
		return result
	}
	if len(active) == 1 {
		owner := active[0]
		if owner.Desired != StateOn || owner.Effective != StateOn {
			result.Code = RefusalDisabled
			return result
		}
		return sealRouteSet(result, []RouteDecision{{Key: owner.Key, Horizon: owner.Horizon, LaneID: owner.LaneID,
			LaneVersion: owner.LaneVersion, CampaignID: owner.CampaignID, ExistingOwner: true}}, true)
	}

	decisions := make([]RouteDecision, 0, len(request.Candidates))
	for _, candidate := range request.Candidates {
		if candidate.Key != request.Key {
			result.Code = RefusalScopeMismatch
			return result
		}
		if !validCandidateValue(candidate) {
			result.Code = RefusalInvalid
			return result
		}
		if !candidate.Eligible || candidate.Desired != StateOn || candidate.Effective != StateOn {
			continue
		}
		decisions = append(decisions, RouteDecision{Key: candidate.Key, Horizon: candidate.Horizon, LaneID: candidate.LaneID,
			LaneVersion: candidate.LaneVersion, EvidenceDigest: candidate.EvidenceDigest, ConfigDigest: candidate.ConfigDigest})
	}
	if len(decisions) == 0 {
		result.Code = RefusalDisabled
		return result
	}
	// 같은 레인이 두 번 들어오면 중재가 자기 자신과 겨루게 되므로 닫아 거절한다.
	seen := make(map[string]bool, len(decisions))
	for _, decision := range decisions {
		key := string(decision.Horizon) + "\x00" + decision.LaneID + "\x00" + decision.LaneVersion
		if seen[key] {
			result.Code = RefusalReconstructionMismatch
			return result
		}
		seen[key] = true
	}
	sort.Slice(decisions, func(i, j int) bool { return routeSetOrderKey(decisions[i]) < routeSetOrderKey(decisions[j]) })
	return sealRouteSet(result, decisions, false)
}

// routeSetOrderKey 는 입력 순서와 무관한 고정 순서를 만든다.
func routeSetOrderKey(decision RouteDecision) string {
	return string(decision.Horizon) + "\x00" + decision.LaneID + "\x00" + decision.LaneVersion
}

func sealRouteSet(result RouteSetResult, decisions []RouteDecision, existingOwner bool) RouteSetResult {
	result.Decisions = decisions
	result.ExistingOwner = existingOwner
	copy(result.seal[:], routeSetSum(decisions, existingOwner))
	return result
}

func routeSetSum(decisions []RouteDecision, existingOwner bool) []byte {
	h := sha256.New()
	writeString(h, routeSetDomain)
	if existingOwner {
		writeString(h, "existing-owner")
	} else {
		writeString(h, "fresh-candidates")
	}
	writeUint64(h, uint64(len(decisions)))
	for _, decision := range decisions {
		writeOwnerKey(h, decision.Key)
		writeString(h, string(decision.Horizon))
		writeString(h, decision.LaneID)
		writeString(h, decision.LaneVersion)
		writeString(h, decision.CampaignID)
		writeString(h, decision.EvidenceDigest)
		writeString(h, decision.ConfigDigest)
		if decision.ExistingOwner {
			writeString(h, "1")
		} else {
			writeString(h, "0")
		}
	}
	return h.Sum(nil)
}
