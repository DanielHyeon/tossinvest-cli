# Branch Test Map: `TestARateLimitedRankingIsReportedAsOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 생성 실패 | 자체 실행 | yes (컴파일) | yes |
| B2 | 429 사상 | 자체 실행 | — (기존 동작) | yes |
