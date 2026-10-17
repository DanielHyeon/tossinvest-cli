package engine

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112 7.3.1 SHADOW 운반(브리프 v3.3 §4).
//
// 반사실 입력은 조정자 관문 **앞** 의 전체 제안이다 — 관문이 멈춘 OFF 가족의 제안이 바로 SHADOW 의 대상 모집단이기 때문이다. 그 묶음은
// authority 구조체(주문 경로의 수신자)에 **넣지 않고** 나란히 가는 별도 값이다: 조정자 → collectMarket → collect → 조립 → evaluate.
// 「관측 없음」 은 빈 묶음이 아니라 **부재 값**(observed=false)이다 — 빈 목록이면 OFF 레인 전부가 NO_INPUT 이 되어 거짓 보고가 된다.

// 핀 env 는 시장마다 하나다(결정 63 — 활성화 핀과 같은 모양). 오늘 생산에서는 비어 있다(미선언 — 파일 I/O 0).
const (
	strategyFamilyShadowKRManifestDigestEnv = "TOSSOS_STRATEGY_FAMILY_SHADOW_KR_MANIFEST_SHA256"
	strategyFamilyShadowUSManifestDigestEnv = "TOSSOS_STRATEGY_FAMILY_SHADOW_US_MANIFEST_SHA256"
)

// strategyShadowBatch 는 한 시장 한 물결의 관문 앞 제안 묶음과 그 shadow 매니페스트 결속 설정이다. 영값이 부재 값이다.
type strategyShadowBatch struct {
	observed bool
	config   strategyshadow.Config
	inputs   []strategyworker.ShadowInput
}

// collect 는 관문 앞 제안 하나를 묶음에 더한다. 조정 루프 안에서 불리므로 panic 거리가 없는 모양 하나(생성자 결과의 append)로 고정한다 —
// 수신자는 조정 함수의 주소 지정 가능한 지역 값이다(브리프 §4 shape 핀).
func (batch *strategyShadowBatch) collect(proposal strategyarbiter.Proposal) {
	batch.inputs = append(batch.inputs, strategyworker.NewShadowInput(proposal))
}

// boundTo 는 관측된 묶음에 그 시장의 결속 설정을 붙인다. 부재 값은 부재 값으로 남는다.
func (batch strategyShadowBatch) boundTo(config strategyshadow.Config) strategyShadowBatch {
	if !batch.observed {
		return strategyShadowBatch{}
	}
	batch.config = config
	return batch
}

// strategyShadowPair 는 두 시장의 묶음이다(authority 짝 옆의 별도 짝).
type strategyShadowPair struct {
	kr, us strategyShadowBatch
}

// forMarket 은 형제 strategyProposalAuthorityPair.forMarket 과 같은 모양이다(시장 비교 · 필드 반환 · 호출 0). 주기 함수의 evaluate 인자
// 자리에서 dispatch **앞**에 평가되므로 이 모양이 「dispatch 앞 shadow 일 0」 의 일부다(AST 핀).
func (pair strategyShadowPair) forMarket(market StrategyMarket) strategyShadowBatch {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyShadowBatch{}
}

// shadowConfig 는 이 시장 shadow 매니페스트의 결속 설정이다 — 결속 다섯(경로 digest · 보정 합의 · 달력 · 위험 정책 env · 빌드)은
// 활성화 적재(loadFamilyActivation)와 **같은 원천**에서 읽고(a112_shadow_binding_test.go 가 생산 활성화 적재기로 대조), 매니페스트 핀만
// shadow 전용 env(TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256)다. 파일은 읽지 않는다(env 와 이미 받은 권한 값뿐) — 매니페스트
// 적재는 shadow 단계가 한다.
func (loader *strategyProposalAuthorityLoader) shadowConfig(market StrategyMarket, schedule strategyScheduleMarketAuthority,
	routes strategyRouteMarketAuthority,
) strategyshadow.Config {
	digestEnv, riskPolicyEnv := strategyFamilyShadowKRManifestDigestEnv, strategyRiskKRManifestDigestEnv
	if market == StrategyMarketUS {
		digestEnv, riskPolicyEnv = strategyFamilyShadowUSManifestDigestEnv, strategyRiskUSManifestDigestEnv
	}
	calibration, _ := strategyMarketCalibrationDigest(routes)
	return strategyshadow.Config{ConfigDir: loader.configDir, Market: strategyRouterMarket(market),
		ManifestDigest: strings.TrimSpace(loader.getenv(digestEnv)), RouteManifestDigest: routes.snapshot.ManifestDigest,
		CalibrationDigest: calibration, CalendarVersion: schedule.calendar.Version, BuildDigest: strategyRuntimeBuildDigest(),
		RiskPolicyDigest: strings.TrimSpace(loader.getenv(riskPolicyEnv))}
}
