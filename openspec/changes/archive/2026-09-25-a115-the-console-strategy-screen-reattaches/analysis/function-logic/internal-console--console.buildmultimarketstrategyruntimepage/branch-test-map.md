# Branch Test Map: `Console.buildMultiMarketStrategyRuntimePage`

구현 후 AST 기준(분기 3, 번호 불변 — B1 조건만 교체).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 부재 신호 false → dormant(Read 0), true/신호 없음 → Read | `TestAnUnconfiguredWrapperStillRendersDormant` · `TestASignallessReaderIsStillWired` · `TestStrategyRuntimeDormantPairIsHonest` | yes — base 에서 부재 wrapper 가 live 로 그려짐·Read 1·질문 0 · 변이 K3·K5·K16 CAUGHT | yes |
| B2 | 구성됐으나 못 읽음 → 도달 불가(NOT_CONFIGURED 아님) | `TestAConfiguredButUnreachableWrapperRendersUnavailable` · `TestStrategyRuntimeReaderFailureAndInvalidProjectionFailClosedWithoutLeakingError` | yes — base 에서 질문 0 | yes |
| B3 | 유효 스냅샷 | `TestStrategyRuntimeMarketsRenderIndependently` · `TestASignallessReaderIsStillWired` | no — 무변경 분기(회귀 핀) | yes |
