# Branch Test Map: `strategyProposalAuthorityLoader.collectMarket`

- Source SHA-256: `a356e5ead7d719e7b791423645a86b2a2f8b2eed26066127eafb1928ec411288`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 16 분기 → 15: 제안 집합 digest 를 손으로 적던 순회(편집 전 B16 `range entries`)를 지우고 A-lite 계약과 같은 함수 `strategyProposalSetDigest(entries)` 를 부른다(A#6 — digest 식 단일 출처). B1~B15 는 편집 전과 같은 분기 · 좌표 이동만.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyproposalauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 314:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 317:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 320:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 325:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 329:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | range at 335:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B7 | if at 337:3 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B8 | if at 349:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | if at 354:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | if at 385:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 392:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 401:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B13 | if at 409:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B14 | if at 419:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B15 | if at 427:2 — 편집 전과 같은 분기(좌표만) | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
