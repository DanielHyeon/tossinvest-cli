package engine

import (
	"context"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.1.2.2 의 관문 절반이다: **네 전략군 레인이 조정자 앞에 선다.**
//
// 오늘 한 시장의 제안은 `coordinateMarketProposals` 에서 곧바로 조정자로 들어간다.
// 여덟 레인은 그 뒤 관측만 하고 아무것도 막지 못한다(5.1.2.1 이 그 상태를 적어
// 두었다). 여기서 그것을 바꾼다 — 다만 **검증된 활성화가 있는 시장에서만** 이다.
//
// **왜 활성화가 있을 때만인가. 두 권위가 갈리는 자리를 여기서 푼다.**
// 5.1.2.1 은 "오늘 여덟을 관문으로 세우면 생산 진입이 0 이 되고 그것은 토글 OFF
// 불변식(§0-2, OFF = upstream 동작)에 어긋난다" 고 적었다. 스펙은 반대로
// "required activation authority 가 missing 이면 broker exposure-raising request
// 는 0건이어야 한다 (SHALL)" 고 적었다. 둘을 함께 만족시키는 유일한 모양이
// **관문을 활성화된 런타임 안에만 세우는 것**이다: 활성화가 없으면 기존 시장
// 단위 경로가 그대로 돌고(= upstream 동작 보존), 활성화가 있으면 그 시장의
// 넷이 각자 판정한다.
//
// design 이 partial 3-of-4 를 시장 전체 OFF 로 못 박았으므로(`design.md:208`),
// 활성화된 시장의 서술자는 언제나 정확히 넷이다. 그래서 관문이 실제로 막는 것은
// **잠긴 레인 하나**이고, 그 가족의 제안만 멈추고 이웃 셋은 계속한다. 그것이
// 이 change 가 사려던 것이다 — "시장 장애 격리이지 전략군 장애 격리가 아니다"
// (`design.md:3`).

// 신뢰 핀은 시장마다 **하나**다 [a112 결정 61]. 앞 판본은 여기에 key id 와 공개키
// 둘을 더 읽었다. 서명을 뺐으므로 배포가 관리할 값이 셋에서 하나로 줄었고, 그 하나가
// 후보 임계값 활성화(`strategyCandidateActivationDigestEnv`)와 같은 모양이다.
const (
	strategyFamilyActivationKRManifestDigestEnv = "TOSSOS_STRATEGY_FAMILY_ACTIVATION_KR_MANIFEST_SHA256"
	strategyFamilyActivationUSManifestDigestEnv = "TOSSOS_STRATEGY_FAMILY_ACTIVATION_US_MANIFEST_SHA256"
)

// strategyFamilyGate 는 한 시장의 네 전략군 레인이 조정자 앞에서 하는 판정이다.
//
// **영값은 "이 시장의 4-가족 런타임이 판정 주체가 아니다" 는 뜻이고, 그것이
// 오늘의 값이다.** 생산에는 4-가족 활성화 매니페스트가 배포돼 있지 않다(측정:
// `~/.config/tossctl` 에 strategy-* 매니페스트 0 건).
type strategyFamilyGate struct {
	activation strategyrouter.FamilyActivation
	lanes      []*strategyworker.Lane
}

// installed 는 이 관문이 실제로 판정하는지다.
//
// 활성화와 레인 **둘 다** 있어야 한다. 하나만 보면 다른 하나가 조용히 빠졌을 때
// 관문이 통과 도장으로 바뀐다 — 레인이 비면 아무 제안도 자기 주인을 못 찾으므로
// 아래 admit 이 전부 거절하게 되고, 그것은 fail-closed 가 아니라 기능 정지다.
func (gate strategyFamilyGate) installed() bool {
	return gate.activation.Verified() && len(gate.lanes) != 0
}

// admit 는 이 봉인된 제안이 조정자에 닿는지를 그 가족의 레인에게 묻는다.
//
// 관문이 서지 않았으면 아무것도 묻지 않고 통과시킨다. 두 번째 반환값은 레인이
// 낸 결과 종류이며 통과할 때는 EMITTED 다 — 관문이 서지 않은 통과는 빈 값이고,
// 그 둘을 구별해야 "왜 통과했는가" 에 답할 수 있다.
//
// 봉투를 레인이 만든 것으로 쓰는 이유: 만들지 않고 통과만 시키면 여덟은 여전히
// 봉투 생산자가 아니고, 엔진의 사본과 레인의 사본이 갈릴 수 있다. 두 사본이
// 같은 값인지는 시험이 값으로 잰다
// (`TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope`).
func (gate strategyFamilyGate) admit(input strategyworker.Input) (strategycoordinator.Envelope, strategyworker.Outcome, bool) {
	if !gate.installed() {
		return strategycoordinator.Envelope{}, "", true
	}
	for _, lane := range gate.lanes {
		if !lane.Owns(input.Proposal) {
			continue
		}
		cycle := lane.Run(gate.activation, input)
		return cycle.Envelope, cycle.Outcome, cycle.Outcome == strategyworker.OutcomeEmitted
	}
	// 어느 레인도 자기 것이라고 하지 않으면 닫는다. 활성화된 런타임에서
	// 주인 없는 제안은 조정자에 들어가서는 안 된다 — 그 상태는 레인 목록과
	// 제안의 가족 유도가 갈렸다는 뜻이고, 갈렸을 때 통과시키는 쪽을 고르면
	// 이 관문이 있는 이유가 사라진다.
	return strategycoordinator.Envelope{}, strategyworker.OutcomeRefused, false
}

// familyGateFor 는 이 시장의 관문을 만든다.
//
// 활성화는 **이 주기에** 읽는다. 값으로 들고 다니지 않는 이유는 만료 때문이다 —
// 구운 승격은 묵은 ON 이 되고, 묵은 OFF 는 안전한 방향이지만 묵은 ON 은 아니다.
func (loader *strategyProposalAuthorityLoader) familyGateFor(ctx context.Context, market StrategyMarket,
	schedule strategyScheduleMarketAuthority, routes strategyRouteMarketAuthority, observedAt time.Time,
) strategyFamilyGate {
	if loader == nil {
		return strategyFamilyGate{}
	}
	// 레인이 없는 경우를 여기서 또 막지 않는다. 그 판정은 아래 `installed` 하나에
	// 있다 — 첫 판본은 `loader.lanes == nil` 을 여기 달고 있었고 반증이 그것을
	// 지워도 아무 색도 안 바뀌었다(레인이 없으면 관문이 서지 않으므로). 같은
	// 규칙을 두 곳에 두면 각자가 상대의 시험을 통과시킨다.
	load := loader.loadActivation
	if load == nil {
		load = loader.loadFamilyActivation
	}
	activation, err := load(ctx, market, schedule, routes, observedAt)
	if err != nil || !activation.Verified() {
		return strategyFamilyGate{}
	}
	return strategyFamilyGate{activation: activation, lanes: loader.lanes.lanesFor(market)}
}

// loadFamilyActivation 은 이 시장의 4-가족 활성화를 읽는다.
//
// **여기서 결속하는 세 값은 이 단계에 실제로 존재하는 사실뿐이다.** 위험 번들과
// ProtectionReady digest 는 이 단계에 아직 없다(둘 다 제안 뒤에 수집된다).
// 없는 사실을 결속하면 그 결속은 어떤 정상 입력으로도 참이 될 수 없고, 그것이
// 이 change 가 이미 한 번 만들었다 고친 "문 없는 fail-closed" 다. 그 둘은
// 매니페스트 몸통에 실려 나가고 그 사실들이 존재하는 단계가 결속한다.
//
// 보정 digest 는 경로 권한이 들고 있는 값에서 읽는다. 이 시장의 항목들이 서로
// 다른 보정을 말하면 읽지 않는다 — 하나를 고르면 그 선택이 고른 사람의 것이 된다.
func (loader *strategyProposalAuthorityLoader) loadFamilyActivation(ctx context.Context, market StrategyMarket,
	schedule strategyScheduleMarketAuthority, routes strategyRouteMarketAuthority, observedAt time.Time,
) (strategyrouter.FamilyActivation, error) {
	digestEnv := strategyFamilyActivationKRManifestDigestEnv
	if market == StrategyMarketUS {
		digestEnv = strategyFamilyActivationUSManifestDigestEnv
	}
	calibration, agreed := strategyMarketCalibrationDigest(routes)
	if !agreed {
		return strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUnavailable
	}
	// 위험은 **정책 매니페스트** digest 로 결속한다 (태스크 8.8.2).
	//
	// 앞 판본은 per-cycle 위험 스냅샷 봉인에 걸었고 그 값은 종목과 파도 벽시계를
	// 품으므로 사람이 미리 서명한 상수가 같아질 수 없었다. 여기서 읽는 것은
	// 운영자가 이미 배포로 핀하는 값(`TOSSOS_RISK_BUCKET_<MARKET>_MANIFEST_SHA256`)
	// 이고, 정책을 다시 서명할 때만 바뀐다 — 그래서 24시간 수명 동안 안정적이고
	// 이 단계에 이미 존재한다.
	//
	// 같은 env 상수를 위험 권한 로더도 쓴다(strategy_risk_authority.go). 두 곳이
	// 같은 이름을 읽으므로 값이 갈릴 수 없다.
	riskPolicyEnv := strategyRiskKRManifestDigestEnv
	if market == StrategyMarketUS {
		riskPolicyEnv = strategyRiskUSManifestDigestEnv
	}
	return strategyrouter.LoadProductionFamilyActivation(ctx, strategyrouter.FamilyActivationConfig{
		ConfigDir: loader.configDir, Market: strategyRouterMarket(market),
		ManifestDigest: strings.TrimSpace(loader.getenv(digestEnv)), ObservedAt: observedAt,
		RouteManifestDigest: routes.snapshot.ManifestDigest, CalibrationDigest: calibration,
		CalendarVersion: schedule.calendar.Version, BuildDigest: strategyRuntimeBuildDigest(),
		RiskPolicyDigest: strings.TrimSpace(loader.getenv(riskPolicyEnv)),
	})
}

// strategyMarketCalibrationDigest 는 이 시장의 모든 경로 항목이 **같은** 보정
// 기준을 말하는지 보고 그 값을 돌려준다.
//
// 항목 하나를 골라 읽지 않는 이유: 매니페스트 몸통에 하나뿐인 값이므로 항목들이
// 갈릴 수 없어야 하는데, 갈렸을 때 하나를 고르면 그 선택이 고른 사람의 것이 되고
// 어긋남은 아무도 보고하지 않는다.
func strategyMarketCalibrationDigest(routes strategyRouteMarketAuthority) (string, bool) {
	// 첫 항목을 기준으로 삼지 않는다. `routes.entries[0]` 은 "이 시장의 제안은
	// 하나" 라는 가정과 **같은 철자**이고(5.5 의 census 가 그 철자를 센다),
	// 실제로 하려는 말은 "항목 전부가 같은 하나에 동의한다" 다. 집합으로 세면
	// 그 말이 그대로 코드가 된다.
	seen := map[string]struct{}{}
	for _, entry := range routes.entries {
		value := strings.TrimSpace(entry.route.Calibration().CalibrationDigest)
		if value == "" {
			return "", false
		}
		seen[value] = struct{}{}
	}
	if len(seen) != 1 {
		return "", false
	}
	for value := range seen {
		return value, true
	}
	return "", false
}
