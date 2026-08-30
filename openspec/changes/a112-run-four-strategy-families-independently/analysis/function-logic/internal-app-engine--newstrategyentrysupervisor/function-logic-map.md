# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `12586e3cf90b708e66988931ad424d7312593bf518f0987a0893bf4f6f4b6fb9`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `494:1`–`571:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `500:3, 507:3, 510:3, 518:3, 524:4, 527:4, 530:4, 534:4, 538:4, 544:4, 547:4, 550:4, 564:4, 568:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 496:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 499:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 503:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 506:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 509:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 513:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 517:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 522:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 523:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 526:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 529:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 532:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 536:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 540:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 546:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 549:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 562:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 563:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 500:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 507:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 509:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 510:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 514:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 516:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 517:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 518:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 521:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 523:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 524:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 527:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 530:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 532:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 533:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 533:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 534:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 536:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 537:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 537:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 538:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 544:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 546:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 547:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 549:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 550:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 554:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 564:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 569:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 569:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
