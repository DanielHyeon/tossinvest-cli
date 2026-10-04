# Branch Test Map: `strategyAccountAuthorityLoader.collectMarket`

- Source SHA-256: `d0d6281292dafcc979edce741a3a2bf98ed348f023267d8198d2436c71ec7291`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** 편집 전 4 분기 → 6: 편집 전 B1(항목 정확히 하나 · 유효)을 활성화 여부로 가르고(B1), 계좌 권한 적재를 **항목(범위)마다** 한다(B4 순회 · B5 범위 유효 · B6 적재 성공). 편집 전 B4(적재 실패 → 시장 전체 AuthorityUnavailable)는 범위별 준비 안 됨으로 바뀌고, 준비된 범위가 없으면 `strategyAccountMarketFromScopes` 가 첫 범위의 사유로 시장 전체를 준비 안 됨으로 둔다(범위 하나면 편집 전과 같은 사유).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--strategyaccountauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 161:2 — 항목 없음, 또는 **활성화 없는 시장**에서 항목이 정확히 하나가 아니거나 무효 → `StrategyAccountProposalNotReady`(토글 OFF = upstream) | `TestProductionFirstLegAuthorityLoaderPairedKRUS`(시장당 하나 — 오늘 경로) | no — 활성화 없는 시장 동작 불변 | yes |
| B2 | if at 164:2 — loader 구성 불완전 → `StrategyAccountInternalFailure` | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 168:2 — 시장이 US 면 계좌 시장 US | 분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | range at 172:2 — **(새)** 항목(소유자 범위)마다 계좌 권한 적재 — 적재 종목은 `entries[0]` 이 아니라 **그 범위의 종목**(A#5) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` · `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` | yes — 변이 X14(첫 항목 종목) CAUGHT(`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`) | yes |
| B5 | if at 176:3 — **(새)** 범위 키 정규화 실패 또는 무효 제안 → 그 범위만 `ProposalNotReady` | 진입 0 — 조립이 무효 제안을 항목에 싣지 않음(시험 seam 으로 못 만듦) | no | 진입 0 |
| B6 | if at 181:4 — 적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비(편집 전 B4 의 반대편); 실패는 그 범위만 `AuthorityUnavailable` | `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`(005930 적재 실패 → 000660 만 거래) | yes — 변이 X09 · X14 CAUGHT | yes |
