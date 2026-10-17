package strategyhandoff

// 소유자 범위의 축을 축마다 못 박는다(2026-09-30 리뷰 — 보이스 B #1 · codex P2-4 · 보이스 C #4).
//
// 범위는 계좌 · 시장 · 종목 · 포지션 세대 넷이고 **horizon · 레인은 축이 아니다**. horizon 이나 레인을 축에 넣으면 같은 종목의
// 두 전략군이 서로 다른 범위로 보여 **둘 다** 상한을 통과한다 — 비보수 방향이다. 앞 판본은 중복 경우의 두 선택이 horizon ·
// 레인까지 같았으므로(영값) 그 변이가 전 스위트를 통과했다.

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func lineageSelection(identity, account string, market strategyrouter.Market, symbol string, generation uint64,
	horizon strategyrouter.Horizon, lane string,
) strategyflow.Result {
	return strategyflow.Result{Lineage: strategyflow.Lineage{Identity: identity, AccountRef: account, Market: market,
		Symbol: symbol, PositionGeneration: generation, Horizon: horizon, LaneID: lane}}
}

// 같은 범위의 두 전략군(horizon · 레인만 다름)은 여전히 한 범위 — 시장 전체 OverCapacity.
func TestTwoFamiliesOnOneScopeAreStillOneScope(t *testing.T) {
	for name, pair := range map[string][2]strategyflow.Result{
		"horizon": {
			lineageSelection("a", "acct", strategyrouter.MarketKR, "005930", 1, strategyrouter.HorizonShort, "lane-a"),
			lineageSelection("b", "acct", strategyrouter.MarketKR, "005930", 1, strategyrouter.HorizonWeekly, "lane-a"),
		},
		"lane": {
			lineageSelection("a", "acct", strategyrouter.MarketKR, "005930", 1, strategyrouter.HorizonShort, "lane-a"),
			lineageSelection("b", "acct", strategyrouter.MarketKR, "005930", 1, strategyrouter.HorizonShort, "lane-b"),
		},
	} {
		handoffs := AdmitEachOwnerScope(true, pair[:])
		if len(handoffs) != 1 || handoffs[0].Refusal() != OverCapacity || handoffs[0].Pending() != 2 {
			t.Fatalf("%s differs only: %d handoffs, first refusal %q — two families on one owner scope must refuse the market",
				name, len(handoffs), handoffs[0].Refusal())
		}
	}
}

// 양성 대조: 계좌만 · 시장만 다르면 다른 범위다(축 하나씩 — 결함이 사는 차원을 바꾼다).
func TestAccountAloneOrMarketAloneMakesAnotherScope(t *testing.T) {
	for name, pair := range map[string][2]strategyflow.Result{
		"account": {
			lineageSelection("a", "acct-1", strategyrouter.MarketKR, "005930", 1, "", ""),
			lineageSelection("b", "acct-2", strategyrouter.MarketKR, "005930", 1, "", ""),
		},
		"market": {
			lineageSelection("a", "acct", strategyrouter.MarketKR, "AAPL", 1, "", ""),
			lineageSelection("b", "acct", strategyrouter.MarketUS, "AAPL", 1, "", ""),
		},
	} {
		handoffs := AdmitEachOwnerScope(true, pair[:])
		if len(handoffs) != 2 || handoffs[0].Refusal() != Admitted || handoffs[1].Refusal() != Admitted {
			t.Fatalf("%s differs only: %d handoffs — that is two owner scopes", name, len(handoffs))
		}
	}
}

// ownerScopeOf 와 strategyrouter.OwnerKey 가 갈라지지 않게 묶는다(보이스 C #4). 이 패키지는 생산 import 허용 목록 때문에
// strategyrouter 를 들이지 못하고 축을 옮겨 적었다 — 옮겨 적은 쪽과 원본 양쪽을 여기서 대조한다.
func TestOwnerScopeAgreesWithTheRouterOwnerKey(t *testing.T) {
	// (가) 원본의 축 집합이 바뀌면 옮겨 적은 쪽을 다시 봐야 한다.
	var fields []string
	for i := 0; i < reflect.TypeOf(strategyrouter.OwnerKey{}).NumField(); i++ {
		fields = append(fields, reflect.TypeOf(strategyrouter.OwnerKey{}).Field(i).Name)
	}
	sort.Strings(fields)
	if strings.Join(fields, ",") != "AccountRef,Market,PositionGeneration,Symbol" {
		t.Fatalf("strategyrouter.OwnerKey fields=%v — ownerScopeOf copies exactly these four; update both together", fields)
	}
	// (나) 유효한 계보 쌍에서 두 판정이 같다: 같은 범위 ⇔ 같은 OwnerKey.
	base := lineageSelection("a", "acct", strategyrouter.MarketKR, "aapl", 1, strategyrouter.HorizonShort, "lane-a")
	variants := map[string]strategyflow.Result{
		"symbol spelling":  lineageSelection("b", " acct ", strategyrouter.MarketKR, " AAPL", 1, strategyrouter.HorizonWeekly, "lane-b"),
		"another symbol":   lineageSelection("b", "acct", strategyrouter.MarketKR, "MSFT", 1, "", ""),
		"another gen":      lineageSelection("b", "acct", strategyrouter.MarketKR, "AAPL", 2, "", ""),
		"another account":  lineageSelection("b", "acct-2", strategyrouter.MarketKR, "AAPL", 1, "", ""),
		"another market":   lineageSelection("b", "acct", strategyrouter.MarketUS, "AAPL", 1, "", ""),
		"account spelling": lineageSelection("b", "acct\t", strategyrouter.MarketKR, "AAPL", 1, "", ""),
		"account case":     lineageSelection("b", "ACCT", strategyrouter.MarketKR, "AAPL", 1, "", ""),
	}
	key := func(result strategyflow.Result) strategyrouter.OwnerKey {
		value, err := strategyrouter.NewOwnerKey(result.Lineage.AccountRef, result.Lineage.Market, result.Lineage.Symbol,
			result.Lineage.PositionGeneration)
		if err != nil {
			t.Fatalf("arrangement: %v", err)
		}
		return value
	}
	for name, other := range variants {
		sameScope := ownerScopeOf(base) == ownerScopeOf(other)
		sameKey := key(base) == key(other)
		if sameScope != sameKey {
			t.Errorf("%s: ownerScopeOf says same=%v, OwnerKey says same=%v", name, sameScope, sameKey)
		}
	}
}
