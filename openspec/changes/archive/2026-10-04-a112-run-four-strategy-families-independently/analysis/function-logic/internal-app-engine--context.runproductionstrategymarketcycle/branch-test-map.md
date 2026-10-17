# Branch Test Map: `runProductionStrategyMarketCycle`

- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(4). evaluate 에 인자 index 5 `fresh.shadow.forMarket(market)` 를 더했다(Manager 판정 (A)). 마지막 문장 dispatch · Args[2] 불변.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--context.runproductionstrategymarketcycle/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 518:2 — refresh 오류 | `TestTheShadowDeferNeverRewritesAFirstCycleError` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B2 | if at 542:2 — 레인 런타임 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B3 | if at 548:2 — evaluate 오류(durable latch) | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B4 | if at 555:2 — dispatch 없는 조립 → nil | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestNeitherTheStepNorTheProjectionShadowsAnOnLane`, `TestOnlyOneShadowStepRunsPerMarket`, `TestTheGapBetweenRecordAndPublishIsUnobservedByDesign` 외 2 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 8개가 진입(측정); 패키지 합집합 진입=True |
