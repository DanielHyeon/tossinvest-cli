package strategyshadow

// a112 7.3.1 SHADOW 매니페스트 적재기(브리프 v3.3 §1 · §3, 결정 63 v3).
//
// 신뢰 앵커는 배포 digest 핀 하나다(결정 61 과 같은 규칙). 적재기는 활성화 적재기와 **같은 검사 순서**를 지킨다 — 미선언 맨 앞 · ctx ·
// 설정 결속 · 파일 읽기 · 핀 · 정규 · 폐기 · 결속 · 수명 · 서술자. 오류는 이 패키지의 sentinel 하나로만 서고(활성화 · 경로 sentinel 과 배타),
// 두 매니페스트는 서로의 적재기에서 거절된다(교차 디코드 양방향).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

var shadowAt = time.Date(2026, 10, 5, 1, 0, 0, 0, time.UTC)

const (
	shadowRouteDigest = "sha256:" + "0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
	shadowRiskDigest  = "sha256:" + "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"
	shadowCalibration = "sha256:calibration-shadow-kr-v1"
	shadowCalendar    = "kr-regular-2026.10"
	shadowBuild       = "tossos-shadow-build-1"
)

func shadowDocument() Document {
	return Document{Market: strategyrouter.MarketKR, Generation: 2,
		RouteManifestDigest: shadowRouteDigest, CalibrationDigest: shadowCalibration, CalendarVersion: shadowCalendar,
		RiskPolicyDigest: shadowRiskDigest, BuildDigest: shadowBuild, Actor: "operator",
		ApprovedAt: shadowAt.Add(-2 * time.Hour), IssuedAt: shadowAt.Add(-time.Hour), ExpiresAt: shadowAt.Add(time.Hour),
		Shadow: []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal}}
}

// install 은 바이트를 로더가 요구하는 자리 · 모드(0400)로 놓고 그 바이트의 핀을 돌려준다.
func install(t *testing.T, data []byte, name string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	return dir, "sha256:" + hex.EncodeToString(digest[:])
}

func shadowConfig(dir, pin string) Config {
	return Config{ConfigDir: dir, Market: strategyrouter.MarketKR, ManifestDigest: pin, ObservedAt: shadowAt,
		RouteManifestDigest: shadowRouteDigest, CalibrationDigest: shadowCalibration, CalendarVersion: shadowCalendar,
		BuildDigest: shadowBuild, RiskPolicyDigest: shadowRiskDigest}
}

func encoded(t *testing.T, document Document) []byte {
	t.Helper()
	data, err := EncodeProductionFamilyShadow(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAValidShadowManifestShadowsExactlyItsFamiliesAndPromotesNothing(t *testing.T) {
	dir, pin := install(t, encoded(t, shadowDocument()), ProductionFamilyShadowFileName(strategyrouter.MarketKR))
	shadow, err := LoadProductionFamilyShadow(context.Background(), shadowConfig(dir, pin))
	if err != nil || !shadow.Verified() {
		t.Fatalf("a valid shadow manifest was refused: verified=%v err=%v", shadow.Verified(), err)
	}
	if shadow.Market() != strategyrouter.MarketKR || shadow.Generation() != 2 || !shadow.ExpiresAt().Equal(shadowAt.Add(time.Hour)) {
		t.Fatalf("shadow facts market=%s generation=%d expires=%s", shadow.Market(), shadow.Generation(), shadow.ExpiresAt())
	}
	shadowed := 0
	for _, lane := range strategyrouter.SharedProductionRouteDescriptors(strategyrouter.MarketKR) {
		want := lane.Family == strategyrouter.FamilyContinuation || lane.Family == strategyrouter.FamilyReversal
		if got := shadow.Shadowed(strategyrouter.MarketKR, lane.Family, lane.LaneID, lane.LaneVersion); got != want {
			t.Errorf("lane %s shadowed=%v, want %v", lane.LaneID, got, want)
		}
		if shadow.Shadowed(strategyrouter.MarketUS, lane.Family, lane.LaneID, lane.LaneVersion) {
			t.Errorf("a KR manifest answered for the US market (lane %s)", lane.LaneID)
		}
		if want {
			shadowed++
		}
	}
	if shadowed != 2 {
		t.Fatalf("shadowed=%d, want 2", shadowed)
	}
	// 영값은 아무 레인도 shadow 하지 않는다.
	var zero FamilyShadow
	if zero.Verified() || zero.Shadowed(strategyrouter.MarketKR, strategyrouter.FamilyContinuation, "kr_short_flow_continuation_v1", "v1") {
		t.Fatal("the zero FamilyShadow must shadow nothing")
	}
}

// 미선언(핀 없음)은 맨 먼저 본다 — ctx 가 nil 이어도, 결속 값이 비어도 답은 미선언이다.
func TestAnEmptyPinIsUndeclaredBeforeAnythingElse(t *testing.T) {
	_, err := LoadProductionFamilyShadow(nil, Config{Market: strategyrouter.MarketKR, ManifestDigest: "  "}) //nolint:staticcheck
	if !errors.Is(err, ErrProductionFamilyShadowUndeclared) {
		t.Fatalf("err=%v, want undeclared", err)
	}
	for _, other := range []error{ErrProductionFamilyShadowUnavailable, ErrProductionFamilyShadowRevoked, ErrProductionFamilyShadowExpired} {
		if errors.Is(err, other) {
			t.Fatalf("undeclared also satisfies %v", other)
		}
	}
}

// 「유효한 shadow manifest 없이」(spec :91) — 핀은 있으나 파일 없음 · 불일치 · 만료 · 폐기 · 결속 불일치 · 서술자 이상. 각 모양이 이 패키지
// sentinel 로 거절되고, 결속 거절은 어긋난 필드 이름을 말한다.
func TestEveryUnusableShadowManifestIsRefusedWithItsOwnSentinel(t *testing.T) {
	name := ProductionFamilyShadowFileName(strategyrouter.MarketKR)
	good := encoded(t, shadowDocument())
	type load func(t *testing.T) (FamilyShadow, error)
	run := func(data []byte, mutate func(*Config)) load {
		return func(t *testing.T) (FamilyShadow, error) {
			dir, pin := install(t, data, name)
			config := shadowConfig(dir, pin)
			if mutate != nil {
				mutate(&config)
			}
			return LoadProductionFamilyShadow(context.Background(), config)
		}
	}
	expired := shadowDocument()
	expired.IssuedAt, expired.ExpiresAt = shadowAt.Add(-2*time.Hour), shadowAt
	revoked := shadowDocument()
	revoked.Revoked = true
	tooLong := shadowDocument()
	tooLong.ExpiresAt = tooLong.IssuedAt.Add(25 * time.Hour)
	body, err := shadowDocument().body()
	if err != nil {
		t.Fatal(err)
	}
	threeOfFour := body
	threeOfFour.Descriptors = append([]productionFamilyShadowDescriptor(nil), body.Descriptors[:3]...)
	duplicate := body
	duplicate.Descriptors = append(append([]productionFamilyShadowDescriptor(nil), body.Descriptors[:3]...), body.Descriptors[0])
	unknownState := body
	unknownState.Descriptors = append([]productionFamilyShadowDescriptor(nil), body.Descriptors...)
	unknownState.Descriptors[0].Shadow = "MAYBE"
	marshal := func(value productionFamilyShadowBody) []byte {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	for _, tc := range []struct {
		name  string
		load  load
		want  error
		field string
	}{
		{"file missing", func(t *testing.T) (FamilyShadow, error) {
			config := shadowConfig(t.TempDir(), "sha256:"+strings.Repeat("0", 64))
			return LoadProductionFamilyShadow(context.Background(), config)
		}, ErrProductionFamilyShadowUnavailable, "manifest file"},
		{"pin mismatch", run(good, func(config *Config) { config.ManifestDigest = "sha256:" + strings.Repeat("1", 64) }),
			ErrProductionFamilyShadowUnavailable, "manifest_digest"},
		{"not canonical", run(append(append([]byte(nil), good...), ' '), nil), ErrProductionFamilyShadowUnavailable, "canonical"},
		{"expired", run(encoded(t, expired), nil), ErrProductionFamilyShadowExpired, "expires_at"},
		{"revoked", run(encoded(t, revoked), nil), ErrProductionFamilyShadowRevoked, "revoked"},
		{"lifetime over 24h", run(encoded(t, tooLong), nil), ErrProductionFamilyShadowUnavailable, "lifetime over maximum"},
		{"route manifest binding", run(good, func(config *Config) { config.RouteManifestDigest = shadowRiskDigest }),
			ErrProductionFamilyShadowUnavailable, "route_manifest_digest"},
		{"calibration binding", run(good, func(config *Config) { config.CalibrationDigest = "sha256:other" }),
			ErrProductionFamilyShadowUnavailable, "calibration_digest"},
		{"calendar binding", run(good, func(config *Config) { config.CalendarVersion = "kr-other" }),
			ErrProductionFamilyShadowUnavailable, "calendar_version"},
		{"build binding", run(good, func(config *Config) { config.BuildDigest = "other-build" }),
			ErrProductionFamilyShadowUnavailable, "build_digest"},
		{"risk policy binding", run(good, func(config *Config) { config.RiskPolicyDigest = shadowRouteDigest }),
			ErrProductionFamilyShadowUnavailable, "risk_policy_digest"},
		{"the market reads its own file", run(good, func(config *Config) { config.Market = strategyrouter.MarketUS }),
			ErrProductionFamilyShadowUnavailable, "manifest file"},
		{"three of four", run(marshal(threeOfFour), nil), ErrProductionFamilyShadowUnavailable, "3 of 4"},
		{"duplicate lane", run(marshal(duplicate), nil), ErrProductionFamilyShadowUnavailable, "duplicate"},
		{"unknown shadow state", run(marshal(unknownState), nil), ErrProductionFamilyShadowUnavailable, "descriptors[0]: shadow"},
	} {
		shadow, err := tc.load(t)
		if shadow.Verified() || !errors.Is(err, tc.want) || err == nil || !strings.Contains(err.Error(), tc.field) {
			t.Errorf("%s: verified=%v err=%v, want %v naming %q", tc.name, shadow.Verified(), err, tc.want, tc.field)
		}
	}
}

// sentinel 배타: shadow 오류는 활성화 · 경로 sentinel 을 만족하지 않고(공유 읽기 오류는 %v 로 접힘), 그 반대도 그렇다.
func TestShadowAndActivationSentinelsAreExclusive(t *testing.T) {
	_, readErr := LoadProductionFamilyShadow(context.Background(), shadowConfig(t.TempDir(), "sha256:"+strings.Repeat("0", 64)))
	if readErr == nil {
		t.Fatal("arrangement: a missing file must fail")
	}
	activation := []error{strategyrouter.ErrProductionFamilyActivationUnavailable, strategyrouter.ErrProductionFamilyActivationRevoked,
		strategyrouter.ErrProductionFamilyActivationExpired, strategyrouter.ErrProductionFamilyActivationUndeclared,
		strategyrouter.ErrProductionRouteUnavailable}
	shadow := []error{ErrProductionFamilyShadowUnavailable, ErrProductionFamilyShadowRevoked, ErrProductionFamilyShadowExpired,
		ErrProductionFamilyShadowUndeclared}
	for _, other := range activation {
		if errors.Is(readErr, other) {
			t.Errorf("a shadow read failure also satisfies %v — fold the shared reader's error with %%v", other)
		}
		for _, mine := range shadow {
			if errors.Is(mine, other) || errors.Is(other, mine) {
				t.Errorf("sentinels %v and %v are not exclusive", mine, other)
			}
		}
	}
}

// 교차 디코드 양방향 거절: 활성화 바이트는 shadow 적재기에서, shadow 바이트는 활성화 적재기에서 — 핀이 그 바이트를 가리켜도.
func TestEachManifestIsRefusedByTheOtherLoader(t *testing.T) {
	activationBytes, err := strategyrouter.EncodeProductionFamilyActivation(strategyrouter.FamilyActivationDocument{
		Market: strategyrouter.MarketKR, Generation: 2, RouteManifestDigest: shadowRouteDigest, CalibrationDigest: shadowCalibration,
		CalendarVersion: shadowCalendar, RiskPolicyDigest: shadowRiskDigest, BuildDigest: shadowBuild, ProtectionReadyMinGeneration: 1,
		Actor: "operator", ApprovedAt: shadowAt.Add(-2 * time.Hour), IssuedAt: shadowAt.Add(-time.Hour), ExpiresAt: shadowAt.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	dir, pin := install(t, activationBytes, ProductionFamilyShadowFileName(strategyrouter.MarketKR))
	if shadow, err := LoadProductionFamilyShadow(context.Background(), shadowConfig(dir, pin)); shadow.Verified() ||
		!errors.Is(err, ErrProductionFamilyShadowUnavailable) {
		t.Fatalf("activation bytes under the shadow pin: verified=%v err=%v", shadow.Verified(), err)
	}
	dir, pin = install(t, encoded(t, shadowDocument()), strategyrouter.ProductionFamilyActivationFileName(strategyrouter.MarketKR))
	activation, err := strategyrouter.LoadProductionFamilyActivation(context.Background(), strategyrouter.FamilyActivationConfig{
		ConfigDir: dir, Market: strategyrouter.MarketKR, ManifestDigest: pin, ObservedAt: shadowAt,
		RouteManifestDigest: shadowRouteDigest, CalibrationDigest: shadowCalibration, CalendarVersion: shadowCalendar,
		BuildDigest: shadowBuild, RiskPolicyDigest: shadowRiskDigest})
	if activation.Verified() || !errors.Is(err, strategyrouter.ErrProductionFamilyActivationUnavailable) {
		t.Fatalf("shadow bytes under the activation pin: verified=%v err=%v", activation.Verified(), err)
	}
}
