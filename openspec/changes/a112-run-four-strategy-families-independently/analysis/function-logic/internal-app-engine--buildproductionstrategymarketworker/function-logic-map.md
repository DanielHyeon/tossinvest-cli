# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `64893ce595e48abb31ed7e6c5a7630ae19373930c9cff148141490444202f888`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- Source range: `377:1`–`417:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `384:3, 396:3, 400:3, 403:3, 406:3, 412:3, 414:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 383:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 394:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 399:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 402:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 405:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 411:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| schedule.forMarket | 392:27 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidate.forMarket | 392:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| route.forMarket | 392:84 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fx.forMarket | 393:3 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposal.forMarket | 393:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.forMarket | 393:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| account.forMarket | 393:86 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 395:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| p.entries.authority.Proposal | 398:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| result.ValidProposal | 399:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| gateway.ObserveStrategyProtection | 402:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToLower | 402:54 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 402:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| gateway.ObserveStrategyEntryGate | 405:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToLower | 405:53 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 405:69 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyWorkerEvidenceDigest | 408:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 411:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| IsZero | 411:64 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.authority.FreshUntil | 411:64 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.authority.FreshUntil | 415:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
