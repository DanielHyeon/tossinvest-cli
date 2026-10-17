# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `9d97e59cf36590ade76b3e4804134b9ee66b3af1f8efa409e2f41222d5579b4f`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `283:1`–`370:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## What the L5 5.3.3 edit changed

이 로트가 이 함수에서 바꾼 것은 **한 줄**이다: 만들어 돌려주는 assembly 에 이 물결의
서명된 일정 권위(`schedule: scheduleAuthority`)를 함께 싣는다. 공개 `Schedule` 스냅숏은
스칼라 관측이라 활성화 세대를 담지 않는데, durable lane latch 의 복구 조건이 바로 그
세대다(`scheduler.Activation.Generation()`).

**스칼라 사본을 만들지 않은 이유.** 세대를 공개 스냅숏에 숫자로 복사하면 그 값이 권위에서
한 다리 멀어지고, 그 한 다리가 "복구 조건이 서명과 갈라지는" 자리가 된다. 권위를 그대로
들고 가면 복구 세대를 읽는 식이 `…restore.Activation.Generation()` 하나로 남고, 그 식은
`TestTheRecoveryGenerationComesFromTheVerifiedActivationAndNothingElse` 가 패키지 전체
열거로 얼려 둔다.

수집 순서·원격 호출·거절 갈래는 하나도 바뀌지 않았다.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Measurement regime (8.7.2 편집 뒤): 몸통 진입 count. engine tagged suite 바이너리(`-coverpkg=./internal/app/engine,./internal/strategyrouter`, -trimpath 없이)를 `systemd-run … MemoryMax=16G` 안에서 실행, 스위트 PASS; 전체 시험 509 개를 하나씩 돈 per-test 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh` · `a872_attribute.py`). 모든 행에서 시험별 합 == 스위트(ATTRIBUTION MISMATCH 0).

Exact AST return positions: 285:3, 312:3, 342:4, 344:3, 360:3, 367:3, 369:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 284:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B2 | if | 290:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B3 | if | 301:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B4 | if | 311:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B5 | if | 332:2 | arm entered 1x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B6 | if | 337:3 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B7 | range | 354:2 | arm entered 2x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B8 | if | 359:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B9 | if | 366:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |

8.7.2 가 더한 것: `if clk != nil { dispatchCycle.now = clk.Now }` 한 분기(Codex 재리뷰 P2 — nil 인터페이스의 메서드 값은 그 자리에서 panic). 그 뒤 분기들은 한 칸씩 밀렸다. 대입 한 줄은 `TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock` 이 구조로 못 박는다(조립 전체를 행동으로 세우려면 원장·게이트웨이·권한 여덟이 필요하다) — 변이 M23 CAUGHT.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 285:45 |
| `collect` | 287:23 |
| `newStrategyScheduleAuthorityLoader` | 287:23 |
| `strings.TrimSpace` | 290:25 |
| `c.Journal.Path` | 290:43 |
| `c.Journal.Path` | 291:17 |
| `filepath.Join` | 292:24 |
| `filepath.Dir` | 292:38 |
| `collect` | 294:24 |
| `newStrategyCandidateAuthorityLoader` | 294:24 |
| `collect` | 295:20 |
| `newStrategyRouteAuthorityLoader` | 295:20 |
| `strings.ToUpper` | 297:21 |
| `strings.TrimSpace` | 297:37 |
| `collect` | 298:17 |
| `newStrategyFXAuthorityLoader` | 298:17 |
| `filepath.Join` | 302:18 |
| `filepath.Dir` | 302:32 |
| `c.productionStrategyLanes` | 310:16 |
| `collect` | 314:23 |
| `withStrategyLanes` | 314:23 |
| `newStrategyProposalAuthorityLoader` | 314:23 |
| `proposalAuthority.ResultAuthority` | 317:21 |
| `collect` | 318:19 |
| `newStrategyRiskAuthorityLoader` | 318:19 |
| `collect` | 320:22 |
| `newStrategyAccountAuthorityLoader` | 320:22 |
| `newProductionStrategyFirstLegAuthorityLoader` | 323:20 |
| `newStrategyFirstLegAdmissionBridge` | 325:20 |
| `newStrategyDispatchCycle` | 326:19 |
| `collectMarket` | 336:12 |
| `newStrategyScheduleAuthorityLoader` | 336:12 |
| `fresh.restore.Activation.Generation` | 340:4 |
| `expected.restore.Activation.Generation` | 340:45 |
| `Equal` | 341:5 |
| `fresh.restore.Activation.ExpiresAt` | 341:5 |
| `expected.restore.Activation.ExpiresAt` | 341:48 |
| `errors.New` | 342:11 |
| `scheduleAuthority.Snapshot` | 351:23 |
| `make` | 353:13 |
| `append` | 355:13 |
| `c.productionStrategyWorker` | 355:29 |
| `NewStrategyEntrySupervisor` | 358:21 |
| `candidateAuthority.Snapshot` | 363:14 |
| `routeAuthority.Snapshot` | 363:52 |
| `fxAuthority.Snapshot` | 363:83 |
| `proposalAuthority.Snapshot` | 363:117 |
| `riskAuthority.Snapshot` | 364:9 |
| `accountAuthority.Snapshot` | 364:44 |
| `c.publishStrategyRuntime` | 366:12 |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.

## 2026-09-27 — 태스크 8.7.2 편집 전 (현재 AST 와 SHA 일치 확인)

`ast.json` 의 SHA-256 은 편집 전 이 워크트리 파일과 같다(재확인). 편집은 한 줄이다: `dispatchCycle.revalidateSchedule = …`
대입 **옆**에 `dispatchCycle.now = clk.Now` 를 둔다 — 분기·호출 순서·반환은 바뀌지 않고 대입 하나가 는다. 이 함수가 dispatch
주기에 실시계를 건네는 유일한 생산 조립이다(생산에서 `newStrategyDispatchCycle` 을 부르는 자리는 이 함수 하나 — grep 으로 확인해
review.md 에 적는다). 이 대입이 빠지면 검증된 가족 활성화를 가진 시장의 주문이 전부 거절된다(fail-closed) — 넓힘이 아니다.
