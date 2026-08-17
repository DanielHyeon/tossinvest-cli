# Function Logic Map: `refreshPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `64893ce595e48abb31ed7e6c5a7630ae19373930c9cff148141490444202f888`
- Signature: `Context.refreshPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `443:1`–`460:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `445:3, 451:3, 455:3, 459:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 444:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 450:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 454:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 445:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Lock | 447:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Unlock | 448:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| UTC | 449:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 449:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 450:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Sub | 450:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.NewPairedStrategyEntryProductionAssembly | 453:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
