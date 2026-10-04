# Branch Test Map: `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening (시험)`

- Source SHA-256: `c1b6888e6bccd1b9ca9e54c43abbae8063f5e9cf12178e98b9170634d4c73979`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** a112 7.3.1 서명 변경(evaluate 6번째 인자 · collect 두 반환값 · collectMarket out 인자)에 맞춘 호출 수정 — 이 시험의 판정 · 단언은 바뀌지 않았다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--testadeclaredactivationthatlapsesrollsitsmarketbackinsteadofwidening/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 129:2 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B2 | if at 178:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B3 | if at 181:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B4 | if at 184:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B5 | switch at 190:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B6 | case at 191:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B7 | if at 192:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B8 | if at 195:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B9 | case at 198:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B10 | if at 199:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B11 | case at 202:4 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B12 | if at 203:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B13 | if at 207:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B14 | if at 211:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B15 | if at 215:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
| B16 | if at 218:5 — 시험 자신의 갈래 | 이 시험 자신 | no — 시험 코드 | yes — 엔진 태그 스위트 PASS |
