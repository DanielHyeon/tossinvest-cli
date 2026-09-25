# Branch Test Map: `Console.buildMultiMarketStrategyRuntimePage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 부재 신호 false → dormant, true/신호 없음 → Read | `TestAnUnconfiguredWrapperStillRendersDormant` · `TestStrategyRuntimeDormantPairIsHonest` | no | no |
| B2 | 구성됐으나 못 읽음 → 도달 불가 | `TestAConfiguredButUnreachableWrapperRendersUnavailable` | no | no |
| B3 | 유효 스냅샷 | `TestStrategyRuntimeMarketsRenderIndependently` | no | no |
