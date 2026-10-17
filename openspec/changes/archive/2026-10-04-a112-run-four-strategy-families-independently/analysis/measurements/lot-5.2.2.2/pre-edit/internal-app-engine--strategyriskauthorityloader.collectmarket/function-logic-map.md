# Function Logic Map: `collectMarket`

- Source: `internal/app/engine/strategy_risk_authority.go`
- Current-base source SHA-256: `8151a106ce66a76adc865520a899a103aaafa767cc66c42d44bed3f979857a55`
- Signature: `strategyRiskAuthorityLoader.collectMarket(params=4, results=1)`
- Source range: `170:1`–`201:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `174:3, 177:3, 180:3, 192:3, 196:3, 198:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 176:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 179:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 183:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 191:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 195:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fail | 177:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 180:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskbucket.LoadProductionRiskSnapshotAuthority | 186:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 192:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bundle.Scope | 194:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 195:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 195:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| scope.AsOf.Equal | 195:87 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 195:126 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bundle.Entries | 195:130 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 196:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 199:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bundle.Digest | 200:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 200:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bundle.Entries | 200:73 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
