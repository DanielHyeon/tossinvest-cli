// Package strategyshadow 는 a112 태스크 7.3.1 의 **shadow 매니페스트**다: 어느 OFF 레인을 읽기 전용 반사실(SHADOW)로 관측할지.
//
// 왜 별도 패키지인가(브리프 v3.3 §2 — Manager 판정 ②). shadow 는 승격의 반대편이다 — 아무것도 켜지 않고, 주문 경로에 아무것도 넘기지
// 않는다. 그 약속을 주석이 아니라 타입이 지키게 하려고 strategyrouter 밖에 둔다: 이 패키지는 strategyrouter 를 import 하지만
// `strategyrouter.FamilyActivation` 의 필드가 전부 비공개라 활성화를 **만들 수 없다**(저장소 전체의 비영 literal 은 활성화 적재기 한 자리 —
// guard_test.go 의 셈). 이 패키지의 생산 소스는 gofmt 정본 digest 로 동결된다(guard_test.go).
//
// 신뢰 앵커는 배포 digest 핀 하나다 [a112 결정 63 v3 — 61 의 (a)·(b) 를 적용]: 정규 바이트 등식은 바이트 동일성만 보장하고, 신뢰는 배포 핀을
// 쓸 권한에서 오며, 생성기(`tools/a112-family-shadow`)만 만든다는 것은 운영 정책이다(기술 증명 아님).
package strategyshadow

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

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const (
	productionFamilyShadowSchema = "strategy-four-family-shadow:v1"
	// domain 은 **종류 표식**이다 — 활성화 매니페스트 바이트가 이 자리에 놓이면 이 값(과 서술자 모양)이 거절한다(교차 디코드).
	productionFamilyShadowDomain       = "TossOS/strategy-four-family-shadow/sha256-pin/v1"
	productionFamilyShadowMaximumBytes = 64 << 10
	// 24시간은 활성화 매니페스트의 수명 상한과 같은 값이다 — 두 매니페스트의 상한이 다르면 사람이 둘 중 하나를 잊는다.
	productionFamilyShadowMaximumLife = 24 * time.Hour
)

var (
	// ErrProductionFamilyShadowUnavailable 은 「이 매니페스트로는 아무 레인도 shadow 할 수 없다」 다: 읽을 수 없음 · 핀 불일치 · 정규 아님 ·
	// 결속 불일치 · 서술자 이상. 공유 읽기 함수의 오류는 문장(`%v`)으로만 싣는다 — 이 sentinel 하나가 유일한 신원이다.
	ErrProductionFamilyShadowUnavailable = errors.New("strategyshadow: family shadow unavailable")
	// ErrProductionFamilyShadowRevoked 는 사람이 폐기 표시를 한 매니페스트다.
	ErrProductionFamilyShadowRevoked = errors.New("strategyshadow: family shadow revoked")
	// ErrProductionFamilyShadowExpired 는 수명이 지난 매니페스트다.
	ErrProductionFamilyShadowExpired = errors.New("strategyshadow: family shadow expired")
	// ErrProductionFamilyShadowUndeclared 는 배포가 이 시장에 shadow 를 **선언하지 않았다**(핀 env 가 빔)는 뜻이다 — 오늘 생산의 값.
	ErrProductionFamilyShadowUndeclared = errors.New("strategyshadow: family shadow undeclared")
)

// ShadowState 는 서술자 한 행의 shadow 여부다. desired/effective 필드는 **없다** — ON 을 값으로 표현할 자리가 0 이다.
type ShadowState string

const (
	ShadowOn  ShadowState = "ON"
	ShadowOff ShadowState = "OFF"
)

// ProductionFamilyShadowFileName 은 닫힌 대응이다(매니페스트나 호출자가 준 경로 조각을 받지 않는다).
func ProductionFamilyShadowFileName(market strategyrouter.Market) string {
	switch market {
	case strategyrouter.MarketKR:
		return "strategy-family-shadow-KR.json"
	case strategyrouter.MarketUS:
		return "strategy-family-shadow-US.json"
	default:
		return ""
	}
}

// Config 는 읽기 전용 경로 · 신뢰 핀 · 결속 값이다. 결속은 활성화와 같은 다섯(경로 매니페스트 · 보정 · 달력 · 빌드 · 위험 정책) —
// ProtectionReady 는 노출이 없는 shadow 에 결속하지 않는다(브리프 §1).
type Config struct {
	ConfigDir      string
	Market         strategyrouter.Market
	ManifestDigest string
	ObservedAt     time.Time

	RouteManifestDigest string
	CalibrationDigest   string
	CalendarVersion     string
	BuildDigest         string
	RiskPolicyDigest    string
}

type productionFamilyShadowDescriptor struct {
	Family      strategyrouter.Family  `json:"family"`
	Horizon     strategyrouter.Horizon `json:"horizon"`
	LaneID      string                 `json:"lane_id"`
	LaneVersion string                 `json:"lane_version"`
	Shadow      ShadowState            `json:"shadow"`
}

type productionFamilyShadowBody struct {
	SchemaVersion string                `json:"schema_version"`
	Domain        string                `json:"domain"`
	Generation    uint64                `json:"generation"`
	Market        strategyrouter.Market `json:"market"`

	RouteManifestDigest string `json:"route_manifest_digest"`
	CalibrationDigest   string `json:"calibration_digest"`
	CalendarVersion     string `json:"calendar_version"`
	RiskPolicyDigest    string `json:"risk_policy_digest"`
	BuildDigest         string `json:"build_digest"`

	Actor      string `json:"actor"`
	ApprovedAt string `json:"approved_at"`
	IssuedAt   string `json:"issued_at"`
	ExpiresAt  string `json:"expires_at"`
	Revoked    bool   `json:"revoked"`

	Descriptors []productionFamilyShadowDescriptor `json:"descriptors"`
}

// Document 는 사람이 채우는 값이다. 서술자는 받지 않는다 — shadow 할 **가족**만 받고 레인 ID · 버전 · 수평선은 검증기가 쓰는 바로 그 표
// (strategyrouter.SharedProductionRouteDescriptors)에서 유도한다. 그래서 도구가 낸 바이트와 검증기가 아는 표는 갈릴 수 없다.
type Document struct {
	Market     strategyrouter.Market
	Generation uint64

	RouteManifestDigest string
	CalibrationDigest   string
	CalendarVersion     string
	RiskPolicyDigest    string
	BuildDigest         string

	Actor      string
	ApprovedAt time.Time
	IssuedAt   time.Time
	ExpiresAt  time.Time
	Revoked    bool

	// Shadow 는 shadow ON 으로 둘 가족이다. 비면 넷 다 OFF — 그것도 정당한 매니페스트다.
	Shadow []strategyrouter.Family
}

// body 는 문서를 정규 몸통으로 옮긴다. 값의 의미는 검사하지 않는다(판정은 LoadProductionFamilyShadow 하나) — 모르는 가족 이름만 막는다.
func (document Document) body() (productionFamilyShadowBody, error) {
	table := strategyrouter.SharedProductionRouteDescriptors(document.Market)
	if len(table) == 0 {
		return productionFamilyShadowBody{}, fmt.Errorf("%w: market %q has no descriptor table", ErrProductionFamilyShadowUnavailable, document.Market)
	}
	shadowed := map[strategyrouter.Family]bool{}
	for _, family := range document.Shadow {
		if !family.Known() {
			return productionFamilyShadowBody{}, fmt.Errorf("%w: shadow: unknown family %q", ErrProductionFamilyShadowUnavailable, family)
		}
		shadowed[family] = true
	}
	descriptors := make([]productionFamilyShadowDescriptor, 0, len(table))
	for _, lane := range table {
		state := ShadowOff
		if shadowed[lane.Family] {
			state = ShadowOn
		}
		descriptors = append(descriptors, productionFamilyShadowDescriptor{Family: lane.Family, Horizon: lane.Horizon,
			LaneID: lane.LaneID, LaneVersion: lane.LaneVersion, Shadow: state})
	}
	// 표 wrapper 가 이미 가족 순이지만 정규 바이트의 정본성은 여기서 한 번 더 못 박는다(입력 · 표 순서와 무관한 바이트).
	sort.Slice(descriptors, func(i, j int) bool { return descriptors[i].Family < descriptors[j].Family })
	return productionFamilyShadowBody{SchemaVersion: productionFamilyShadowSchema, Domain: productionFamilyShadowDomain,
		Generation: document.Generation, Market: document.Market,
		RouteManifestDigest: document.RouteManifestDigest, CalibrationDigest: document.CalibrationDigest,
		CalendarVersion: document.CalendarVersion, RiskPolicyDigest: document.RiskPolicyDigest, BuildDigest: document.BuildDigest,
		Actor: document.Actor, ApprovedAt: document.ApprovedAt.UTC().Format(time.RFC3339Nano),
		IssuedAt: document.IssuedAt.UTC().Format(time.RFC3339Nano), ExpiresAt: document.ExpiresAt.UTC().Format(time.RFC3339Nano),
		Revoked: document.Revoked, Descriptors: descriptors}, nil
}

// EncodeProductionFamilyShadow 는 이 빌드가 받아들이는 정규 바이트를 낸다. 바이트는 아무것도 켜지 못한다 — 배포가
// `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` 을 그 바이트로 핀해야 shadow 가 선다. 부르는 자리는 도구 하나(guard_test.go).
func EncodeProductionFamilyShadow(input Document) ([]byte, error) {
	body, err := input.body()
	if err != nil {
		return nil, err
	}
	return json.Marshal(body)
}

type shadowLaneKey struct {
	family      strategyrouter.Family
	laneID      string
	laneVersion string
}

// FamilyShadow 는 검증된 shadow 매니페스트다. **영값이 안전한 값이다** — 필드가 비공개라 이 패키지 밖에서는 영값만 만들 수 있고, 영값은
// 아무 레인도 shadow 하지 않는다. 승격 · 활성화 · desired/effective 를 돌려주는 메서드는 없다.
type FamilyShadow struct {
	market     strategyrouter.Market
	generation uint64
	expiresAt  time.Time
	shadowed   map[shadowLaneKey]bool
}

// Verified 는 이 값이 검증된 매니페스트에서 왔는지다(전부 OFF 인 검증 매니페스트도 Verified).
func (shadow FamilyShadow) Verified() bool { return shadow.generation != 0 }

func (shadow FamilyShadow) Market() strategyrouter.Market { return shadow.market }
func (shadow FamilyShadow) Generation() uint64            { return shadow.generation }
func (shadow FamilyShadow) ExpiresAt() time.Time          { return shadow.expiresAt.UTC() }

// Shadowed 는 매니페스트가 이 레인을 shadow ON 으로 두었는지다. 시장을 인자로 받아 대조한다 — KR 매니페스트에 US 레인을 물으면 거짓.
func (shadow FamilyShadow) Shadowed(market strategyrouter.Market, family strategyrouter.Family, laneID, laneVersion string) bool {
	if !shadow.Verified() || market != shadow.market {
		return false
	}
	return shadow.shadowed[shadowLaneKey{family: family, laneID: laneID, laneVersion: laneVersion}]
}

// LoadProductionFamilyShadow 는 한 시장의 shadow 매니페스트를 읽는다. 아무것도 쓰지 않는다.
//
// 검사 순서는 활성화 적재기(strategyrouter.LoadProductionFamilyActivation)와 같다 — 미선언 맨 앞 · ctx · 설정 결속 · 파일 읽기 · 핀 ·
// 정규 · 폐기 · 결속 · 수명 · 서술자. 두 적재기의 순서는 양쪽 AST 로 못 박힌다(브리프 §3).
func LoadProductionFamilyShadow(ctx context.Context, config Config) (FamilyShadow, error) {
	// 미선언을 맨 먼저 본다 — 핀이 없는 시장(오늘 생산)의 답이 주기의 사정(취소된 ctx · 빈 결속 값)으로 바뀌지 않게.
	if strings.TrimSpace(config.ManifestDigest) == "" {
		return FamilyShadow{}, fmt.Errorf("%w: manifest_digest pin is empty", ErrProductionFamilyShadowUndeclared)
	}
	if ctx == nil {
		return FamilyShadow{}, fmt.Errorf("%w: context is nil", ErrProductionFamilyShadowUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return FamilyShadow{}, err
	}
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	owner, ownerOK := strategyrouter.SharedProductionRouteOwnerUID()
	name := ProductionFamilyShadowFileName(config.Market)
	// 설정 결속: 분기 하나, 어긋난 필드 전부를 모은다(각 항은 부작용 없는 순수 비교).
	if fields := failedFields(
		fieldCheck{"owner_uid", !ownerOK},
		fieldCheck{"market", name == ""},
		fieldCheck{"config_dir", !filepath.IsAbs(config.ConfigDir)},
		fieldCheck{"observed_at", config.ObservedAt.IsZero()},
		fieldCheck{"manifest_digest", !strategyrouter.SharedProductionRouteDigestValid(config.ManifestDigest)},
		fieldCheck{"calibration_digest", !strategyrouter.SharedProductionRouteIdentity(config.CalibrationDigest)},
		fieldCheck{"route_manifest_digest", !strategyrouter.SharedProductionRouteDigestValid(config.RouteManifestDigest)},
		fieldCheck{"risk_policy_digest", !strategyrouter.SharedProductionRouteDigestValid(config.RiskPolicyDigest)},
		fieldCheck{"calendar_version", !strategyrouter.SharedProductionRouteIdentity(config.CalendarVersion)},
		fieldCheck{"build_digest", !strategyrouter.SharedProductionRouteIdentity(config.BuildDigest)},
	); len(fields) != 0 {
		return FamilyShadow{}, fmt.Errorf("%w: config binding: %s", ErrProductionFamilyShadowUnavailable, strings.Join(fields, ", "))
	}
	data, err := strategyrouter.SharedReadProductionRouteFile(filepath.Join(config.ConfigDir, name), owner, 0o400,
		productionFamilyShadowMaximumBytes)
	// 읽기 결함은 공유 읽기 함수의 오류를 문장으로(`%v`) 싣는다 — 경로 sentinel 을 사슬에 두면 오류가 두 신원을 갖는다.
	if err != nil {
		return FamilyShadow{}, fmt.Errorf("%w: manifest file %s: %v", ErrProductionFamilyShadowUnavailable, name, err)
	}
	if strategyrouter.SharedProductionRouteDigest(data) != config.ManifestDigest {
		return FamilyShadow{}, fmt.Errorf("%w: manifest_digest: the pinned digest does not match the file bytes", ErrProductionFamilyShadowUnavailable)
	}
	manifest, err := decodeProductionFamilyShadow(data)
	if err != nil {
		return FamilyShadow{}, err
	}
	// 폐기는 핀 · 정규성 뒤에 본다 — 배포가 핀하지 않은 파일이 「폐기됐다」 는 답을 내지 않게.
	if manifest.Revoked {
		return FamilyShadow{}, fmt.Errorf("%w: revoked=true", ErrProductionFamilyShadowRevoked)
	}
	shadowed, expires, err := validateProductionFamilyShadow(manifest, config)
	if err != nil {
		return FamilyShadow{}, err
	}
	if err := ctx.Err(); err != nil {
		return FamilyShadow{}, err
	}
	return FamilyShadow{market: manifest.Market, generation: manifest.Generation, expiresAt: expires, shadowed: shadowed}, nil
}

// decodeProductionFamilyShadow 는 바이트가 정확히 이 구조체의 정규 직렬화와 같기를 요구한다(unknown field · 중복 키 · 뒤 데이터 ·
// 필드 순서 · 공백을 한 등식이 함께 거절). 활성화 바이트는 필드 모양이 달라 여기서 거절된다(교차 디코드).
func decodeProductionFamilyShadow(data []byte) (productionFamilyShadowBody, error) {
	if len(data) == 0 || len(data) > productionFamilyShadowMaximumBytes {
		return productionFamilyShadowBody{}, fmt.Errorf("%w: manifest size %d bytes is outside 1..%d", ErrProductionFamilyShadowUnavailable,
			len(data), productionFamilyShadowMaximumBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionFamilyShadowBody
	if err := decoder.Decode(&manifest); err != nil {
		return productionFamilyShadowBody{}, fmt.Errorf("%w: manifest json: %v", ErrProductionFamilyShadowUnavailable, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionFamilyShadowBody{}, fmt.Errorf("%w: trailing data after the manifest document", ErrProductionFamilyShadowUnavailable)
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionFamilyShadowBody{}, fmt.Errorf("%w: manifest bytes are not this build's canonical serialization", ErrProductionFamilyShadowUnavailable)
	}
	return manifest, nil
}

// validateProductionFamilyShadow 는 결속 · 수명 · 서술자 집합을 본다. 서술자는 정확히 그 시장의 네 레인이어야 한다(개수 · 중복 · 미지 ·
// 드리프트 · shadow 값 전부 거절).
func validateProductionFamilyShadow(body productionFamilyShadowBody, config Config) (map[shadowLaneKey]bool, time.Time, error) {
	if fields := failedFields(
		fieldCheck{"schema_version", body.SchemaVersion != productionFamilyShadowSchema},
		fieldCheck{"domain", body.Domain != productionFamilyShadowDomain},
		fieldCheck{"generation", body.Generation == 0},
		fieldCheck{"market", body.Market != config.Market},
		fieldCheck{"route_manifest_digest", body.RouteManifestDigest != config.RouteManifestDigest},
		fieldCheck{"calibration_digest", body.CalibrationDigest != config.CalibrationDigest},
		fieldCheck{"calendar_version", body.CalendarVersion != config.CalendarVersion},
		fieldCheck{"build_digest", body.BuildDigest != config.BuildDigest},
		fieldCheck{"risk_policy_digest", body.RiskPolicyDigest != config.RiskPolicyDigest},
		fieldCheck{"actor", !strategyrouter.SharedProductionRouteIdentity(body.Actor)},
	); len(fields) != 0 {
		return nil, time.Time{}, fmt.Errorf("%w: body binding: %s", ErrProductionFamilyShadowUnavailable, strings.Join(fields, ", "))
	}
	approved, okApproved := strategyrouter.SharedProductionRouteTime(body.ApprovedAt)
	issued, okIssued := strategyrouter.SharedProductionRouteTime(body.IssuedAt)
	expires, okExpires := strategyrouter.SharedProductionRouteTime(body.ExpiresAt)
	now := config.ObservedAt.UTC()
	if fields := failedFields(
		fieldCheck{"approved_at", !okApproved},
		fieldCheck{"issued_at", !okIssued},
		fieldCheck{"expires_at", !okExpires},
		fieldCheck{"issued_at before approved_at", issued.Before(approved)},
		fieldCheck{"issued_at after observed_at", issued.After(now)},
		fieldCheck{"issued_at not before expires_at", !issued.Before(expires)},
		fieldCheck{"lifetime over maximum", expires.Sub(issued) > productionFamilyShadowMaximumLife},
	); len(fields) != 0 {
		return nil, time.Time{}, fmt.Errorf("%w: lifetime: %s", ErrProductionFamilyShadowUnavailable, strings.Join(fields, ", "))
	}
	// 만료는 경계 1ns 에서 등호 쪽이 만료다(`!now.Before(expires)`) — 활성화의 familyActivationRemaining 과 같은 규칙.
	if !now.Before(expires) {
		return nil, time.Time{}, fmt.Errorf("%w: expires_at %s is not after %s", ErrProductionFamilyShadowExpired,
			expires.UTC().Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	}
	want := map[string]strategyrouter.SharedLaneDescriptor{}
	for _, lane := range strategyrouter.SharedProductionRouteDescriptors(body.Market) {
		want[lane.LaneID] = lane
	}
	shadowed := make(map[shadowLaneKey]bool, len(want))
	seen := make(map[shadowLaneKey]bool, len(want))
	// 서술자 거절은 위치와 필드 이름만 말한다 — lane_id 원문은 매니페스트 작성자의 임의 값이라 싣지 않는다.
	for index, descriptor := range body.Descriptors {
		table, known := want[descriptor.LaneID]
		if fields := failedFields(
			fieldCheck{"lane_id", !known},
			fieldCheck{"family", known && table.Family != descriptor.Family},
			fieldCheck{"horizon", known && table.Horizon != descriptor.Horizon},
			fieldCheck{"lane_version", known && table.LaneVersion != descriptor.LaneVersion},
			fieldCheck{"shadow", descriptor.Shadow != ShadowOn && descriptor.Shadow != ShadowOff},
		); len(fields) != 0 {
			return nil, time.Time{}, fmt.Errorf("%w: descriptors[%d]: %s", ErrProductionFamilyShadowUnavailable, index, strings.Join(fields, ", "))
		}
		key := shadowLaneKey{family: descriptor.Family, laneID: descriptor.LaneID, laneVersion: descriptor.LaneVersion}
		if seen[key] {
			return nil, time.Time{}, fmt.Errorf("%w: descriptors[%d]: duplicate lane_id", ErrProductionFamilyShadowUnavailable, index)
		}
		seen[key] = true
		shadowed[key] = descriptor.Shadow == ShadowOn
	}
	if len(seen) != len(want) {
		return nil, time.Time{}, fmt.Errorf("%w: descriptors: %d of %d lanes", ErrProductionFamilyShadowUnavailable, len(seen), len(want))
	}
	return shadowed, expires, nil
}

// fieldCheck · failedFields 는 복합 결속의 진단이다 — 첫 실패만 말하면 운영자가 하나를 고치고 다음 거절을 또 만난다.
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
