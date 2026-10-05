package verifylive

// reconcile_m0_layer_test.go — a121 RED 로트 처분 ⑥: M0 제외는 현재 PendingCleanup 의 withoutM0ManualReconcile 와
// 겹쳐(M0Unsettled 의 부모 ⊆ m0ManualReconcileIDs) 선택 층 시험으로는 가려진다. 층을 내려 withoutM0Unsettled 를 직접
// 불러, PendingCleanup 이 거르지 않은 입력에서도 M0 가 가리키는 부모가 빠짐을 단언한다(변이 「M0 필터 제거」 를 잡는 자리).

import "testing"

func TestWithoutM0UnsettledDropsTheNamedParentAtItsOwnLayer(t *testing.T) {
	entries := []Entry{{Kind: KindM0Checkpoint, M0Checkpoint: &M0Checkpoint{
		Kind: "parent-created", ClientOrderID: "client-1", ParentConditionalID: "CO-M0", Symbol: rcSymbol, Market: MarketKR}}}
	in := []Artifact{
		{Kind: KindConditional, ID: "CO-M0", Symbol: rcSymbol},
		{Kind: KindConditional, ID: "CO-OTHER", Symbol: rcSymbol},
	}
	got, err := withoutM0Unsettled(entries, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "CO-OTHER" {
		t.Fatalf("withoutM0Unsettled kept %v, want only CO-OTHER", rcIDs(got))
	}
}

func TestWithoutM0UnsettledRefusesAnAmbiguousOwner(t *testing.T) {
	var entries []Entry
	for _, client := range []string{"client-a", "client-b"} {
		entries = append(entries, Entry{Kind: KindM0Checkpoint, M0Checkpoint: &M0Checkpoint{
			Kind: "pending-create", ClientOrderID: client, Symbol: rcSymbol, Market: MarketKR}})
	}
	_, err := withoutM0Unsettled(entries, []Artifact{{Kind: KindConditional, ID: "CO-1"}})
	rcRequireRefusal(t, err, RefuseM0Unresolved)
}
