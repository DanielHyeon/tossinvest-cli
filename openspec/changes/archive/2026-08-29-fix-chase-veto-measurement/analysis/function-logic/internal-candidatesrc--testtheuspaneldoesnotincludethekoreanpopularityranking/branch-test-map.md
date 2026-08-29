# Branch Test Map: `TestTheUSPanelDoesNotIncludeTheKoreanPopularityRanking`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | KR에 인기 순위 | 자체 실행 | yes (컴파일) | yes |
| B2 | US에 인기 순위 없음 | 자체 실행 | — (기존 동작) | yes |
| B3 | US 패널은 비지 않는다 | 자체 실행 | — (기존 동작) | yes |
