# Branch Test Map: `projection`

- Source SHA-256: `713fd68268cb36340d9c990c2d213c56a3b9d72c1342f83e2b845f09254978ef`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 편집 전 2 분기 → 3: 런타임 시계로 `shadowObservationUsable`(파도 등식 ∧ 미만료 ∧ 나이 상한)을 판정해 쓸 수 있는 shadow 관측을 넘긴다(B3 새).
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--strategylaneruntime.projection/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 33:2 — 런타임 nil | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | range at 42:2 — 레인 순회 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestTheGapBetweenRecordAndPublishIsUnobservedByDesign` 외 4 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 10개가 진입(측정); 패키지 합집합 진입=True |
| B3 | if at 47:3 — **(새)** 쓸 수 있는 shadow 관측 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestTheGapBetweenRecordAndPublishIsUnobservedByDesign`, `TestTheNextWaveClearsAnUnusableShadow` 외 3 | yes — `analysis/measurements/lot-7.3.1-shadow/red-7.3.1-shadow.log`(편집 전 기호 없음 · 컴파일 RED) | yes — shadow 시험 9개가 진입(측정); 패키지 합집합 진입=True |
