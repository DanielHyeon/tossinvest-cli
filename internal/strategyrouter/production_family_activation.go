package strategyrouter

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// 이 파일은 태스크 8.7.1 이다: **서명된 4-가족 활성화 매니페스트.**
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

const (
	productionFamilyActivationSchema       = "strategy-four-family-activation:v1"
	productionFamilyActivationDomain       = "TossOS/strategy-four-family-activation/ed25519/v1"
	productionFamilyActivationAlgorithm    = "Ed25519"
	productionFamilyActivationMaximumBytes = 64 << 10
	// 24시간은 고른 값이 아니라 같은 저장소의 활성화 매니페스트에서 읽은 값이다
	// (`internal/scheduler/production_activation.go` 의 productionActivationMaximumLife).
	// 두 활성화의 수명 상한이 다르면 사람이 둘 중 하나를 잊는다.
	productionFamilyActivationMaximumLife = 24 * time.Hour
)

var (
	// ErrProductionFamilyActivationUnavailable 는 "이 매니페스트로는 아무것도
	// 승격할 수 없다" 는 뜻이다. 읽을 수 없거나, 서명이 안 맞거나, 결속이
	// 어긋나거나, 네 서술자가 정확히 서지 않은 경우 전부 여기로 온다.
	ErrProductionFamilyActivationUnavailable = errors.New("strategyrouter: four-family activation unavailable")
	// ErrProductionFamilyActivationRevoked 는 사람이 폐기 표시를 한 매니페스트다.
	ErrProductionFamilyActivationRevoked = errors.New("strategyrouter: four-family activation revoked")
	// ErrProductionFamilyActivationExpired 는 수명이 지난 매니페스트다.
	//
	// 셋을 따로 두는 이유: 운영자가 할 일이 다르다. 어긋남은 배포를 고치는
	// 일이고, 폐기는 사람의 결정이며, 만료는 다시 서명하는 일이다. 하나로
	// 뭉치면 "왜 안 켜지는가" 에 답할 수 없다.
	ErrProductionFamilyActivationExpired = errors.New("strategyrouter: four-family activation expired")
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
// 알고 있는 사실들이다. 서명 키도, desired-state writer 도, 토글도 없다.
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
	TrustedKeyID   string
	TrustedKey     ed25519.PublicKey
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
	SchemaVersion      string `json:"schema_version"`
	Domain             string `json:"domain"`
	SignatureAlgorithm string `json:"signature_algorithm"`
	KeyID              string `json:"key_id"`
	Generation         uint64 `json:"generation"`
	Market             Market `json:"market"`

	// 태스크 8.7 이 이름을 부른 다섯 값이 전부 서명 안에 있다 (태스크 8.8.2 정정).
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

type productionFamilyActivationManifest struct {
	productionFamilyActivationBody
	Signature string `json:"signature"`
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
// ed25519 서명 검증을 통과하는 것이다.
type FamilyActivation struct {
	market     Market
	generation uint64
	actor      string
	expiresAt  time.Time
	// protectionReadyMinGeneration 은 사람이 승인한 보호 자세의 **하한**이다.
	// 등식이 아닌 이유는 ProtectionReady 가 살아 있는 상태이기 때문이다 —
	// 등식을 걸면 어떤 정상 입력으로도 참이 될 수 없다(태스크 8.8.2).
	protectionReadyMinGeneration uint64
	// state 는 서명이 말한 (desired, effective) 를 서술자마다 그대로 담는다.
	// "승격된 것만" 담지 않는 이유: desired ON / effective OFF 는 운영자가 보는
	// 다른 상태이고, 승격만 담으면 그 구별이 사라진다.
	state map[familyLaneKey]productionFamilyActivationDescriptor
}

// Verified 는 이 값이 검증된 매니페스트에서 왔는지다.
//
// 이 시장의 4-가족 런타임이 판정 주체가 되었는가와 같은 뜻이다. 승격이 하나도
// 없는 검증된 매니페스트도 Verified 다 — 사람이 서명해서 전부 OFF 로 둔 것과
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

// Desired 와 Effective 는 서명이 이 레인에 대해 말한 상태다.
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

// LoadProductionFamilyActivation 은 한 시장의 서명된 4-가족 활성화를 읽는다.
//
// 아무것도 쓰지 않고, 아무것도 만들지 않는다. 실패는 전부 "승격 없음" 이고
// 그것이 곧 오늘의 값이다.
func LoadProductionFamilyActivation(ctx context.Context, config FamilyActivationConfig) (FamilyActivation, error) {
	if ctx == nil {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	if err := ctx.Err(); err != nil {
		return FamilyActivation{}, err
	}
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	owner, ownerOK := productionRouteOwnerUID()
	name := ProductionFamilyActivationFileName(config.Market)
	if !ownerOK || name == "" || !filepath.IsAbs(config.ConfigDir) || config.ObservedAt.IsZero() ||
		!productionRouteDigestValid(config.ManifestDigest) || !productionRouteIdentity(config.TrustedKeyID) ||
		len(config.TrustedKey) != ed25519.PublicKeySize || !productionRouteIdentity(config.CalibrationDigest) ||
		!productionRouteDigestValid(config.RouteManifestDigest) ||
		!productionRouteDigestValid(config.RiskPolicyDigest) ||
		!productionRouteIdentity(config.CalendarVersion) || !productionRouteIdentity(config.BuildDigest) {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	data, err := readProductionRouteFile(filepath.Join(config.ConfigDir, name), owner, 0o400,
		productionFamilyActivationMaximumBytes)
	if err != nil || productionRouteDigest(data) != config.ManifestDigest {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	manifest, err := decodeProductionFamilyActivation(data)
	if err != nil {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	canonical, err := json.Marshal(manifest.productionFamilyActivationBody)
	if err != nil {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	if err != nil || base64.StdEncoding.EncodeToString(signature) != manifest.Signature ||
		len(signature) != ed25519.SignatureSize ||
		!ed25519.Verify(config.TrustedKey, canonical, signature) {
		return FamilyActivation{}, ErrProductionFamilyActivationUnavailable
	}
	// 폐기와 만료를 서명 **뒤에** 본다. 앞에서 보면 서명 없는 파일이 "폐기됐다"
	// 는 답을 낼 수 있고, 그러면 운영자가 자기가 쓴 적 없는 결정을 보게 된다.
	if manifest.Revoked {
		return FamilyActivation{}, ErrProductionFamilyActivationRevoked
	}
	state, err := validateProductionFamilyActivation(manifest.productionFamilyActivationBody, config)
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
func decodeProductionFamilyActivation(data []byte) (productionFamilyActivationManifest, error) {
	if len(data) == 0 || len(data) > productionFamilyActivationMaximumBytes {
		return productionFamilyActivationManifest{}, ErrProductionFamilyActivationUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionFamilyActivationManifest
	if err := decoder.Decode(&manifest); err != nil {
		return productionFamilyActivationManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionFamilyActivationManifest{}, ErrProductionFamilyActivationUnavailable
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionFamilyActivationManifest{}, ErrProductionFamilyActivationUnavailable
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
	if body.SchemaVersion != productionFamilyActivationSchema || body.Domain != productionFamilyActivationDomain ||
		body.SignatureAlgorithm != productionFamilyActivationAlgorithm || body.KeyID != config.TrustedKeyID ||
		body.Generation == 0 || body.Market != config.Market || !validMarket(body.Market) ||
		body.RouteManifestDigest != config.RouteManifestDigest ||
		body.CalibrationDigest != config.CalibrationDigest || body.CalendarVersion != config.CalendarVersion ||
		body.BuildDigest != config.BuildDigest ||
		body.RiskPolicyDigest != config.RiskPolicyDigest ||
		body.ProtectionReadyMinGeneration == 0 ||
		!productionRouteIdentity(body.Actor) {
		return nil, ErrProductionFamilyActivationUnavailable
	}
	approved, okApproved := productionRouteTime(body.ApprovedAt)
	issued, okIssued := productionRouteTime(body.IssuedAt)
	expires, okExpires := productionRouteTime(body.ExpiresAt)
	now := config.ObservedAt.UTC()
	if !okApproved || !okIssued || !okExpires || issued.Before(approved) || issued.After(now) ||
		!issued.Before(expires) || expires.Sub(issued) > productionFamilyActivationMaximumLife {
		return nil, ErrProductionFamilyActivationUnavailable
	}
	if !now.Before(expires) {
		return nil, ErrProductionFamilyActivationExpired
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
	for _, descriptor := range body.Descriptors {
		table, known := want[descriptor.LaneID]
		if !known || table.Family != descriptor.Family || table.Horizon != descriptor.Horizon ||
			table.LaneVersion != descriptor.LaneVersion ||
			!validDesiredState(descriptor.Desired) || !validDesiredState(descriptor.Effective) {
			return nil, ErrProductionFamilyActivationUnavailable
		}
		// effective ON 은 desired ON 없이 설 수 없다. 반대는 정당하다 —
		// 사람이 켜기로 했지만 아직 서지 않은 상태다.
		if descriptor.Effective == StateOn && descriptor.Desired != StateOn {
			return nil, ErrProductionFamilyActivationUnavailable
		}
		key := familyLaneKey{family: descriptor.Family, laneID: descriptor.LaneID, laneVersion: descriptor.LaneVersion}
		if _, duplicate := state[key]; duplicate {
			return nil, ErrProductionFamilyActivationUnavailable
		}
		state[key] = descriptor
	}
	if len(state) != len(want) {
		return nil, ErrProductionFamilyActivationUnavailable
	}
	return state, nil
}
