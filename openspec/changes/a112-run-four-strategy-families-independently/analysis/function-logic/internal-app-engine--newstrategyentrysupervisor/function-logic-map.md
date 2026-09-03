# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `51840da714b49651bee5292a3b51f2814f98b7e2ee5e6996088fb9cceba14d2a`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `528:1`–`605:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `534:3, 541:3, 544:3, 552:3, 558:4, 561:4, 564:4, 568:4, 572:4, 578:4, 581:4, 584:4, 598:4, 602:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 530:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 533:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 537:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 540:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 543:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 547:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 551:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 556:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 557:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 560:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 563:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 566:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 570:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 574:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 580:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 583:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 596:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 597:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 534:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 541:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 543:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 544:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 548:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 550:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 551:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 552:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 555:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 557:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 558:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 561:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 564:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 566:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 567:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 567:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 568:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 570:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 571:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 571:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 572:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 578:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 580:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 581:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 583:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 584:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 588:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 598:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 603:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 603:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
