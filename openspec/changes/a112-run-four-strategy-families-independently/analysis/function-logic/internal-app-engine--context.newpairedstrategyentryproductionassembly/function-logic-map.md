# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `296:1`–`377:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- shadow 필드 대입 자리는 AST 핀.

## Branches and early returns

- Exact AST return nodes: `298:3, 325:3, 351:3, 367:3, 374:3, 376:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 297:2 | Context nil |
| B2 | if | 303:2 | 원장 경로 |
| B3 | if | 314:2 | evidence 경로 |
| B4 | if | 324:2 | 레인 런타임 오류 |
| B5 | if | 345:2 | 시계 → dispatch now |
| B6 | range | 361:2 | 시장 worker 순회 |
| B7 | if | 366:2 | supervisor 생성 오류 |
| B8 | if | 373:2 | 투영 발행 오류 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 298:45 |
| `collect` | 300:23 |
| `newStrategyScheduleAuthorityLoader` | 300:23 |
| `strings.TrimSpace` | 303:25 |
| `c.Journal.Path` | 303:43 |
| `c.Journal.Path` | 304:17 |
| `filepath.Join` | 305:24 |
| `filepath.Dir` | 305:38 |
| `collect` | 307:24 |
| `newStrategyCandidateAuthorityLoader` | 307:24 |
| `collect` | 308:20 |
| `newStrategyRouteAuthorityLoader` | 308:20 |
| `strings.ToUpper` | 310:21 |
| `strings.TrimSpace` | 310:37 |
| `collect` | 311:17 |
| `newStrategyFXAuthorityLoader` | 311:17 |
| `filepath.Join` | 315:18 |
| `filepath.Dir` | 315:32 |
| `c.productionStrategyLanes` | 323:16 |
| `collect` | 327:40 |
| `withStrategyLanes` | 327:40 |
| `newStrategyProposalAuthorityLoader` | 327:40 |
| `proposalAuthority.ResultAuthority` | 330:21 |
| `collect` | 331:19 |
| `newStrategyRiskAuthorityLoader` | 331:19 |
| `collect` | 333:22 |
| `newStrategyAccountAuthorityLoader` | 333:22 |
| `newProductionStrategyFirstLegAuthorityLoader` | 336:20 |
| `newStrategyFirstLegAdmissionBridge` | 338:20 |
| `newStrategyDispatchCycle` | 339:19 |
| `collectMarket` | 350:12 |
| `newStrategyScheduleAuthorityLoader` | 350:12 |
| `strategyScheduleStillMatchesAdmission` | 351:10 |
| `scheduleAuthority.Snapshot` | 358:23 |
| `make` | 360:13 |
| `append` | 362:13 |
| `c.productionStrategyWorker` | 362:29 |
| `NewStrategyEntrySupervisor` | 365:21 |
| `candidateAuthority.Snapshot` | 370:14 |
| `routeAuthority.Snapshot` | 370:52 |
| `fxAuthority.Snapshot` | 370:83 |
| `proposalAuthority.Snapshot` | 370:117 |
| `riskAuthority.Snapshot` | 371:9 |
| `accountAuthority.Snapshot` | 371:44 |
| `c.publishStrategyRuntime` | 373:12 |

## State mutations and fallbacks

- 조립 값 · 발행(편집 전과 같음).

## Safety conclusion

- High-risk(조정 · 제안 권한 · 조립) 경로의 편집은 운반뿐이다 — 조정 · admit · Submit · Arbitrate · dispatch 의 입력 · 순서 · 반환은 편집 전과 같고(차등 dispatch 시험 · 변이 S01~S08), shadow 값은 authority 구조체에 들어가지 않는다(census ② `TestOnlyTheAllowedFunctionsEverTouchAShadowType`).
