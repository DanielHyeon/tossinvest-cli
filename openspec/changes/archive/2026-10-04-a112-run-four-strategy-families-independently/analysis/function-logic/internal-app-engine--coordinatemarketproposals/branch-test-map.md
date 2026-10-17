# Branch Test Map: `coordinateMarketProposals`

- Source SHA-256: `590318d7267c1ffd5398c4c9878b9e70486633dc42033e19466e3474cf6da01b`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(8). 반환에 shadow 묶음을 더했다: 시작에 `shadow := strategyShadowBatch{observed: true}`, 안쪽 루프 `gate.admit` 앞에 수집 helper 문장 하나, 계보 충돌 return 은 부재 값, 정상 return 만 수집 묶음(브리프 v3.3 §4 ⓐⓑⓒ).
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--coordinatemarketproposals/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 66:2 — 경로(종목) 순회 | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 11 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 17개가 진입(측정); 패키지 합집합 진입=True |
| B2 | if at 68:3 — 레인 없는 종목 → 거절 계수 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B3 | range at 77:3 — 그 종목의 레인 제안 순회 — **편집: admit 앞에서 shadow.collect** | `TestAFailedCycleDiscardsTheShadowAtOnceAndALateStepCannotRevive`, `TestAFailingShadowStepChangesNothingButItsOwnObservation`, `TestALateNilCycleIsAFailureAtTheLimit`, `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`, `TestARestartNeverRestoresShadow`, `TestAStuckShadowStepNeitherLatchesNorRefreshesTheOtherMarket` 외 11 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 17개가 진입(측정); 패키지 합집합 진입=True |
| B4 | if at 96:4 — 관문이 멈춤 → gated 기록 | `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination`, `TestTheCoordinatorCarriesEveryPreGateProposalAsTheShadowBatch` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 2개가 진입(측정); 패키지 합집합 진입=True |
| B5 | if at 101:4 — 관문 EMITTED → 레인 봉투 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B6 | if at 105:4 — 조정자 Submit 거절 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B7 | if at 109:4 — 계보 신원 충돌 → **부재 값**으로 반환(편집) | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B8 | if at 124:3 — 범위 전부 관문에 빼앗김 → erasedScopes | `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination`, `TestTheCoordinatorCarriesEveryPreGateProposalAsTheShadowBatch` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 2개가 진입(측정); 패키지 합집합 진입=True |
