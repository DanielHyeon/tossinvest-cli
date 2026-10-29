# Branch Test Map: `TestCandidateFilterCardsContainExactMarketMetricMatrix`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `.veto-matrix` 반복 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
| B2 | native table의 `aria-readonly` 잔존 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | `review-red-output.txt` FAIL | 대상 GREEN 통과 |
| B3 | `scope=col` registry 시장 수와 다름 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | not-applicable | 대상 GREEN 통과 |
| B4 | `scope=row` 시장 합집합 key 수와 다름 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | P2 RED에서 기존 첫 시장 길이 계산 | `p2-green-output.txt`로 `candidateFilterMatrixFromMarkets(markets).Rows` 기준 GREEN |
| B5 | 합집합 registry key 반복·누락 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | P2 RED에서 US 전용 key 누락 | 대상 GREEN 통과 |
| B6 | 합집합 registry key의 정확한 `<code>` 텍스트가 한 번이 아님 | `TestCandidateFilterCardsContainExactMarketMetricMatrix` | P2 RED에서 US 전용 key 누락 | 대상 GREEN 통과 |
