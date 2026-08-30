# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `17ad4c0c684b74686dd1e80b256a06971802afa26bcfd300dbeac9bd5f7e0496`
- Signature: `NewStrategyEntrySupervisor(params=1, results=2)`
- Source range: `495:1`–`572:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `501:3, 508:3, 511:3, 519:3, 525:4, 528:4, 531:4, 535:4, 539:4, 545:4, 548:4, 551:4, 565:4, 569:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 497:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 500:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 504:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 507:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 510:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 514:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 518:2 | planned targeted RED before any edit; not run by L0 |
| B8 | range | 523:2 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 524:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 527:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 530:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 533:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 537:3 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 541:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 547:3 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 550:3 | planned targeted RED before any edit; not run by L0 |
| B17 | range | 563:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 564:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| fmt.Errorf | 501:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 508:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 510:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 511:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clock.System | 515:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| clk.Now | 517:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.IsZero | 518:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 519:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 522:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyMarket | 524:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 525:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 528:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 531:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 533:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| now.Before | 534:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyDigest | 534:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 535:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 537:124 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| validStrategyWorkerRefusal | 538:96 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 538:151 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 539:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 545:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.AuthorityExpiresAt.IsZero | 547:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 548:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| descriptor.RestartNotBefore.IsZero | 550:123 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 551:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 555:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fmt.Errorf | 565:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 570:62 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 570:91 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
