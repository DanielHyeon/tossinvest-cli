package strategyprojection

// a112 태스크 7.3 — 투영 envelope 의 additive 자식 `lanes[8]` · `coordinators[2]`(골든 operator_compatibility.additive_children).
//
// 여기서 재는 것은 계약의 모양이다: 개수 · 고정 순서 · 열쇠 · enum · null 짝 · JSON 이름 · 깊은 복사. 엔진이 실제 관측을 싣는지는
// internal/app/engine 의 a112_lane_coordinator_projection_test.go 가 잰다.

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// a112GoldenPath 는 동결 골든이다. strategyworker/golden_contract_test.go 와 같은 파일을 직접 읽는다(옮겨 적지 않음).
const a112GoldenPath = "../../openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens/four-family-runtime-v1.json"

var a112At = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators(t *testing.T) {
	for name, snapshot := range map[string]Snapshot{"dormant": DormantSnapshot(a112At), "unavailable": UnavailableSnapshot(a112At)} {
		if err := Validate(snapshot); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(snapshot.Lanes) != 8 || len(snapshot.Coordinators) != 2 {
			t.Fatalf("%s: lanes=%d coordinators=%d, want 8 and 2", name, len(snapshot.Lanes), len(snapshot.Coordinators))
		}
		for _, lane := range snapshot.Lanes {
			// 미관측 레인은 추론한 사실을 싣지 않는다 — 건강 상태조차 null(엔진이 레인을 세우기 전이거나 엔진에 닿지 못함).
			if lane.Desired != StateOff || lane.Effective != StateOff || lane.Runtime != LaneRuntimeUnobserved || lane.Health != nil ||
				lane.CycleGeneration != 0 || lane.Trigger != nil || lane.PolicyVersion != nil || lane.CycleDeadlineMS != nil {
				t.Fatalf("%s: lane %+v is not OFF/OFF/UNOBSERVED and unobserved", name, lane)
			}
		}
		for index, market := range []Market{MarketKR, MarketUS} {
			coordinator := snapshot.Coordinators[index]
			if coordinator.Market != market || coordinator.Reason != nil || coordinator.Ready || coordinator.Selected == nil ||
				len(coordinator.Selected) != 0 || coordinator.GatedOutcomes == nil {
				t.Fatalf("%s: coordinator %d = %+v, want unobserved %s with empty (not null) lists", name, index, coordinator, market)
			}
		}
	}
}

// 기본 레인 표는 동결 골든의 서술자 목록 그 자체여야 한다(순서 포함). 생산 레인 목록과의 등식은 strategyworker 쪽 시험이 잰다.
func TestTheDefaultLaneTableIsTheFrozenGoldenDescriptorList(t *testing.T) {
	raw, err := os.ReadFile(a112GoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Descriptors []struct {
			Market      string `json:"market"`
			Family      string `json:"family"`
			LaneID      string `json:"lane_id"`
			LaneVersion string `json:"lane_version"`
			Horizon     string `json:"horizon"`
			Desired     string `json:"desired"`
			Effective   string `json:"effective"`
			Runtime     string `json:"runtime"`
		} `json:"descriptors"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	lanes := DormantSnapshot(a112At).Lanes
	if len(golden.Descriptors) != len(lanes) {
		t.Fatalf("golden descriptors=%d default lanes=%d", len(golden.Descriptors), len(lanes))
	}
	for index, want := range golden.Descriptors {
		got := lanes[index]
		if string(got.Market) != want.Market || got.Family != want.Family || got.LaneID != want.LaneID || got.LaneVersion != want.LaneVersion ||
			got.Horizon != want.Horizon || string(got.Desired) != want.Desired || string(got.Effective) != want.Effective ||
			string(got.Runtime) != want.Runtime {
			t.Fatalf("default lane %d = %+v, golden %+v", index, got, want)
		}
	}
}

// 중재 거절 어휘는 골든 refusal_enums.arbitration 여섯 그대로다(지어내지 않음).
func TestTheArbitrationRefusalVocabularyIsTheGoldenSix(t *testing.T) {
	raw, err := os.ReadFile(a112GoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		RefusalEnums struct {
			Arbitration []string `json:"arbitration"`
		} `json:"refusal_enums"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	if got := ArbitrationRefusals(); len(golden.RefusalEnums.Arbitration) != 6 || !reflect.DeepEqual(got, golden.RefusalEnums.Arbitration) {
		t.Fatalf("projection arbitration codes=%v, golden=%v", got, golden.RefusalEnums.Arbitration)
	}
}

// a112ObservedSnapshot 은 엔진이 낼 법한 관측된 자식들을 단 유효한 스냅숏이다(거절 표의 기준점).
func a112ObservedSnapshot() Snapshot {
	snapshot := DormantSnapshot(a112At)
	health, trigger, start, outcome := LaneHealthy, LaneTriggerEnqueued, LaneStartAdmitted, LaneOutcomeDormant
	version, deadline, due := "a112-runtime-policy-v1", int64(30000), a112At.Add(time.Minute)
	for index := range snapshot.Lanes {
		lane := &snapshot.Lanes[index]
		lane.Health, lane.PolicyVersion, lane.CycleDeadlineMS = &health, &version, &deadline
		lane.CycleGeneration, lane.Trigger, lane.Start, lane.Outcome, lane.NextDueAt = 3, &trigger, &start, &outcome, &due
	}
	reason, digest := "READY", strings.Repeat("a", 64)
	kr := &snapshot.Coordinators[0]
	kr.Reason, kr.Ready, kr.RoutedCount, kr.ProposedCount, kr.ProposalSetDigest = &reason, true, 2, 2, &digest
	config, score, calibration := "sha256:config-1", "arbitration-score:v1", "sha256:calibration-v1"
	kr.Selected = []SelectedScopeProjection{
		{Symbol: "005930", LaneID: "kr_short_flow_continuation_v1", CampaignID: "campaign-1", LineageIdentity: "strategy-lineage:v1:sha256:" + strings.Repeat("1", 64),
			ConfigDigest: &config, ScoreVersion: &score, CalibrationDigest: &calibration},
		{Symbol: "000660", LaneID: "kr_short_flow_continuation_v1", CampaignID: "campaign-2", LineageIdentity: "strategy-lineage:v1:sha256:" + strings.Repeat("2", 64)},
	}
	// 레인 하나는 REFUSED 와 그 골든 코드를 싣는다(짝 규칙의 통과 쪽).
	refused, code := LaneOutcomeRefused, "ARBITRATION_SEAL_MISMATCH"
	snapshot.Lanes[6].Outcome, snapshot.Lanes[6].Refusal = &refused, &code
	return snapshot
}

func TestValidateRefusesMalformedLaneAndCoordinatorChildren(t *testing.T) {
	if err := Validate(a112ObservedSnapshot()); err != nil {
		t.Fatalf("baseline observed snapshot: %v", err)
	}
	str := func(value string) *string { return &value }
	for name, mutate := range map[string]func(*Snapshot){
		"seven lanes":                    func(s *Snapshot) { s.Lanes = s.Lanes[:7] },
		"lanes swapped":                  func(s *Snapshot) { s.Lanes[0], s.Lanes[1] = s.Lanes[1], s.Lanes[0] },
		"lane key rewritten":             func(s *Snapshot) { s.Lanes[2].LaneID = "kr_short_absorption_reversal_v2" },
		"lane horizon rewritten":         func(s *Snapshot) { s.Lanes[4].Horizon = "SHORT" },
		"lane desired invalid":           func(s *Snapshot) { s.Lanes[0].Desired = "MAYBE" },
		"lane runtime invented":          func(s *Snapshot) { s.Lanes[0].Runtime = "OBSERVED" },
		"lane health invalid":            func(s *Snapshot) { value := LaneHealth("SICK"); s.Lanes[0].Health = &value },
		"observed lane without policy":   func(s *Snapshot) { s.Lanes[0].PolicyVersion = nil },
		"observed lane without deadline": func(s *Snapshot) { s.Lanes[0].CycleDeadlineMS = nil },
		"non-positive deadline":          func(s *Snapshot) { value := int64(0); s.Lanes[0].CycleDeadlineMS = &value },
		"generation zero with trigger":   func(s *Snapshot) { s.Lanes[0].CycleGeneration = 0 },
		"trigger without generation": func(s *Snapshot) {
			s.Lanes[0].Trigger, s.Lanes[0].Start, s.Lanes[0].Outcome = nil, nil, nil
		},
		"disabled trigger with start": func(s *Snapshot) {
			value := LaneTriggerDisabled
			s.Lanes[0].Trigger, s.Lanes[0].Outcome = &value, nil
		},
		"enqueued without start": func(s *Snapshot) { s.Lanes[0].Start, s.Lanes[0].Outcome = nil, nil },
		"outcome without admission": func(s *Snapshot) {
			value := LaneStartTooSoon
			s.Lanes[0].Start = &value
		},
		"abnormal without admission": func(s *Snapshot) {
			value := LaneStartBackoff
			s.Lanes[0].Start, s.Lanes[0].Outcome, s.Lanes[0].Abnormal = &value, nil, true
		},
		"refusal on a dormant outcome": func(s *Snapshot) { s.Lanes[0].Refusal = str("ARBITRATION_SEAL_MISMATCH") },
		"refused without its code":     func(s *Snapshot) { s.Lanes[6].Refusal = nil },
		"refusal code invented":        func(s *Snapshot) { s.Lanes[6].Refusal = str("NOT_MINE") },
		"refusal on an unobserved lane": func(s *Snapshot) {
			s.Lanes[0] = DormantSnapshot(a112At).Lanes[0]
			s.Lanes[0].Refusal = str("ARBITRATION_TIE")
		},
		"selected config with space": func(s *Snapshot) { s.Coordinators[0].Selected[0].ConfigDigest = str("sha256: x") },
		"selected calibration empty": func(s *Snapshot) { s.Coordinators[0].Selected[0].CalibrationDigest = str("") },
		"negative pending":           func(s *Snapshot) { s.Lanes[0].Pending = -1 },
		"zero next due":              func(s *Snapshot) { s.Lanes[0].NextDueAt = &time.Time{} },
		"digest with space":          func(s *Snapshot) { s.Lanes[0].EvidenceDigest = str("sha256: abc") },
		"first failure control":      func(s *Snapshot) { s.Lanes[0].FirstFailure = str("boom\x00") },
		"first failure empty":        func(s *Snapshot) { s.Lanes[0].FirstFailure = str("  ") },
		"unobserved lane has fact": func(s *Snapshot) {
			s.Lanes[0].Health, s.Lanes[0].Trigger, s.Lanes[0].Start, s.Lanes[0].Outcome = nil, nil, nil, nil
		},
		"one coordinator":             func(s *Snapshot) { s.Coordinators = s.Coordinators[:1] },
		"coordinators swapped":        func(s *Snapshot) { s.Coordinators[0], s.Coordinators[1] = s.Coordinators[1], s.Coordinators[0] },
		"coordinator reason invented": func(s *Snapshot) { s.Coordinators[0].Reason = str("FINE") },
		"unobserved coordinator has count": func(s *Snapshot) {
			s.Coordinators[1].RoutedCount = 1
		},
		"negative count":               func(s *Snapshot) { s.Coordinators[0].RefusedCount = -1 },
		"arbitration code invented":    func(s *Snapshot) { s.Coordinators[0].ArbitrationRefusal = str("ARBITRATION_UNHAPPY") },
		"gated outcome invented":       func(s *Snapshot) { s.Coordinators[0].GatedOutcomes = []string{"SLEEPY"} },
		"gated outcomes unsorted":      func(s *Snapshot) { s.Coordinators[0].GatedOutcomes = []string{"REFUSED", "DORMANT"} },
		"gated outcomes duplicated":    func(s *Snapshot) { s.Coordinators[0].GatedOutcomes = []string{"DORMANT", "DORMANT"} },
		"gated outcomes null":          func(s *Snapshot) { s.Coordinators[0].GatedOutcomes = nil },
		"selected null":                func(s *Snapshot) { s.Coordinators[0].Selected = nil },
		"selected empty symbol":        func(s *Snapshot) { s.Coordinators[0].Selected[1].Symbol = "" },
		"selected identity with space": func(s *Snapshot) { s.Coordinators[0].Selected[0].LineageIdentity = "a b" },
		"set digest noncanonical":      func(s *Snapshot) { s.Coordinators[0].ProposalSetDigest = str("sha256:" + strings.Repeat("a", 64)) },
	} {
		snapshot := Clone(a112ObservedSnapshot())
		mutate(&snapshot)
		if err := Validate(snapshot); err == nil {
			t.Errorf("%s: Validate accepted a malformed child", name)
		}
	}
}

func TestCloneDeepCopiesLaneAndCoordinatorChildren(t *testing.T) {
	original := a112ObservedSnapshot()
	failure, digest := "lane fault", "sha256:"+strings.Repeat("e", 64)
	original.Lanes[1].FirstFailure, original.Lanes[1].EvidenceDigest, original.Lanes[1].SnapshotDigest = &failure, &digest, &digest
	restart := a112At.Add(time.Hour)
	original.Lanes[1].RestartNotBefore = &restart
	refusal := "ARBITRATION_TIE"
	original.Coordinators[0].ArbitrationRefusal = &refusal
	original.Coordinators[0].GatedOutcomes = []string{"DORMANT"}
	// 기준은 Clone 이 아니라 직렬화 바이트다 — 기준을 Clone 으로 만들면 얕은 Clone 이 원본 · 기준 둘 다와 포인터를 나눠
	// 같이 움직여 통과한다(변이 P12 첫 판 SURVIVED 의 원인).
	want, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	copied := Clone(original)
	*copied.Lanes[0].Health, *copied.Lanes[0].PolicyVersion, *copied.Lanes[0].CycleDeadlineMS = LaneLatched, "x", 1
	*copied.Lanes[0].Trigger, *copied.Lanes[0].Start, *copied.Lanes[0].Outcome = LaneTriggerFull, LaneStartBackoff, LaneOutcomeRefused
	*copied.Lanes[0].NextDueAt = a112At
	*copied.Lanes[1].FirstFailure, *copied.Lanes[1].EvidenceDigest, *copied.Lanes[1].SnapshotDigest = "x", "x", "x"
	*copied.Lanes[1].RestartNotBefore = a112At
	*copied.Lanes[6].Refusal = "x"
	*copied.Coordinators[0].Selected[0].ConfigDigest, *copied.Coordinators[0].Selected[0].ScoreVersion = "x", "x"
	*copied.Coordinators[0].Selected[0].CalibrationDigest = "x"
	*copied.Coordinators[0].Reason, *copied.Coordinators[0].ProposalSetDigest, *copied.Coordinators[0].ArbitrationRefusal = "x", "x", "x"
	copied.Coordinators[0].Selected[0].Symbol, copied.Coordinators[0].GatedOutcomes[0] = "x", "x"
	copied.Lanes[2].Pending = 9
	got, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("mutating the clone changed the original's lane or coordinator children — Clone must copy deeply")
	}
}

// JSON 이름이 계약이다(OpenAPI 와 운영자가 읽는 이름). httpapi 의 OpenAPI 시험이 이 이름 집합을 스키마와 대조한다.
func TestLaneAndCoordinatorJSONNamesAreTheContract(t *testing.T) {
	raw, err := json.Marshal(a112ObservedSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Lanes        []map[string]json.RawMessage `json:"lanes"`
		Coordinators []map[string]json.RawMessage `json:"coordinators"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	var selected []map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Coordinators[0]["selected"], &selected); err != nil {
		t.Fatal(err)
	}
	keys := func(value map[string]json.RawMessage) string {
		out := make([]string, 0, len(value))
		for key := range value {
			out = append(out, key)
		}
		sort.Strings(out)
		return strings.Join(out, ",")
	}
	if got := keys(envelope.Lanes[0]); got != strings.Join(LaneJSONFields(), ",") {
		t.Fatalf("lane JSON names=%s, want %v", got, LaneJSONFields())
	}
	if got := keys(envelope.Coordinators[0]); got != strings.Join(CoordinatorJSONFields(), ",") {
		t.Fatalf("coordinator JSON names=%s, want %v", got, CoordinatorJSONFields())
	}
	if got := keys(selected[0]); got != strings.Join(SelectedScopeJSONFields(), ",") {
		t.Fatalf("selected JSON names=%s, want %v", got, SelectedScopeJSONFields())
	}
}
