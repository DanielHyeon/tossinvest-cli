# Branch Test Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(4). worker 의 Cycle 클로저를 `c.productionStrategyCycle(clk, market)`(shadow 래퍼 — 성공 플래그 · 경과 판정 · 실패 폐기 defer · nil 뒤 비동기 시작)로 바꿨다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--context.newrefreshingpairedstrategyentrysupervisor/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 386:2 — Context · 시계 nil | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | if at 390:2 — 진입 관문 없음 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B3 | range at 394:2 — 시장 순회 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B4 | if at 406:2 — supervisor 생성 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
