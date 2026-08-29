package strategyproposal

import "testing"

// 이 파일에는 빌드 태그가 없다 — CI 가 실제로 돌려야 하는 검사이기 때문이다.

// 리뷰 지적 C2: `For` 는 모호한 종목을 거절하면서 이유를 밖에 알려주지 않는다.
// 부르는 쪽이 "제안이 없었다"와 "고를 수 없어서 닫았다"를 구별하지 못하면
// 한 종목의 fail-closed 가 시장 수준의 fail-open 이 된다. `Ambiguous` 는 그
// 구별을 밖에서 할 수 있게 해 준다.
func TestAmbiguousSeparatesNoProposalFromTooManyProposals(t *testing.T) {
	batch := ProductionBatchAuthority{values: map[string]ProductionAuthority{
		batchKey("AAPL", "us_short_participation_continuation_v1"): {snapshotID: "one"},
		batchKey("AAPL", "us_short_dislocation_reversal_v1"):       {snapshotID: "two"},
		batchKey("MSFT", "us_short_participation_continuation_v1"): {snapshotID: "three"},
	}}
	if !batch.Ambiguous("AAPL") {
		t.Fatal("a symbol with two family proposals was not reported ambiguous")
	}
	if batch.Ambiguous("MSFT") {
		t.Fatal("a symbol with exactly one proposal was reported ambiguous")
	}
	if batch.Ambiguous("TSLA") {
		t.Fatal("a symbol with no proposal at all was reported ambiguous")
	}
	// 두 종목 다 하나씩이면 어느 쪽도 모호하지 않다. 즉 Ambiguous 는
	// "시장에 제안이 둘"이 아니라 "한 종목에 가족이 둘"만 잡는다.
	single := ProductionBatchAuthority{values: map[string]ProductionAuthority{
		batchKey("AAPL", "us_short_participation_continuation_v1"): {snapshotID: "one"},
		batchKey("MSFT", "us_short_dislocation_reversal_v1"):       {snapshotID: "two"},
	}}
	if single.Ambiguous("AAPL") || single.Ambiguous("MSFT") {
		t.Fatal("two different symbols with one lane each were treated as ambiguous")
	}
	// 접두사가 겹치는 이웃 종목이 남의 레인을 빌려 모호해지지 않는다.
	prefix := ProductionBatchAuthority{values: map[string]ProductionAuthority{
		batchKey("AAP", "us_short_participation_continuation_v1"):  {snapshotID: "one"},
		batchKey("AAPL", "us_short_dislocation_reversal_v1"):       {snapshotID: "two"},
		batchKey("AAPL", "us_short_participation_continuation_v1"): {snapshotID: "three"},
	}}
	if prefix.Ambiguous("AAP") {
		t.Fatal("AAP borrowed AAPL's lanes and was reported ambiguous")
	}
	if !prefix.Ambiguous("AAPL") {
		t.Fatal("AAPL's own two lanes were not reported ambiguous")
	}
	// 종목 이름의 공백과 대소문자는 For/LanesFor 와 같은 규칙으로 다듬는다.
	if !batch.Ambiguous("  aapl ") {
		t.Fatal("Ambiguous did not canonicalise the symbol the way LanesFor does")
	}
}
