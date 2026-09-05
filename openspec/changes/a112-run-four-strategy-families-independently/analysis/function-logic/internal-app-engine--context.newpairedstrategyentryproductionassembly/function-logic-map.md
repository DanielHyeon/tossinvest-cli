# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Current-base source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`
- Signature: `Context.NewPairedStrategyEntryProductionAssembly(params=2, results=2)`
- Source range: `266:1`–`334:2`
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

- Exact AST return nodes: `285:3, 312:3, 335:4, 337:3, 353:3, 360:3, 362:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if| 284:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B2 | if| 290:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B3 | if| 301:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B4 | if| 311:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B5 | if| 330:3 | no coverage block for this arm (engine tagged suite); no coverage block for this arm (engine untagged suite); no per-test profile in the attribution set entered it |
| B6 | range| 347:2 | arm entered 2x (engine tagged suite); arm entered 2x (engine untagged suite); `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B7 | if| 352:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |
| B8 | if | 359:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no per-test profile in the attribution set entered it |


**태스크 5.1.2.2 가 분기 하나를 더했고 그래서 옛 B4~B7 이 B5~B8 이 되었다.**
새 B4(311:2)는 `c.productionStrategyLanes` 의 오류다. 레인을 **제안 수집 앞**에
세우는 것이 이 태스크가 옮긴 순서이고, 그 순서가 안전이다 — 관문이 레인의 잠금을
읽어 판정하므로(5.3.3 의 durable latch), 뒤에 세우면 재시작 뒤 첫 주기에 잠긴 레인이
열린 것으로 읽힌다. 그 창은 한 주기뿐이라 행동 시험이 우연히 잡지 못하므로 순서는
`TestTheLanesAreBuiltBeforeTheProposalsAreCollected` 가 구조로 못 박는다.

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| errors.New | 285:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 287:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 287:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 290:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 290:43 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.Journal.Path | 291:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 292:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 292:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 294:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyCandidateAuthorityLoader | 294:24 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 295:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRouteAuthorityLoader | 295:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.ToUpper | 297:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 297:37 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 298:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFXAuthorityLoader | 298:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Join | 302:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| filepath.Dir | 302:32 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.productionStrategyLanes | 310:16 | 태스크 5.1.2.2 — 이 프로세스의 여덟 레인을 **제안 수집 앞**에 세운다. 원장에서 durable latch 를 읽으므로 오류를 낼 수 있고, 그 오류가 B4 다 |
| collect | 314:23 | 태스크 5.1.2.2 — 그 레인을 제안 로더에 붙인다. 붙이지 않으면 4-가족 관문이 서지 않고 조정은 오늘과 같은 경로로 돈다 |
| withStrategyLanes | 314:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyProposalAuthorityLoader | 314:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.ResultAuthority | 317:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 318:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyRiskAuthorityLoader | 318:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collect | 320:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyAccountAuthorityLoader | 320:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newProductionStrategyFirstLegAuthorityLoader | 323:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyFirstLegAdmissionBridge | 325:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyDispatchCycle | 326:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| collectMarket | 329:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| newStrategyScheduleAuthorityLoader | 329:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.Generation | 333:4 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.Generation | 333:45 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| Equal | 334:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fresh.restore.Activation.ExpiresAt | 334:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| expected.restore.Activation.ExpiresAt | 334:48 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| errors.New | 335:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| scheduleAuthority.Snapshot | 344:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 346:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 348:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.productionStrategyWorker | 348:29 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| NewStrategyEntrySupervisor | 351:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| candidateAuthority.Snapshot | 356:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| routeAuthority.Snapshot | 356:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fxAuthority.Snapshot | 356:83 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| proposalAuthority.Snapshot | 356:117 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| riskAuthority.Snapshot | 357:9 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| accountAuthority.Snapshot | 357:44 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| c.publishStrategyRuntime | 359:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
