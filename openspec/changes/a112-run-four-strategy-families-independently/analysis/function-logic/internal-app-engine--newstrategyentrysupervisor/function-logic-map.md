# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `1c2432d0f49db59209fc147f57a0c68d30d15596e68642aff8356ea29b0d69d5`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `507:1`–`584:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `513:3, 520:3, 523:3, 531:3, 537:4, 540:4, 543:4, 547:4, 551:4, 557:4, 560:4, 563:4, 577:4, 581:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 509:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 512:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 516:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 519:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 522:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 526:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 530:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 535:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 536:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 539:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 542:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 545:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 549:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 553:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 559:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 562:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 575:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 576:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 513:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 520:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 522:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 523:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 527:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 529:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 530:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 531:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 534:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 536:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 537:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 540:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 543:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 545:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 546:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 546:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 547:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 549:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 550:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 550:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 551:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 557:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 559:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 560:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 562:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 563:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 567:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 577:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 582:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 582:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
