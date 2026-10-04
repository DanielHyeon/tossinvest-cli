# Branch Test Map: `validProductionRiskPolicyContents`

- Source SHA-256: `9a74db4abd523da823e02e716258d428d3f385647c3d18121e744f9fac45119e`; AST branch locations are authoritative.
- Revision: **modified (a112 6.1, 2026-10-01).** a112 6.1(Manager 판정 (C)): strategy 항목 순회 안에 B9 를 더했다 — 레인의 family 를 strategyrouter 정본 표(`ProductionLaneFamily`)에서 유도하고, 해소되지 않는 레인 또는 한 risk_id 를 두 family 가 공유하면 false(정책 전체 거절 → 시장 위험 미준비, 결함 등급). B10~B14 는 편집 전 B9~B13(번호 이동).
- 편집 전 번들: `analysis/measurements/lot-6.1/pre-edit/internal-riskbucket--validproductionriskpolicycontents/`. 변이 원장 `analysis/measurements/lot-6.1/mutation-6.1.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 268:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B2 | range at 271:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B3 | if at 272:3 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B4 | range at 276:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B5 | if at 277:3 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B6 | if at 281:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B7 | range at 288:2 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B8 | if at 290:3 — 편집 전과 같은 분기 | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B9 | if at 296:3 — **(새)** family 해소 불가 레인 · 한 risk_id 를 두 family 가 공유 → false(정책 거절 — `ErrProductionRiskSnapshotUnavailable`, 범위 국소 아님) | `internal/riskbucket/a112_family_risk_binding_test.go` `TestARiskIDSharedByTwoFamiliesRefusesThePolicy`(대조: 다른 risk_id 면 수락) · `TestAStrategyEntryWhoseLaneHasNoFamilyRefusesThePolicy` · `TestTheFamilyIsResolvedInThePolicysOwnMarket` | yes — `analysis/measurements/lot-6.1/red-6.1.log`(편집 전 err=nil) · 변이 V01 · V02 · V03 · V04 CAUGHT(`analysis/measurements/lot-6.1/mutation-6.1.tsv`) | yes |
| B10 | if at 300:3 — 편집 전과 같은 분기(번호 이동) | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B11 | range at 306:2 — 편집 전과 같은 분기(번호 이동) | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B12 | if at 307:3 — 편집 전과 같은 분기(번호 이동) | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B13 | if at 310:3 — 편집 전과 같은 분기(번호 이동) | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
| B14 | if at 313:3 — 편집 전과 같은 분기(번호 이동) | 분기 불변 — 편집 전 번들(`lot-6.1/pre-edit/`) | no — 이 로트가 바꾸지 않음 | 편집 전 측정 |
