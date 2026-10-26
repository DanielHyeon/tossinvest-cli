# Branch Test Map: `candidateFilterDescriptorEqual`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path: `Market` 외 field 전체 비교 | `TestCandidateFilterMatrixKeepsCategoryDifferenceSeparate` | 기존 나열 구현도 `Category`를 이미 비교해 이 단독 RED는 없음 | `p2-green-output.txt`에서 Category 차이 공통 병합 금지 확인 |
