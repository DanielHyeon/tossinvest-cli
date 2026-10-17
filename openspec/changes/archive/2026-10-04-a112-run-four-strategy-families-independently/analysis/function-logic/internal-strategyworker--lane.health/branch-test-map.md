# Branch Test Map: `Lane.Health`

- Source SHA-256: `b869921551c2058b80fa03d2bc9c601f3ea306234d918ee41cde019312d529ab`; AST branch locations are authoritative.
- Revision: **modified (a112 7.5, 2026-10-01).** 편집 전 2 분기 → 0: 판정 본문을 `healthLocked` 로 옮기고 Health 는 잠금 + 위임. Status 가 같은 잠금 안에서 같은 판정을 쓴다(판정 한 벌).
- 편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/internal-strategyworker--lane.health/`. 변이 원장 `analysis/measurements/lot-7.5/mutation-7.5.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — 잠금을 잡고 healthLocked 의 판정을 돌려줌(LATCHED · DEGRADED · HEALTHY — 판정은 healthLocked 로 이동, 값 불변) | `TestTheLaneStatusIsTheRowItsAccessorsRead` · `TestAnOrdinaryFailureCountsAndLatchesExactlyAtTheThreshold` · `TestEveryProductionLaneIsBornHealthyAndUnlatched` | no — 시험 코드 | yes |
