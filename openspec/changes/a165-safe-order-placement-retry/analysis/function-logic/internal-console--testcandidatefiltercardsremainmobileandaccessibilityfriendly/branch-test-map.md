# Branch Test Map: `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `.veto-matrix` 부재·복수 | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` | not-applicable | 대상 GREEN 통과 |
| B2 | `scope=col` registry 시장 수와 다름 | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` | not-applicable | 대상 GREEN 통과 |
| B3 | `scope=row` 시장 합집합 key 수와 다름 | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` | P2 RED에서 첫 시장 key 수를 사용 | `p2-green-output.txt`에서 행렬 합집합 기준 GREEN |
| B4 | 접근성·모바일 marker 목록 반복 | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` | not-applicable | 대상 GREEN 통과 |
| B5 | 시장 deep link·region·`dl` marker 누락 | `TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly` | not-applicable | 대상 GREEN 통과 |
