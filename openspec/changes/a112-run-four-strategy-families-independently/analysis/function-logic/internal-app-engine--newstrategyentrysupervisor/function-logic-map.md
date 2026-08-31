# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `4e457c677157b2f8c73f813f8250575657b6beedddc1ad467db209a35579986d`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `509:1`–`586:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `515:3, 522:3, 525:3, 533:3, 539:4, 542:4, 545:4, 549:4, 553:4, 559:4, 562:4, 565:4, 579:4, 583:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 511:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 514:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 518:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 521:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 524:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 528:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 532:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 537:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 538:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 541:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 544:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 547:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 551:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 555:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 561:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 564:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 577:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 578:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 515:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 522:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 524:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 525:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 529:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 531:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 532:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 533:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 536:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 538:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 539:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 542:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 545:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 547:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 548:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 548:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 549:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 551:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 552:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 552:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 553:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 559:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 561:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 562:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 564:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 565:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 569:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 579:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 584:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 584:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
