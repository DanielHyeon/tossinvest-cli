# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `1c2432d0f49db59209fc147f57a0c68d30d15596e68642aff8356ea29b0d69d5`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `262:1`–`330:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `264:3, 302:4, 304:3, 320:3, 327:3, 329:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 263:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 269:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 280:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 297:3 | planned targeted RED before any edit; not run by L0 |
| B5 | range | 314:2 | planned targeted RED before any edit; not run by L0 |
| B6 | if | 319:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 326:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 264:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 266:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 266:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 269:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 269:43 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 270:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 271:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 271:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 273:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyCandidateAuthorityLoader | 273:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 274:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRouteAuthorityLoader | 274:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 276:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 276:37 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 277:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFXAuthorityLoader | 277:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 281:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 281:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 283:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyProposalAuthorityLoader | 283:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.ResultAuthority | 285:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 286:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRiskAuthorityLoader | 286:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 288:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyAccountAuthorityLoader | 288:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newProductionStrategyFirstLegAuthorityLoader | 291:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFirstLegAdmissionBridge | 293:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyDispatchCycle | 294:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collectMarket | 296:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 296:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.Generation | 300:4 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.Generation | 300:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Equal | 301:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.ExpiresAt | 301:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.ExpiresAt | 301:48 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 302:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| scheduleAuthority.Snapshot | 311:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 313:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 315:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.productionStrategyWorker | 315:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| NewStrategyEntrySupervisor | 318:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidateAuthority.Snapshot | 323:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| routeAuthority.Snapshot | 323:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fxAuthority.Snapshot | 323:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.Snapshot | 323:117 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.Snapshot | 324:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accountAuthority.Snapshot | 324:44 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.publishStrategyRuntime | 326:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
