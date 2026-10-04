//go:build tossos_testseams

package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 0.5 리뷰 유지#1: shadowConfig 는 loadFamilyActivation 의 결속 절반을 손으로 옮겨 적은 사본임. 둘을 묶는 시험이 없어
// 세 필드를 한꺼번에 바꾼 변이(M1 — 달력을 desired 쪽에서 · US 가 KR 위험 env · 보정을 빈 값)가 엔진 스위트 전체를 통과했음.
//
// 활성화 적재기는 바꾸지 않고(편집 0) **생산 활성화 적재기 자체를 판정기로** 씀: shadowConfig 가 낸 결속 값으로 활성화 매니페스트를
// 써서 핀한 뒤 loadFamilyActivation 이 그것을 검증된 활성화로 받아들이는지 봄. 활성화 적재기는 결속 다섯(경로 digest · 보정 · 달력 ·
// 위험 정책 · 빌드)을 자기 원천과 대조하므로, 한 필드라도 다른 원천에서 읽으면 거절됨.
//
// 픽스처는 갈림이 보이게 원천을 일부러 벌림: 달력 스냅숏 버전 ≠ desired 의 달력 버전, KR · US 위험 핀 · shadow 핀이 서로 다름.
func TestTheShadowBindingIsTheActivationBindingFieldByField(t *testing.T) {
	now := time.Date(2026, 10, 5, 1, 2, 3, 0, time.UTC)
	routes := arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	schedules := routeReadySchedulePair(now)
	pins := map[string]string{
		strategyRiskKRManifestDigestEnv:         "sha256:" + strings.Repeat("1", 64),
		strategyRiskUSManifestDigestEnv:         "sha256:" + strings.Repeat("2", 64),
		strategyFamilyShadowKRManifestDigestEnv: "sha256:" + strings.Repeat("3", 64),
		strategyFamilyShadowUSManifestDigestEnv: "sha256:" + strings.Repeat("4", 64),
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			loader := testStrategyProposalLoader(t)
			env := map[string]string{}
			for name, value := range pins {
				env[name] = value
			}
			base := loader.getenv
			loader.getenv = func(name string) string {
				if value, ok := env[name]; ok {
					return value
				}
				return base(name)
			}
			schedule := schedules.forMarket(market)
			// 달력 원천 둘을 벌림 — 같으면 「desired 쪽에서 읽는다」 변이가 보이지 않음.
			schedule.calendar.Version = "calendar-snapshot-" + string(market)
			if schedule.calendar.Version == schedule.desired.CalendarVersion {
				t.Fatal("arrangement: the two calendar sources must differ")
			}
			route := routes.forMarket(market)
			// 활성화 적재기는 sha256 형식의 경로 digest 만 받음 — 시장마다 다른 값(시장 바꿔치기도 거절되게).
			route.snapshot.ManifestDigest = "sha256:" + strings.Repeat(map[StrategyMarket]string{StrategyMarketKR: "5", StrategyMarketUS: "6"}[market], 64)
			config := loader.shadowConfig(market, schedule, route)

			// 활성화 적재기가 대조하지 않는 셋은 값으로 직접 대조함(설정 디렉터리 · 시장 · shadow 핀 env).
			shadowEnv, riskEnv := strategyFamilyShadowKRManifestDigestEnv, strategyRiskKRManifestDigestEnv
			activationEnv := strategyFamilyActivationKRManifestDigestEnv
			if market == StrategyMarketUS {
				shadowEnv, riskEnv = strategyFamilyShadowUSManifestDigestEnv, strategyRiskUSManifestDigestEnv
				activationEnv = strategyFamilyActivationUSManifestDigestEnv
			}
			if config.ConfigDir != loader.configDir || config.Market != strategyRouterMarket(market) ||
				config.ManifestDigest != pins[shadowEnv] || config.RiskPolicyDigest != pins[riskEnv] {
				t.Fatalf("shadow binding dir=%q market=%s pin=%q risk=%q — want this market's own sources",
					config.ConfigDir, config.Market, config.ManifestDigest, config.RiskPolicyDigest)
			}
			// 빈 값은 「같은 원천」 이 아니라 「원천 없음」 — 둘 다 비어도 같다고 답하지 않게 먼저 막음.
			for name, value := range map[string]string{"route": config.RouteManifestDigest, "calibration": config.CalibrationDigest,
				"calendar": config.CalendarVersion, "build": config.BuildDigest} {
				if strings.TrimSpace(value) == "" {
					t.Fatalf("shadow binding %s is empty", name)
				}
			}

			data, err := strategyrouter.EncodeProductionFamilyActivation(strategyrouter.FamilyActivationDocument{
				Market: config.Market, Generation: 1,
				RouteManifestDigest: config.RouteManifestDigest, CalibrationDigest: config.CalibrationDigest,
				CalendarVersion: config.CalendarVersion, RiskPolicyDigest: config.RiskPolicyDigest,
				BuildDigest: config.BuildDigest, ProtectionReadyMinGeneration: 1,
				Actor: "test-human", ApprovedAt: now.Add(-time.Hour), IssuedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour),
				On: []strategyrouter.Family{strategyrouter.FamilyContinuation},
			})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(config.ConfigDir, strategyrouter.ProductionFamilyActivationFileName(config.Market))
			if err := os.WriteFile(path, data, 0o400); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, 0o400); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			env[activationEnv] = "sha256:" + hex.EncodeToString(sum[:])

			activation, err := loader.loadFamilyActivation(context.Background(), market, schedule, route, now)
			if err != nil || !activation.Verified() {
				t.Fatalf("the activation loader refused a manifest bound with the shadow binding (err=%v) — "+
					"shadowConfig reads a binding field from a different source than loadFamilyActivation", err)
			}
		})
	}
}
