package engine

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112 태스크 7.3 — 여덟 레인 · 두 조정자를 읽기 전용 투영의 additive 자식으로 옮기는 함수들.
//
// 여기서는 아무것도 바꾸지 않는다. 레인은 읽기 접근자로만(Health · Pending · Dropped · … — 각자 레인 잠금 아래 값 하나), 관측 기록은
// 런타임 읽기 잠금 아래에서, 조정 결과는 이미 발행된 조립 값에서 읽는다. 투입(Offer) · 실패(Fail) · 기록(record)을 부르는 순간
// 「화면을 보는 일」이 레인을 움직이게 된다 — 시험이 Read 전후 레인 상태 동일을 단언하고 그런 변이를 잡는다.

// strategyLaneEvidenceDigest 는 이 레인이 이번 물결에 받은 봉인된 제안의 근거 digest 다. 시장 레코드(strategyProjectionFromAssembly)와
// 같은 선택: 레인 근거가 있으면 그것, 없으면 후보 근거. 입력이 없으면 빈 값.
func strategyLaneEvidenceDigest(input strategyworker.Input) string {
	lineage := input.Proposal.Result.Lineage
	if lineage.LaneEvidenceDigest != "" {
		return lineage.LaneEvidenceDigest
	}
	return lineage.CandidateEvidenceDigest
}

// projection 은 여덟 레인의 투영을 생산 목록 순서로 돌려준다.
//
// 건강 · 잠금 · 칸 · 기한은 레인에서 **지금** 읽고, 물결 · 트리거 · 시작 · 결과 · desired/effective · digest 는 마지막 관측에서 읽는다.
// 아직 돌지 않은 레인(물결 0)은 관측 사실 없이 건강 · 정책만 싣는다. 잠금 순서는 런타임 → 레인이다(evaluate 의 lanesFor 와 같음).
func (runtime *strategyLaneRuntime) projection() []strategyprojection.LaneRuntimeProjection {
	if runtime == nil {
		return nil
	}
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	values := make([]strategyprojection.LaneRuntimeProjection, 0, len(runtime.lanes))
	for _, lane := range runtime.lanes {
		observation, seen := runtime.observed[lane.Key()]
		values = append(values, strategyLaneProjection(lane, observation, seen && observation.Wave > 0))
	}
	return values
}

func strategyLaneProjection(lane *strategyworker.Lane, observation strategyLaneObservation, observed bool,
) strategyprojection.LaneRuntimeProjection {
	key, policy := lane.Key(), lane.Policy()
	health := strategyprojection.LaneHealth(lane.Health())
	version, deadline := policy.Version(), policy.CycleDeadline().Milliseconds()
	value := strategyprojection.LaneRuntimeProjection{Market: strategyprojection.Market(key.Market), Family: string(key.Family),
		LaneID: key.LaneID, LaneVersion: key.LaneVersion, Horizon: string(lane.Horizon()),
		Desired: strategyprojection.StateOff, Effective: strategyprojection.StateOff,
		Runtime: strategyprojection.LaneRuntime(lane.Runtime()), Health: &health,
		ConsecutiveFailures: lane.ConsecutiveFailures(), LatchRevision: lane.LatchRevision(),
		FirstFailure: strategyprojection.NormalizedText(lane.FirstFailure()),
		Pending:      lane.Pending(), Dropped: lane.Dropped(), Abandoned: lane.Abandoned(),
		PolicyVersion: &version, CycleDeadlineMS: &deadline,
		NextDueAt: projectionTime(lane.NextDue()), RestartNotBefore: projectionTime(lane.RestartNotBefore())}
	if !observed {
		return value
	}
	value.CycleGeneration = observation.Wave
	value.Desired, value.Effective = strategyprojection.State(observation.Desired), strategyprojection.State(observation.Effective)
	trigger := strategyprojection.LaneTrigger(observation.Trigger)
	value.Trigger = &trigger
	if observation.Trigger == strategyworker.TriggerEnqueued {
		start := strategyprojection.LaneStart(observation.Start)
		value.Start = &start
		if observation.Start == strategyworker.StartAdmitted {
			value.Abnormal = observation.Abnormal
			if observation.Outcome != "" {
				outcome := strategyprojection.LaneOutcome(observation.Outcome)
				value.Outcome = &outcome
			}
			// 거절 코드는 REFUSED 결과에만 싣는다(판정 (A) — Validate 짝). 다른 결과의 Cycle.Refusal 은 빈 값이지만 그 사실에
			// 기대지 않고 결과로 가른다.
			if observation.Outcome == strategyworker.OutcomeRefused {
				value.Refusal = projectionOptional(string(observation.Refusal))
			}
		}
	}
	value.SnapshotDigest = projectionOptional(observation.SnapshotDigest)
	value.EvidenceDigest = projectionOptional(observation.EvidenceDigest)
	return value
}

// strategyCoordinatorProjection 은 한 시장의 제안 조정 스냅숏과 승인된 소유자 범위 전부다.
//
// selected 는 주문 경로(runProductionStrategyMarketCycle)와 **같은** handoff 목록(dispatchHandoffs)에서 승인되고 봉인이 유효한 범위다
// (R4 — 서명 활성화된 두 범위 시장은 둘 다). 시장 레코드가 첫 범위를 고르는 순회와 같은 술어이고, 그 순회는 바꾸지 않는다.
// 사유가 비어 있으면(조립이 이 시장의 조정 스냅숏을 만들지 않음) 미관측이다.
func strategyCoordinatorProjection(market StrategyMarket, authority strategyProposalMarketAuthority) strategyprojection.CoordinatorProjection {
	value := strategyprojection.CoordinatorProjection{Market: strategyprojection.Market(market), GatedOutcomes: []string{},
		Selected: []strategyprojection.SelectedScopeProjection{}}
	snapshot := authority.snapshot
	if snapshot.Reason == "" {
		return value
	}
	reason := string(snapshot.Reason)
	value.Reason, value.Ready = &reason, snapshot.Ready
	value.RoutedCount, value.ProposedCount, value.RefusedCount = snapshot.RoutedCount, snapshot.ProposedCount, snapshot.RefusedCount
	value.GatedCount, value.QueueDropCount = snapshot.GatedCount, snapshot.QueueDropCount
	value.GatedOutcomes = append(value.GatedOutcomes, snapshot.GatedOutcomes...)
	value.ArbitrationRefusal = projectionOptional(snapshot.ArbitrationRefusal)
	if digest := projectionDigest(snapshot.ProposalSetDigest); digest != "" {
		value.ProposalSetDigest = &digest
	}
	for _, handoff := range authority.dispatchHandoffs() {
		scoped, admitted := handoff.Single()
		if !admitted || !scoped.ValidProposal() {
			continue
		}
		lineage := scoped.Lineage
		selected := strategyprojection.SelectedScopeProjection{Symbol: lineage.Symbol, LaneID: lineage.LaneID,
			CampaignID: lineage.CampaignID, LineageIdentity: lineage.Identity, ConfigDigest: projectionOptional(lineage.ConfigDigest)}
		if route, ok := authority.routeForSelectedScope(lineage); ok {
			calibration := route.route.Calibration()
			selected.ScoreVersion = projectionOptional(calibration.ScoreVersion)
			selected.CalibrationDigest = projectionOptional(calibration.CalibrationDigest)
		}
		value.Selected = append(value.Selected, selected)
	}
	return value
}

// routeForSelectedScope 는 승인된 범위의 경로 권한(중재 보정이 사는 곳)을 되찾는다(태스크 7.3 calibration 계보, 판정 (A)).
//
// 범위 해소는 1차 레그 봉인과 **같은** 선택기(`authorityForOwnerScope` — 소유자 범위로 정확히 하나)를 그대로 쓰고, 고른 항목의 봉인
// 신원이 건너온 계보와 같을 때만 그 항목의 경로를 돌려준다. 선택기를 여기 다시 쓰면 판정이 둘이 된다. 하나로 못 되찾으면 false —
// 투영은 보정을 null 로 둔다(추론하지 않음). 읽기만 한다.
func (authority strategyProposalMarketAuthority) routeForSelectedScope(lineage strategyflow.Lineage) (strategyRouteEntryAuthority, bool) {
	chosen, ok := authority.authorityForOwnerScope(lineage)
	if !ok || chosen.Proposal().Lineage.Identity != lineage.Identity {
		return strategyRouteEntryAuthority{}, false
	}
	// 같은 소유자 범위의 항목은 위 선택기가 정확히 하나로 보장했으므로, 그 신원을 가진 항목이 바로 그것이다.
	for _, entry := range authority.entries {
		if entry.authority.Proposal().Lineage.Identity == lineage.Identity {
			return entry.route, true
		}
	}
	return strategyRouteEntryAuthority{}, false
}

func projectionOptional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func projectionTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	value = value.UTC()
	return &value
}
