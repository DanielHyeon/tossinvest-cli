package strategyproposal

import "testing"

// 이 파일에는 빌드 태그가 없다 — CI 가 실제로 돌려야 하는 검사이기 때문이다.
// 한 종목이 여러 가족을 동시에 제안할 수 있게 되면서 담는 열쇠가 (종목, 레인)이 되었다.
// For 는 그 종목의 제안이 정확히 하나일 때만 돌려주고, 둘 이상이면 고르지 않고 거절한다.
// 임의로 하나를 집으면 그게 곧 "평가 전에 가족을 고른 것"이 되기 때문이다.

const (
	continuationLaneForTest = "kr_short_continuation_v1"
	reversalLaneForTest     = "kr_short_reversal_v1"
	weeklyLaneForTest       = "kr_weekly_value_v1"
)

func batchWith(entries map[string]string) ProductionBatchAuthority {
	values := make(map[string]ProductionAuthority, len(entries))
	for key, id := range entries {
		values[key] = ProductionAuthority{snapshotID: id}
	}
	return ProductionBatchAuthority{values: values}
}

// 두 가족이 같은 종목을 제안하면 For 는 아무것도 돌려주지 않아야 한다.
func TestForRefusesASymbolThatHasMoreThanOneLane(t *testing.T) {
	authority := batchWith(map[string]string{
		batchKey("AAPL", continuationLaneForTest): "continuation",
		batchKey("AAPL", reversalLaneForTest):     "reversal",
	})
	got, ok := authority.For("AAPL")
	if ok {
		t.Fatalf("두 가족이 제안했는데 %q 하나를 골라 줬다", got.SnapshotID())
	}
	if got.SnapshotID() != "" {
		t.Fatalf("거절인데 값이 새어 나왔다: %q", got.SnapshotID())
	}
}

// 제안이 정확히 하나면 그것을 돌려준다.
func TestForReturnsTheLaneWhenExactlyOneExists(t *testing.T) {
	authority := batchWith(map[string]string{batchKey("AAPL", reversalLaneForTest): "reversal"})
	got, ok := authority.For("aapl")
	if !ok || got.SnapshotID() != "reversal" {
		t.Fatalf("하나뿐인 제안을 못 돌려줬다: ok=%v id=%q", ok, got.SnapshotID())
	}
}

// 없는 종목은 당연히 거절이다.
func TestForRefusesAnUnknownSymbol(t *testing.T) {
	authority := batchWith(map[string]string{batchKey("AAPL", reversalLaneForTest): "reversal"})
	if _, ok := authority.For("MSFT"); ok {
		t.Fatal("담지 않은 종목을 돌려줬다")
	}
}

// LanesFor 는 모든 가족을 레인 이름 순서로, 매번 같은 순서로 돌려줘야 한다.
// map 순회는 Go 가 일부러 뒤섞으므로, 정렬을 빼면 이 반복 검사가 거의 확실히 걸린다.
func TestLanesForReturnsEveryLaneInTheSameOrderEveryTime(t *testing.T) {
	authority := batchWith(map[string]string{
		batchKey("AAPL", weeklyLaneForTest):       "weekly",
		batchKey("AAPL", continuationLaneForTest): "continuation",
		batchKey("AAPL", reversalLaneForTest):     "reversal",
	})
	want := []string{"continuation", "reversal", "weekly"} // 레인 이름 오름차순
	for round := 0; round < 64; round++ {
		values := authority.LanesFor("AAPL")
		if len(values) != len(want) {
			t.Fatalf("%d 회차: 가족 %d 개를 돌려줬는데 %d 개여야 한다", round, len(values), len(want))
		}
		for index, id := range want {
			if values[index].SnapshotID() != id {
				t.Fatalf("%d 회차: %d 번째가 %q 인데 %q 여야 한다", round, index, values[index].SnapshotID(), id)
			}
		}
	}
}

// 종목 이름이 앞부분만 같은 다른 종목의 제안을 끌어오면 안 된다.
// 열쇠를 이을 때 넣은 구분자(\x00)가 이것을 막는다.
func TestLanesForDoesNotLeakIntoASymbolThatMerelySharesAPrefix(t *testing.T) {
	authority := batchWith(map[string]string{
		batchKey("AAP", reversalLaneForTest):  "short-symbol",
		batchKey("AAPL", reversalLaneForTest): "long-symbol",
	})
	values := authority.LanesFor("AAP")
	if len(values) != 1 || values[0].SnapshotID() != "short-symbol" {
		t.Fatalf("앞부분만 같은 종목까지 끌어왔다: %d 개", len(values))
	}
	got, ok := authority.For("AAP")
	if !ok || got.SnapshotID() != "short-symbol" {
		t.Fatalf("앞부분이 같다는 이유로 하나뿐인 제안이 거절되었다: ok=%v id=%q", ok, got.SnapshotID())
	}
}
