# Function Logic Map: `collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go`
- Current-base source SHA-256: `7d5d9b410e53463ca271c40edbd5c637fb37ed6e16a3e6b8528ab1a56bdfffcb`
- Signature: `strategyRouteAuthorityLoader.collectMarket(params=4, results=1)`
- Source range: `140:1`–`215:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `145:3, 148:3, 151:3, 154:3, 157:3, 162:3, 173:4, 185:3, 203:44, 205:3, 212:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 147:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 150:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 153:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 156:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 161:2 | planned targeted RED before any edit; not run by L0 |
| B6 | for | 170:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 172:3 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 184:2 | planned targeted RED before any edit; not run by L0 |
| B9 | range | 189:2 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 191:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 197:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 204:2 | planned targeted RED before any edit; not run by L0 |
| B13 | range | 209:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fail | 148:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 151:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 153:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 154:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 157:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 159:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 159:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| DecodeString | 160:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| base64.StdEncoding.Strict | 160:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| base64.StdEncoding.EncodeToString | 161:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 161:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 162:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 164:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 164:30 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 165:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 165:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 166:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 166:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 167:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 167:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 168:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 168:61 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 169:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 169:57 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 170:26 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.At | 171:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Valid | 172:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Market | 172:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 172:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Symbol | 172:78 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 173:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Symbol | 175:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 176:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Symbol | 176:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 177:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.load | 179:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyRouterMarket | 180:42 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ed25519.PublicKey | 181:36 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| batch.ManifestDigest | 184:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 186:58 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 186:99 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| batch.For | 190:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Symbol | 190:30 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.Request | 195:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyrouter.Route | 196:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyRouterMarket | 197:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| approved.Symbol | 197:127 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 201:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| sort.Slice | 203:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entries.approved.Symbol | 203:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entries.approved.Symbol | 203:82 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 204:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 206:58 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| sha256.New | 208:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| h.Write | 210:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| call | 210:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.approved.Symbol | 210:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.route.OwnerDigest | 210:60 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidates.approved.Len | 213:113 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 214:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| hex.EncodeToString | 214:106 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| h.Sum | 214:125 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
