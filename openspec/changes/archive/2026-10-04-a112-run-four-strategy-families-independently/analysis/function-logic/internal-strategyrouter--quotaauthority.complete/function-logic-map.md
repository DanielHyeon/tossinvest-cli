# Function Logic Map: `Complete`

- Source: `internal/strategyrouter/quota.go`
- Current-base source SHA-256: `f76c9e4fa25fc664efcc918ab5e1f588051dd5119149298778cf3d0a6d4a26c9`
- Signature: `QuotaAuthority.Complete(params=3, results=1)`
- Source range: `244:1`–`270:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `246:3, 249:3, 252:3, 259:3, 263:3, 269:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 245:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 248:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 251:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 258:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 262:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 266:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| validMarket | 245:26 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validHorizon | 245:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| capabilitySeal | 251:50 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.mu.Lock | 255:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.mu.Unlock | 256:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
