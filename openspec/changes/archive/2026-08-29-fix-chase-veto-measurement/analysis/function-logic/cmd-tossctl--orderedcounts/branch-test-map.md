# Branch Test Map: `orderedCounts`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 선언 순서대로, 0도 칸을 차지한다 | `cmd/tossctl/candidate_test.go`의 밴드 렌더 · `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne`(같은 규칙의 상류) | — (동작 무변경) | yes |
| B2 | 밴드가 없는 code는 'none' | **커버 없음** | no | no |
