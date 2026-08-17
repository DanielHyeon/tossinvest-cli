# Function Logic Map: `buildLaneInput`

- Source: `internal/strategyproposal/production.go`
- Current-base source SHA-256: `6cc7474d631e24c1daee677743fdbcc942787e9ae6874ed318cd3550326803b3`
- Signature: `buildLaneInput(params=6, results=3)`
- Source range: `269:1`–`382:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `274:3, 278:4, 293:5, 295:4, 300:4, 302:3, 306:4, 318:4, 332:5, 334:4, 338:4, 340:3, 343:3, 347:3, 373:4, 375:3, 379:3, 381:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 273:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 276:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 277:3 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 289:3 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 292:4 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 299:3 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 304:2 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 305:3 | planned targeted RED before any edit; not run by L0 |
| B9 | if | 309:3 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 313:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 317:3 | planned targeted RED before any edit; not run by L0 |
| B12 | if | 329:3 | planned targeted RED before any edit; not run by L0 |
| B13 | if | 331:4 | planned targeted RED before any edit; not run by L0 |
| B14 | if | 337:3 | planned targeted RED before any edit; not run by L0 |
| B15 | if | 342:2 | planned targeted RED before any edit; not run by L0 |
| B16 | if | 346:2 | planned targeted RED before any edit; not run by L0 |
| B17 | if | 350:2 | planned targeted RED before any edit; not run by L0 |
| B18 | if | 370:2 | planned targeted RED before any edit; not run by L0 |
| B19 | if | 372:3 | planned targeted RED before any edit; not run by L0 |
| B20 | if | 378:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| parseTime | 270:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| parseTime | 271:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| parseTime | 272:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| int64 | 281:86 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| continuationlane.BuildProductionKRProposalAuthority | 291:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.ContinuationKR | 295:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 295:39 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| continuationlane.BuildProductionUSProposalAuthority | 298:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.ContinuationUS | 302:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 302:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| reversalStructure | 316:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| time.Duration | 325:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| reversallane.BuildProductionKRProposalAuthority | 330:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.ReversalKR | 334:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 334:35 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| reversallane.BuildProductionUSProposalAuthority | 336:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.ReversalUS | 340:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 340:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| isWeeklyLane | 342:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| journalRO.WeeklyMarketReservation | 345:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 345:79 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| weeklyvaluelane.BuildProductionKRProposalAuthority | 371:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.WeeklyKR | 375:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 375:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| weeklyvaluelane.BuildProductionUSProposalAuthority | 377:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyflow.WeeklyUS | 381:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| a.Request | 381:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
