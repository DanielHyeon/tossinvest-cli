package strategyrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 이 파일은 태스크 8.7.1 이다: **4-가족 활성화 매니페스트** (신뢰 앵커는 8.8.3
// [결정 61] 이 서명에서 외부 digest pin 으로 바꿨다 — 아래 const 블록 참조).
//
// 왜 별도 매니페스트인가. 같은 패키지의 `strategy-lane-authority-<MARKET>.json`
// 도 이미 "시장마다 정확히 네 가족"을 서명으로 못 박는다(태스크 4.3,
// `validProductionRouteCandidates`). 그러나 그 매니페스트의 `Desired`/`Effective`
// 는 **scope(종목·세대) 마다** 있다. 여덟 `FamilyWorker` 의 열쇠에는 종목이 없다
// (골든 `worker_key_fields` 가 네 필드를 얼렸고 종목은 그중에 없다). 종목별 행을
// worker 하나의 상태로 접으려면 "모든 scope 에서 ON 이면 ON" 같은 규칙을 **지어내야**
// 하고, 지어낸 규칙은 계약이 아니다. 그리고 태스크 8.7 이 이름을 부른 build ·
// ProtectionReady digest 는 그 매니페스트에 없다.
//
// design.md:210 이 "새 runtime activation 은 exact 4-family-aware signed manifest
// 가 필요하다" 고 적고, tasks 8.7 이 "**separate** human-approved operating
// activation" 이라고 적은 것이 같은 말이다.
//
// **왜 이 패키지인가.** design.md:221 이 "exact four-per-market manifest" 를
// `internal/strategyrouter` 에 배정했고, 여덟 worker 가 사는
// `internal/strategyworker` 는 이 패키지를 이미 자기 허용 폐포에 들여오고 있다
// (`dependency_closure_test.go`). 활성화를 새 패키지에 두면 그 폐포를 넓혀야 하고,
// `internal/scheduler` 에 두면 desired-state **writer** 가 폐포에 들어와 스펙의
// "worker dependency closure 에 activation/toggle writer 가 없어야 한다" 를 깬다.
//
// **승격은 저장하는 상태가 아니라 주기마다 건네는 값이다.** 활성화에는 만료가
// 있고(아래 24시간 상한) 레인은 프로세스 수명이다. 승격을 레인 생성 시점에
// 구우면 묵은 ON 이 생긴다 — 묵은 OFF 는 안전한 방향이지만 묵은 ON 은 아니다.

// **신뢰 앵커는 서명이 아니라 외부 digest pin 이다 [a112 결정 61, 2026-09-04].**
//
// 앞 판본은 ed25519 서명을 요구했다. 사람이 그것을 빼기로 정했고 근거는
// `design.md` §8 의 같은 날짜 개정 블록에 있다. 요약하면 셋이다: 같은 저장소에
// 서명 없는 사람 승인 활성화 선례가 이미 있고(후보 임계값 활성화는
// `<PREFIX>_<MARKET>_ACTIVATION_SHA256` 하나로 지킨다), digest pin 이 서명과
// **독립적으로** 파일 치환을 이미 막고 있으며, 서명이 더 사는 승인자·배포자 분리는
// 개인키가 배포 호스트 밖에 있을 때만 참인데 이 배포는 그렇지 않다.
//
// **서명을 빼도 사람이 손으로 쓸 수는 없다.** 아래 정규 바이트 등식은 서명과 무관한
// 별개의 문이고 그대로 남는다. 매니페스트는 여전히 생성기가 만들며
// (`tools/a112-family-activation`), 달라진 것은 그 생성기가 비밀을 갖지 않는다는 것뿐이다.
const (
	productionFamilyActivationSchema = "strategy-four-family-activation:v1"
	// domain 은 서명 도메인 분리자가 아니라 **종류 표식**으로 남는다. 다른 종류의
	// 매니페스트 바이트가 이 자리에 놓였을 때 그것을 거절하는 것이 여전히 이 값이다.
	productionFamilyActivationDomain       = "TossOS/strategy-four-family-activation/sha256-pin/v1"
	productionFamilyActivationMaximumBytes = 64 << 10
	// 24시간은 고른 값이 아니라 같은 저장소의 활성화 매니페스트에서 읽은 값이다
	// (`internal/scheduler/production_activation.go` 의 productionActivationMaximumLife).
	// 두 활성화의 수명 상한이 다르면 사람이 둘 중 하나를 잊는다.
	productionFamilyActivationMaximumLife = 24 * time.Hour
)

var (
	// ErrProductionFamilyActivationUnavailable 는 "이 매니페스트로는 아무것도
	// 승격할 수 없다" 는 뜻이다. 읽을 수 없거나, 배포가 핀한 digest 와 다르거나, 결속이
	// 어긋나거나, 네 서술자가 정확히 서지 않은 경우 전부 여기로 온다.
	ErrProductionFamilyActivationUnavailable = errors.New("strategyrouter: four-family activation unavailable")
	// ErrProductionFamilyActivationRevoked 는 사람이 폐기 표시를 한 매니페스트다.
	ErrProductionFamilyActivationRevoked = errors.New("strategyrouter: four-family activation revoked")
	// ErrProductionFamilyActivationExpired 는 수명이 지난 매니페스트다.
	//
	// 셋을 따로 두는 이유: 운영자가 할 일이 다르다. 어긋남은 배포를 고치는
	// 일이고, 폐기는 사람의 결정이며, 만료는 다시 발급하는 일이다. 하나로
	// 뭉치면 "왜 안 켜지는가" 에 답할 수 없다.
	ErrProductionFamilyActivationExpired = errors.New("strategyrouter: four-family activation expired")
	// ErrProductionFamilyActivationUndeclared 는 배포가 이 시장에 4-가족 활성화를
	// **선언하지 않았다**는 뜻이다 — 핀 env 가 비어 있다 (태스크 8.7.2).
	//
	// 위 셋과 방향이 반대다. 셋은 "선언은 했는데 쓸 수 없다" 이고 엔진은 그 시장의
	// 네 가족을 OFF 로 되돌린다. 이것은 "4-가족 런타임이 배포되지 않았다" 이고 엔진은
	// 기존 시장 단위 경로를 그대로 쓴다 — 토글 OFF = upstream 동작이다. 둘을 한
	// 값으로 뭉치면 사람이 끈 가족이 활성화 만료·폐기 뒤 기존 경로로 되살아난다.
	ErrProductionFamilyActivationUndeclared = errors.New("strategyrouter: four-family activation undeclared")
)

// ProductionFamilyActivationFileName 은 닫힌 대응이다. 매니페스트나 호출자가 준
// 경로 조각을 절대 받지 않는다.
func ProductionFamilyActivationFileName(market Market) string {
	switch market {
	case MarketKR:
		return "strategy-family-activation-KR.json"
	case MarketUS:
		return "strategy-family-activation-US.json"
	default:
		return ""
	}
}

// FamilyActivationConfig 는 읽기 전용 경로와 신뢰 핀, 그리고 이 단계에서 이미
// 알고 있는 사실들이다. desired-state writer 도, 토글도 없다.
//
// ManifestDigest 가 **유일한 신뢰 앵커**다(결정 61). 배포가 env 로 핀하고 이 파일은
// 그 핀과 바이트가 같을 때만 읽힌다.
//
// 여기 있는 세 digest(보정·달력·빌드)는 **제안 수집 시점에 존재하는 값**이다.
// 위험 번들과 ProtectionReady digest 는 그 단계에 아직 없으므로 결속을 여기서
// 하지 않고, 검증된 값이 그 둘을 실어 내보내 뒤 단계가 결속한다(RiskBundleDigest,
// ProtectionReadyDigest). 존재하지 않는 사실을 결속하면 그 결속은 어떤 정상
// 입력으로도 참이 될 수 없고, 그것은 문 없는 fail-closed 다.
type FamilyActivationConfig struct {
	ConfigDir      string
	Market         Market
	ManifestDigest string
	ObservedAt     time.Time

	// RouteManifestDigest 는 이 시장의 서명된 경로 권한 전체를 가리키는 값이다.
	// 보정 digest 도 그 몸통 안에 있으므로 이것을 결속하면 네-가족 후보 행렬과
	// 보정 기준이 함께 못 박힌다. 보정 digest 를 **따로도** 결속하는 이유는
	// 담김에 의한 결속은 유도이고, 8.7 이 이름을 부른 값은 보정이기 때문이다.
	RouteManifestDigest string
	CalibrationDigest   string
	CalendarVersion     string
	BuildDigest         string
	// RiskPolicyDigest 는 이 시장의 서명된 위험 정책 매니페스트 digest 다
	// (`TOSSOS_RISK_BUCKET_<MARKET>_MANIFEST_SHA256`). 운영자가 이미 관리하는
	// 값이고 정책을 다시 서명할 때만 바뀐다.
	RiskPolicyDigest string
}

type productionFamilyActivationDescriptor struct {
	Family      Family       `json:"family"`
	Horizon     Horizon      `json:"horizon"`
	LaneID      string       `json:"lane_id"`
	LaneVersion string       `json:"lane_version"`
	Desired     DesiredState `json:"desired"`
	Effective   DesiredState `json:"effective"`
}

type productionFamilyActivationBody struct {
	SchemaVersion string `json:"schema_version"`
	Domain        string `json:"domain"`
	Generation    uint64 `json:"generation"`
	Market        Market `json:"market"`

	// 태스크 8.7 이 이름을 부른 다섯 값이 전부 매니페스트 몸통 안에 있다 (태스크 8.8.2 정정).
	//
	// **앞 판본은 위험과 ProtectionReady 를 per-cycle 스냅샷 봉인에 걸었고, 그
	// 결속은 어떤 정상 입력으로도 참이 될 수 없었다.** `RiskSnapshotAuthorityBundle`
	// 의 digest 는 `Symbol` 과 `AsOf`(파도의 벽시계)를 품으므로 파도마다·종목마다
	// 바뀐다. 사람이 미리 서명하는 상수가 같아질 수 없다. ProtectionReady 도
	// 살아 있는 readiness 스냅샷의 신원이라 같은 문제가 있다.
	//
	// 그래서 둘을 **서명 가능한 값**으로 바꿨다.
	//
	//   - 위험: 운영자가 이미 관리하는 **위험 정책 매니페스트** digest
	//     (`TOSSOS_RISK_BUCKET_<MARKET>_MANIFEST_SHA256`). 정책을 다시 서명할
	//     때만 바뀌므로 활성화 수명(24h) 동안 안정적이고, `docs/operations.md`
	//     가 이미 그 값을 운영 절차로 문서화하고 있다.
	//   - ProtectionReady: 등식이 아니라 **하한 세대**. 살아 있는 상태에 등식을
	//     걸면 문 없는 fail-closed 가 되고, 하한은 "내가 승인한 것보다 오래된
	//     보호 자세로는 켜지 마라"를 그대로 말한다. 세대는 단조 증가하므로
	//     하한은 안전 방향으로만 어긋난다.
	//
	// 다섯 중 넷은 아래 validateProductionFamilyActivation 이 이 단계에서
	// 결속한다. ProtectionReady 하한만 값으로 나가고, 그 사실이 존재하는
	// **주문 경로**가 결속한다 — 앞 판본은 그것을 화면만 바꾸는 자리에 두었다.
	RouteManifestDigest          string `json:"route_manifest_digest"`
	CalibrationDigest            string `json:"calibration_digest"`
	CalendarVersion              string `json:"calendar_version"`
	RiskPolicyDigest             string `json:"risk_policy_digest"`
	BuildDigest                  string `json:"build_digest"`
	ProtectionReadyMinGeneration uint64 `json:"protection_ready_min_generation"`

	Actor      string `json:"actor"`
	ApprovedAt string `json:"approved_at"`
	IssuedAt   string `json:"issued_at"`
	ExpiresAt  string `json:"expires_at"`
	Revoked    bool   `json:"revoked"`

	Descriptors []productionFamilyActivationDescriptor `json:"descriptors"`
}

// FamilyActivationDocument 는 사람이 채우는 값들이다 [a112 결정 61].
//
// **서술자 넷은 여기 없다.** 운영자가 레인 ID·버전·수평선을 손으로 적으면 표가
// 바뀔 때 조용히 옛 이름을 적게 되고, 그런 매니페스트는 "미지의 레인" 으로
// 거절되는데 그 거절은 배포 시각에야 보인다. 대신 켤 **가족**만 받고 나머지는
// 검증기가 쓰는 바로 그 표(`productionRouteDescriptors`)에서 유도한다 — 그래서
// 도구가 낸 바이트와 검증기가 아는 표는 갈릴 수 없다.
//
// desired 와 effective 를 따로 받지 않는 이유: 이 매니페스트가 곧 권위이므로
// effective 를 desired 보다 늦출 별도 기전이 없다. 둘을 나누면 만들 수 있는
// 상태가 늘고 그중 하나만 의미가 있다.
type FamilyActivationDocument struct {
	Market     Market
	Generation uint64

	RouteManifestDigest          string
	CalibrationDigest            string
	CalendarVersion              string
	RiskPolicyDigest             string
	BuildDigest                  string
	ProtectionReadyMinGeneration uint64

	Actor      string
	ApprovedAt time.Time
	IssuedAt   time.Time
	ExpiresAt  time.Time
	Revoked    bool

	// On 은 desired/effective 를 함께 ON 으로 둘 가족들이다. 비면 넷 다 OFF —
	// 그것도 정당한 매니페스트다(사람이 만들어 전부 껐다 ≠ 매니페스트가 없다).
	On []Family
}

// body 는 문서를 이 빌드의 정규 몸통으로 옮긴다.
//
// 여기서 값을 검사하지 않는다. 검사는 LoadProductionFamilyActivation 하나가 하고,
// 도구가 낸 바이트도 그 문을 지나야 한다 — 두 곳에서 검사하면 각자가 상대의 시험을
// 통과시킨다(5.3.3 이 원장에서 만난 것과 같은 모양). 여기서 막는 것은 **모르는
// 가족 이름** 하나뿐인데, 그것은 표에서 유도할 수 없어 조용히 빠지기 때문이다.
func (document FamilyActivationDocument) body() (productionFamilyActivationBody, error) {
	want := productionRouteDescriptors(document.Market)
	if len(want) == 0 {
		return productionFamilyActivationBody{}, fmt.Errorf("%w: market %q has no descriptor table", ErrProductionFamilyActivationUnavailable, document.Market)
	}
	on := map[Family]bool{}
	for _, family := range document.On {
		if !family.Known() {
			return productionFamilyActivationBody{}, fmt.Errorf("%w: on: unknown family %q", ErrProductionFamilyActivationUnavailable, family)
		}
		on[family] = true
	}
	descriptors := make([]productionFamilyActivationDescriptor, 0, len(want))
	for laneID, table := range want {
		state := StateOff
		if on[table.Family] {
			state = StateOn
		}
		descriptors = append(descriptors, productionFamilyActivationDescriptor{Family: table.Family,
			Horizon: table.Horizon, LaneID: laneID, LaneVersion: table.LaneVersion,
			Desired: state, Effective: state})
	}
	// 가족 이름으로 정렬한다 — 같은 패키지가 이미 쓰는 순서다(production.go 의
	// candidate 정렬). map 순회 순서에 기대면 같은 입력이 실행마다 다른 바이트를
	// 내고, 그러면 digest 핀이 무엇을 가리키는지 아무도 미리 말할 수 없다.
	sort.Slice(descriptors, func(i, j int) bool { return descriptors[i].Family < descriptors[j].Family })
	return productionFamilyActivationBody{
		SchemaVersion: productionFamilyActivationSchema, Domain: productionFamilyActivationDomain,
		Generation: document.Generation, Market: document.Market,
		RouteManifestDigest: document.RouteManifestDigest, CalibrationDigest: document.CalibrationDigest,
		CalendarVersion: document.CalendarVersion, RiskPolicyDigest: document.RiskPolicyDigest,
		BuildDigest:                  document.BuildDigest,
		ProtectionReadyMinGeneration: document.ProtectionReadyMinGeneration,
		Actor:                        document.Actor,
		ApprovedAt:                   document.ApprovedAt.UTC().Format(time.RFC3339Nano),
		IssuedAt:                     document.IssuedAt.UTC().Format(time.RFC3339Nano),
		ExpiresAt:                    document.ExpiresAt.UTC().Format(time.RFC3339Nano),
		Revoked:                      document.Revoked,
		Descriptors:                  descriptors,
	}, nil
}

// EncodeProductionFamilyActivation 은 이 빌드가 받아들이는 **정규 바이트**를 낸다.
//
// 존재 이유는 정본이 하나여야 하기 때문이다 [a112 결정 61]. 매니페스트를 만드는
// 도구가 구조체를 다시 선언하면 두 정의가 갈릴 수 있고, 갈리는 순간 도구가 낸
// 바이트를 검증기가 거절한다 — 그리고 그 어긋남을 잡는 시험은 어디에도 없다
// (같은 패키지의 로더 시험들은 전부 검증기 자신의 직렬화기로 바이트를 만든다).
// 그래서 도구는 이 함수를 부르고, 커밋된 골든 바이트가 드리프트 검출기가 된다
// (`production_family_activation_golden_test.go`).
//
// **이 함수는 아무것도 승격하지 못한다.** 바이트를 낼 뿐이고, 그 바이트가 무언가를
// 켜려면 배포가 `TOSSOS_STRATEGY_FAMILY_ACTIVATION_<MARKET>_MANIFEST_SHA256` 을
// 그 바이트로 핀해야 한다. 그 핀은 이 프로세스가 쓸 수 없다. 생산 코드가 이 함수를
// 부르지 않는다는 것은 `dependency_test.go` 의 참조 셈이 지킨다.
func EncodeProductionFamilyActivation(input FamilyActivationDocument) ([]byte, error) {
	body, err := input.body()
	if err != nil {
		return nil, err
	}
	return json.Marshal(body)
}

// familyLaneKey 는 활성화가 승격을 색인하는 열쇠다.
//
// `strategyworker.Key` 를 쓰지 않는 이유는 방향이다 — 그 패키지가 이 패키지를
// 들여오므로 반대 방향 import 는 순환이다. 시장은 값 자체가 들고 있으므로 여기
// 열쇠에는 없다.
type familyLaneKey struct {
	family      Family
	laneID      string
	laneVersion string
}

// FamilyActivation 은 정확한 매니페스트 검증 뒤에만 발급되는 불투명한 권한이다.
//
// **영값이 안전한 값이다.** 필드가 전부 비공개이므로 이 패키지 밖에서는 영값만
// 만들 수 있고, 영값은 아무것도 승격하지 않는다. 그래서 "켜진 worker 를 아무나
// 만들 수 있는가" 를 셈 시험으로 지킬 필요가 없다 — 승격을 얻는 유일한 길이
// LoadProductionFamilyActivation 을 통과하는 것이고, 그 함수는 배포가 핀한 digest 와
// 바이트가 같은 파일 없이는 아무것도 발급하지 않는다 [결정 61].
//
// EncodeProductionFamilyActivation 이 exported 라는 것은 이 성질을 깨지 않는다.
// 그것은 **바이트**를 낼 뿐이고, 바이트가 무언가를 켜려면 이 프로세스가 쓸 수 없는
// env 핀이 그 바이트를 가리켜야 한다.
type FamilyActivation struct {
	market     Market
	generation uint64
	actor      string
	expiresAt  time.Time
	// protectionReadyMinGeneration 은 사람이 승인한 보호 자세의 **하한**이다.
	// 등식이 아닌 이유는 ProtectionReady 가 살아 있는 상태이기 때문이다 —
	// 등식을 걸면 어떤 정상 입력으로도 참이 될 수 없다(태스크 8.8.2).
	protectionReadyMinGeneration uint64
	// state 는 매니페스트가 말한 (desired, effective) 를 서술자마다 그대로 담는다.
	// "승격된 것만" 담지 않는 이유: desired ON / effective OFF 는 운영자가 보는
	// 다른 상태이고, 승격만 담으면 그 구별이 사라진다.
	state map[familyLaneKey]productionFamilyActivationDescriptor
}

// Verified 는 이 값이 검증된 매니페스트에서 왔는지다.
//
// 이 시장의 4-가족 런타임이 판정 주체가 되었는가와 같은 뜻이다. 승격이 하나도
// 없는 검증된 매니페스트도 Verified 다 — 사람이 만들어 전부 OFF 로 둔 것과
// 매니페스트가 아예 없는 것은 다른 상태다.
// 시장을 여기서 다시 보지 않는 이유는 닿지 않기 때문이다. 이 값을 만드는 길은
// 아래 LoadProductionFamilyActivation 과 태그 아래 test seam 둘뿐이고, 둘 다
// 파일 이름 대응(`ProductionFamilyActivationFileName`)으로 시장을 이미 검증한다.
// 첫 판본은 `&& validMarket(activation.market)` 를 달고 있었고 반증이 그것을
// 지워도 아무 색도 안 바뀌었다 — 닿지 않는 방어는 지키는 것이 없다.
func (activation FamilyActivation) Verified() bool {
	return activation.generation != 0
}

func (activation FamilyActivation) Market() Market       { return activation.market }
func (activation FamilyActivation) Generation() uint64   { return activation.generation }
func (activation FamilyActivation) Actor() string        { return activation.actor }
func (activation FamilyActivation) ExpiresAt() time.Time { return activation.expiresAt.UTC() }

// ProtectionReadyMinGeneration 은 사람이 승인한 보호 자세의 하한이다.
//
// 이 값만 값으로 나간다. 나머지 넷은 제안 수집 시점에 존재하므로 그 단계가
// 결속했고, 이것은 **주문 경로**가 결속한다 — 보호 세대는 주문을 내려는 순간에만
// 존재하는 사실이다. 앞 판본은 이 결속을 worker 서술자를 만드는 자리에 두었고,
// 그 자리는 화면만 바꾸므로 주문을 하나도 막지 못했다(8.5 리뷰, 태스크 8.8.2).
func (activation FamilyActivation) ProtectionReadyMinGeneration() uint64 {
	return activation.protectionReadyMinGeneration
}

// LeaseCeiling 은 주문 lease 의 수명 상한을 이 활성화의 남은 수명으로 깎는다 (태스크 8.7.2).
//
// **줄이기만 한다.** 검증되지 않은 값(영값 = 4-가족 런타임이 판정하지 않는 시장)은
// 상한을 그대로 돌려주고, 검증된 값은 `min(상한, 남은 수명)` 을 돌려준다. 늘리는 길은
// 없다 — 이 함수가 늘릴 수 있으면 활성화가 오히려 주문을 오래 살리는 권한이 된다.
//
// 수명이 다 됐으면 만료 오류다. 그 판정은 적재가 쓰는 바로 그 함수
// (`familyActivationRemaining`)이므로 두 자리가 같은 순간에 같은 답을 낸다.
//
// 왜 필요한가: 활성화는 파도 시각에 검증되고 주문 lease 는 그 뒤에 발급된다. 상한을
// 깎지 않으면 만료 직전에 검증된 활성화가 만료 뒤 최대 30초까지 SUBMITTING 을 허락한다.
// 스케줄 활성화는 같은 이유로 이미 lease 를 자기 수명으로 깎는다(`strategyDispatchCycle.dispatch`).
func (activation FamilyActivation) LeaseCeiling(now time.Time, ceiling time.Duration) (time.Duration, error) {
	if !activation.Verified() {
		return ceiling, nil
	}
	remaining, err := familyActivationRemaining(activation.expiresAt, now)
	if err != nil {
		return 0, err
	}
	return min(ceiling, remaining), nil
}

// familyActivationRemaining 은 이 패키지의 **유일한** 만료 판정이다.
//
// 만료 시각에 정확히 닿은 순간부터 만료다(`!now.Before(expires)`). 적재
// (`validateProductionFamilyActivation`)와 lease 상한(`LeaseCeiling`)이 둘 다 이것을
// 부른다 — 판정을 두 곳에 적으면 경계 1ns 에서 둘이 갈릴 수 있고, 각자가 상대의
// 시험을 통과시킨다.
func familyActivationRemaining(expires, now time.Time) (time.Duration, error) {
	if !now.Before(expires) {
		return 0, fmt.Errorf("%w: expires_at %s is not after %s", ErrProductionFamilyActivationExpired,
			expires.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano))
	}
	return expires.Sub(now), nil
}

// Desired 와 Effective 는 매니페스트가 이 레인에 대해 말한 상태다.
//
// 시장을 인자로 받아 이 활성화의 시장과 대조한다. 받지 않고 값의 시장을 그냥
// 쓰면, KR 활성화를 US 레인에 물어본 호출자가 KR 의 답을 받는다.
func (activation FamilyActivation) Desired(market Market, family Family, laneID, laneVersion string) DesiredState {
	return activation.lookup(market, family, laneID, laneVersion).Desired
}

func (activation FamilyActivation) Effective(market Market, family Family, laneID, laneVersion string) DesiredState {
	return activation.lookup(market, family, laneID, laneVersion).Effective
}

// lookup 은 없는 것을 StateOff 로 돌려준다. 영값도 여기로 온다.
func (activation FamilyActivation) lookup(market Market, family Family,
	laneID, laneVersion string,
) productionFamilyActivationDescriptor {
	off := productionFamilyActivationDescriptor{Desired: StateOff, Effective: StateOff}
	if !activation.Verified() || market != activation.market || activation.state == nil {
		return off
	}
	descriptor, known := activation.state[familyLaneKey{family: family, laneID: laneID, laneVersion: laneVersion}]
	if !known {
		return off
	}
	return descriptor
}

// LoadProductionFamilyActivation 은 한 시장의 4-가족 활성화를 읽는다.
//
// 아무것도 쓰지 않고, 아무것도 만들지 않는다. 실패는 전부 "승격 없음" 이고
// 그것이 곧 오늘의 값이다.
//
// 승격을 얻는 길은 문 셋을 **함께** 지나는 것이다 [a112 결정 61]: 파일이 현재
// UID 소유의 `0400` 정규 파일이고, 바이트가 배포가 핀한 digest 와 같고, 그 바이트가
// 이 빌드의 정규 직렬화와 한 바이트도 다르지 않아야 한다. 앞 판본은 여기에 ed25519
// 서명 검증을 하나 더 두었다 — 사람이 그것을 뺐고, digest 핀이 파일 치환을 막는
// 역할은 그대로다.
func LoadProductionFamilyActivation(ctx context.Context, config FamilyActivationConfig) (FamilyActivation, error) {
	// 미선언을 **맨 먼저** 본다 (태스크 8.7.2). ctx 나 다른 결속 값보다 앞인 이유:
	// 핀이 없는 시장(오늘 생산)의 답이 주기의 사정 — 취소된 ctx, 비어 있는 보정 값 —
	// 에 따라 "선언했는데 쓸 수 없다" 로 바뀌면, 엔진이 그 시장의 네 가족을 OFF 로
	// 되돌려 기존 경로를 닫는다. 미배포 여부는 설정 사실 하나로만 정한다.
	if strings.TrimSpace(config.ManifestDigest) == "" {
		return FamilyActivation{}, fmt.Errorf("%w: manifest_digest pin is empty", ErrProductionFamilyActivationUndeclared)
	}
	if ctx == nil {
		return FamilyActivation{}, fmt.Errorf("%w: context is nil", ErrProductionFamilyActivationUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return FamilyActivation{}, err
	}
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	owner, ownerOK := productionRouteOwnerUID()
	name := ProductionFamilyActivationFileName(config.Market)
	// 설정 결속(a112 8.8.4 항목 1, 판정 Q-B1=(c)): 분기는 하나 그대로이고, 그 안에서 결속 필드를 **전부** 비교해 어긋난 이름을 모은다.
	// 각 항은 순수 비교 · 부작용 없는 검사라(단락 평가에 기대는 항 없음 — 2026-10-04 확인) 무조건 평가해도 판정이 같다.
	if fields := failedFields(
		fieldCheck{"owner_uid", !ownerOK},
		fieldCheck{"market", name == ""},
		fieldCheck{"config_dir", !filepath.IsAbs(config.ConfigDir)},
		fieldCheck{"observed_at", config.ObservedAt.IsZero()},
		fieldCheck{"manifest_digest", !productionRouteDigestValid(config.ManifestDigest)},
		fieldCheck{"calibration_digest", !productionRouteIdentity(config.CalibrationDigest)},
		fieldCheck{"route_manifest_digest", !productionRouteDigestValid(config.RouteManifestDigest)},
		fieldCheck{"risk_policy_digest", !productionRouteDigestValid(config.RiskPolicyDigest)},
		fieldCheck{"calendar_version", !productionRouteIdentity(config.CalendarVersion)},
		fieldCheck{"build_digest", !productionRouteIdentity(config.BuildDigest)},
	); len(fields) != 0 {
		return FamilyActivation{}, fmt.Errorf("%w: config binding: %s", ErrProductionFamilyActivationUnavailable, strings.Join(fields, ", "))
	}
	data, err := readProductionRouteFile(filepath.Join(config.ConfigDir, name), owner, 0o400,
		productionFamilyActivationMaximumBytes)
	// 읽기 결함과 핀 불일치는 다른 종류다(Manager 판정 2026-10-04): 결함을 불일치로 보이면 운영자가 I/O 장애를 매니페스트 오류로 읽는다.
	// 읽기 결함은 읽기 함수의 오류를 **문장으로** 싣고(`%v`), 불일치만 필드 이름으로 말한다. 사슬에는 활성화 sentinel 하나만 둔다 —
	// 둘째 `%w` 는 공유 읽기 함수의 sentinel(ErrProductionRouteUnavailable)까지 만족시켜 오류가 두 신원을 갖게 했다(a112 8.5 보이스 2 P2-1).
	if err != nil {
		return FamilyActivation{}, fmt.Errorf("%w: manifest file %s: %v", ErrProductionFamilyActivationUnavailable, name, err)
	}
	if productionRouteDigest(data) != config.ManifestDigest {
		return FamilyActivation{}, fmt.Errorf("%w: manifest_digest: the pinned digest does not match the file bytes", ErrProductionFamilyActivationUnavailable)
	}
	// 해석 거절은 해석기가 이미 이유를 붙여 sentinel 로 감쌌다 — 그대로 돌려준다.
	manifest, err := decodeProductionFamilyActivation(data)
	if err != nil {
		return FamilyActivation{}, err
	}
	// 폐기와 만료를 핀·정규성 검사 **뒤에** 본다. 앞에서 보면 배포가 핀하지 않은
	// 파일이 "폐기됐다" 는 답을 낼 수 있고, 그러면 운영자가 자기가 쓴 적 없는
	// 결정을 보게 된다. (앞 판본은 이 자리가 서명 뒤였다 — 같은 이유였다.)
	if manifest.Revoked {
		return FamilyActivation{}, fmt.Errorf("%w: revoked=true", ErrProductionFamilyActivationRevoked)
	}
	state, err := validateProductionFamilyActivation(manifest, config)
	if err != nil {
		return FamilyActivation{}, err
	}
	if err := ctx.Err(); err != nil {
		return FamilyActivation{}, err
	}
	expires, _ := productionRouteTime(manifest.ExpiresAt)
	return FamilyActivation{market: manifest.Market, generation: manifest.Generation, actor: manifest.Actor,
		expiresAt: expires, protectionReadyMinGeneration: manifest.ProtectionReadyMinGeneration, state: state}, nil
}

// decodeProductionFamilyActivation 은 바이트가 정확히 이 구조체의 정규 직렬화와
// 같기를 요구한다.
//
// 그 한 등식이 unknown field, 중복 키, 뒤에 붙은 JSON, 그리고 필드 순서·공백을
// 바꾼 사본까지 함께 거절한다. 검사를 하나씩 적으면 빠뜨린 것이 조용히 통과한다 —
// 같은 패키지의 decodeProductionRouteManifest 가 같은 이유로 같은 모양이다.
//
// **서명이 사라진 뒤 이 등식이 하는 일이 늘었다** [a112 결정 61]. 앞 판본에서는
// 서명이 바이트 전체를 함께 묶었으므로 여기서 놓친 이형이 있어도 서명 검증이
// 막았다. 이제는 이 등식과 digest 핀 둘뿐이다 — 그래서 사람이 만든 바이트가
// 실제로 이 문을 지나는지를 커밋된 골든이 따로 잰다
// (`production_family_activation_golden_test.go`).
func decodeProductionFamilyActivation(data []byte) (productionFamilyActivationBody, error) {
	if len(data) == 0 || len(data) > productionFamilyActivationMaximumBytes {
		return productionFamilyActivationBody{}, fmt.Errorf("%w: manifest size %d bytes is outside 1..%d", ErrProductionFamilyActivationUnavailable,
			len(data), productionFamilyActivationMaximumBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionFamilyActivationBody
	if err := decoder.Decode(&manifest); err != nil {
		return productionFamilyActivationBody{}, fmt.Errorf("%w: manifest json: %w", ErrProductionFamilyActivationUnavailable, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionFamilyActivationBody{}, fmt.Errorf("%w: trailing data after the manifest document", ErrProductionFamilyActivationUnavailable)
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionFamilyActivationBody{}, fmt.Errorf("%w: manifest bytes are not this build's canonical serialization", ErrProductionFamilyActivationUnavailable)
	}
	return manifest, nil
}

// validateProductionFamilyActivation 은 결속과 서술자 집합을 본다.
//
// 서술자는 **정확히 이 시장의 네 개**여야 한다. 개수, 중복, 미지의 레인,
// 가족·수평선·버전 드리프트가 전부 거절이다. design.md:208 이 "missing,
// duplicate, unknown, partial 3-of-4, legacy 3-lane ON manifest 는 새 권위로
// 자동 승격하지 않고 해당 새 runtime 을 OFF 로 유지한다" 고 적은 그대로다.
//
// 표는 여기서 다시 적지 않고 productionRouteDescriptors 를 그대로 쓴다. 옮겨
// 적으면 두 표가 갈릴 수 있고, 갈리는 순간 한쪽 매니페스트로 켜진 레인이 다른
// 쪽에서는 미지의 레인이 된다.
func validateProductionFamilyActivation(body productionFamilyActivationBody,
	config FamilyActivationConfig,
) (map[familyLaneKey]productionFamilyActivationDescriptor, error) {
	// 몸통 결속(복합 — 분기 하나, 어긋난 필드 전부를 모음; 항은 전부 값 필드의 순수 비교).
	if fields := failedFields(
		fieldCheck{"schema_version", body.SchemaVersion != productionFamilyActivationSchema},
		fieldCheck{"domain", body.Domain != productionFamilyActivationDomain},
		fieldCheck{"generation", body.Generation == 0},
		fieldCheck{"market", body.Market != config.Market || !validMarket(body.Market)},
		fieldCheck{"route_manifest_digest", body.RouteManifestDigest != config.RouteManifestDigest},
		fieldCheck{"calibration_digest", body.CalibrationDigest != config.CalibrationDigest},
		fieldCheck{"calendar_version", body.CalendarVersion != config.CalendarVersion},
		fieldCheck{"build_digest", body.BuildDigest != config.BuildDigest},
		fieldCheck{"risk_policy_digest", body.RiskPolicyDigest != config.RiskPolicyDigest},
		fieldCheck{"protection_ready_min_generation", body.ProtectionReadyMinGeneration == 0},
		fieldCheck{"actor", !productionRouteIdentity(body.Actor)},
	); len(fields) != 0 {
		return nil, fmt.Errorf("%w: body binding: %s", ErrProductionFamilyActivationUnavailable, strings.Join(fields, ", "))
	}
	approved, okApproved := productionRouteTime(body.ApprovedAt)
	issued, okIssued := productionRouteTime(body.IssuedAt)
	expires, okExpires := productionRouteTime(body.ExpiresAt)
	now := config.ObservedAt.UTC()
	// 수명(복합 — 같은 모양). 해석 실패한 시각은 영값이라 뒤 비교도 순수하다(공황 없음).
	if fields := failedFields(
		fieldCheck{"approved_at", !okApproved},
		fieldCheck{"issued_at", !okIssued},
		fieldCheck{"expires_at", !okExpires},
		fieldCheck{"issued_at before approved_at", issued.Before(approved)},
		fieldCheck{"issued_at after observed_at", issued.After(now)},
		fieldCheck{"issued_at not before expires_at", !issued.Before(expires)},
		fieldCheck{"lifetime over maximum", expires.Sub(issued) > productionFamilyActivationMaximumLife},
	); len(fields) != 0 {
		return nil, fmt.Errorf("%w: lifetime: %s", ErrProductionFamilyActivationUnavailable, strings.Join(fields, ", "))
	}
	// 만료 판정은 lease 상한과 같은 함수 하나다 (태스크 8.7.2).
	if _, err := familyActivationRemaining(expires, now); err != nil {
		return nil, err
	}
	// 표가 비어 있는 시장을 여기서 다시 막지 않는다. 그 문은 위에서 이미
	// 닫혀 있다(`ProductionFamilyActivationFileName` 이 "" 를 주면 곧바로 거절).
	// 첫 판본은 `len(want) == 0` 를 여기 달고 있었고 반증이 그것을 지워도 아무
	// 색도 안 바뀌었다. 닿지 않는 방어 대신 **닫아 두는 등식**을 시험한다:
	// 파일 이름이 있는 시장의 집합 == 서술자 표가 있는 시장의 집합
	// (`TestEveryMarketWithAnActivationFileNameHasADescriptorTable`).
	want := productionRouteDescriptors(body.Market)
	// 아래 두 판정만 남긴 것은 측정 결과다. 첫 판본은 셋이었다 —
	// `len(body.Descriptors) == len(want)`(입력 개수) · 중복 거절(유일성) ·
	// `len(state) == len(want)`(완전성). 셋 중 **아무 둘이면 충분**하므로 반증이
	// 셋 다 살아남았다: 각자가 상대의 시험을 통과시킨다(5.3.3 이 원장에서 만난
	// 것과 같은 모양). 개수 검사가 나머지 둘의 재진술이라 그것을 지웠다.
	// 남은 둘은 서로 다른 성질이고 각각 다른 입력에서만 짐을 진다:
	// 중복 거절은 `[c,r,w,b,b]`(다섯 중 넷이 다 있음)를, 완전성은 `[c,r,w]`를.
	state := make(map[familyLaneKey]productionFamilyActivationDescriptor, len(want))
	// 서술자 거절은 위치(`descriptors[i]`)와 필드 이름만 말한다 — lane_id 원문은 싣지 않는다. 핀이 맞는 파일이라도 그 문자열은 매니페스트
	// 작성자의 임의 값이고(개행 포함), 오류 문장은 로그 · 화면으로 간다(a112 8.5 응답 로트 ③ — codex r2 P2).
	for index, descriptor := range body.Descriptors {
		table, known := want[descriptor.LaneID]
		// 모르는 레인이면 표 값이 영값이라 표 대조 셋은 `known &&` 로 묶는다 — 판정은 앞 판(`!known || …`)과 같고 이름만 정확해진다.
		if fields := failedFields(
			fieldCheck{"lane_id", !known},
			fieldCheck{"family", known && table.Family != descriptor.Family},
			fieldCheck{"horizon", known && table.Horizon != descriptor.Horizon},
			fieldCheck{"lane_version", known && table.LaneVersion != descriptor.LaneVersion},
			fieldCheck{"desired", !validDesiredState(descriptor.Desired)},
			fieldCheck{"effective", !validDesiredState(descriptor.Effective)},
		); len(fields) != 0 {
			return nil, fmt.Errorf("%w: descriptors[%d]: %s", ErrProductionFamilyActivationUnavailable, index, strings.Join(fields, ", "))
		}
		// effective ON 은 desired ON 없이 설 수 없다. 반대는 정당하다 —
		// 사람이 켜기로 했지만 아직 서지 않은 상태다.
		if descriptor.Effective == StateOn && descriptor.Desired != StateOn {
			return nil, fmt.Errorf("%w: descriptors[%d]: effective ON without desired ON", ErrProductionFamilyActivationUnavailable, index)
		}
		key := familyLaneKey{family: descriptor.Family, laneID: descriptor.LaneID, laneVersion: descriptor.LaneVersion}
		if _, duplicate := state[key]; duplicate {
			return nil, fmt.Errorf("%w: descriptors[%d]: duplicate lane_id", ErrProductionFamilyActivationUnavailable, index)
		}
		state[key] = descriptor
	}
	if len(state) != len(want) {
		return nil, fmt.Errorf("%w: descriptors: %d of %d lanes", ErrProductionFamilyActivationUnavailable, len(state), len(want))
	}
	return state, nil
}

// fieldCheck · failedFields 는 복합 결속의 진단이다(a112 8.8.4 항목 1, Q-B1=(c)). 결속 판정은 호출자의 분기 하나(`len(fields) != 0`)가
// 하고, 이 함수는 그 분기가 무엇 때문에 섰는지를 모든 어긋난 필드로 말한다 — 첫 실패만 말하면 운영자가 하나를 고치고 다음 거절을 또 만난다.
type fieldCheck struct {
	name   string
	failed bool
}

func failedFields(checks ...fieldCheck) []string {
	var fields []string
	for _, check := range checks {
		if check.failed {
			fields = append(fields, check.name)
		}
	}
	return fields
}
