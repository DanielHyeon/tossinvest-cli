package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

// a112ScopedAccount 는 시험이 손으로 만드는 시장 계좌 권한에 그 범위의 목록 원소를 붙인다(a112 5.2.2.2 — 1차 레그 권한은 범위 목록에서만
// 계좌 권한을 고른다; 생산 적재기는 늘 목록을 채운다). result 는 그 시장의 조립 제안.
func a112ScopedAccount(value strategyAccountMarketAuthority, result strategyflow.Result, authority strategyaccount.Authority) strategyAccountMarketAuthority {
	key, _ := strategyOwnerKeyOf(result.Lineage)
	value.scopes = []strategyAccountScopeAuthority{{key: key, authority: authority, ready: true, reason: StrategyAccountReady}}
	return value
}
