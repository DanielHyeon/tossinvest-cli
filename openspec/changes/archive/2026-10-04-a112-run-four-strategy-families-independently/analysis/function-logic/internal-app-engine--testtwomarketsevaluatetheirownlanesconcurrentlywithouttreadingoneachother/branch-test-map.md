# Branch Test Map: `TestTwoMarketsEvaluateTheirOwnLanesConcurrentlyWithoutTreadingOnEachOther (시험)`

- Source SHA-256: `bf27865d0f81e8a13f1443fa18f697676ed951c8054bffb5f2355da7aa1295de`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** a112 7.3.1 서명 변경(evaluate 6번째 인자 · collect 두 반환값 · collectMarket out 인자)에 맞춘 호출 수정 — 이 시험의 판정 · 단언은 바뀌지 않았다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--testtwomarketsevaluatetheirownlanesconcurrentlywithouttreadingoneachother/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 401:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B2 | range at 408:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B3 | for at 413:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B4 | if at 416:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B5 | range at 426:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B6 | if at 429:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B7 | if at 432:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
