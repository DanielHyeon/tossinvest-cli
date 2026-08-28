# Branch Test Map: `strategyRouteAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go`; file SHA-256 `2c6b43decf3a2706bb7a6d0d71428587612c5a011aac086d7c058220bb78fb98`. AST branch positions are authoritative.
- Rows carry measured counts. engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Tests whose individual coverage profile entered at least one arm: `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal`.

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 147:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if at 150:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if at 153:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if at 156:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if at 161:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | for at 170:2 | arm entered 8x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B7 | if at 172:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if at 184:2 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal` |
| B9 | range at 189:2 | arm entered 7x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B10 | if at 191:3 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B11 | if at 199:3 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B12 | if at 207:2 | arm never entered: count 0 in every profile measured for this function |
| B13 | range at 212:2 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
