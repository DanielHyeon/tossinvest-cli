# Branch Test Map: `WTSPopular`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 비양수 size가 30이 된다 | **커버 없음** | no | no |
| B2 | nil clock이 시스템 시계가 된다 | `TestThePopularityRankingReportsNoTradingFigures` · `TestThePopularityRankingFallsBackToTheProductCode` · `TestThePopularityRankingRefusesAMarketItCannotSee` | — (신규 인자) | yes |

**정직한 커버리지 기록**: B1은 이 change 이전부터 있던 기본값 분기이고 테스트가 없다.
생산 호출부는 `Panel`의 리터럴 30 하나뿐이다.
