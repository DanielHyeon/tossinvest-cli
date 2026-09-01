//go:build tossos_testseams

package strategyflow

import "fmt"

// ResultWithRestatedStopProvenanceForTest 는 이미 봉인된 제안의 **손절 가격
// 출처만** 바꾼 제안을 만든다. 숫자(`priceMinor`)는 건드리지 않는다.
//
// 왜 필요한가. `PriceProvenance` 는 여덟 필드이고 `executionTermsIdentity` 는
// 여덟 개를 다 해시한다. 그런데 `AcceptedResultForAuthorityTest` 는 숫자 말고
// 일곱 개를 상수나 시장·시각에서 만들므로, 그 seam 만으로는 **숫자를 고정한 채
// 출처만 바꾼 제안**을 만들 수 없다. 그래서 8차 적대 리뷰가 잰 것처럼 비교를
// 숫자 셋으로만 좁혀도 시험이 전부 초록이었다.
//
// 이것은 지어낸 위협이 아니다. `continuationlane/execution_terms.go` 는 같은
// 숫자에 대해 신선한 후보와 저장된 권한에서 **서로 다른 출처**를 만든다
// (`saved-effective-stop`/`stop-state-v1`). 그 패키지는 위조된 손절 출처를
// 이미 위협으로 다루고 시험도 있다. 1차 진입 backstop 에는 없었다.
//
// lineage 는 건드리지 않으므로 `Lineage.Identity` 는 그대로다 — 바뀌는 축은
// terms identity 하나뿐이다.
func ResultWithRestatedStopProvenanceForTest(result Result, source, version, digest string) (Result, error) {
	if !result.ValidProposal() || source == "" || version == "" || digest == "" {
		return Result{}, fmt.Errorf("strategyflow stop provenance test seam: invalid input")
	}
	terms := result.ExecutionTerms
	terms.stop.source, terms.stop.version, terms.stop.digest = source, version, digest
	terms.identity = ""
	if !validExecutionTermsFields(terms) {
		return Result{}, fmt.Errorf("strategyflow stop provenance test seam: invalid execution terms")
	}
	terms.identity = executionTermsIdentity(terms)
	result.ExecutionTerms = terms
	result = sealProposalResult(result)
	if !result.ValidProposal() {
		return Result{}, fmt.Errorf("strategyflow stop provenance test seam: result did not reseal")
	}
	return result, nil
}
