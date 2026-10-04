# Branch Test Map: `strategyRiskAuthorityLoader.collectMarket`

- Source SHA-256: `bd5589d0c7d35de8294af8646be3502615fc0509d43a27c19931fe3fcc093d10`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 5 분기 → 6: 번들 하나 적재(편집 전 B4 적재 실패 · B5 범위 불일치 → 시장 AuthorityUnavailable)를 결과 권한의 범위마다(B4 순회 · B5 키 · B6 적재 성공)로 바꿨다. 준비된 범위가 없으면 `strategyRiskMarketFromScopes` 가 시장 전체를 AuthorityUnavailable 로 둔다(편집 전 사유 그대로).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyriskauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 192:2 — 결과 권한 준비 안 됨 → LaneNotReady | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 195:2 — 환율 준비 안 됨 → FXNotReady | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 199:2 — 시장이 US 면 버킷 시장 US | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | range at 205:2 — **(새)** 결과 권한의 범위마다 번들 하나(`result.results()` — 활성화 없는 시장은 하나) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(범위 번들 digest 둘) | yes — 변이 X12(첫 범위만) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B5 | if at 208:3 — **(새)** 범위 키 정규화 성공 시에만 적재 — 실패면 그 범위 AuthorityUnavailable | 진입 0(거짓 갈래) — 조립 결과는 항상 유효 키 | no | 진입 0 |
| B6 | if at 215:4 — 적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비(편집 전 B4 · B5 의 반대편) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` · `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal`(**a127 `82080177` 에서 제거됨** — 다리 제거 조건 이행, 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal`: 실원장 → 범위 준비)(실원장 → 전 범위 준비 안 됨) | yes — X12 | yes |
