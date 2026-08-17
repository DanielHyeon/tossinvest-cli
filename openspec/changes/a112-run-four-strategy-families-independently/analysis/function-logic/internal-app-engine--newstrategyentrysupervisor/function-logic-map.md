# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `64893ce595e48abb31ed7e6c5a7630ae19373930c9cff148141490444202f888`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `491:1`–`568:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `497:3, 504:3, 507:3, 515:3, 521:4, 524:4, 527:4, 531:4, 535:4, 541:4, 544:4, 547:4, 561:4, 565:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 493:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 496:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 500:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 503:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 506:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 510:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 514:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 519:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 520:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 523:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 526:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 529:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 533:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 537:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 543:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 546:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 559:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 560:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 497:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 504:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 506:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 507:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 511:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 513:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 514:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 515:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 518:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 520:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 521:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 524:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 527:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 529:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 530:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 530:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 531:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 533:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 534:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 534:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 535:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 541:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 543:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 544:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 546:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 547:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 551:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 561:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 566:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 566:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
