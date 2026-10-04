# Branch Test Map: `TestStrategyProposalAuthorityKeepsMarketFailureLocal (시험)`

- Source SHA-256: `d05d11be42f3538802baec08718c10782cea1007f608bac95ad206b16a9b57b0`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** a112 7.3.1 서명 변경(evaluate 6번째 인자 · collect 두 반환값 · collectMarket out 인자)에 맞춘 호출 수정 — 이 시험의 판정 · 단언은 바뀌지 않았다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--teststrategyproposalauthoritykeepsmarketfailurelocal/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 53:3 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B2 | if at 59:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B3 | if at 62:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
