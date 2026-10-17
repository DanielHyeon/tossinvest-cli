# Branch Test Map: `productionStrategyWorker`

- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(1). 마지막 인자(cycle)를 `c.productionStrategyCycle(clk, market)` 로 바꿨다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--context.productionstrategyworker/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 420:2 — Context nil → 빈 worker | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
