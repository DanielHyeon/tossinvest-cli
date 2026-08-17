# Function Logic Map: `runProductionStrategyMarketCycle`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `64893ce595e48abb31ed7e6c5a7630ae19373930c9cff148141490444202f888`
- Signature: `Context.runProductionStrategyMarketCycle(params=3, results=1)`
- Source range: `419:1`–`441:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `422:3, 426:3, 431:3, 434:3, 438:3, 440:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 421:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 425:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 430:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 433:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 437:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| c.refreshPairedStrategyEntryProductionAssembly | 420:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.proposals.forMarket | 424:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 425:33 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposal.entries.authority.Proposal | 428:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.CurrentPositionCampaignCAS | 429:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 429:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.dispatch.dispatch | 436:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.Is | 437:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
