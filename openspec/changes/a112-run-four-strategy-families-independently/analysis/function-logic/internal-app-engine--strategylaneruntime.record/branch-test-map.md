# Branch Test Map: `record`

- Source SHA-256: `333970fa5e15db9741cc10da836922c79763f4b1b87bf62ddbc1af2fba9c6462`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(4). 파도 증가 바로 뒤 같은 잠금에서 `shadowCells[market] = {wave, batch, activation}` 을 덮어쓴다(⑤).
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--strategylaneruntime.record/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 347:2 — 관측 없음 → 물결 아님(칸도 그대로) | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | if at 352:2 — 파도 맵 지연 생성 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
| B3 | if at 355:2 — 파도 증가(포화) | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
| B4 | range at 359:2 — 관측 덮어쓰기 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
