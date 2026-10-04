# Branch Test Map: `evaluate`

- Source SHA-256: `0935960f05276aa2bf972734943b879a5c56b3525f2083aeed9f5b2e3855c620`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(8). 6번째 인자 shadow 를 받아 record 로 넘기기만 한다(④ — 호출 · 순회 0).
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--strategylaneruntime.evaluate/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 221:2 — 런타임 nil | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | if at 227:2 — 복구 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B3 | range at 247:2 — 레인 순회(동시) | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
| B4 | range at 249:3 — 입력 찾기 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
| B5 | if at 250:4 — 레인 소유 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B6 | range at 263:2 — panic 수거 순회 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 7 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 13개가 진입(측정); 패키지 합집합 진입=True |
| B7 | if at 264:3 — panic 재던짐 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B8 | if at 271:2 — latch 저장 오류 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
