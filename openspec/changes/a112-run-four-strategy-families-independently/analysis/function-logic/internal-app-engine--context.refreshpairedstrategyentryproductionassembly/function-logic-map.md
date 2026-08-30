# Function Logic Map: `refreshPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `12586e3cf90b708e66988931ad424d7312593bf518f0987a0893bf4f6f4b6fb9`
- Signature: `Context.refreshPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `446:1`–`463:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `448:3, 454:3, 458:3, 462:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 447:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 453:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 457:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 448:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Lock | 450:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Unlock | 451:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| UTC | 452:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 452:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 453:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Sub | 453:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.NewPairedStrategyEntryProductionAssembly | 456:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
