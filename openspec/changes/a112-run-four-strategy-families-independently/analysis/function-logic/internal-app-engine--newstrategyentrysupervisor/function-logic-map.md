# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `66150078e25dfad6d1fec322b955e5f23e3aad77f0525321867a500e0960f58f`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `525:1`–`602:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `531:3, 538:3, 541:3, 549:3, 555:4, 558:4, 561:4, 565:4, 569:4, 575:4, 578:4, 581:4, 595:4, 599:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 527:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 530:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 534:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 537:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 540:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 544:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 548:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 553:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 554:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 557:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 560:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 563:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 567:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 571:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 577:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 580:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 593:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 594:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 531:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 538:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 540:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 541:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 545:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 547:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 548:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 549:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 552:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 554:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 555:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 558:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 561:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 563:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 564:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 564:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 565:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 567:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 568:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 568:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 569:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 575:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 577:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 578:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 580:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 581:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 585:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 595:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 600:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 600:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
