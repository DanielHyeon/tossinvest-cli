package strategyprojection

// a112 태스크 7.3 — envelope 의 additive 자식 `lanes[8]` · `coordinators[2]`.
//
// 골든 four-family-runtime-v1 의 operator_compatibility 가 정한 모양이다: 기존 시장 레코드는 그대로 두고(legacy reader 무변), 고정 결정적
// 순서의 자식 두 묶음을 더한다. 이 표면은 읽기 전용이다 — 여기에는 활성화 · 주문 · 원장 · 레인 상태를 바꾸는 방법이 없다.
//
// 이 패키지는 import 0 인 잎이라 레인 목록을 스스로 들고 있다(defaultLaneTable). 옮겨 적은 표이므로 양쪽을 못 박는다:
// 골든 쪽은 이 패키지 시험이 골든 파일을 직접 읽어, 생산 레인 쪽은 strategyworker 시험이 ProductionLanes 와 대조한다.
// 어휘(트리거 · 시작 · 결과 · 건강 · 조정자 사유)도 같은 식으로 census 시험이 원본 상수 전부와 대조한다 — 원본이 값을 하나 더하면
// 여기서 모르는 값이 되어 Validate 가 스냅숏 전체를 거절하므로, 빠짐도 지어냄도 시험 실패로 드러난다.

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// LaneHealth 는 레인 하나의 건강 상태다(strategyworker.LaneHealth 와 같은 어휘).
type LaneHealth string

const (
	LaneHealthy  LaneHealth = "HEALTHY"
	LaneDegraded LaneHealth = "DEGRADED"
	LaneLatched  LaneHealth = "LATCHED"
)

// LaneTrigger 는 마지막 관측 물결에서 레인 칸에 투입을 넣은 결과다.
type LaneTrigger string

const (
	LaneTriggerEnqueued LaneTrigger = "ENQUEUED"
	LaneTriggerDisabled LaneTrigger = "DISABLED"
	LaneTriggerFull     LaneTrigger = "FULL"
)

// LaneStart 는 투입이 들어간 물결에서 사이클을 열었는지, 못 열었다면 왜인지다.
type LaneStart string

const (
	LaneStartAdmitted  LaneStart = "ADMITTED"
	LaneStartLatched   LaneStart = "LATCHED"
	LaneStartInFlight  LaneStart = "IN_FLIGHT"
	LaneStartBackoff   LaneStart = "BACKOFF"
	LaneStartTooSoon   LaneStart = "TOO_SOON"
	LaneStartNoTrigger LaneStart = "NO_TRIGGER"
)

// LaneOutcome 은 열린 사이클의 결과 종류다.
type LaneOutcome string

const (
	LaneOutcomeEmitted LaneOutcome = "EMITTED"
	LaneOutcomeDormant LaneOutcome = "DORMANT"
	LaneOutcomeRefused LaneOutcome = "REFUSED"
	LaneOutcomeLatched LaneOutcome = "LATCHED"
)

// LaneRuntime 은 worker 의 runtime 상태다. 정의된 값은 골든의 UNOBSERVED 하나뿐이다(판정 Q3 — 골든 밖 어휘를 지어내지 않음).
// 관측된 사실은 health · trigger · start · outcome 이 나른다. SHADOW 같은 값은 태스크 7.3.1 의 몫이다.
type LaneRuntime string

const LaneRuntimeUnobserved LaneRuntime = "UNOBSERVED"

// LaneRuntimeProjection 은 레인 하나의 읽기 전용 상태다.
//
// 미관측(health null): 이 프로세스가 레인 런타임을 아직 세우지 않았거나(첫 생산 주기 전) 엔진에 닿지 못한 스냅숏이다. 그때는
// 열쇠와 골든 기본값(OFF/OFF/UNOBSERVED) 말고 아무 사실도 싣지 않는다 — 건강조차 추론하지 않는다.
//
// CycleGeneration(판정 Q1=(B)): 엔진이 시장마다 센 evaluate 물결 번호 중 이 레인이 **마지막으로 관측된** 번호다. 프로세스 수명이고
// 재시작하면 0 부터 다시 센다. 0 은 「이 프로세스에서 아직 관측 없음」이다. 같은 시장의 다른 레인보다 작으면 그 레인은 묵었다.
//
// Desired · Effective 는 마지막 관측 물결이 받은 서명 활성화가 이 레인 열쇠에 대해 말한 값이다(관측 시점 계산 — 활성화에는 만료가
// 있으므로 다음 물결에서 바뀔 수 있다). 기한(판정 Q2)은 기존 읽기 접근자만 쓴다: 서버 정책의 사이클 마감(ms) · 다음 카덴스 시각 ·
// 재시작 backoff 시각(영값이면 null).
type LaneRuntimeProjection struct {
	Market              Market       `json:"market"`
	Family              string       `json:"family"`
	LaneID              string       `json:"laneId"`
	LaneVersion         string       `json:"laneVersion"`
	Horizon             string       `json:"horizon"`
	Desired             State        `json:"desired"`
	Effective           State        `json:"effective"`
	Runtime             LaneRuntime  `json:"runtime"`
	Health              *LaneHealth  `json:"health"`
	ConsecutiveFailures uint64       `json:"consecutiveFailures"`
	LatchRevision       uint64       `json:"latchRevision"`
	FirstFailure        *string      `json:"firstFailure"`
	CycleGeneration     uint64       `json:"cycleGeneration"`
	Trigger             *LaneTrigger `json:"trigger"`
	Start               *LaneStart   `json:"start"`
	Outcome             *LaneOutcome `json:"outcome"`
	// Refusal 은 마지막 관측 물결의 결과가 REFUSED 일 때만 있는 골든 중재 코드다(refusal_enums.arbitration 여섯 중 하나 — 태스크 7.3
	// 「first refusal」, Manager 판정 (A)). REFUSED 가 아니면 null.
	Refusal          *string    `json:"refusal"`
	Abnormal         bool       `json:"abnormal"`
	Pending          int        `json:"pending"`
	Dropped          uint64     `json:"dropped"`
	Abandoned        uint64     `json:"abandoned"`
	PolicyVersion    *string    `json:"policyVersion"`
	CycleDeadlineMS  *int64     `json:"cycleDeadlineMs"`
	NextDueAt        *time.Time `json:"nextDueAt"`
	RestartNotBefore *time.Time `json:"restartNotBefore"`
	SnapshotDigest   *string    `json:"snapshotDigest"`
	EvidenceDigest   *string    `json:"evidenceDigest"`
}

// CoordinatorProjection 은 시장 하나의 제안 조정 결과다(골든 coordinator_key_fields=[market]).
//
// Reason null 은 미관측이다(조립이 아직 발행되지 않았거나 엔진에 닿지 못함). Selected 는 주문 경로와 **같은** handoff 목록에서 승인된
// 소유자 범위 전부다 — 서명 활성화된 두 범위 시장은 둘 다 보인다(R4). 시장 레코드(Markets)는 오늘처럼 첫 범위만 싣는다.
type CoordinatorProjection struct {
	Market             Market                    `json:"market"`
	Reason             *string                   `json:"reason"`
	Ready              bool                      `json:"ready"`
	RoutedCount        int                       `json:"routedCount"`
	ProposedCount      int                       `json:"proposedCount"`
	RefusedCount       int                       `json:"refusedCount"`
	GatedCount         int                       `json:"gatedCount"`
	GatedOutcomes      []string                  `json:"gatedOutcomes"`
	QueueDropCount     uint64                    `json:"queueDropCount"`
	ArbitrationRefusal *string                   `json:"arbitrationRefusal"`
	ProposalSetDigest  *string                   `json:"proposalSetDigest"`
	Selected           []SelectedScopeProjection `json:"selected"`
}

// SelectedScopeProjection 은 승인된 소유자 범위 하나다. LineageIdentity 는 봉인된 계보의 SHA-256 신원이다 — 계좌 성분은 해시 입력
// 안에만 있고 원문은 싣지 않는다(review 7.3 「계좌 성분」).
//
// ConfigDigest 는 계보의 설정 digest, ScoreVersion · CalibrationDigest 는 그 범위 항목의 경로 권한이 봉인한 중재 보정이다(태스크 7.3
// 「config/calibration lineage」, Manager 판정 (A)). 값이 없거나 범위 항목을 하나로 되찾지 못하면 null — 추론하지 않는다.
type SelectedScopeProjection struct {
	Symbol            string  `json:"symbol"`
	LaneID            string  `json:"laneId"`
	CampaignID        string  `json:"campaignId"`
	LineageIdentity   string  `json:"lineageIdentity"`
	ConfigDigest      *string `json:"configDigest"`
	ScoreVersion      *string `json:"scoreVersion"`
	CalibrationDigest *string `json:"calibrationDigest"`
}

// laneKey 는 골든 서술자 한 행이다(worker_key_fields + horizon).
type laneKey struct {
	market                       Market
	family, laneID, version, hor string
}

// defaultLaneTable 은 골든 descriptors 순서 그대로의 여덟 레인이다(생산 ProductionWorkers 순서와 같음 — 시험이 양쪽을 잰다).
var defaultLaneTable = [8]laneKey{
	{MarketKR, "CONTINUATION", "kr_short_flow_continuation_v1", "v1", "SHORT"},
	{MarketUS, "CONTINUATION", "us_short_participation_continuation_v1", "v1", "SHORT"},
	{MarketKR, "REVERSAL", "kr_short_absorption_reversal_v1", "v1", "SHORT"},
	{MarketUS, "REVERSAL", "us_short_dislocation_reversal_v1", "v1", "SHORT"},
	{MarketKR, "WEEKLY_VALUE", "kr_weekly_disclosure_value_v1", "v1", "WEEKLY"},
	{MarketUS, "WEEKLY_VALUE", "us_weekly_disclosure_value_v1", "v1", "WEEKLY"},
	{MarketKR, "BREAKOUT_RETEST", "kr_short_breakout_retest_v1", "v1", "SHORT"},
	{MarketUS, "BREAKOUT_RETEST", "us_short_breakout_retest_v1", "v1", "SHORT"},
}

// defaultLanes 는 미관측 여덟 레인이다(골든 기본값).
func defaultLanes() []LaneRuntimeProjection {
	lanes := make([]LaneRuntimeProjection, 0, len(defaultLaneTable))
	for _, key := range defaultLaneTable {
		lanes = append(lanes, LaneRuntimeProjection{Market: key.market, Family: key.family, LaneID: key.laneID, LaneVersion: key.version,
			Horizon: key.hor, Desired: StateOff, Effective: StateOff, Runtime: LaneRuntimeUnobserved})
	}
	return lanes
}

// defaultCoordinators 는 미관측 두 조정자다. 목록은 null 이 아니라 빈 배열이다(계약을 한 모양으로).
func defaultCoordinators() []CoordinatorProjection {
	return []CoordinatorProjection{
		{Market: MarketKR, GatedOutcomes: []string{}, Selected: []SelectedScopeProjection{}},
		{Market: MarketUS, GatedOutcomes: []string{}, Selected: []SelectedScopeProjection{}},
	}
}

// LaneJSONFields · CoordinatorJSONFields · SelectedScopeJSONFields 는 계약 JSON 이름(사전순)이다. 이 패키지 시험이 실제 직렬화와,
// httpapi 시험이 OpenAPI 스키마와 대조한다.
func LaneJSONFields() []string {
	return []string{"abandoned", "abnormal", "consecutiveFailures", "cycleDeadlineMs", "cycleGeneration", "desired", "dropped", "effective",
		"evidenceDigest", "family", "firstFailure", "health", "horizon", "laneId", "laneVersion", "latchRevision", "market", "nextDueAt",
		"outcome", "pending", "policyVersion", "refusal", "restartNotBefore", "runtime", "snapshotDigest", "start", "trigger"}
}

func CoordinatorJSONFields() []string {
	return []string{"arbitrationRefusal", "gatedCount", "gatedOutcomes", "market", "proposalSetDigest", "proposedCount", "queueDropCount",
		"ready", "reason", "refusedCount", "routedCount", "selected"}
}

func SelectedScopeJSONFields() []string {
	return []string{"calibrationDigest", "campaignId", "configDigest", "laneId", "lineageIdentity", "scoreVersion", "symbol"}
}

// LaneTriggers · LaneStarts · LaneOutcomes · LaneHealths 는 받는 어휘다(census 시험이 strategyworker 상수 전부와 대조).
func LaneTriggers() []string {
	return []string{string(LaneTriggerEnqueued), string(LaneTriggerDisabled), string(LaneTriggerFull)}
}

func LaneStarts() []string {
	return []string{string(LaneStartAdmitted), string(LaneStartLatched), string(LaneStartInFlight), string(LaneStartBackoff),
		string(LaneStartTooSoon), string(LaneStartNoTrigger)}
}

func LaneOutcomes() []string {
	return []string{string(LaneOutcomeEmitted), string(LaneOutcomeDormant), string(LaneOutcomeRefused), string(LaneOutcomeLatched)}
}

func LaneHealths() []string {
	return []string{string(LaneHealthy), string(LaneDegraded), string(LaneLatched)}
}

// CoordinatorReasons 는 엔진 StrategyProposalReason 상수 전부다(엔진 census 시험이 대조).
func CoordinatorReasons() []string {
	return []string{"READY", "ROUTE_NOT_READY", "FX_NOT_READY", "PROPOSAL_AUTHORITY_INVALID", "NO_ACCEPTED_PROPOSAL", "INTERNAL_FAILURE",
		"ARBITRATION_REFUSED", "PROPOSAL_QUEUE_OVERFLOW", "PROPOSAL_PRODUCTION_FAULT", "FAMILY_GATE_CLOSED"}
}

// ArbitrationRefusals 는 골든 refusal_enums.arbitration 여섯 그대로다.
func ArbitrationRefusals() []string {
	return []string{"ARBITRATION_UNCALIBRATED", "ARBITRATION_TIE", "ARBITRATION_MULTIPLE_OWNER", "ARBITRATION_STALE_OWNER",
		"ARBITRATION_STALE_ENVELOPE", "ARBITRATION_SEAL_MISMATCH"}
}

// maxTextBytes 는 자유 문장(firstFailure) 상한이다.
const maxTextBytes = 512

// NormalizedText 는 자유 문장을 투영이 받는 모양으로 만든다: 제어 문자는 공백으로, 앞뒤 공백 제거, 512 바이트에서 룬 경계로 자름.
// 남는 것이 없으면 nil. 규칙이 Validate(validText)와 한 곳에 있어야 엔진이 만든 값을 Validate 가 거절하는 일이 없다.
func NormalizedText(value string) *string {
	value = strings.ToValidUTF8(value, "�")
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
	value = strings.TrimSpace(value)
	for len(value) > maxTextBytes {
		_, size := utf8.DecodeLastRuneInString(value)
		value = value[:len(value)-size]
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func validText(value string) bool {
	normalized := NormalizedText(value)
	return normalized != nil && *normalized == value
}

func validateLanes(lanes []LaneRuntimeProjection) error {
	if len(lanes) != len(defaultLaneTable) {
		return errors.New("lanes must hold exactly eight lanes")
	}
	for index, lane := range lanes {
		key := defaultLaneTable[index]
		if lane.Market != key.market || lane.Family != key.family || lane.LaneID != key.laneID || lane.LaneVersion != key.version ||
			lane.Horizon != key.hor {
			return errors.New("lanes out of the fixed production order")
		}
		if err := validateLane(lane); err != nil {
			return errors.New("lane " + key.laneID + ": " + err.Error())
		}
	}
	return nil
}

func validateLane(lane LaneRuntimeProjection) error {
	if !validState(lane.Desired) || !validState(lane.Effective) || lane.Runtime != LaneRuntimeUnobserved {
		return errors.New("invalid desired, effective or runtime")
	}
	// 거절 코드는 REFUSED 결과에만 있고 REFUSED 결과에는 반드시 있다(판정 (A)). 결과와 무관하게 싣는 생산자는 여기서 거절된다.
	refused := lane.Outcome != nil && *lane.Outcome == LaneOutcomeRefused
	if (lane.Refusal != nil) != refused || lane.Refusal != nil && !member(*lane.Refusal, ArbitrationRefusals()) {
		return errors.New("refusal must be a golden arbitration code exactly when the outcome is REFUSED")
	}
	if lane.Health == nil {
		// 미관측: 열쇠와 기본값 말고 아무 사실도 없어야 한다.
		if lane.Desired != StateOff || lane.Effective != StateOff || lane.ConsecutiveFailures != 0 || lane.LatchRevision != 0 ||
			lane.FirstFailure != nil || lane.CycleGeneration != 0 || lane.Trigger != nil || lane.Start != nil || lane.Outcome != nil ||
			lane.Abnormal || lane.Pending != 0 || lane.Dropped != 0 || lane.Abandoned != 0 || lane.PolicyVersion != nil ||
			lane.CycleDeadlineMS != nil || lane.NextDueAt != nil || lane.RestartNotBefore != nil || lane.SnapshotDigest != nil ||
			lane.EvidenceDigest != nil {
			return errors.New("unobserved lane carries inferred facts")
		}
		return nil
	}
	if !member(string(*lane.Health), LaneHealths()) {
		return errors.New("invalid health")
	}
	if lane.PolicyVersion == nil || !validIdentity(*lane.PolicyVersion) || lane.CycleDeadlineMS == nil || *lane.CycleDeadlineMS <= 0 {
		return errors.New("observed lane lacks its server policy")
	}
	if lane.FirstFailure != nil && !validText(*lane.FirstFailure) || lane.Pending < 0 ||
		lane.NextDueAt != nil && lane.NextDueAt.IsZero() || lane.RestartNotBefore != nil && lane.RestartNotBefore.IsZero() ||
		lane.SnapshotDigest != nil && !validIdentity(*lane.SnapshotDigest) || lane.EvidenceDigest != nil && !validIdentity(*lane.EvidenceDigest) {
		return errors.New("noncanonical lane value")
	}
	// 마지막 관측 물결의 사슬: 물결 없음 ⇔ 트리거 없음, 투입이 들어간 물결만 시작이 있고, 연 사이클만 결과 · 비정상이 있다.
	if (lane.CycleGeneration == 0) != (lane.Trigger == nil) {
		return errors.New("cycle generation and trigger disagree")
	}
	if lane.Trigger == nil {
		if lane.Start != nil || lane.Outcome != nil || lane.Abnormal || lane.SnapshotDigest != nil || lane.EvidenceDigest != nil ||
			lane.Desired != StateOff || lane.Effective != StateOff {
			return errors.New("unobserved cycle carries facts")
		}
		return nil
	}
	if !member(string(*lane.Trigger), LaneTriggers()) {
		return errors.New("invalid trigger")
	}
	if (*lane.Trigger == LaneTriggerEnqueued) != (lane.Start != nil) {
		return errors.New("start without an enqueued trigger")
	}
	if lane.Start != nil && !member(string(*lane.Start), LaneStarts()) {
		return errors.New("invalid start")
	}
	admitted := lane.Start != nil && *lane.Start == LaneStartAdmitted
	if !admitted && (lane.Outcome != nil || lane.Abnormal) {
		return errors.New("outcome without an admitted cycle")
	}
	if lane.Outcome != nil && !member(string(*lane.Outcome), LaneOutcomes()) {
		return errors.New("invalid outcome")
	}
	return nil
}

func validateCoordinators(coordinators []CoordinatorProjection) error {
	markets := []Market{MarketKR, MarketUS}
	if len(coordinators) != len(markets) {
		return errors.New("coordinators must hold exactly KR and US")
	}
	for index, coordinator := range coordinators {
		if coordinator.Market != markets[index] {
			return errors.New("coordinators out of the fixed KR, US order")
		}
		if err := validateCoordinator(coordinator); err != nil {
			return errors.New("coordinator " + string(coordinator.Market) + ": " + err.Error())
		}
	}
	return nil
}

func validateCoordinator(coordinator CoordinatorProjection) error {
	if coordinator.GatedOutcomes == nil || coordinator.Selected == nil {
		return errors.New("lists must be arrays, not null")
	}
	if coordinator.Reason == nil {
		if coordinator.Ready || coordinator.RoutedCount != 0 || coordinator.ProposedCount != 0 || coordinator.RefusedCount != 0 ||
			coordinator.GatedCount != 0 || len(coordinator.GatedOutcomes) != 0 || coordinator.QueueDropCount != 0 ||
			coordinator.ArbitrationRefusal != nil || coordinator.ProposalSetDigest != nil || len(coordinator.Selected) != 0 {
			return errors.New("unobserved coordinator carries inferred facts")
		}
		return nil
	}
	if !member(*coordinator.Reason, CoordinatorReasons()) {
		return errors.New("invalid reason")
	}
	if coordinator.RoutedCount < 0 || coordinator.ProposedCount < 0 || coordinator.RefusedCount < 0 || coordinator.GatedCount < 0 {
		return errors.New("negative count")
	}
	for index, outcome := range coordinator.GatedOutcomes {
		if !member(outcome, LaneOutcomes()) || index > 0 && coordinator.GatedOutcomes[index-1] >= outcome {
			return errors.New("gated outcomes must be distinct sorted lane outcomes")
		}
	}
	if coordinator.ArbitrationRefusal != nil && !member(*coordinator.ArbitrationRefusal, ArbitrationRefusals()) {
		return errors.New("invalid arbitration refusal")
	}
	if coordinator.ProposalSetDigest != nil && !validDigest(*coordinator.ProposalSetDigest) {
		return errors.New("noncanonical proposal set digest")
	}
	for _, selected := range coordinator.Selected {
		if !validIdentity(selected.Symbol) || !validIdentity(selected.LaneID) || !validIdentity(selected.CampaignID) ||
			!validIdentity(selected.LineageIdentity) || !optionalIdentity(selected.ConfigDigest) ||
			!optionalIdentity(selected.ScoreVersion) || !optionalIdentity(selected.CalibrationDigest) {
			return errors.New("noncanonical selected scope")
		}
	}
	return nil
}

func optionalIdentity(value *string) bool { return value == nil || validIdentity(*value) }

func member(value string, values []string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func cloneLanes(lanes []LaneRuntimeProjection) []LaneRuntimeProjection {
	if lanes == nil {
		return nil
	}
	out := make([]LaneRuntimeProjection, len(lanes))
	for index, lane := range lanes {
		lane.Health = clonePointer(lane.Health)
		lane.FirstFailure = cloneString(lane.FirstFailure)
		lane.Trigger, lane.Start, lane.Outcome = clonePointer(lane.Trigger), clonePointer(lane.Start), clonePointer(lane.Outcome)
		lane.Refusal = cloneString(lane.Refusal)
		lane.PolicyVersion, lane.CycleDeadlineMS = cloneString(lane.PolicyVersion), clonePointer(lane.CycleDeadlineMS)
		lane.NextDueAt, lane.RestartNotBefore = cloneTime(lane.NextDueAt), cloneTime(lane.RestartNotBefore)
		lane.SnapshotDigest, lane.EvidenceDigest = cloneString(lane.SnapshotDigest), cloneString(lane.EvidenceDigest)
		out[index] = lane
	}
	return out
}

func cloneCoordinators(coordinators []CoordinatorProjection) []CoordinatorProjection {
	if coordinators == nil {
		return nil
	}
	out := make([]CoordinatorProjection, len(coordinators))
	for index, coordinator := range coordinators {
		coordinator.Reason, coordinator.ArbitrationRefusal = cloneString(coordinator.Reason), cloneString(coordinator.ArbitrationRefusal)
		coordinator.ProposalSetDigest = cloneString(coordinator.ProposalSetDigest)
		if coordinator.GatedOutcomes != nil {
			coordinator.GatedOutcomes = append([]string{}, coordinator.GatedOutcomes...)
		}
		if coordinator.Selected != nil {
			selected := make([]SelectedScopeProjection, len(coordinator.Selected))
			for at, scope := range coordinator.Selected {
				scope.ConfigDigest, scope.ScoreVersion = cloneString(scope.ConfigDigest), cloneString(scope.ScoreVersion)
				scope.CalibrationDigest = cloneString(scope.CalibrationDigest)
				selected[at] = scope
			}
			coordinator.Selected = selected
		}
		out[index] = coordinator
	}
	return out
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copyValue := value.UTC()
	return &copyValue
}
