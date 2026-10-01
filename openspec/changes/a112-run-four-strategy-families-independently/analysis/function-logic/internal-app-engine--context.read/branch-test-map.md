# Branch Test Map: `Context.Read`

- Source SHA-256: `4302edefe72942bd1f4f7f4aa51b3c03e26ef97c13c3d8d52a0b6be94a2eef9f`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 편집 전 7 분기 → 8: runtime identity 뒤 · 감독자 부재 검사 앞에 B4(레인 런타임이 있으면 여덟 레인을 지금 상태로 덧씌움)를 더했다. 나머지 일곱은 번호만 밀렸다(편집 전 B4~B7 → B5~B8).
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-app-engine--context.read/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 25:2 — nil receiver/context → 오류 | `TestStrategyRuntimeReadWithoutAStoreStaysAnError` | no — 이 로트가 바꾸지 않음 | yes |
| B2 | if at 31:2 — store 부재 → 오류 | `TestStrategyRuntimeReadWithoutAStoreStaysAnError` | no — 이 로트가 바꾸지 않음 | yes |
| B3 | if at 35:2 — store 읽기 오류 → 그대로 | `TestStrategyRuntimeReadOnAFailedStoreInventsNothing` | no — 이 로트가 바꾸지 않음 | yes |
| B4 | if at 56:2 — **(새)** 레인 런타임 있음 → `lanes = runtime.projection()`(읽기만), 없으면 저장소의 미관측 기본값 | `a112_lane_coordinator_projection_test.go` `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure` · `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`(있음) · `TestAProcessWithoutLanesProjectsTheEightUnobservedDefaults`(없음) | yes — 편집 전 컴파일 실패(`red-7.3.log`), 변이 P03 CAUGHT(`analysis/measurements/lot-7.3/mutation-7.3.tsv`) | yes |
| B5 | if at 59:2 — 감독자 부재 → 반환(편집 전 B4) | `TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest` | no — 이 로트가 바꾸지 않음 | yes |
| B6 | range at 62:2 — KR · US latch 검사(편집 전 B5) | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` | no — 이 로트가 바꾸지 않음 | yes |
| B7 | if at 64:3 — latch 안 된 시장 건너뛰기(편집 전 B6) | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` | no — 이 로트가 바꾸지 않음 | yes |
| B8 | if at 68:3 — CURRENT 시장만 덮기(편집 전 B7) | `TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket` | no — 이 로트가 바꾸지 않음 | yes |
