# Branch Test Map: `strategyRouteAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go`; file SHA-256 `85e97cf96bd416ee52555042478ac3c184d745e61e7bb5a25aaf28128afe8fb9`. AST branch positions are authoritative.
- Rows carry measured counts. engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 148:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if at 151:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if at 154:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if at 157:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if at 162:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | for at 171:2 | arm entered 8x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B7 | if at 173:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if at 187:2 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal` |
| B9 | range at 192:2 | arm entered 7x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B10 | if at 194:3 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B11 | if at 202:3 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B12 | if at 210:2 | arm never entered: count 0 in every profile measured for this function |
| B13 | range at 215:2 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.

> 재기준화: a112 게이트 준비(2026-10-04): a127 82080177 이 이 함수 본문을 편집(분기 · return 구조 동일, 호출 · 줄 이동) — 좌표 · 호출 표는 새 AST, 분기 의미 · a112 시험 인용은 그대로. a127 편집의 증거는 아카이브 번들 `openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/function-logic/internal-app-engine--strategyrouteauthorityloader.collectmarket/`(편집 뒤 FLM/BTM · a127 시험) — 분기 (id · 종류) · return 수 동일, 좌표는 id 끼리 사상, 호출 표는 새 AST.
