# Function Logic Map (편집 전): `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `64f1cc0b85ecf5693dc5df0622b0f19f665697ea1b6f7616eb6755598f546e97`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `293:1`–`380:2`
- AST evidence: `ast.json` — 편집 **전**(a112 6.3 잔여 (c), Manager 판정 2026-10-01).

## Inputs and invariants

- 편집 계획: 재검증 클로저의 drift 비교(활성화 세대 · 만료 · 매니페스트 digest · 달력 버전 · desired revision · 준비 · 활성화 존재)를 의미 무변경으로 순수 함수 strategyScheduleStillMatchesAdmission 으로 옮긴다 — 클로저는 수집 + 그 함수 호출 두 문장. 조건식 철자 동일(이동 전후 AST 영수증).

## Branches and early returns

- Exact AST return nodes: `295:3`, `322:3`, `352:4`, `354:3`, `370:3`, `377:3`, `379:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 294:2 | `if c == nil {` |
| B2 | if | 300:2 | `if c.Journal != nil && strings.TrimSpace(c.Journal.Path()) != "" {` |
| B3 | if | 311:2 | `if journalPath != "" {` |
| B4 | if | 321:2 | `if err != nil {` |
| B5 | if | 342:2 | `if clk != nil {` |
| B6 | if | 347:3 | `if !fresh.snapshot.Ready // fresh.restore.Activation == nil // expected.restore.Activation == nil //` |
| B7 | range | 364:2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B8 | if | 369:2 | `if err != nil {` |
| B9 | if | 376:2 | `if err := c.publishStrategyRuntime(assembly); err != nil {` |

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
| `collectMarket` | 346:12 |
| `newStrategyScheduleAuthorityLoader` | 346:12 |
| `fresh.restore.Activation.Generation` | 350:4 |
| `expected.restore.Activation.Generation` | 350:45 |
| `Equal` | 351:5 |
| `fresh.restore.Activation.ExpiresAt` | 351:5 |
| `expected.restore.Activation.ExpiresAt` | 351:48 |
| `errors.New` | 352:11 |
| `scheduleAuthority.Snapshot` | 361:23 |
| `make` | 363:13 |
| `append` | 365:13 |
| `c.productionStrategyWorker` | 365:29 |
| `NewStrategyEntrySupervisor` | 368:21 |
| `candidateAuthority.Snapshot` | 373:14 |
| `routeAuthority.Snapshot` | 373:52 |
| `fxAuthority.Snapshot` | 373:83 |
| `proposalAuthority.Snapshot` | 373:117 |
| `riskAuthority.Snapshot` | 374:9 |
| `accountAuthority.Snapshot` | 374:44 |
| `c.publishStrategyRuntime` | 376:12 |

## Safety conclusion

- High-risk 인접(주문 경로 최종 권한 재검사) — 의미 무변경 이동. 거절 집합 · 통과 집합 불변.
