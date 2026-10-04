# Branch Test Map: `cloneLanes`

- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(2). shadowOutcome 포인터도 깊은 복사한다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-strategyprojection--clonelanes/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 435:2 — nil | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | range at 439:2 — 레인 순회 | `TestCloneCarriesRuntimeIdentityWithoutSharingIt`, `TestCloneDeepCopiesLaneAndCoordinatorChildren`, `TestStorePublishesOnlyValidatedImmutablePairedSnapshots`, `TestTheShadowVocabularyAndTheWireShape`, `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 5개가 진입(측정); 패키지 합집합 진입=True |
