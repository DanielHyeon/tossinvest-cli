# Function Logic Map: `validProductionRouteCandidates`

- Source: `internal/strategyrouter/production.go`
- Current-base source SHA-256: `eafb36f41e2c07b85737692afa20fac968123481c812237f8678ad7a140bb520`
- Signature: `validProductionRouteCandidates(params=2, results=1)`
- Source range: `373:1`–`389:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `376:3, 384:4, 388:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 375:2 | planned targeted RED before any edit; not run by L0 |
| B2 | range | 379:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 381:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| productionRouteDescriptors | 374:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 375:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 375:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 378:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 378:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validDesiredState | 382:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validDesiredState | 382:42 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| productionRouteIdentity | 383:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| productionRouteIdentity | 383:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 388:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
