# Function Logic Map: `Context.NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `293:1`–`374:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.3).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 재검증의 drift 판정은 `strategyScheduleStillMatchesAdmission` 하나 — 클로저 모양(수집 + 그 함수 반환)과 판정 조건식 철자를 `TestTheScheduleDriftJudgementMovedVerbatim` 이 양쪽 못 박음, 축별 거절은 `TestTheScheduleDriftJudgementRefusesEveryAxis`(변이 A1 · A4~A9 CAUGHT, A2 · A3 는 행동상 동등 — 철자 핀만 잡음, `analysis/measurements/lot-6.3/mutation-6.3.tsv`).

## Branches and early returns

- Exact AST return nodes: `295:3, 322:3, 348:3, 364:3, 371:3, 373:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 294:2 | nil Context → 오류 |
| B2 | if | 300:2 | 원장 경로 있음 → 후보 · 원장 경로 |
| B3 | if | 311:2 | 원장 경로 → 근거 경로 |
| B4 | if | 321:2 | 레인 세우기 실패 → 오류 |
| B5 | if | 342:2 | 시계 있음 → dispatch now 배선 |
| B6 | range | 358:2 | KR · US worker 만들기 |
| B7 | if | 363:2 | 감독자 생성 실패 → 오류 |
| B8 | if | 370:2 | 투영 발행 실패 → 오류 |

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

## State mutations and fallbacks

- 편집 전과 같다(조립은 원장을 쓰지 않음, 투영 발행만).

## Safety conclusion

- High-risk 인접(주문 경로 최종 권한 재검사의 판정) — 의미 무변경 이동. 거절 · 통과 집합 불변(조건식 철자 동일).
