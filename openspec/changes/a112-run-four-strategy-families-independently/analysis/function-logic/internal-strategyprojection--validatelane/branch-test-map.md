# Branch Test Map: `validateLane`

- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 편집 전 15 분기 → 17: runtime 어휘 {UNOBSERVED, SHADOW}(B1 조건), SHADOW 교차 규칙(B2 새) · shadowOutcome 짝(B3 새). 이후 번호 둘씩 밀림.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-strategyprojection--validatelane/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 299:2 — desired/effective/runtime 어휘 | `TestValidateRefusesMalformedLaneAndCoordinatorChildren`, `TestValidateRefusesShadowOutsideTheOneCrossRule` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 2개가 진입(측정); 패키지 합집합 진입=True |
| B2 | if at 305:2 — **(새)** SHADOW ⇒ OFF/OFF ∧ health ∧ cycleGeneration>0 | `TestValidateRefusesShadowOutsideTheOneCrossRule` | yes — `analysis/measurements/lot-7.3.1-shadow/red-7.3.1-shadow.log`(편집 전 기호 없음 · 컴파일 RED) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B3 | if at 308:2 — **(새)** shadowOutcome ⇔ SHADOW | `TestValidateRefusesShadowOutsideTheOneCrossRule` | yes — `analysis/measurements/lot-7.3.1-shadow/red-7.3.1-shadow.log`(편집 전 기호 없음 · 컴파일 RED) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B4 | if at 313:2 — 거절 코드 짝 | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B5 | if at 316:2 — 미관측 | `TestStorePublishesOnlyValidatedImmutablePairedSnapshots`, `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 2개가 진입(측정); 패키지 합집합 진입=True |
| B6 | if at 318:3 — 미관측 사실 추론 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B7 | if at 327:2 — 건강 어휘 | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B8 | if at 330:2 — 정책 | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B9 | if at 333:2 — 정규 값 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B10 | if at 339:2 — 물결 ⇔ 트리거 | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B11 | if at 342:2 — 트리거 없음 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B12 | if at 343:3 — 트리거 없는 사실 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B13 | if at 349:2 — 트리거 어휘 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B14 | if at 352:2 — 시작 ⇔ ENQUEUED | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B15 | if at 355:2 — 시작 어휘 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B16 | if at 359:2 — 결과 ⇔ ADMITTED | `TestValidateRefusesMalformedLaneAndCoordinatorChildren` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B17 | if at 362:2 — 결과 어휘 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
