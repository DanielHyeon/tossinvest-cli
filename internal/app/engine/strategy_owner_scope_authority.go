package engine

import (
	"context"
	"errors"

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

// strategyRiskScopeAuthority 는 한 소유자 범위의 서명 위험 번들이다. ready 가 아니면 reason 이 그 범위의 공개 사유이고 cause 가 적재기가
// 돌려준 **원인 그대로**다(5.2.2.2 리뷰 수리 — 원인을 접지 않고 운반해 범위 국소 거절과 결함을 1차 레그가 타입으로 가름).
type strategyRiskScopeAuthority struct {
	key    strategyrouter.OwnerKey
	bundle riskbucket.RiskSnapshotAuthorityBundle
	ready  bool
	reason StrategyRiskReason
	cause  error
}

// strategyAccountScopeAuthority 는 한 소유자 범위(종목)의 계좌 권한이다. cause 는 위험 쪽과 같은 뜻.
type strategyAccountScopeAuthority struct {
	key       strategyrouter.OwnerKey
	authority strategyaccount.Authority
	ready     bool
	reason    StrategyAccountReason
	cause     error
}

// riskScopeCause 는 준비되지 않은 위험 범위의 원인과 그것이 **범위 국소 거절**인지다. 범위 국소는 적재기 원천의 신원
// (riskbucket.ErrProductionRiskScopeRefused — 서명 정책 밖 종목 · scope latch)이 있을 때뿐이고, 원장 결함 · 무결성 · 범위 항목 부재
// (조립이 같은 결과 집합으로 모은 권한에 그 범위가 없음 = 불일치)는 결함이다.
func (authority strategyRiskMarketAuthority) riskScopeCause(key strategyrouter.OwnerKey) (bool, error) {
	for _, scope := range authority.scopes {
		if scope.key != key {
			continue
		}
		if scope.cause == nil {
			return false, errors.New("production risk authority scope is not ready without a cause")
		}
		return errors.Is(scope.cause, riskbucket.ErrProductionRiskScopeRefused), scope.cause
	}
	return false, errors.New("production risk authority holds no entry for this owner scope")
}

// accountScopeCause 는 계좌 쪽 같은 분류다. 계좌 적재기는 원장을 읽지 않고(서명 매니페스트 파일만) 그 실패는 범위 국소이며, 결함은
// ctx 종료(취소 · 기한)와 범위 항목 부재뿐이다(Manager 조건 ④).
func (authority strategyAccountMarketAuthority) accountScopeCause(key strategyrouter.OwnerKey) (bool, error) {
	for _, scope := range authority.scopes {
		if scope.key != key {
			continue
		}
		if scope.cause == nil {
			return false, errors.New("production account authority scope is not ready without a cause")
		}
		fault := errors.Is(scope.cause, context.Canceled) || errors.Is(scope.cause, context.DeadlineExceeded)
		return !fault, scope.cause
	}
	return false, errors.New("production account authority holds no entry for this owner scope")
}

// strategyOwnerKeyOf 는 계보의 소유자 범위 정규형이다(봉인의 선택과 같은 키).
func strategyOwnerKeyOf(lineage strategyflow.Lineage) (strategyrouter.OwnerKey, bool) {
	key, err := strategyrouter.NewOwnerKey(lineage.AccountRef, lineage.Market, lineage.Symbol, lineage.PositionGeneration)
	return key, err == nil
}

// strategyScopeRefusal 은 **그 소유자 범위에만** 해당하는 거절이다(Manager 판정 J4 — 문구가 아니라 타입으로 분류). 주문 경로의 반복은 이
// 타입의 거절만 건너뛰고 다음 범위로 가며(각각 기록), 그 밖의 오류(원장 · Gateway · 중앙 무결성 · 위조 의심)는 주기를 멈춘다.
//
// 이 타입을 만드는 자리는 1차 레그 권한의 둘뿐이다(census — a112_scope_refusal_census_test.go): 그 범위의 위험 · 계좌 권한이 **범위 국소
// 원인**(riskScopeCause · accountScopeCause 가 가름)으로 준비되지 않았을 때. 원장 결함 · ctx · 범위 항목 부재는 이 타입이 아니다 — 넓히면
// 원장 · Gateway 오류가 「범위 거절」로 오분류되어 한 범위의 고장 뒤에도 주기가 주문을 계속 낸다(5.2.2.2 리뷰 A #1 · codex #2 가 실측).
type strategyScopeRefusal struct {
	scope  strategyrouter.OwnerKey
	detail string
	cause  error
}

func (refusal *strategyScopeRefusal) Error() string {
	message := "strategy owner scope refused (" + refusal.scope.AccountRef + "/" + string(refusal.scope.Market) + "/" +
		refusal.scope.Symbol + "): " + refusal.detail
	if refusal.cause != nil {
		message += ": " + refusal.cause.Error()
	}
	return message
}

// Unwrap 은 범위 국소 원인(적재기의 신원)을 사슬에 남김 — 기록에서 원인이 사라지지 않게.
func (refusal *strategyScopeRefusal) Unwrap() error { return refusal.cause }
