# Function Logic Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- Signature: `Context.NewRefreshingPairedStrategyEntrySupervisor(params=1, results=2)`
- Source range: `342:1`–`366:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `373:3, 381:5, 389:3, 394:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 372:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B2 | range| 376:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | if| 388:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 373:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 375:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 378:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.runProductionStrategyMarketCycle | 381:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| NewStrategyEntrySupervisor | 385:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.Lock | 391:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.strategyProjectionMu.Unlock | 393:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
