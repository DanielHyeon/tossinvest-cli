# Branch Test Map: `Tracker.AdjustmentApplied`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (357) `if` — if t.adjusted == nil | 아래 통합 근거 | n/a | yes |
| B2 | (360) `range` — for _, symbol := range symbols | 아래 통합 근거 | n/a | yes |
| B3 | (362) `if` — if key == "" | 아래 통합 근거 | n/a | yes |

추가 통합 근거: `TestTheCycleAfterAConvergenceReleasesTheBlock`(드라이버 두 주기),
`TestTheCreditCarriesTheComparisonItWasComputedFrom`(비교 as-of 전달).
