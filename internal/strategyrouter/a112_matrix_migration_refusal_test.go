package strategyrouter

// a112 4.5(-c) 감사 보강(Manager 판정 2026-10-04): 매트릭스 이행 거절을 **적재기 층** 에서 KR · US 둘 다 잰다.
//
// 감사(audit-3.8-4.5.md 4.5-c) 시점의 핀은 둘로 갈려 있었다 — 세 가족 후보 집합의 거절은 `validProductionRouteCandidates` 단위
// 시험뿐이었고(적재기가 그 함수를 부르는지는 따로 안 잼), 활성화 매니페스트의 세 가족 거절은 KR 만 돌았다. 여기서는:
//   ① 현재 스키마 · 올바른 서명의 경로 매니페스트가 레거시 세 가족(breakout 없음) 후보만 실으면 `LoadProductionRouteAuthority` 가 거절한다.
//   ② 활성화 매니페스트가 세 가족 서술자만 실으면 두 시장 모두 Unavailable 이고 검증된 활성화가 없다.
// 양성 대조: 같은 fixture 의 네 가족 매니페스트는 두 시장 모두 적재된다(거절이 fixture 결함이 아님).

import (
	"context"
	"errors"
	"testing"
)

func TestAFreshlySignedThreeFamilyRouteManifestIsNotRouteAuthorityInEitherMarket(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			if _, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market]); err != nil {
				t.Fatalf("arrangement: the four-family manifest is refused: %v", err)
			}
			body := fixture.body(market)
			candidates := body.Scopes[0].Candidates
			var legacy []productionRouteCandidate
			for _, candidate := range candidates {
				if candidate.Family != FamilyBreakoutRetest {
					legacy = append(legacy, candidate)
				}
			}
			if len(legacy) != 3 || len(candidates) != 4 {
				t.Fatalf("arrangement: legacy=%d of %d candidates — the fixture no longer has exactly one breakout lane", len(legacy), len(candidates))
			}
			body.Scopes[0].Candidates = legacy
			fixture.write(t, market, body) // 현재 스키마 · 새 서명 · 새 digest 핀 — 어긋난 것은 가족 수 하나뿐
			if authority, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market]); err == nil {
				t.Fatalf("a legacy three-family manifest loaded as route authority: %d candidates", len(authority.Request().Candidates))
			}
		})
	}
}

func TestAThreeFamilyActivationManifestPromotesNothingInEitherMarket(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			if activation, err := LoadProductionFamilyActivation(context.Background(), fixture.write(t, market, fixture.body(market))); err != nil || !activation.Verified() {
				t.Fatalf("arrangement: the four-family %s activation does not verify: %v", market, err)
			}
			fixture = newFamilyActivationFixture(t)
			body := fixture.body(market)
			body.Descriptors = fixture.descriptors(market, nil)[:3] // 골든 순서의 넷째 = breakout
			activation, err := LoadProductionFamilyActivation(context.Background(), fixture.write(t, market, body))
			if !errors.Is(err, ErrProductionFamilyActivationUnavailable) || activation.Verified() {
				t.Fatalf("%s three-family activation: err=%v verified=%v, want Unavailable and nothing verified", market, err, activation.Verified())
			}
		})
	}
}
