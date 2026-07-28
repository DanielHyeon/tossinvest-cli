# Branch Test Map: `wtsPopular.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | US 스캔은 클라이언트를 부르기 전에 거부 | `TestThePopularityRankingRefusesAMarketItCannotSee`(`f.gotSize != 0`까지 확인) | — (기존 동작) | yes |
| B2 | 읽기 실패는 래핑되고 기억은 그대로 | `TestTheWTSMemoryCarriesTheSameTwoConditions`(간접) + fake err 경로 | — (기존 동작) | yes |
| B3 | 행마다 요청 수와 3-상태 | `TestTheWTSPopularityListReportsTheSameTwoFacts` · `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes | yes |
| B4 | Symbol도 ProductCode도 없는 행 | **커버 없음** | no | no |

**정직한 커버리지 기록**: B4는 이 change 이전부터 있던 분기다. `TestThePopularityRankingFallsBackToTheProductCode`는
ProductCode **있는** 경우만 만든다. 둘 다 빈 행을 만드는 fixture는 없다.
