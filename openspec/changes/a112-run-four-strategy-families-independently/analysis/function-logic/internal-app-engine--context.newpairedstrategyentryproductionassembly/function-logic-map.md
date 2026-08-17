# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `64893ce595e48abb31ed7e6c5a7630ae19373930c9cff148141490444202f888`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `261:1`–`329:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `263:3, 301:4, 303:3, 319:3, 326:3, 328:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 262:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 268:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 279:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 296:3 | planned targeted RED before any edit; not run by L0 |
| B5 | range | 313:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 318:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 325:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 263:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 265:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 265:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 268:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 268:43 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 269:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 270:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 270:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 272:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyCandidateAuthorityLoader | 272:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 273:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRouteAuthorityLoader | 273:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 275:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 275:37 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 276:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFXAuthorityLoader | 276:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 280:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 280:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 282:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyProposalAuthorityLoader | 282:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.ResultAuthority | 284:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 285:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRiskAuthorityLoader | 285:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 287:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyAccountAuthorityLoader | 287:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newProductionStrategyFirstLegAuthorityLoader | 290:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFirstLegAdmissionBridge | 292:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyDispatchCycle | 293:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collectMarket | 295:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 295:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.Generation | 299:4 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.Generation | 299:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Equal | 300:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.ExpiresAt | 300:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.ExpiresAt | 300:48 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 301:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| scheduleAuthority.Snapshot | 310:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 312:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 314:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.productionStrategyWorker | 314:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| NewStrategyEntrySupervisor | 317:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidateAuthority.Snapshot | 322:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| routeAuthority.Snapshot | 322:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fxAuthority.Snapshot | 322:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.Snapshot | 322:117 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.Snapshot | 323:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accountAuthority.Snapshot | 323:44 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.publishStrategyRuntime | 325:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
