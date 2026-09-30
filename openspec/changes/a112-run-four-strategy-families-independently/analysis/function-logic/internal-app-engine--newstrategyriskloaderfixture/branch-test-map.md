# Branch Test Map: `newStrategyRiskLoaderFixture`

- Source SHA-256: `bb13286b7859e266ec7b0a3717d866e9627e758b8f3a8cbf43dd1a56898480cd`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** a112 5.2.2.2: 본문을 `newStrategyRiskLoaderFixtureWith(t, nil)` 위임 한 줄로 바꿨다 — KR 서명 위험 정책에 종목을 더할 수 있는 판을 새 함수로 떼고, 기존 호출자는 같은 fixture 를 받는다(extraKR 이 nil 이면 편집 전과 같은 매니페스트). 분기 없음.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--newstrategyriskloaderfixture/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — `newStrategyRiskLoaderFixtureWith(t, nil)` 위임 | `TestStrategyRiskAuthorityLoaderPairedKRUSSameWave` · `TestStrategyRiskAuthorityLoaderPreservesPeerOnMarketFailure` 등 기존 호출자 전부(엔진 태그 스위트 PASS) | no — 시험 코드 | yes |
