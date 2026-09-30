package riskcalc_test

// a091 tasks 3.3a — 엔진이 「계좌 보유 0」을 가르는 판정(Bound == Holdings ∧ "0")의 근거를 핀함(design D3 ③ 5판).
//
//   (⇒, 정확)  Holdings 한정 0 이면 신선한 보유가 0 이다.
//   (⇐, 조건부) 신선한 보유 0 은 매도가능도 신선하고 로컬 수량이 유효할 때만 Holdings 한정 0 을 낸다. 매도가능이 없거나 낡으면
//              다른 한정 항(NoSnapshot · Stale)이 붙고, 로컬 수량이 비정상이면 오류다 — 엔진은 그 칸들을 critical 로 남긴다(과보고 방향).
//
// riskcalc 는 편집하지 않는다. 이 표가 깨지면 엔진의 ③ 배제가 틀린 칸을 거르게 된다.

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func TestA091AHoldingsBoundZeroMeansTheAccountHoldsNone(t *testing.T) {
	t.Parallel()
	stale := 11 * time.Second // AccountSnapshotStaleness(10s) 초과
	type cell struct {
		holdings  *riskcalc.QuantitySnapshot
		sellable  *riskcalc.QuantitySnapshot
		local     string
		wantErr   bool
		wantBound string // wantErr 가 아니면
		wantZero  bool
	}
	for name, c := range map[string]cell{
		// 신선 보유 0 × 매도가능 칸
		"held 0, sellable fresh 0, local 0":    {snap("0", time.Second), snap("0", time.Second), "0", false, riskcalc.FloorBoundHoldings, true},
		"held 0, sellable fresh 5, local 0":    {snap("0", time.Second), snap("5", time.Second), "0", false, riskcalc.FloorBoundHoldings, true},
		"held 0, sellable fresh 0, local 3":    {snap("0", time.Second), snap("0", time.Second), "3", false, riskcalc.FloorBoundHoldings, true},
		"held 0, sellable absent (leaks to ②)": {snap("0", time.Second), nil, "0", false, riskcalc.FloorBoundNoSnapshot, true},
		"held 0, sellable stale (leaks to ②)":  {snap("0", time.Second), snap("0", stale), "0", false, riskcalc.FloorBoundStaleSnapshot, true},
		"held 0, local invalid (leaks to ①)":   {snap("0", time.Second), snap("0", time.Second), "-1", true, "", false},
		"held 0 but stale (leaks to ②)":        {snap("0", stale), snap("0", time.Second), "0", false, riskcalc.FloorBoundStaleSnapshot, true},
		// 보유 양수 칸 — Holdings 한정 0 이 나와서는 안 됨
		"held 5, sellable 0":                     {snap("5", time.Second), snap("0", time.Second), "0", false, riskcalc.FloorBoundSellable, true},
		"held 5, local 5":                        {snap("5", time.Second), snap("5", time.Second), "5", false, riskcalc.FloorBoundLocalSells, true},
		"held 5, sellable 5, local 0 (not zero)": {snap("5", time.Second), snap("5", time.Second), "0", false, riskcalc.FloorBoundHoldings, false},
	} {
		got, err := riskcalc.ConfirmedFloorQuantity(riskcalc.ConfirmedFloorInputs{
			Now: floorNow, Symbol: "005930", Holdings: c.holdings, Sellable: c.sellable, LocalOpenSellQuantity: c.local,
		})
		if c.wantErr {
			if err == nil {
				t.Errorf("%s: err = nil, want an error (the engine reads it as a floor it could not compute)", name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got.Bound != c.wantBound || (got.Quantity == "0") != c.wantZero {
			t.Errorf("%s: floor = %q bound %q, want zero=%v bound %q", name, got.Quantity, got.Bound, c.wantZero, c.wantBound)
		}
		// (⇒): Holdings 한정 0 이면 신선한 보유가 0 이어야 함.
		if got.Bound == riskcalc.FloorBoundHoldings && got.Quantity == "0" && c.holdings.Quantity != "0" {
			t.Errorf("%s: a holdings-bound zero from a non-zero holding", name)
		}
	}
}
