# Function Logic Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `51840da714b49651bee5292a3b51f2814f98b7e2ee5e6996088fb9cceba14d2a`
- Signature: `Context.NewRefreshingPairedStrategyEntrySupervisor(params=1, results=2)`
- Source range: `338:1`–`362:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `340:3, 348:5, 356:3, 361:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 339:2 | planned targeted RED before any edit; not run by L0 |
| B2 | range | 343:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 355:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 340:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 342:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 345:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.runProductionStrategyMarketCycle | 348:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| NewStrategyEntrySupervisor | 352:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.Lock | 358:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.Unlock | 360:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
