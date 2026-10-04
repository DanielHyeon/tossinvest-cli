# Branch Test Map: `Read`

- Source SHA-256: `c9cdc65328bc621f04b109f103d4111d774a877c96565ec5d7cb1b3ee7cd60dc`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 편집 전 8 분기 → 9: supervisor 가 그 시장 평가를 abandon 으로 기록했으면 그 시장 레인의 SHADOW 를 지운다(R1 두 시계 — B7 새). 이후 갈래 번호가 하나씩 밀렸다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--context.read/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 25:2 — Context · ctx nil | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B2 | if at 31:2 — 투영 저장소 없음 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B3 | if at 35:2 — 저장소 읽기 오류 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B4 | if at 56:2 — 레인 런타임 있음 → 레인 덧씌움 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestTheGapBetweenRecordAndPublishIsUnobservedByDesign` 외 4 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 10개가 진입(측정); 패키지 합집합 진입=True |
| B5 | if at 59:2 — supervisor 없음 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestTheGapBetweenRecordAndPublishIsUnobservedByDesign` 외 4 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 10개가 진입(측정); 패키지 합집합 진입=True |
| B6 | range at 62:2 — 시장 순회 | `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B7 | if at 66:3 — **(새)** abandon 기록 시장 → SHADOW 지움 | `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow` | yes — `analysis/measurements/lot-7.3.1-shadow/red-7.3.1-shadow.log`(편집 전 기호 없음 · 컴파일 RED) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B8 | if at 69:3 — worker 없음 · 잠기지 않음(편집 전 B7) | `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B9 | if at 73:3 — 현재 시장 레코드 → 실패 표시(편집 전 B8) | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
