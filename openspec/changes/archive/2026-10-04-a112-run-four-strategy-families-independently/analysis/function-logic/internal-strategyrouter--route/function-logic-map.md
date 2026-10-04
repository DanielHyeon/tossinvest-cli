# Function Logic Map: `Route`

- Source: `internal/strategyrouter/router.go`
- Current-base source SHA-256: `21fde4fa78acf9988c95e479b375dbb75308f4cb0ccb08435185ff9339c5d6fc`
- Signature: `Route(params=1, results=1)`
- Source range: `195:1`–`292:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `199:3, 204:3, 208:3, 212:3, 216:3, 222:4, 226:4, 234:3, 238:3, 242:3, 246:3, 250:3, 256:4, 259:3, 267:4, 271:4, 284:3, 288:3, 291:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 197:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 202:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 206:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 210:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 214:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 219:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 220:3 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 224:3 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 228:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 232:2 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 236:2 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 240:2 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 244:2 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 248:2 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 252:2 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 254:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 264:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 265:3 | planned targeted RED before any edit; not run by L0 |
| B19 | if | 269:3 | planned targeted RED before any edit; not run by L0 |
| B20 | if | 273:3 | planned targeted RED before any edit; not run by L0 |
| B21 | if | 276:3 | planned targeted RED before any edit; not run by L0 |
| B22 | else | 278:10 | planned targeted RED before any edit; not run by L0 |
| B23 | if | 278:10 | planned targeted RED before any edit; not run by L0 |
| B24 | if | 282:2 | planned targeted RED before any edit; not run by L0 |
| B25 | if | 286:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validOwnerKey | 197:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.EvaluatedAt.IsZero | 197:74 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ownerSnapshotSeal | 202:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validOwnerKey | 202:54 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.EvaluatedAt.Before | 214:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| request.EvaluatedAt.Before | 214:57 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 218:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validOwner | 224:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 229:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 232:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validMarketRecord | 236:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| EvaluateMarketLifecycle | 248:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 252:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validCandidateValue | 269:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
