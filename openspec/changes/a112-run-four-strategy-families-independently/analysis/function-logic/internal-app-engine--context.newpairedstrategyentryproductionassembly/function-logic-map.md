# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `293:1`–`380:2`
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

Exact AST return positions: 295:3, 322:3, 352:4, 354:3, 370:3, 377:3, 379:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 294:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B2 | if | 300:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B3 | if | 311:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B4 | if | 321:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B5 | if | 342:2 | arm entered 1x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B6 | if | 347:3 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B7 | range | 364:2 | arm entered 2x (engine tagged suite, post-edit); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B8 | if | 369:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |
| B9 | if | 376:2 | arm not entered (engine tagged suite, post-edit); no per-test profile entered it |

8.7.2 가 더한 것: `if clk != nil { dispatchCycle.now = clk.Now }` 한 분기(Codex 재리뷰 P2 — nil 인터페이스의 메서드 값은 그 자리에서 panic). 그 뒤 분기들은 한 칸씩 밀렸다. 대입 한 줄은 `TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock` 이 구조로 못 박는다(조립 전체를 행동으로 세우려면 원장·게이트웨이·권한 여덟이 필요하다) — 변이 M23 CAUGHT.

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

> **5.6.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일의 다른 함수 편집으로 +12줄 이동 · 파일 해시만 바뀜). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 편집으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1 리뷰 수리(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 의 몸통 이동으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
