# Branch Test Map: `wtsPopular.Read`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | US 스캔은 클라이언트를 부르기 전에 거부 | `TestThePopularityRankingRefusesAMarketItCannotSee`(`f.gotSize != 0`까지 확인) | — (기존 동작) | yes |
| B2 | 읽기 실패는 래핑되고 기억은 그대로 | `TestTheWTSMemoryCarriesTheSameTwoConditions`(간접) + fake err 경로 | — (기존 동작) | yes |
| B3 | 행마다 요청 수와 3-상태 | `TestTheWTSPopularityListReportsTheSameTwoFacts` · `TestTheWTSMemoryCarriesTheSameTwoConditions` | yes | yes |
| B4 | Symbol도 ProductCode도 없는 행 | `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps` | yes (2026-07-28: `whole`을 옛 형태로 되돌리면 실패) | yes |

**2026-07-28 해소**: `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps`가 Symbol도
ProductCode도 없는 행을 만든다. 그 전까지는 `TestThePopularityRankingFallsBackToTheProductCode`가
ProductCode **있는** 경우만 만들었다.
