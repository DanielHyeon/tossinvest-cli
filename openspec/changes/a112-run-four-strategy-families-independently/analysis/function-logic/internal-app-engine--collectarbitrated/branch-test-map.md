# Branch Test Map: `collectArbitrated (시험)`

- Source SHA-256: `44c3a7a4270c2c5b5e965da58dadff94b51df649d3aaf90d786141b1b189bb78`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** a112 7.3.1 서명 변경(evaluate 6번째 인자 · collect 두 반환값 · collectMarket out 인자)에 맞춘 호출 수정 — 이 시험의 판정 · 단언은 바뀌지 않았다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--collectarbitrated/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — 분기 없음 — 시험 본문 | 이 시험 자신(엔진 태그 스위트 PASS) | no — 시험 코드 | yes |
