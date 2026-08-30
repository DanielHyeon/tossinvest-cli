# Function Logic Map: `refreshPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `1c2432d0f49db59209fc147f57a0c68d30d15596e68642aff8356ea29b0d69d5`
- Signature: `Context.refreshPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `459:1`–`476:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `461:3, 467:3, 471:3, 475:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 460:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 466:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 470:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 461:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Lock | 463:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Unlock | 464:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| UTC | 465:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 465:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 466:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Sub | 466:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.NewPairedStrategyEntryProductionAssembly | 469:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
