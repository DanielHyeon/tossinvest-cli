package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 5.2.2.2 — 하류 권한의 소유자 범위 목록(Manager 판정 J1: 기존 시장 권한 안의 범위별 목록, 새 타입 계층 없음).
//
// 활성화 없는 시장은 목록 원소가 하나이고(오늘의 단일 범위 경로 그대로 — 토글 OFF = upstream), 서명 활성화된 시장은 조정자가 고른 소유자
// 범위마다 원소 하나다. 1차 레그 권한은 봉인과 같은 범위 키(strategyrouter.OwnerKey 정규형)로 위험 · 계좌 권한을 **각각** 다시 고른다 —
// 봉투의 값으로 폴백하지 않는다(J3 ①).

// strategyRiskScopeAuthority 는 한 소유자 범위의 서명 위험 번들이다. ready 가 아니면 reason 이 그 범위만의 거절 사유다.
type strategyRiskScopeAuthority struct {
	key    strategyrouter.OwnerKey
	bundle riskbucket.RiskSnapshotAuthorityBundle
	ready  bool
	reason StrategyRiskReason
}

// strategyAccountScopeAuthority 는 한 소유자 범위(종목)의 계좌 권한이다.
type strategyAccountScopeAuthority struct {
	key       strategyrouter.OwnerKey
	authority strategyaccount.Authority
	ready     bool
	reason    StrategyAccountReason
}

// strategyOwnerKeyOf 는 계보의 소유자 범위 정규형이다(봉인의 선택과 같은 키).
func strategyOwnerKeyOf(lineage strategyflow.Lineage) (strategyrouter.OwnerKey, bool) {
	key, err := strategyrouter.NewOwnerKey(lineage.AccountRef, lineage.Market, lineage.Symbol, lineage.PositionGeneration)
	return key, err == nil
}

// strategyScopeRefusal 은 **그 소유자 범위에만** 해당하는 거절이다(Manager 판정 J4 — 문구가 아니라 타입으로 분류). 주문 경로의 반복은 이
// 타입의 거절만 건너뛰고 다음 범위로 가며(각각 기록), 그 밖의 오류(원장 · Gateway · 중앙 무결성 · 위조 의심)는 주기를 멈춘다.
//
// 이 타입을 만드는 자리는 1차 레그 권한의 범위별 위험 · 계좌 권한 부재뿐이다(census — a112_owner_scope_trading_test.go). 범위를 넓히면
// 원장 · Gateway 오류가 「범위 거절」로 오분류되어 한 범위의 고장 뒤에도 주기가 주문을 계속 낸다 — 이 분류 경계가 이 설계의 안전선이다.
type strategyScopeRefusal struct {
	scope  strategyrouter.OwnerKey
	detail string
}

func (refusal *strategyScopeRefusal) Error() string {
	return "strategy owner scope refused (" + refusal.scope.AccountRef + "/" + string(refusal.scope.Market) + "/" +
		refusal.scope.Symbol + "): " + refusal.detail
}
