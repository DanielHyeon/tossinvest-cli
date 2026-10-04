# Branch Test Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source SHA-256: `2c546898bc4178d0ee44238d51fb8e05a4c908cee721041ee24713c4c98ac385`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 1 분기 → 4: 시장 단위 handoff 하나(`dispatchHandoff().Single()`) 대신 주문 경로와 같은 목록(`dispatchHandoffs`)을 순회하고(B1), 하나라도 거절 · 무효면 시장 준비 안 됨(B2 — 편집 전 B1), 목록이 비면 준비 안 됨(B3), 범위가 둘 이상이거나 활성화 시장이면 범위별 결과를 싣는다(B4).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyproposalauthoritypair.resultauthority/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 192:3 — **(새)** handoff 목록 순회 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` | yes — 변이 X13 CAUGHT | yes |
| B2 | if at 194:4 — handoff 거절 또는 무효 제안 → 시장 결과 권한 준비 안 됨(편집 전 B1 — 목록의 일부만 넘기지 않음) | `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` | no — 동작 보존 | yes |
| B3 | if at 199:3 — **(새)** 목록 없음 → 준비 안 됨 | 진입 0 — dispatchHandoffs 는 항상 하나 이상을 돌려준다(거절 handoff 포함) | no | 진입 0 |
| B4 | if at 203:3 — **(새)** 범위 둘 이상 또는 활성화 시장 → 범위별 결과(`scoped`)를 싣는다 — 위험 적재기가 범위마다 번들을 만든다 | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` | yes — 변이 X13(`if false`) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
