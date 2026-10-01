# Branch Test Map: `TestOnlyThePackageLevelStepEverRunsInsideALane (시험)`

- Source SHA-256: `406bd29b9c2ce881d31fe69f638c9f6aee89a442e0adbcd709d56aac4f6ab98d`; AST branch locations are authoritative.
- Revision: **modified (a112 7.5, 2026-10-01).** 자리 철자를 `runtime.laneStepFor(lane, promotion)` 로 바꾸고, laneStepFor 정의 전부를 세는 대조(정의 둘 · 생산 본문 한 줄 · 태그)를 `assertLaneStepForIsTheProductionStepOutsideTestSeams` 로 더했다 — 핀 강화(Manager 판정 (A)).
- 편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/internal-app-engine--testonlythepackagelevelstepeverrunsinsidealane/`. 변이 원장 `analysis/measurements/lot-7.5/mutation-7.5.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 195:2 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
| B2 | range at 197:3 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
| B3 | if at 200:5 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
| B4 | if at 204:5 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
| B5 | if at 207:5 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
| B6 | if at 227:2 — RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조 | 이 시험 자신 | no — 시험 코드 | yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`) |
