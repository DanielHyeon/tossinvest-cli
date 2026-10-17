# Branch Test Map: `NewPairedStrategyEntryProductionAssembly`

- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(8). collect 의 둘째 값을 `shadowAuthority` 로 받아 조립 리터럴의 별개 필드 `shadow` 에 담는다 — dispatch · worker · 결과 권한에는 넘기지 않는다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--context.newpairedstrategyentryproductionassembly/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 297:2 — Context nil | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | if at 303:2 — 원장 경로 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B3 | if at 314:2 — evidence 경로 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B4 | if at 324:2 — 레인 런타임 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B5 | if at 345:2 — 시계 → dispatch now | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B6 | range at 361:2 — 시장 worker 순회 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B7 | if at 366:2 — supervisor 생성 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B8 | if at 373:2 — 투영 발행 오류 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
