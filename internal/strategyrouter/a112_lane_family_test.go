package strategyrouter

import "testing"

// a112 6.1 — lane → family 해소의 정본은 `productionRouteDescriptors` 표 하나다. `ProductionLaneFamily` 는 그 표를 **유도**하고(손 복사 금지),
// 이 시험은 표 자체의 완전성(시장마다 네 전략군이 정확히 한 번씩 — 8 = 4 × 2)과 유도가 표와 한 글자도 다르지 않음을 잰다.
func TestProductionLaneFamilyIsDerivedFromTheCompleteRouteTable(t *testing.T) {
	families := []Family{FamilyContinuation, FamilyReversal, FamilyWeeklyValue, FamilyBreakoutRetest}
	total := 0
	for _, market := range []Market{MarketKR, MarketUS} {
		table := productionRouteDescriptors(market)
		seen := map[Family]int{}
		for laneID, descriptor := range table {
			seen[descriptor.Family]++
			total++
			if got, ok := ProductionLaneFamily(market, laneID); !ok || got != descriptor.Family {
				t.Fatalf("%s lane %s: derived family %q ok=%v, table says %q", market, laneID, got, ok, descriptor.Family)
			}
		}
		for _, family := range families {
			if seen[family] != 1 {
				t.Fatalf("%s route table holds family %s %d time(s), want exactly once (table=%v)", market, family, seen[family], table)
			}
		}
		if len(table) != len(families) {
			t.Fatalf("%s route table has %d lanes, want %d", market, len(table), len(families))
		}
	}
	if total != 8 {
		t.Fatalf("route table lanes=%d, want 8 (4 families × 2 markets)", total)
	}
	// 표 밖의 레인 · 다른 시장의 레인은 해소되지 않는다(fail-closed — 모르는 레인의 family 를 지어내지 않음).
	for _, market := range []Market{MarketKR, MarketUS} {
		other := MarketUS
		if market == MarketUS {
			other = MarketKR
		}
		for laneID := range productionRouteDescriptors(other) {
			if _, ok := ProductionLaneFamily(market, laneID); ok {
				t.Fatalf("%s resolved %s's lane %s", market, other, laneID)
			}
		}
		if _, ok := ProductionLaneFamily(market, "no-such-lane"); ok {
			t.Fatalf("%s resolved an unknown lane", market)
		}
	}
}
