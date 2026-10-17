# Branch Test Map: `newStrategyLaneRuntime`

- Source SHA-256: `0935960f05276aa2bf972734943b879a5c56b3525f2083aeed9f5b2e3855c620`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(1). shadow 맵 다섯(cells · epochs · observed · inFlight · skipped)을 생성자에서 만든다 — nil 맵 대입 panic 이 주기 경로의 오류를 덮지 않게(v3.3 N5).
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--newstrategylaneruntime/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 106:2 — 시계 nil → nil 런타임 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
