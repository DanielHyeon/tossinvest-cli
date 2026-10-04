# Function Logic Map: `strategyRouteAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go` (141-221)
- Function: `strategyRouteAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyRouteAuthorityLoader.collectMarket(params=4, results=1)`
- File SHA-256: `85e97cf96bd416ee52555042478ac3c184d745e61e7bb5a25aaf28128afe8fb9`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 13.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Loads one market's signed route authority and turns each approved candidate into a routed scope. Task 4.3.1 replaced the pre-evaluation single-winner `strategyrouter.Route` with `strategyrouter.RouteSet` here, so every eligible family reaches the coordinator instead of one winner chosen before evaluation. The market argument selects the manifest; nothing else crosses in.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 146:3, 149:3, 152:3, 155:3, 158:3, 163:3, 174:4, 188:3, 209:44, 211:3, 218:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 148:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if | 151:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if | 154:2 | arm never entered: count 0 in every profile measured for this function |
| B4 | if | 157:2 | arm never entered: count 0 in every profile measured for this function |
| B5 | if | 162:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | for | 171:2 | arm entered 8x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B7 | if | 173:3 | arm never entered: count 0 in every profile measured for this function |
| B8 | if | 187:2 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal` |
| B9 | range | 192:2 | arm entered 7x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B10 | if | 194:3 | arm entered 1x (engine suite); entered by `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B11 | if | 202:3 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |
| B12 | if | 210:2 | arm never entered: count 0 in every profile measured for this function |
| B13 | range | 215:2 | arm entered 6x (engine suite); entered by `TestStrategyRouteAuthorityKeepsMarketFailureLocal`, `TestStrategyRouteAuthorityLoadsKRUSConcurrently`, `TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fail` | 149:10 |
| `fail` | 152:10 |
| `candidates.approved.Len` | 154:5 |
| `fail` | 155:10 |
| `fail` | 158:10 |
| `strings.TrimSpace` | 160:13 |
| `loader.getenv` | 160:31 |
| `DecodeString` | 161:14 |
| `base64.StdEncoding.Strict` | 161:14 |
| `base64.StdEncoding.EncodeToString` | 162:19 |
| `len` | 162:72 |
| `fail` | 163:10 |
| `strings.TrimSpace` | 165:12 |
| `loader.getenv` | 165:30 |
| `strings.TrimSpace` | 166:11 |
| `loader.getenv` | 166:29 |
| `make` | 167:13 |
| `candidates.approved.Len` | 167:52 |
| `make` | 168:10 |
| `candidates.approved.Len` | 168:32 |
| `make` | 169:13 |
| `candidates.approved.Len` | 169:61 |
| `make` | 170:20 |
| `candidates.approved.Len` | 170:57 |
| `candidates.approved.Len` | 171:26 |
| `candidates.approved.At` | 172:19 |
| `approved.Valid` | 173:14 |
| `approved.Market` | 173:34 |
| `string` | 173:55 |
| `approved.Symbol` | 173:78 |
| `fail` | 174:11 |
| `approved.Symbol` | 176:8 |
| `append` | 177:13 |
| `approved.Symbol` | 177:74 |
| `append` | 178:20 |
| `loader.load` | 180:16 |
| `strategyRouterMarket` | 181:42 |
| `ed25519.PublicKey` | 182:36 |
| `batch.ManifestDigest` | 187:19 |
| `candidates.approved.Len` | 189:58 |
| `candidates.approved.Len` | 189:99 |
| `batch.For` | 193:20 |
| `approved.Symbol` | 193:30 |
| `authority.Request` | 198:14 |
| `strategyrouter.RouteSet` | 201:13 |
| `routed.Valid` | 202:52 |
| `len` | 202:70 |
| `strategyRouterMarket` | 203:26 |
| `approved.Symbol` | 203:80 |
| `append` | 207:13 |
| `sort.Slice` | 209:2 |
| `entries.approved.Symbol` | 209:51 |
| `entries.approved.Symbol` | 209:82 |
| `len` | 210:5 |
| `candidates.approved.Len` | 212:58 |
| `sha256.New` | 214:7 |
| `h.Write` | 216:10 |
| `(unnamed)` | 216:18 |
| `entry.approved.Symbol` | 216:25 |
| `entry.route.OwnerDigest` | 216:60 |
| `candidates.approved.Len` | 219:113 |
| `len` | 220:17 |
| `hex.EncodeToString` | 220:106 |
| `h.Sum` | 220:125 |

## State mutations and fallbacks

- AST assignments: 26. Defers: 0. Goroutine statements: 0.
- Appends to this market's own result slice only. No journal write, no broker call, no shared mutable state — the paired KR/US loaders are independent by construction.

## Safety conclusion

- Read-only over an already-signed manifest. A refusal is counted locally and never widens exposure; the function cannot admit a lane the manifest did not sign. Task 4.3.2's AST guard (`strategy_route_authority_guard_test.go`) pins that this file resolves the router import and never calls `Route` through it.
