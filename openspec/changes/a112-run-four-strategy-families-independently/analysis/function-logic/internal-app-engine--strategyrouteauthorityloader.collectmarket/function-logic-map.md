# Function Logic Map: `strategyRouteAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go` (140-218)
- Function: `strategyRouteAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyRouteAuthorityLoader.collectMarket(params=4, results=1)`
- File SHA-256: `2c6b43decf3a2706bb7a6d0d71428587612c5a011aac086d7c058220bb78fb98`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 13.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Loads one market's signed route authority and turns each approved candidate into a routed scope. Task 4.3.1 replaced the pre-evaluation single-winner `strategyrouter.Route` with `strategyrouter.RouteSet` here, so every eligible family reaches the coordinator instead of one winner chosen before evaluation. The market argument selects the manifest; nothing else crosses in.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 145:3, 148:3, 151:3, 154:3, 157:3, 162:3, 173:4, 185:3, 206:44, 208:3, 215:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 147:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if | 150:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if | 153:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if | 156:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if | 161:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | for | 170:2 | arm entered 8x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B7 | if | 172:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if | 184:2 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal` |
| B9 | range | 189:2 | arm entered 7x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B10 | if | 191:3 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B11 | if | 199:3 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B12 | if | 207:2 | arm never entered: count 0 in every profile measured for this function |
| B13 | range | 212:2 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fail` | 148:10 |
| `fail` | 151:10 |
| `candidates.approved.Len` | 153:5 |
| `fail` | 154:10 |
| `fail` | 157:10 |
| `strings.TrimSpace` | 159:13 |
| `loader.getenv` | 159:31 |
| `DecodeString` | 160:14 |
| `base64.StdEncoding.Strict` | 160:14 |
| `base64.StdEncoding.EncodeToString` | 161:19 |
| `len` | 161:72 |
| `fail` | 162:10 |
| `strings.TrimSpace` | 164:12 |
| `loader.getenv` | 164:30 |
| `strings.TrimSpace` | 165:11 |
| `loader.getenv` | 165:29 |
| `make` | 166:13 |
| `candidates.approved.Len` | 166:52 |
| `make` | 167:10 |
| `candidates.approved.Len` | 167:32 |
| `make` | 168:13 |
| `candidates.approved.Len` | 168:61 |
| `make` | 169:20 |
| `candidates.approved.Len` | 169:57 |
| `candidates.approved.Len` | 170:26 |
| `candidates.approved.At` | 171:19 |
| `approved.Valid` | 172:14 |
| `approved.Market` | 172:34 |
| `string` | 172:55 |
| `approved.Symbol` | 172:78 |
| `fail` | 173:11 |
| `approved.Symbol` | 175:8 |
| `append` | 176:13 |
| `approved.Symbol` | 176:74 |
| `append` | 177:20 |
| `loader.load` | 179:16 |
| `strategyRouterMarket` | 180:42 |
| `ed25519.PublicKey` | 181:36 |
| `batch.ManifestDigest` | 184:19 |
| `candidates.approved.Len` | 186:58 |
| `candidates.approved.Len` | 186:99 |
| `batch.For` | 190:20 |
| `approved.Symbol` | 190:30 |
| `authority.Request` | 195:14 |
| `strategyrouter.RouteSet` | 198:13 |
| `routed.Valid` | 199:52 |
| `len` | 199:70 |
| `strategyRouterMarket` | 200:26 |
| `approved.Symbol` | 200:80 |
| `append` | 204:13 |
| `sort.Slice` | 206:2 |
| `entries.approved.Symbol` | 206:51 |
| `entries.approved.Symbol` | 206:82 |
| `len` | 207:5 |
| `candidates.approved.Len` | 209:58 |
| `sha256.New` | 211:7 |
| `h.Write` | 213:10 |
| `(unnamed)` | 213:18 |
| `entry.approved.Symbol` | 213:25 |
| `entry.route.OwnerDigest` | 213:60 |
| `candidates.approved.Len` | 216:113 |
| `len` | 217:17 |
| `hex.EncodeToString` | 217:106 |
| `h.Sum` | 217:125 |

## State mutations and fallbacks

- AST assignments: 26. Defers: 0. Goroutine statements: 0.
- Appends to this market's own result slice only. No journal write, no broker call, no shared mutable state — the paired KR/US loaders are independent by construction.

## Safety conclusion

- Read-only over an already-signed manifest. A refusal is counted locally and never widens exposure; the function cannot admit a lane the manifest did not sign. Task 4.3.2's AST guard (`strategy_route_authority_guard_test.go`) pins that this file resolves the router import and never calls `Route` through it.
