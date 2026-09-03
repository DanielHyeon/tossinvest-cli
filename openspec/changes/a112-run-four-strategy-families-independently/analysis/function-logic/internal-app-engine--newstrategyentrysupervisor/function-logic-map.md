# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `545:1`–`622:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `551:3, 558:3, 561:3, 569:3, 575:4, 578:4, 581:4, 585:4, 589:4, 595:4, 598:4, 601:4, 615:4, 619:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 547:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 550:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 554:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 557:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 560:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 564:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 568:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 573:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 574:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 577:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 580:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 583:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 587:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 591:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 597:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 600:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 613:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 614:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 551:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 558:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 560:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 561:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 565:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 567:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 568:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 569:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 572:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 574:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 575:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 578:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 581:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 583:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 584:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 584:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 585:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 587:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 588:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 588:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 589:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 595:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 597:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 598:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 600:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 601:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 605:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 615:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 620:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 620:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
