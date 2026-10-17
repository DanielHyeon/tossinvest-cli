# Function Logic Map (편집 전): `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `293:1`–`374:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §4 · §5: collect 의 shadow 짝을 받아 조립 리터럴(:366)에 별개 필드로 둠; proposalAuthority 분배(ResultAuthority · 계좌 · 1차 레그 · dispatch cycle · worker)는 무변경.

## Branches and early returns

- Exact AST return nodes: `295:3`, `322:3`, `348:3`, `364:3`, `371:3`, `373:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 294:2 | `if c == nil {` |
| B2 | if | 300:2 | `if c.Journal != nil && strings.TrimSpace(c.Journal.Path()) != "" {` |
| B3 | if | 311:2 | `if journalPath != "" {` |
| B4 | if | 321:2 | `if err != nil {` |
| B5 | if | 342:2 | `if clk != nil {` |
| B6 | range | 358:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B7 | if | 363:2 | `if err != nil {` |
| B8 | if | 370:2 | `if err := c.publishStrategyRuntime(assembly); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 295:45 |
| `collect` | 297:23 |
| `newStrategyScheduleAuthorityLoader` | 297:23 |
| `strings.TrimSpace` | 300:25 |
| `c.Journal.Path` | 300:43 |
| `c.Journal.Path` | 301:17 |
| `filepath.Join` | 302:24 |
| `filepath.Dir` | 302:38 |
| `collect` | 304:24 |
| `newStrategyCandidateAuthorityLoader` | 304:24 |
| `collect` | 305:20 |
| `newStrategyRouteAuthorityLoader` | 305:20 |
| `strings.ToUpper` | 307:21 |
| `strings.TrimSpace` | 307:37 |
| `collect` | 308:17 |
| `newStrategyFXAuthorityLoader` | 308:17 |
| `filepath.Join` | 312:18 |
| `filepath.Dir` | 312:32 |
| `c.productionStrategyLanes` | 320:16 |
| `collect` | 324:23 |
| `withStrategyLanes` | 324:23 |
| `newStrategyProposalAuthorityLoader` | 324:23 |
| `proposalAuthority.ResultAuthority` | 327:21 |
| `collect` | 328:19 |
| `newStrategyRiskAuthorityLoader` | 328:19 |
| `collect` | 330:22 |
| `newStrategyAccountAuthorityLoader` | 330:22 |
| `newProductionStrategyFirstLegAuthorityLoader` | 333:20 |
| `newStrategyFirstLegAdmissionBridge` | 335:20 |
| `newStrategyDispatchCycle` | 336:19 |
| `collectMarket` | 347:12 |
| `newStrategyScheduleAuthorityLoader` | 347:12 |
| `strategyScheduleStillMatchesAdmission` | 348:10 |
| `scheduleAuthority.Snapshot` | 355:23 |
| `make` | 357:13 |
| `append` | 359:13 |
| `c.productionStrategyWorker` | 359:29 |
| `NewStrategyEntrySupervisor` | 362:21 |
| `candidateAuthority.Snapshot` | 367:14 |
| `routeAuthority.Snapshot` | 367:52 |
| `fxAuthority.Snapshot` | 367:83 |
| `proposalAuthority.Snapshot` | 367:117 |
| `riskAuthority.Snapshot` | 368:9 |
| `accountAuthority.Snapshot` | 368:44 |
| `c.publishStrategyRuntime` | 370:12 |

## Safety conclusion

- High-risk(조립) — 기존 분배 · 오류 갈래 무변경, shadow 필드는 dispatch · worker 생성에 전달되지 않음(census ②).
