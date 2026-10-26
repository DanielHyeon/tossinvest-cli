# Branch Test Map: `TestCandidateFilterCardsContainExactMarketMetricMatrix`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `.veto-matrix` 반복 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
| B2 | native table의 `aria-readonly` 잔존 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | `review-red-output.txt` FAIL | 대상 GREEN 통과 |
| B3 | `scope=col` registry 시장 수와 다름 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
| B4 | `scope=row` registry key 수와 다름 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
| B5 | registry key 반복·누락 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
