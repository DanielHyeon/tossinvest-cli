# Function Logic Map: `refreshPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `4e457c677157b2f8c73f813f8250575657b6beedddc1ad467db209a35579986d`
- Signature: `Context.refreshPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `461:1`–`478:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `463:3, 469:3, 473:3, 477:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 462:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 468:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 472:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 463:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Lock | 465:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyRefreshMu.Unlock | 466:8 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| UTC | 467:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 467:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 468:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Sub | 468:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.NewPairedStrategyEntryProductionAssembly | 471:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
