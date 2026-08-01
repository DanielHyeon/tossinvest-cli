// Package operatorview turns persisted domain evidence into transport-neutral
// operator-facing values. It does not evaluate an exit policy or read a clock.
package operatorview

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

const dash = "—"

// Source is one already-selected persisted snapshot and its read-time evidence.
// Callers decide freshness against their own approved bound before constructing
// this value; the adapter only applies the fail-closed display contract.
type Source struct {
	Snapshot          *exitpolicy.ExitLineSnapshot
	UnknownReason     string
	StaleReason       string
	RemainingQuantity string
	ObservationSource string
	ObservedAt        string
	EffectiveSource   string
}

// ExitLineView is the one representation shared by every operator transport.
// Empty domain values are normalized once here so an HTML or JSON adapter cannot
// independently turn missing evidence into zero.
type ExitLineView struct {
	Status     string
	StatusText string
	Reason     string

	EntryPrice        string
	InitialStop       string
	CurrentProtection string
	NextTarget        string
	NextProtection    string
	ProjectedQuantity string
	ObservedPrice     string
	HighWater         string
	Stage             string
	ActionText        string

	Policy             string
	DecisionID         string
	SnapshotID         string
	ObservationID      string
	ObservationSource  string
	EvaluatedAt        string
	EffectiveSource    string
	PositionGeneration string

	OneShare      bool
	OneShareText  string
	FinalExitText string
}

func (v ExitLineView) Fresh() bool   { return v.Status == "fresh" }
func (v ExitLineView) Stale() bool   { return v.Status == "stale" }
func (v ExitLineView) Unknown() bool { return v.Status == "unknown" }

// BuildExitLine maps a persisted snapshot without recomputing any price, rung,
// action, quantity, or identity. Stale and unknown evidence keeps provenance but
// hides actionable values behind em dashes.
func BuildExitLine(source Source) ExitLineView {
	view := emptyView()
	view.ObservationSource = dashIfEmpty(source.ObservationSource)
	view.EvaluatedAt = dashIfEmpty(source.ObservedAt)
	view.EffectiveSource = dashIfEmpty(source.EffectiveSource)

	if source.Snapshot == nil {
		view.Status, view.StatusText = "unknown", "근거 없음"
		view.Reason = reasonText(source.UnknownReason)
		return view
	}

	line := *source.Snapshot
	view.Policy = policyText(line.Policy)
	view.DecisionID = dashIfEmpty(line.DecisionID)
	view.SnapshotID = dashIfEmpty(line.SnapshotID)
	view.ObservationID = dashIfEmpty(line.ObservationID)
	view.PositionGeneration = generationText(line.PositionGeneration)

	if strings.TrimSpace(source.StaleReason) != "" {
		view.Status, view.StatusText = "stale", "오래된 평가"
		view.Reason = reasonText(source.StaleReason)
		return view
	}

	view.Status, view.StatusText = "fresh", "평가 완료"
	view.EntryPrice = dashIfEmpty(line.EntryPrice)
	view.InitialStop = dashIfEmpty(line.InitialStop)
	view.CurrentProtection = dashIfEmpty(line.CurrentProtection)
	view.NextTarget = dashIfEmpty(line.NextTarget)
	view.NextProtection = dashIfEmpty(line.NextProtection)
	view.ProjectedQuantity = projectedText(line)
	view.ObservedPrice = dashIfEmpty(line.ObservedPrice)
	view.HighWater = dashIfEmpty(line.HighWater)
	view.Stage = stageText(line)
	view.ActionText = actionText(line)

	if strings.TrimSpace(source.RemainingQuantity) == "1" && line.StateOnly && line.ProjectedQuantity == "0" {
		view.OneShare = true
		view.OneShareText = "중간 매도 없음 · 보호선 승격"
		view.FinalExitText = "최종 익절·손절 시 1주 전량"
	}
	return view
}

func emptyView() ExitLineView {
	return ExitLineView{
		EntryPrice: dash, InitialStop: dash, CurrentProtection: dash,
		NextTarget: dash, NextProtection: dash, ProjectedQuantity: dash,
		ObservedPrice: dash, HighWater: dash, Stage: dash, ActionText: dash,
		Policy: dash, DecisionID: dash, SnapshotID: dash, ObservationID: dash,
		ObservationSource: dash, EvaluatedAt: dash, EffectiveSource: dash,
		PositionGeneration: dash,
	}
}

func dashIfEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return dash
	}
	return strings.TrimSpace(value)
}

func policyText(identity exitpolicy.PolicyIdentity) string {
	id, version := strings.TrimSpace(identity.ID), strings.TrimSpace(identity.Version)
	switch {
	case id == "":
		return dash
	case version == "":
		return id
	default:
		return id + " · " + version
	}
}

func generationText(generation int64) string {
	if generation < 0 {
		return dash
	}
	return formatInt(generation)
}

func stageText(line exitpolicy.ExitLineSnapshot) string {
	if line.ActiveRung >= 0 {
		return "rung " + formatInt(int64(line.ActiveRung))
	}
	if strings.TrimSpace(string(line.RatchetLevel)) != "" {
		return string(line.RatchetLevel)
	}
	return dash
}

func projectedText(line exitpolicy.ExitLineSnapshot) string {
	if line.StateOnly && strings.TrimSpace(line.ProjectedQuantity) == "0" {
		return "0 (주문 없음)"
	}
	return dashIfEmpty(line.ProjectedQuantity)
}

func actionText(line exitpolicy.ExitLineSnapshot) string {
	switch line.Action {
	case exitpolicy.ActionBaselineBreach, exitpolicy.ActionLadderStop:
		return "보호선 이탈 · 전량 청산"
	case exitpolicy.ActionRatchetPartial, exitpolicy.ActionLadderPartial:
		if ratio := percentText(line.Ratio); ratio != "" {
			return "중간 익절 " + ratio
		}
		return "중간 익절"
	case exitpolicy.ActionLadderTakeProfit:
		return "최종 익절 · 전량 청산"
	case exitpolicy.ActionLadderHoldStopPromoted:
		return "매도 없이 보호선 승격"
	case exitpolicy.ActionNone:
		return "대기"
	default:
		return dashIfEmpty(string(line.Action))
	}
}

func percentText(ratio string) string {
	switch strings.TrimSpace(ratio) {
	case "0.25":
		return "25%"
	case "0.4", "0.40":
		return "40%"
	case "1", "1.0":
		return "100%"
	default:
		return ""
	}
}

func reasonText(reason string) string {
	switch strings.TrimSpace(reason) {
	case "observation_older_than_limit":
		return "평가 시각이 표시 허용 범위를 지났다"
	case "observation_in_future":
		return "평가 시각이 현재보다 미래다"
	case "invalid_observed_at":
		return "평가 시각을 확인할 수 없다"
	case "not_evaluated_yet":
		return "아직 exit 평가가 기록되지 않았다"
	case "no_saved_evaluation":
		return "저장된 exit 평가를 찾을 수 없다"
	case "legacy_snapshot_absent", "legacy_event":
		return "이전 원장에는 exit snapshot 근거가 없다"
	case "invalid_stored_snapshot", "invalid_effective_snapshot":
		return "저장된 exit snapshot의 무결성을 확인할 수 없다"
	case "legacy_policy_identity_unknown", "legacy_adoption_context_required":
		return "이전 원장의 exit 정책 정보를 확인할 수 없다"
	case "partial_snapshot_tuple", "partial_evaluated_tuple", "partial_seed_tuple", "partial_tuple":
		return "exit snapshot 정보가 일부만 기록되어 표시하지 않는다"
	case "partial_policy_tuple", "invalid_policy_identity":
		return "exit 정책 정보가 불완전하여 표시하지 않는다"
	case "flattened_snapshot_mismatch":
		return "exit snapshot 요약과 원본이 일치하지 않는다"
	case "ambiguous_exit_evidence":
		return "하나의 주문에 서로 다른 exit 판단 근거가 연결되어 표시하지 않는다"
	case "exit_evidence_unlinked":
		return "주문과 exit 판단 근거의 연결을 확인할 수 없다"
	case "lineage_cycle":
		return "주문 변경 계보가 순환하여 근거를 표시하지 않는다"
	case "lineage_ambiguous":
		return "주문 변경 계보가 여러 갈래여서 근거를 표시하지 않는다"
	case "lineage_depth_exceeded":
		return "주문 변경 계보가 표시 한도를 넘어 근거를 표시하지 않는다"
	case "lineage_scope_mismatch":
		return "주문 변경 계보의 계좌 또는 거래일이 일치하지 않아 표시하지 않는다"
	case "invalid_event_evidence":
		return "exit 판단 근거의 무결성을 확인할 수 없다"
	case "":
		return "exit snapshot 근거를 확인할 수 없다"
	default:
		return "exit snapshot 근거를 안전하게 표시할 수 없다"
	}
}

func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [20]byte
	pos := len(digits)
	for value > 0 {
		pos--
		digits[pos] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		pos--
		digits[pos] = '-'
	}
	return string(digits[pos:])
}
