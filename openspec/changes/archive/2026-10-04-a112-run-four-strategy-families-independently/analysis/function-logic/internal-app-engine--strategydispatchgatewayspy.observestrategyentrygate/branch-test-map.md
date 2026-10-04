# Branch Test Map: `strategyDispatchGatewaySpy.ObserveStrategyEntryGate`

- Source SHA-256: `b0b9734d75c5e4fafa2b2033bd9c3d9660af813aa7d485b1840d2a1f8ebcd958`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** a112 5.2.2.2: 종목 단위 진입 관문 거절(`failEntryGateSymbol`)을 더했다(B2) — worker 승격이 한 범위의 관문 거절로 다른 범위를 굶기지 않는지 재려고. 시장 단위 거절(B1)은 불변.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategydispatchgatewayspy.observestrategyentrygate/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 50:2 — 시장 단위 거절(편집 불변) | `TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` 등 기존 사용처 | no — 시험 코드 | yes |
| B2 | if at 53:2 — **(새)** 종목 단위 거절 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | no — 시험 코드 | yes |
