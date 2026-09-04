# Branch Test Map: `refreshPairedStrategyEntryProductionAssembly`

- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`; AST branch locations are authoritative.
- L5 5.2.1 이 이 함수 본문을 바꿨다. 아래 RED 는 편집 **전** 소스에서 실제로 실행해
  얻은 결과이고, GREEN 은 편집 뒤의 결과다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 526:2 — `c` 또는 시계가 없다 | `TestARefreshWithoutAClockRefusesBeforeItCanMintAWave` | n/a (편집 전과 동작 같음; 이 로트가 처음 시험을 붙였다) | yes |
| B2 | if at 534:2 — 창 안의 캐시 | `TestTheCacheWindowStillMeasuresFromTheWaveStart`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` | n/a (창 의미는 바꾸지 않았다; 이 로트가 처음 시험을 붙였다) | yes |
| B3 | if at 537:2 — 지도자가 아니면 파도에 합류 | `TestTwoMarketsRideOneAuthorityWaveInsteadOfTakingTurns`, `TestAFailedWaveReachesEveryMarketAndIsNeverCached`, `TestAMarketWaitingOnTheWaveLeavesWhenItsOwnCycleIsCancelled`, `TestAPanickingWaveNeverStrandsTheMarketsWaitingOnIt`, `TestAFailedWaveCarriesNoAssemblyForEitherMarket` | **no — 정직하게 적는다: 이 갈래도 그것을 재는 시험도 편집과 함께 생겼다. 편집 전 소스에서는 컴파일되지 않으므로 "빨갛다"가 아니라 "없다"이다. 이 갈래의 이빨은 RED 가 아니라 반증 배터리가 지운다(M1·M11·M12) — 셋 다 CAUGHT** | yes |
| 본문 | 511:9 — 지도자가 잠금 **밖에서** 수집한다 | `TestTheRemoteAuthorityWaveNeverRunsUnderTheSharedAssemblyMutex`, `TestExactlyOneFunctionRunsTheRemoteAuthorityWave`, `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` | **yes — 두 셈 시험 모두 편집 전 소스에서 실패했다(기록: review.md 5.2.1 절)** | yes |

RED 는 손으로 옮겨 적은 것이 아니라 편집 전 `go test` 출력이다: 셈 시험은
`Context.refreshPairedStrategyEntryProductionAssembly 가 공유 assembly mutex 를 들고
원격 권한 파도를 돌린다` 로 실패했고, 자리 셈은 `got: …refreshPaired…` 로 실패했다.
