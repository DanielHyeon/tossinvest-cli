# a100 M-A 준비 — FLM 재동결 증거 (2026-11-10)

작성: Opus Teammate(준비 로트 C). 브로커·네트워크 호출 0, `make sdd-sync`·`make sdd-check`·`make gate`·
전체 시험 미실행. worktree `/home/daniel/tossos-work/a100-ma-prep`(ext4), branch `docs/a100-ma-prep`.
도구 판: git 2.43.0 · go1.26.5 linux/amd64 · Python 3.12.3. `go run`은 `GOPROXY=off GOFLAGS=-mod=readonly`로
돌려 모듈 내려받기를 막았다.

## 결론

- `AST source hash is stale`: **15 → 4**. 남은 4개는 위치 분기 ID가 밀린 번들이라 손대지 않았다(아래 「멈춘 곳」).
- `AST hash does not match modified function revision`: **11 → 4**(같은 4개).
- `missing evidence for modified function`: **491 → 491**. base `882a0b49` 뒤 1,378(+이 로트 2) 커밋이 바꾼
  기존 함수가 창 안에 들어온 결과다. 번들 재생성으로 줄지 않는다.
- `check_analysis.py` 종료 코드는 전후 모두 **1**이다. 이 로트는 PASS를 주장하지 않는다.

## base 재고정(`capture_change_base.py`)은 실행하지 않았다 — 근거

1. `tools/sdd/capture_change_base.py`는 `base-commit.txt`가 이미 있으면 `change base already captured`로
   거부한다(`capture()` 안의 `if target.exists()`). 재고정은 그 파일을 지우는 사람 절차다.
2. `docs/WORKFLOW.md` 「사람 승인 base 재고정」은 셋을 **전부** 요구한다: (1) 귀속 실측 — 옛 base 이후 이
   change 의 자기 Go 커밋이 0 이거나, 아니면 그 커밋들이 고친 기존 함수가 전부 fresh 번들로 덮였음을 실측,
   (2) Manager 또는 사람의 승인을 `review.md`·`tasks.md`에 기록, (3) `base-commit.txt` 한 파일 단독 커밋.
3. 귀속 실측의 입력(이 로트가 센 값): `882a0b49..6dbb752f6`에서 이 change 디렉터리를 만지며 `.go`를 고친
   비병합 커밋은 **3개**다 — `016da6245`(2026-08-15 M0), `30694fe56`(2026-10-05 T1), `6f06eee8e`(2026-10-17 a144
   — a100 `tasks.md` 한 줄과 함께 커밋됨). 0이 아니므로 (1)의 「전부 덮였음」 실측이 필요하다. 이 로트는 그
   실측을 하지 않았다.
4. tasks 0.1(`tasks.md:46-47`)은 재고정 뒤 `make sdd-sync`를 다시 돌리라고 적는다. 이 로트는 그 명령을
   실행할 권한이 없다.

⇒ 재고정은 Manager 결정과 사람 절차 대상이다. 여기서 멈췄다.

## 실행한 명령

```bash
# 전(HEAD 6dbb752f6, 이 로트의 문서 커밋 전)
python3 tools/logic-map/check_analysis.py --change a100-wire-fill-to-broker-protection

# stale 15개 번들의 새 AST — 저장소 밖 scratch 로 먼저 뽑아 옛 AST 와 대조
export GOPROXY=off GOFLAGS=-mod=readonly
go run ./tools/logic-map --file cmd/tossctl/engine.go --func engineRuntime > $SCRATCH/fresh/cmd-tossctl--engineruntime.ast.json
go run ./tools/logic-map --file cmd/tossctl/verify.go --func newVerifyRunCmd > $SCRATCH/fresh/cmd-tossctl--newverifyruncmd.ast.json
go run ./tools/logic-map --file cmd/tossctl/verify.go --func runVerifyRun > $SCRATCH/fresh/cmd-tossctl--runverifyrun.ast.json
go run ./tools/logic-map --file internal/app/engine/gateway.go --func buildGateway > $SCRATCH/fresh/internal-app-engine--buildgateway.ast.json
go run ./tools/logic-map --file internal/app/engine/exitloop.go --func ExitObserver.record > $SCRATCH/fresh/internal-app-engine--exitobserver.record.ast.json
go run ./tools/logic-map --file internal/app/engine/exitloop.go --func ExitObserver.submit > $SCRATCH/fresh/internal-app-engine--exitobserver.submit.ast.json
go run ./tools/logic-map --file internal/app/engine/interlock.go --func runInterlock > $SCRATCH/fresh/internal-app-engine--runinterlock.ast.json
go run ./tools/logic-map --file internal/execgw/gateway_test.go --func TestUnreachableBrokerIsNotDispatched > $SCRATCH/fresh/internal-execgw--testunreachablebrokerisnotdispatched.ast.json
go run ./tools/logic-map --file internal/journal/apply_hook.go --func Journal.OpenExitStates > $SCRATCH/fresh/internal-journal--journal.openexitstates.ast.json
go run ./tools/logic-map --file internal/journal/exit_snapshot.go --func scanExitStateResult > $SCRATCH/fresh/internal-journal--scanexitstateresult.ast.json
go run ./tools/logic-map --file internal/official/client.go --func Client.doRequest > $SCRATCH/fresh/internal-official--client.dorequest.ast.json
go run ./tools/logic-map --file internal/official/client.go --func Client.send > $SCRATCH/fresh/internal-official--client.send.ast.json
go run ./tools/logic-map --file internal/verifylive/report.go --func BuildProgress > $SCRATCH/fresh/internal-verifylive--buildprogress.ast.json
go run ./tools/logic-map --file internal/verifylive/report.go --func BuildReport > $SCRATCH/fresh/internal-verifylive--buildreport.ast.json
go run ./tools/logic-map --file internal/verifylive/report.go --func Progress.WriteText > $SCRATCH/fresh/internal-verifylive--progress.writetext.ast.json

# 대조 규칙: 옛/새 AST 의 (분기 id, kind) 열이 같으면 SAME, 다르면 SHIFTED.
# SAME 11개만 ast.json 을 새 것으로 바꾸고(옛 파일에 "revision" 키가 있으면 그 값을 보존),
# function-logic-map.md·branch-test-map.md 의 `Source: `file` (a-b)` 줄 범위만 새 start-end 로 바꿨다.
# SHIFTED 4개는 번들을 그대로 두고 새 AST 만 analysis/ma-prep-2026-11-10/shifted-ast/ 에 남겼다.

# 후(같은 워킹트리, 번들 갱신 후·커밋 전; 새 파일이 모두 있는 상태로 두 번 돌려 결과가 같음을 확인)
python3 tools/logic-map/check_analysis.py --change a100-wire-fill-to-broker-protection
```

## 대조 결과 (옛 AST → 새 AST)

```text
cmd-tossctl--engineruntime: shape=SAME branches 6->6 range 649-739 -> 655-752 offsets-differ rev=None body_len 90->97
cmd-tossctl--newverifyruncmd: shape=SAME branches 0->0 range 131-217 -> 132-218 same-offsets rev=current body_len 86->86
cmd-tossctl--runverifyrun: shape=SAME branches 17->17 range 272-400 -> 273-401 same-offsets rev=current body_len 128->128
internal-app-engine--buildgateway: shape=SHIFTED branches 6->10 range 234-359 -> 250-386 offsets-differ rev=None body_len 125->136
   old: B1:if B2:if B3:if B4:if B5:if B6:if
   new: B1:if B2:if B3:if B4:if B5:if B6:if B7:if B8:if B9:if B10:if
internal-app-engine--exitobserver.record: shape=SHIFTED branches 16->19 range 1232-1362 -> 1314-1455 offsets-differ rev=None body_len 130->141
   old: B1:if B2:if B3:if B4:else B5:if B6:if B7:else B8:if B9:if B10:else B11:if B12:if B13:if B14:if B15:if B16:if
   new: B1:if B2:if B3:if B4:else B5:if B6:else B7:if B8:if B9:else B10:if B11:if B12:if B13:else B14:if B15:if B16:if B17:if B18:if B19:if
internal-app-engine--exitobserver.submit: shape=SHIFTED branches 13->17 range 1402-1487 -> 1495-1616 offsets-differ rev=None body_len 85->121
   old: B1:if B2:if B3:if B4:if B5:if B6:switch B7:case B8:case B9:case B10:case B11:if B12:case B13:if
   new: B1:if B2:if B3:if B4:if B5:if B6:if B7:if B8:switch B9:case B10:case B11:if B12:case B13:case B14:if B15:case B16:if B17:if
internal-app-engine--runinterlock: shape=SAME branches 3->3 range 376-407 -> 390-421 same-offsets rev=None body_len 31->31
internal-execgw--testunreachablebrokerisnotdispatched: shape=SAME branches 4->4 range 727-757 -> 729-759 same-offsets rev=None body_len 30->30
internal-journal--journal.openexitstates: shape=SAME branches 4->4 range 621-644 -> 631-654 same-offsets rev=None body_len 23->23
internal-journal--scanexitstateresult: shape=SAME branches 22->22 range 574-702 -> 577-705 same-offsets rev=None body_len 128->128
internal-official--client.dorequest: shape=SHIFTED branches 2->3 range 191-207 -> 203-222 offsets-differ rev=current body_len 16->19
   old: B1:if B2:if
   new: B1:if B2:if B3:if
internal-official--client.send: shape=SAME branches 9->9 range 320-366 -> 335-381 same-offsets rev=current body_len 46->46
internal-verifylive--buildprogress: shape=SAME branches 7->7 range 315-343 -> 345-374 same-offsets rev=current body_len 28->29
internal-verifylive--buildreport: shape=SAME branches 10->10 range 166-216 -> 169-220 same-offsets rev=current body_len 50->51
internal-verifylive--progress.writetext: shape=SAME branches 8->8 range 346-379 -> 377-411 offsets-differ rev=current body_len 33->34
```

`offsets-differ`는 분기 kind 열은 같지만 함수 시작 기준 분기 위치가 달라졌다는 뜻이다. 그런 SAME 번들
두 개(`cmd-tossctl--engineruntime`, `internal-verifylive--progress.writetext`)는 분기 ID가 밀리지 않아
갱신했지만, **서술(호출 표·불변식)을 새 본문과 다시 대조하지는 않았다.** 예: `engineRuntime`은 base 뒤
`ExitObserverOptions`에 전용 로거 셋이 들어오고 `Auxiliary` 목록이 1개에서 5개로 늘었다(`git diff
882a0b49 HEAD -- cmd/tossctl/engine.go`). 이 둘은 T4-B(3.9) 착수 전 서술 재검토 대상이다. 다른 SAME 번들도
같은 이유로 서술은 재검토하지 않았다 — 이 로트는 해시·범위 재동결만 한다.

## 멈춘 곳 — 위치 분기 ID가 밀린 번들 (Manager 판단 대상, 손으로 재번호하지 않음)

| 번들 | 옛 분기 | 새 분기 | 옛 범위 → 새 범위 |
| --- | --- | --- | --- |
| `internal-app-engine--buildgateway` | 6 | 10 | 234-359 → 250-386 |
| `internal-app-engine--exitobserver.record` | 16 | 19 | 1232-1362 → 1314-1455 |
| `internal-app-engine--exitobserver.submit` | 13 | 17 | 1402-1487 → 1495-1616 |
| `internal-official--client.dorequest` | 2 | 3 | 191-207 → 203-222 |

새 AST는 `shifted-ast/<번들>.ast.json`에 있다. Branch Test Map 재번호는 옛/새 `ast.json`을 정렬해
행을 옮기는 작업이고 FLM 서술도 다시 써야 한다. 이 넷은 a100의 High-risk 편집 대상(4.4·4.5·0.2a)이라
Pre-Edit 선언(`pre-edit-gate.md` 1·4·5절, 0.2a)과 함께 다시 봐야 한다.

## 원문 출력 — 전 (HEAD 6dbb752f6)

```text
[logic-map] a100-wire-fill-to-broker-protection: base 882a0b490b0b → working tree (no landed-commit.txt in HEAD) required 534 function(s) — judged at HEAD 6dbb752f63fe
[logic-map] a100-wire-fill-to-broker-protection: the target is the working tree, so this window also holds 1378 commit(s) that landed after the base and every existing function they changed is required here too — but 2 of this change's `revision: current` bundle(s) (internal-app-engine--runinterlock, internal-journal--scanexitstateresult) record the source as it stood at the base, not at HEAD, so every landing the gate could compute is one where those functions still equal the base and the narrowed window would require none of them; refresh those bundles against the current source instead
[logic-map] cmd-tossctl--engineruntime: AST source hash is stale: cmd/tossctl/engine.go
[logic-map] cmd-tossctl--newverifyruncmd: AST source hash is stale: cmd/tossctl/verify.go
[logic-map] cmd-tossctl--runverifyrun: AST source hash is stale: cmd/tossctl/verify.go
[logic-map] internal-app-engine--buildgateway: AST source hash is stale: internal/app/engine/gateway.go
[logic-map] internal-app-engine--exitobserver.record: AST source hash is stale: internal/app/engine/exitloop.go
[logic-map] internal-app-engine--exitobserver.submit: AST source hash is stale: internal/app/engine/exitloop.go
[logic-map] internal-app-engine--runinterlock: AST source hash is stale: internal/app/engine/interlock.go
[logic-map] internal-execgw--testunreachablebrokerisnotdispatched: AST source hash is stale: internal/execgw/gateway_test.go
[logic-map] internal-journal--journal.openexitstates: AST source hash is stale: internal/journal/apply_hook.go
[logic-map] internal-journal--scanexitstateresult: AST source hash is stale: internal/journal/exit_snapshot.go
[logic-map] internal-official--client.dorequest: AST source hash is stale: internal/official/client.go
[logic-map] internal-official--client.send: AST source hash is stale: internal/official/client.go
[logic-map] internal-verifylive--buildprogress: AST source hash is stale: internal/verifylive/report.go
[logic-map] internal-verifylive--buildreport: AST source hash is stale: internal/verifylive/report.go
[logic-map] internal-verifylive--progress.writetext: AST source hash is stale: internal/verifylive/report.go
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061RunAndConsoleVerificationDoNotBuildBrokerWhileTheProfileBudgetIsOccupied
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061VerificationRefusesBeforeBrokerWhenItsIntentCannotBePublished
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061RunVerifyAbortPublishesItsMarkerBeforeWaitingForTheProfileBudget
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061AbortListAndRefreshedEmptyRecordNeverBuildABroker
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061ActiveVerificationExclusionRefusesAbortWithoutErasingItsMarker
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061AbortReloadsOutstandingTargetsAfterExclusiveAdmission
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061MetadataAndAccountReadsShareOneOfficialClient
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_engine_registers_a_sender_test.go:TestProductionRuntimeStartsExactlyOneAlertDeliverer
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_operator_command_names_a_person_test.go:TestTheAlertCommandsAreWiredUnderEngine
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_operator_command_names_a_person_test.go:TestTheCommandsRefuseWhenNoEngineIsRunning
[logic-map] missing evidence for modified function cmd/tossctl/a108_the_aggregate_snapshot_outlives_the_strategy_read_test.go:a108SnapshotReader
[logic-map] missing evidence for modified function cmd/tossctl/a111_httpapi_exit_freshness_test.go:a111NewHTTPRouteSession
[logic-map] missing evidence for modified function cmd/tossctl/candidate.go:candidateWiring
[logic-map] missing evidence for modified function cmd/tossctl/candidate_test.go:withCandidateFixture
[logic-map] missing evidence for modified function cmd/tossctl/candidatepanel.go:buildCandidatePanel
[logic-map] missing evidence for modified function cmd/tossctl/console.go:runConsole
[logic-map] missing evidence for modified function cmd/tossctl/console.go:newConsoleBroker
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.resolve
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.instrumentMetadata
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.TypedMarketCalendar
[logic-map] missing evidence for modified function cmd/tossctl/console.go:lazyHoldings.Holdings
[logic-map] missing evidence for modified function cmd/tossctl/console.go:lazyOrders.Orders
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleRelaunch
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleVerifyStarter
[logic-map] missing evidence for modified function cmd/tossctl/console_openapi.go:defaultConsoleOpenAPIDeps
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheOrdersSeamCarriesEachListsOutcomeSeparately
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce
[logic-map] missing evidence for modified function cmd/tossctl/doctor_openapi.go:computeDoctorOpenAPISummary
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:newEngineCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:newEngineRunCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:runEngineRun
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:reportStrategyProjectionDegraded
[logic-map] cmd-tossctl--engineruntime: AST hash does not match modified function revision
[logic-map] missing evidence for modified function cmd/tossctl/engine_account_seq_recovery_test.go:TestActualEngineRecoveryStillFailsClosedOnASnapshot429
[logic-map] missing evidence for modified function cmd/tossctl/engine_alerts.go:newEngineAlertsAckCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine_assembly.go:assembleEngine
[logic-map] missing evidence for modified function cmd/tossctl/engine_assembly_test.go:TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver
[logic-map] missing evidence for modified function cmd/tossctl/engine_reconcile.go:buildReconcileResolveRuntime
[logic-map] missing evidence for modified function cmd/tossctl/engine_test.go:TestTheExitObserverDefersToFillDetection
[logic-map] missing evidence for modified function cmd/tossctl/flatten.go:runFlatten
[logic-map] missing evidence for modified function cmd/tossctl/flatten.go:buildFlattenWiring
[logic-map] missing evidence for modified function cmd/tossctl/help_convention_test.go:TestMutatingAnnotationOnTradeCommands
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:runHTTPAPI
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:strategyRuntimeReaderFor
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:openHTTPAPIResources
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:publishHTTPAPISnapshots
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.readPositionBrokerSnapshot
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.Performance
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.Snapshot
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader_test.go:TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources
[logic-map] missing evidence for modified function cmd/tossctl/init.go:runInitOfficialFlow
[logic-map] missing evidence for modified function cmd/tossctl/instrument_names.go:lazyInstrumentNames.Names
[logic-map] missing evidence for modified function cmd/tossctl/instrument_names.go:buildConsoleAccountBroker
[logic-map] missing evidence for modified function cmd/tossctl/limitsettings.go:consoleLimitSettings.Save
[logic-map] missing evidence for modified function cmd/tossctl/limitsettings.go:quotedOrUnset
[logic-map] missing evidence for modified function cmd/tossctl/main.go:main
[logic-map] missing evidence for modified function cmd/tossctl/marketschedule_test.go:TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef
[logic-map] missing evidence for modified function cmd/tossctl/marketschedule_test.go:TestConsoleBrokerTypedMarketCalendarFailsClosed
[logic-map] missing evidence for modified function cmd/tossctl/mcp.go:newMCPCmd
[logic-map] missing evidence for modified function cmd/tossctl/monitor.go:newMonitorCmd
[logic-map] missing evidence for modified function cmd/tossctl/openapi.go:validateOpenAPICredentials
[logic-map] missing evidence for modified function cmd/tossctl/openapi.go:newOpenAPICmd
[logic-map] missing evidence for modified function cmd/tossctl/root.go:newRootCmd
[logic-map] missing evidence for modified function cmd/tossctl/root.go:newAppContext
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:newSoakAttestCmd
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:runSoakRun
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:runSoakAttest
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:buildSoakWiring
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:soakSurveyedBase
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:pointSoakAt
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:TestSoakAttestWritesAVerifiableAttestation
[logic-map] missing evidence for modified function cmd/tossctl/system_update_wiring_test.go:TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:newVerifyCmd
[logic-map] cmd-tossctl--newverifyruncmd: AST hash does not match modified function revision
[logic-map] cmd-tossctl--runverifyrun: AST hash does not match modified function revision
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:runVerifyAbort
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:buildVerifyBroker
[logic-map] missing evidence for modified function cmd/tossctl/verify_test.go:pointVerifyAt
[logic-map] missing evidence for modified function cmd/tossctl/verify_test.go:seedVerifyRecord
[logic-map] missing evidence for modified function internal/app/app.go:New
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestNotificationSettingsAreAudited
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestTheAuditTrailCarriesNoNotificationSecret
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestARefusedNotificationBlockIsAudited
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestExistingAuditEntriesKeepTheirOrder
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.judgeHoldings
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.adopt
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.alertUnmanaged
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.checkExternalIncrease
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:StartAlertControlServer
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:AlertControlServer.Close
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:publishPrivateDescriptor
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.cycle
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.deliverOne
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.release
[logic-map] missing evidence for modified function internal/app/engine/auxiliary.go:Context.AlertDeliverer
[logic-map] missing evidence for modified function internal/app/engine/engine.go:NewOrderPath
[logic-map] missing evidence for modified function internal/app/engine/engine.go:NewContext
[logic-map] missing evidence for modified function internal/app/engine/engine_test.go:seedExpiredUnconsumedReservation
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:NewExitObserver
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.Run
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.ObserveOnce
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.workingSet
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.openState
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.observe
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.checkOutage
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.judge
[logic-map] internal-app-engine--exitobserver.record: AST hash does not match modified function revision
[logic-map] internal-app-engine--exitobserver.submit: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.release
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.clearTheSymbol
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.applyFloor
[logic-map] missing evidence for modified function internal/app/engine/exitloop_test.go:TestABreachDisplacesAnOutstandingTakeProfit
[logic-map] missing evidence for modified function internal/app/engine/exitwiring.go:Context.ExitObserver
[logic-map] missing evidence for modified function internal/app/engine/gateway.go:bindApplyHooks
[logic-map] internal-app-engine--buildgateway: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/app/engine/guardian_production_test.go:TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:riskPolicyFromAutomationGate
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:newProductionRiskGuardian
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:defaultProductionGuardianFactory
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring_internal_test.go:TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:newAutomationStatus
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:acceptanceDetail
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:verifyGate
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:recordGateSettings
[logic-map] missing evidence for modified function internal/app/engine/interlock_test.go:stubGuardian.ExposureLimits
[logic-map] missing evidence for modified function internal/app/engine/position_policy_command.go:NewPositionPolicyCommandService
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:StartPositionPolicyRuntimeServer
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:PositionPolicyRuntimeServer.Close
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:writePositionPolicyRuntimeDescriptor
[logic-map] missing evidence for modified function internal/app/engine/position_policy_transport.go:StartPositionPolicyCommandServer
[logic-map] missing evidence for modified function internal/app/engine/position_policy_transport.go:writePositionPolicyDescriptor
[logic-map] missing evidence for modified function internal/app/engine/reconcileloop.go:NewReconcileDriver
[logic-map] missing evidence for modified function internal/app/engine/reconcileloop.go:Context.ReconcileDriver
[logic-map] missing evidence for modified function internal/app/engine/runtime.go:Runtime.escalate
[logic-map] missing evidence for modified function internal/app/engine/runtime_wiring.go:Context.Recovery
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:strategyAccountAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:newProductionStrategyFirstLegAuthorityLoader
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:strategyFirstLegBindingDigest
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority_test.go:TestProductionFirstLegAuthorityLoaderPairedKRUS
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle.go:newStrategyDispatchCycle
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle.go:strategyDispatchCycle.dispatch
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:strategyDispatchGatewaySpy.ObserveStrategyEntryGate
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:pairedStrategyDispatchCycleFixture
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.NewPairedStrategyEntryProductionAssembly
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.NewRefreshingPairedStrategyEntrySupervisor
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.productionStrategyWorker
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:buildProductionStrategyMarketWorker
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.runProductionStrategyMarketCycle
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.refreshPairedStrategyEntryProductionAssembly
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:NewStrategyEntrySupervisor
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:StrategyEntrySupervisor.Snapshot
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:StrategyEntrySupervisor.runMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission.go:strategyFirstLegAdmissionBridge.admit
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:pairedStrategyFirstLegResults
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:allSixStrategyFirstLegResults
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityPair.ResultAuthority
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityLoader.collect
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:TestStrategyProposalAuthorityLoadsKRUSConcurrently
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:TestStrategyProposalAuthorityKeepsMarketFailureLocal
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:proposalRoutePair
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority.go:strategyRiskAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority_test.go:newStrategyRiskLoaderFixture
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority_test.go:createStrategyRiskLoaderJournal
[logic-map] missing evidence for modified function internal/app/engine/strategy_route_authority.go:strategyRouteAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_runtime_projection.go:Context.Read
[logic-map] missing evidence for modified function internal/app/engine/strategy_runtime_projection.go:strategyProjectionFromAssembly
[logic-map] missing evidence for modified function internal/app/engine/tracer.go:NewTracer
[logic-map] missing evidence for modified function internal/app/engine/tracer_test.go:newTracer
[logic-map] missing evidence for modified function internal/audit/audit.go:Log.Record
[logic-map] missing evidence for modified function internal/client/account.go:Client.ListPendingOrders
[logic-map] missing evidence for modified function internal/client/completed_orders.go:Client.ListCompletedOrdersRange
[logic-map] missing evidence for modified function internal/client/completed_orders_test.go:TestListCompletedOrdersParsesStructuredFields
[logic-map] missing evidence for modified function internal/client/orders_parse_test.go:TestListPendingOrdersParsesStructuredFields
[logic-map] missing evidence for modified function internal/client/watchlist.go:Client.ListWatchlist
[logic-map] missing evidence for modified function internal/client/watchlist_manage.go:Client.ListWatchlistGroups
[logic-map] missing evidence for modified function internal/client/watchlist_test.go:TestListWatchlistFromFixtures
[logic-map] missing evidence for modified function internal/client/watchlist_test.go:TestListWatchlistSendsTypesFilter
[logic-map] missing evidence for modified function internal/config/engine.go:mergeEngine
[logic-map] missing evidence for modified function internal/config/limits.go:GuardianLimits.Validate
[logic-map] missing evidence for modified function internal/config/limits_io.go:Service.LoadRawEngineGate
[logic-map] missing evidence for modified function internal/config/limits_io.go:Service.SaveEngineGateLimits
[logic-map] missing evidence for modified function internal/config/limits_io.go:gateMembersOf
[logic-map] missing evidence for modified function internal/config/limits_io.go:spliceGateLimits
[logic-map] missing evidence for modified function internal/console/a079_quarantine_release_test.go:TestReleaseErrorsAreMappedToOperatorLanguage
[logic-map] missing evidence for modified function internal/console/chrome.go:Console.chromeFor
[logic-map] missing evidence for modified function internal/console/chrome.go:holdingsSnapshot.freshness
[logic-map] missing evidence for modified function internal/console/chrome.go:ordersSnapshot.freshness
[logic-map] missing evidence for modified function internal/console/chrome.go:signalsView.freshness
[logic-map] missing evidence for modified function internal/console/console.go:Console.Serve
[logic-map] missing evidence for modified function internal/console/console.go:Console.settle
[logic-map] missing evidence for modified function internal/console/console.go:Console.routes
[logic-map] missing evidence for modified function internal/console/console_test.go:TestTheDashboardReportsAnUnstartedMachineWithoutFailing
[logic-map] missing evidence for modified function internal/console/data.go:Console.readAttestation
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:TestTheHeadingStepsAreDistinguishable
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:declaredFontSize
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:declaredBodyFontSize
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:TestTheSummaryCostsNoBrokerCall
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.handleQuarantineReleasePreview
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.handleQuarantineReleaseApply
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.writeQuarantineError
[logic-map] missing evidence for modified function internal/console/optimization.go:Console.handleOptimization
[logic-map] missing evidence for modified function internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsContainExactMarketMetricMatrix
[logic-map] missing evidence for modified function internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed
[logic-map] missing evidence for modified function internal/console/optimization_view.go:optimizationCategoryViews
[logic-map] missing evidence for modified function internal/console/pages.go:Console.handleDashboard
[logic-map] missing evidence for modified function internal/console/pages.go:Console.handleReport
[logic-map] missing evidence for modified function internal/console/performance_history.go:Console.handlePerformanceHistory
[logic-map] missing evidence for modified function internal/console/performance_history_test.go:fakePerformanceReader.Dashboard
[logic-map] missing evidence for modified function internal/console/run.go:Console.startRun
[logic-map] missing evidence for modified function internal/console/settings_cadence_test.go:TestTheSubmitContractIsUnchanged
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestEveryCardShowsWhatChangesAndWhen
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestTheCurrencyConsequenceIsRenderedAndNotFolded
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestTheSettingsScreensAskForNoTyping
[logic-map] missing evidence for modified function internal/console/settings_limits.go:settingsPage.LimitRows
[logic-map] missing evidence for modified function internal/console/settings_limits.go:Console.handleSettingsLimitPreset
[logic-map] missing evidence for modified function internal/console/settings_limits.go:Console.handleSettingsLimits
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:fakeLimits.Save
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:TestThePresetControlsAskForNoTyping
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:TestTheAdvancedFormSavesIndividualValues
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:limitCurrent
[logic-map] missing evidence for modified function internal/console/settings_tabs.go:Console.strategyRuntimeSummary
[logic-map] missing evidence for modified function internal/console/static_test.go:registeredRoutes
[logic-map] missing evidence for modified function internal/console/static_test.go:TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate
[logic-map] missing evidence for modified function internal/console/strategy_runtime_multimarket.go:Console.buildMultiMarketStrategyRuntimePage
[logic-map] missing evidence for modified function internal/console/strategy_runtime_multimarket.go:multiMarketStrategyRuntimePage.project
[logic-map] missing evidence for modified function internal/console/trading_views_test.go:TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA
[logic-map] missing evidence for modified function internal/execgw/a098_sender_down_reason_test.go:TestTheSenderDownReasonIsRegisteredInTheEnumeration
[logic-map] missing evidence for modified function internal/execgw/classify.go:classifyMutation
[logic-map] missing evidence for modified function internal/execgw/failclosed.go:AllReasonCodes
[logic-map] missing evidence for modified function internal/execgw/gateway.go:New
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.submit
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkSymbolFree
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkEntry
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkReservation
[logic-map] missing evidence for modified function internal/execgw/gateway_test.go:holdExposure
[logic-map] internal-execgw--testunreachablebrokerisnotdispatched: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/execgw/modegate.go:EntryGate.ProjectOperatingMode
[logic-map] missing evidence for modified function internal/execgw/observation_test.go:newGuardianOn
[logic-map] missing evidence for modified function internal/execgw/replay.go:Gateway.ReplayInDoubt
[logic-map] missing evidence for modified function internal/execgw/retry.go:Retrier.Query
[logic-map] missing evidence for modified function internal/execgw/retry.go:Retrier.escalateCredentialFailure
[logic-map] missing evidence for modified function internal/execgw/retry.go:EntryGate.Clear
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:NewRiskGuardian
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.LimitsDigest
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.IssueEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.escalateFor
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalAccountBaseGuardianPairedKRUS
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:qFinalCampaignFirstLegIssuanceFixture
[logic-map] missing evidence for modified function internal/execgw/riskguardian_first_leg.go:RiskGuardian.PrecheckQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/execgw/riskguardian_first_leg.go:RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal.go:RiskGuardian.PrecheckQFinalEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal.go:RiskGuardian.IssuePrecheckedQFinalEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal_test.go:TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal_test.go:qFinalKRRequest
[logic-map] missing evidence for modified function internal/execgw/riskguardian_test.go:newGuardian
[logic-map] missing evidence for modified function internal/execgw/riskguardian_test.go:mustGuardian
[logic-map] missing evidence for modified function internal/execgw/roundtrip.go:Gateway.confirmCreatedOrder
[logic-map] missing evidence for modified function internal/execgw/strategy_entry_gate_authority.go:Gateway.ObserveStrategyEntryGate
[logic-map] missing evidence for modified function internal/execgw/strategy_entry_gate_authority.go:Gateway.withStrategyEntryGateAuthority
[logic-map] missing evidence for modified function internal/httpapi/router.go:router.read
[logic-map] missing evidence for modified function internal/httpapi/strategy_runtime.go:StrategyRuntimeSnapshotFunc
[logic-map] missing evidence for modified function internal/journal/alert_claim.go:Journal.settleUnderClaim
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:Journal.runApplyHooks
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:Journal.ResolveExitProposal
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:ApplyExitFill
[logic-map] missing evidence for modified function internal/journal/exit_snapshot.go:knownArmSuppression
[logic-map] missing evidence for modified function internal/journal/exit_state.go:Journal.recordExitJudgementTx
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestARefusalReArmsTheLevel
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestResolvingNothingIsNotAnError
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestACancelledRungIsProposableAgain
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestARefusedProposalIsReArmedAfterARestart
[logic-map] missing evidence for modified function internal/journal/fills.go:Journal.RecordFill
[logic-map] missing evidence for modified function internal/journal/fills.go:Journal.LiveOrdersForSymbol
[logic-map] missing evidence for modified function internal/journal/issuance_test.go:issueRequest
[logic-map] missing evidence for modified function internal/journal/observation_crash_linux_test.go:crashChildObservationGap
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.TransitionOperatingMode
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.RestoreOperatingModeProjection
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.OperatingModeHistory
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:currentModeFromRow
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:scanOperatingMode
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:Journal.LinkCampaignOrder
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:ApplyPositionCampaignFill
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:updateCampaignSuccessorRemaining
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignExposureBlockedInTx
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignQuantity
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignRemaining
[logic-map] missing evidence for modified function internal/journal/position_campaign_test.go:TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce
[logic-map] missing evidence for modified function internal/journal/readonly_test.go:TestTheReadOnlyHandleHasNoWriteMethods
[logic-map] missing evidence for modified function internal/journal/reservations.go:reserveRows
[logic-map] missing evidence for modified function internal/journal/reservations_test.go:exposureReserve
[logic-map] missing evidence for modified function internal/journal/reservations_test.go:TestFractionalShareAmountsSurviveTheLedger
[logic-map] missing evidence for modified function internal/journal/risk_bucket.go:Journal.CommitRiskBucketAdmission
[logic-map] missing evidence for modified function internal/journal/risk_bucket.go:ensureRiskBucketEntryScopeClean
[logic-map] missing evidence for modified function internal/journal/risk_bucket_admission_test.go:TestRiskBucketAdmissionRejectsImmutablePolicyAndSnapshotCollision
[logic-map] missing evidence for modified function internal/journal/risk_bucket_admission_test.go:seedExistingRiskReservation
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:loadRiskBucketOrderAuthority
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:riskBucketOrderRemaining
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:loadRiskBucketFillTransition
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:verifyRiskBucketStateDigest
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketAmbiguousSidecarLatchesAllApplicableReservationsWithoutDroppingFill
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketActualAndReleaseRequireExactOwnerDecisionScopeForCollidingOrderID
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:commitFreshRiskBucketAdmissionTx
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:insertFreshRiskBucketReservations
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:Journal.RevalidateQFinalAdmission
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance_test.go:qFinalIssueFixture
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner.go:Journal.applyRiskBucketOwnerBindingInTx
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner.go:Journal.releaseRiskBucketOwner
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner_test.go:seedRiskBucketOwnerLifecycle
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner_test.go:closeRiskBucketOwnerLifecycle
[logic-map] missing evidence for modified function internal/journal/schema_test.go:TestSchemaTablesAndColumns
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime.go:linkConfirmedStrategyCampaignTx
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime.go:Journal.RequireCurrentStrategyDispatchTransportAuthority
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime_test.go:prepareStrategyDispatchLease
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime_test.go:openStrategyDispatchV25Journal
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic.go:Journal.RecordQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic_test.go:TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic_test.go:firstLegAtomicFixture
[logic-map] missing evidence for modified function internal/journal/strategy_lineage.go:insertExactStrategyDecision
[logic-map] missing evidence for modified function internal/journal/strategy_lineage.go:verifyStrategyRiskBinding
[logic-map] missing evidence for modified function internal/journal/strategy_weekly_first_leg_test.go:weeklyFirstLegFixture
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:ProjectAcceptedStrategyflowLineage
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:verifyProjectionRiskIntent
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:strategyflowDecisionIdentity
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:strategyflowLineageFromProjection
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:decodeStrategyflowRiskBinding
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestVerifyStrategyflowRiskBindingV2Compatibility
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:strategyflowFirstLegFixture
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestRecordQFinalCampaignFirstLegRejectsStrategyflowTamperWithoutRows
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestVerifyAcceptedStrategyflowProjectionRejectsOuterCanonicalTamper
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:freezeTradeOutcomeTx
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:computeTradeOutcome
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:adoptedRoundTripLegs
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:Journal.BackfillTradeOutcome
[logic-map] missing evidence for modified function internal/mcp/server.go:toolResult
[logic-map] missing evidence for modified function internal/mcp/server.go:trim
[logic-map] missing evidence for modified function internal/mcp/server_test.go:TestToolResultKeepsSmallPayload
[logic-map] missing evidence for modified function internal/measure/counterfactual.go:gridPolicy
[logic-map] missing evidence for modified function internal/measure/reconstruct_test.go:issueEntry
[logic-map] missing evidence for modified function internal/obs/a096_one_send_per_condition_test.go:TestAcknowledgeCannotClearTheGateMidSend
[logic-map] missing evidence for modified function internal/obs/a099_round4_test.go:TestAHeldRowIsNotWhispered
[logic-map] missing evidence for modified function internal/obs/log.go:NewLogger
[logic-map] missing evidence for modified function internal/obs/mode.go:Notifier.AnnounceOperatingMode
[logic-map] missing evidence for modified function internal/obs/mode_test.go:TestTheTransitionLogLineIsCountable
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.publishBestEffort
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.notifyCritical
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.claimAndDeliver
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.escalate
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.deliver
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.logLeaseLost
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.logClaimHeld
[logic-map] missing evidence for modified function internal/obs/notifier.go:renderFields
[logic-map] missing evidence for modified function internal/obs/ntfy.go:Ntfy.Publish
[logic-map] missing evidence for modified function internal/official/candle_reads.go:adaptCandles
[logic-map] missing evidence for modified function internal/official/candle_reads_test.go:TestAdaptCandlesEmpty
[logic-map] missing evidence for modified function internal/official/client.go:Client.hcTransport
[logic-map] internal-official--client.dorequest: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/official/market_reads.go:adaptPrices
[logic-map] missing evidence for modified function internal/official/reads.go:Client.accountsLocked
[logic-map] missing evidence for modified function internal/official/stock_reads.go:adaptStocks
[logic-map] missing evidence for modified function internal/official/token.go:tokenManager.exchange
[logic-map] missing evidence for modified function internal/ops/wts_operations.go:wtsOperations
[logic-map] missing evidence for modified function internal/performance/attribution_persistence.go:validateAttributionGeneration
[logic-map] missing evidence for modified function internal/performance/attribution_persistence_test.go:TestAttributionEvidenceRejectsHeadEnvelopeMismatch
[logic-map] missing evidence for modified function internal/positioncampaign/campaign.go:TransitionLeg
[logic-map] missing evidence for modified function internal/positioncampaign/command.go:ValidateCommand
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:Replay
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:replayCampaignTransition
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:replayLegFillTransition
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:remainingFromCap
[logic-map] missing evidence for modified function internal/positioncampaign/review_regression_test.go:TestProductionCoreContainsNoLaneRatiosOrBrokerDependency
[logic-map] missing evidence for modified function internal/positioncampaign/watermark.go:LegLedger.Observe
[logic-map] missing evidence for modified function internal/reconcile/recovery.go:Recovery.Run
[logic-map] missing evidence for modified function internal/risk/chain_test.go:usdPolicy
[logic-map] missing evidence for modified function internal/risk/chain_test.go:TestOrderAboveThePerOrderCapIsRefused
[logic-map] missing evidence for modified function internal/risk/contract_test.go:TestQuantityAboveTheRiskBudgetIsRefused
[logic-map] missing evidence for modified function internal/risk/input.go:DefaultPolicy
[logic-map] missing evidence for modified function internal/risk/input.go:Policy.Validate
[logic-map] missing evidence for modified function internal/riskbucket/fill.go:recomputeOverageLatches
[logic-map] missing evidence for modified function internal/riskbucket/fill.go:cloneFillState
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:LoadProductionRiskSnapshotAuthority
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:validProductionRiskPolicyContents
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:bindProductionRiskInputs
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:loadProductionRiskEntries
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:readProductionRiskUsage
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:aggregateProductionRiskUsage
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority_test.go:newProductionRiskFixture
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority_test.go:createProductionRiskDB
[logic-map] missing evidence for modified function internal/scheduler/budget.go:BudgetCoordinator.TryAcquire
[logic-map] missing evidence for modified function internal/scheduler/budget.go:BudgetCoordinator.Complete
[logic-map] missing evidence for modified function internal/soak/attest.go:Summary.Evaluate
[logic-map] missing evidence for modified function internal/soak/attest.go:BuildAttestation
[logic-map] missing evidence for modified function internal/soak/attest_test.go:TestBuildAttestationRefusesAnIncompleteSoak
[logic-map] missing evidence for modified function internal/soak/record.go:OpenRecorder
[logic-map] missing evidence for modified function internal/soak/record.go:Recorder.Append
[logic-map] missing evidence for modified function internal/soak/record.go:LoadCycles
[logic-map] missing evidence for modified function internal/strategycandle/adapter_test.go:officialPage
[logic-map] missing evidence for modified function internal/strategyengine/lane_test.go:verifiedBarFixtureAt
[logic-map] missing evidence for modified function internal/strategyengine/wiring_test.go:TestA047ShipsNoRuntimeOrderOrExitWiring
[logic-map] missing evidence for modified function internal/strategyengine/wiring_test.go:assertConsoleOnlyCallsDormantDescriptor
[logic-map] missing evidence for modified function internal/strategyevidence/model.go:kindSupportsMarket
[logic-map] missing evidence for modified function internal/strategyevidence/model.go:validateTypedPayload
[logic-map] missing evidence for modified function internal/strategyevidence/official_source.go:NewOfficialAdapter
[logic-map] missing evidence for modified function internal/strategyevidence/official_source.go:OfficialAdapter.collectSEC
[logic-map] missing evidence for modified function internal/strategyevidence/projection.go:Project
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:SourcePolicy.validate
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:NewAdapter
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:Adapter.Fetch
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestIncompleteOrUnverifiedPolicyMakesZeroCalls
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterKeepsCredentialOutOfRequestMetadata
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterRejectsRuntimeBoundExpansionBeforeTransport
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterEnforcesResponseBytesWindowAndRetryAfter
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest
[logic-map] missing evidence for modified function internal/strategyflow/adapters.go:defaultRegistry
[logic-map] missing evidence for modified function internal/strategyflow/adapters.go:proposalRegistry
[logic-map] missing evidence for modified function internal/strategyflow/canonical_projection_test.go:acceptedProjectionFixture
[logic-map] missing evidence for modified function internal/strategyflow/canonical_projection_test.go:TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether
[logic-map] missing evidence for modified function internal/strategyflow/execution_terms_test.go:TestEvaluateSealsExactExecutionTerms
[logic-map] missing evidence for modified function internal/strategyflow/execution_terms_test.go:TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:Evaluate
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:Propose
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:evaluateWith
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestPairedRegistryCoversKRUSContinuationReversalAndWeekly
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestCompleteLineageSealDetectsMutation
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:inputFor
[logic-map] missing evidence for modified function internal/strategyflow/production_integration_test.go:TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators
[logic-map] missing evidence for modified function internal/strategyflow/registry.go:LaneInput.matches
[logic-map] missing evidence for modified function internal/strategymarket/bars.go:aggregateClosedKRXFiveMinute
[logic-map] missing evidence for modified function internal/strategymarket/bars_test.go:minutes
[logic-map] missing evidence for modified function internal/strategymarket/bars_test.go:TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:DormantSnapshot
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:UnavailableSnapshot
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:Clone
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:Validate
[logic-map] missing evidence for modified function internal/strategyprojection/projection_test.go:currentPair
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/a108_publication_is_total_test.go:TestProjectionLivenessClausesEachDecideOnTheirOwn
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport.go:Client.Read
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:reclaimStaleControlDirectory
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:verifyStaleSocketShape
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:projectionSocketAccepts
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:Dial
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix_test.go:TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.Len
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.ManifestDigest
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.For
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:LoadProductionAuthorityBatch
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:buildLaneInput
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:validScopes
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:laneMatchesMarket
[logic-map] missing evidence for modified function internal/strategyproposal/production_test.go:productionFixture
[logic-map] missing evidence for modified function internal/strategyproposal/production_testseam.go:ProductionBatchAuthorityForTest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:ProductionRouteAuthority.OwnerDigest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:LoadProductionRouteAuthorityBatch
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:verifyProductionRouteManifest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:validProductionRouteCandidates
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:productionRouteDescriptors
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:openProductionRouteSnapshot
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:loadProductionRouteOwnersFrom
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthorityFailureIsMarketLocal
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthoritySelectsEverySignedSymbolScope
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:newProductionRouteFixture
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:productionRouteFixture.body
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:productionRouteFixture.write
[logic-map] missing evidence for modified function internal/verifylive/record.go:OpenRecorder
[logic-map] missing evidence for modified function internal/verifylive/record.go:Recorder.Append
[logic-map] missing evidence for modified function internal/verifylive/record.go:LoadEntries
[logic-map] missing evidence for modified function internal/verifylive/record.go:Artifact.terminal
[logic-map] missing evidence for modified function internal/verifylive/record_test.go:TestRecordCarriesNoUnmaskedAccountNumber
[logic-map] internal-verifylive--buildreport: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/verifylive/report.go:Report.WriteText
[logic-map] internal-verifylive--buildprogress: AST hash does not match modified function revision
[logic-map] internal-verifylive--progress.writetext: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:StrategyflowKRFixture
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:StrategyflowUSFixture
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:strategyflowEvaluation
rc=1
```

## 원문 출력 — 후 (번들 11개 갱신, 커밋 전 워킹트리)

```text
[logic-map] a100-wire-fill-to-broker-protection: base 882a0b490b0b → working tree (no landed-commit.txt in HEAD) required 534 function(s) — judged at HEAD c03928b639df
[logic-map] a100-wire-fill-to-broker-protection: the target is the working tree, so this window also holds 1380 commit(s) that landed after the base and every existing function they changed is required here too — `--record-landing` cannot narrow it: the working tree has uncommitted changes to tracked files — commit them first, because a recorded landing points at a commit and step 5 would then never compare those edits
[logic-map] internal-app-engine--buildgateway: AST source hash is stale: internal/app/engine/gateway.go
[logic-map] internal-app-engine--exitobserver.record: AST source hash is stale: internal/app/engine/exitloop.go
[logic-map] internal-app-engine--exitobserver.submit: AST source hash is stale: internal/app/engine/exitloop.go
[logic-map] internal-official--client.dorequest: AST source hash is stale: internal/official/client.go
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061RunAndConsoleVerificationDoNotBuildBrokerWhileTheProfileBudgetIsOccupied
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061VerificationRefusesBeforeBrokerWhenItsIntentCannotBePublished
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061RunVerifyAbortPublishesItsMarkerBeforeWaitingForTheProfileBudget
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061AbortListAndRefreshedEmptyRecordNeverBuildABroker
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061ActiveVerificationExclusionRefusesAbortWithoutErasingItsMarker
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061AbortReloadsOutstandingTargetsAfterExclusiveAdmission
[logic-map] missing evidence for modified function cmd/tossctl/a061_history_instrument_names_test.go:TestA061MetadataAndAccountReadsShareOneOfficialClient
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_engine_registers_a_sender_test.go:TestProductionRuntimeStartsExactlyOneAlertDeliverer
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_operator_command_names_a_person_test.go:TestTheAlertCommandsAreWiredUnderEngine
[logic-map] missing evidence for modified function cmd/tossctl/a098_the_operator_command_names_a_person_test.go:TestTheCommandsRefuseWhenNoEngineIsRunning
[logic-map] missing evidence for modified function cmd/tossctl/a108_the_aggregate_snapshot_outlives_the_strategy_read_test.go:a108SnapshotReader
[logic-map] missing evidence for modified function cmd/tossctl/a111_httpapi_exit_freshness_test.go:a111NewHTTPRouteSession
[logic-map] missing evidence for modified function cmd/tossctl/candidate.go:candidateWiring
[logic-map] missing evidence for modified function cmd/tossctl/candidate_test.go:withCandidateFixture
[logic-map] missing evidence for modified function cmd/tossctl/candidatepanel.go:buildCandidatePanel
[logic-map] missing evidence for modified function cmd/tossctl/console.go:runConsole
[logic-map] missing evidence for modified function cmd/tossctl/console.go:newConsoleBroker
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.resolve
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.instrumentMetadata
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleBroker.TypedMarketCalendar
[logic-map] missing evidence for modified function cmd/tossctl/console.go:lazyHoldings.Holdings
[logic-map] missing evidence for modified function cmd/tossctl/console.go:lazyOrders.Orders
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleRelaunch
[logic-map] missing evidence for modified function cmd/tossctl/console.go:consoleVerifyStarter
[logic-map] missing evidence for modified function cmd/tossctl/console_openapi.go:defaultConsoleOpenAPIDeps
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestOneRefreshAsksTheOpenGroupAndTheClosedGroupSeparatelyAndTheLiveOneWhole
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestTheOrdersSeamCarriesEachListsOutcomeSeparately
[logic-map] missing evidence for modified function cmd/tossctl/console_test.go:TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce
[logic-map] missing evidence for modified function cmd/tossctl/doctor_openapi.go:computeDoctorOpenAPISummary
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:newEngineCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:newEngineRunCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:runEngineRun
[logic-map] missing evidence for modified function cmd/tossctl/engine.go:reportStrategyProjectionDegraded
[logic-map] missing evidence for modified function cmd/tossctl/engine_account_seq_recovery_test.go:TestActualEngineRecoveryStillFailsClosedOnASnapshot429
[logic-map] missing evidence for modified function cmd/tossctl/engine_alerts.go:newEngineAlertsAckCmd
[logic-map] missing evidence for modified function cmd/tossctl/engine_assembly.go:assembleEngine
[logic-map] missing evidence for modified function cmd/tossctl/engine_assembly_test.go:TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver
[logic-map] missing evidence for modified function cmd/tossctl/engine_reconcile.go:buildReconcileResolveRuntime
[logic-map] missing evidence for modified function cmd/tossctl/engine_test.go:TestTheExitObserverDefersToFillDetection
[logic-map] missing evidence for modified function cmd/tossctl/flatten.go:runFlatten
[logic-map] missing evidence for modified function cmd/tossctl/flatten.go:buildFlattenWiring
[logic-map] missing evidence for modified function cmd/tossctl/help_convention_test.go:TestMutatingAnnotationOnTradeCommands
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:runHTTPAPI
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:strategyRuntimeReaderFor
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:openHTTPAPIResources
[logic-map] missing evidence for modified function cmd/tossctl/httpapi.go:publishHTTPAPISnapshots
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.readPositionBrokerSnapshot
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.Performance
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader.go:httpAPIReader.Snapshot
[logic-map] missing evidence for modified function cmd/tossctl/httpapi_reader_test.go:TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources
[logic-map] missing evidence for modified function cmd/tossctl/init.go:runInitOfficialFlow
[logic-map] missing evidence for modified function cmd/tossctl/instrument_names.go:lazyInstrumentNames.Names
[logic-map] missing evidence for modified function cmd/tossctl/instrument_names.go:buildConsoleAccountBroker
[logic-map] missing evidence for modified function cmd/tossctl/limitsettings.go:consoleLimitSettings.Save
[logic-map] missing evidence for modified function cmd/tossctl/limitsettings.go:quotedOrUnset
[logic-map] missing evidence for modified function cmd/tossctl/main.go:main
[logic-map] missing evidence for modified function cmd/tossctl/marketschedule_test.go:TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef
[logic-map] missing evidence for modified function cmd/tossctl/marketschedule_test.go:TestConsoleBrokerTypedMarketCalendarFailsClosed
[logic-map] missing evidence for modified function cmd/tossctl/mcp.go:newMCPCmd
[logic-map] missing evidence for modified function cmd/tossctl/monitor.go:newMonitorCmd
[logic-map] missing evidence for modified function cmd/tossctl/openapi.go:validateOpenAPICredentials
[logic-map] missing evidence for modified function cmd/tossctl/openapi.go:newOpenAPICmd
[logic-map] missing evidence for modified function cmd/tossctl/root.go:newRootCmd
[logic-map] missing evidence for modified function cmd/tossctl/root.go:newAppContext
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:newSoakAttestCmd
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:runSoakRun
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:runSoakAttest
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:buildSoakWiring
[logic-map] missing evidence for modified function cmd/tossctl/soak.go:soakSurveyedBase
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:pointSoakAt
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing
[logic-map] missing evidence for modified function cmd/tossctl/soak_test.go:TestSoakAttestWritesAVerifiableAttestation
[logic-map] missing evidence for modified function cmd/tossctl/system_update_wiring_test.go:TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:newVerifyCmd
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:runVerifyAbort
[logic-map] missing evidence for modified function cmd/tossctl/verify.go:buildVerifyBroker
[logic-map] missing evidence for modified function cmd/tossctl/verify_test.go:pointVerifyAt
[logic-map] missing evidence for modified function cmd/tossctl/verify_test.go:seedVerifyRecord
[logic-map] missing evidence for modified function internal/app/app.go:New
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestNotificationSettingsAreAudited
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestTheAuditTrailCarriesNoNotificationSecret
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestARefusedNotificationBlockIsAudited
[logic-map] missing evidence for modified function internal/app/engine/a074_notification_audit_test.go:TestExistingAuditEntriesKeepTheirOrder
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.judgeHoldings
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.adopt
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.alertUnmanaged
[logic-map] missing evidence for modified function internal/app/engine/adoption.go:ReconcileDriver.checkExternalIncrease
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:StartAlertControlServer
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:AlertControlServer.Close
[logic-map] missing evidence for modified function internal/app/engine/alert_control_transport_unix.go:publishPrivateDescriptor
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.cycle
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.deliverOne
[logic-map] missing evidence for modified function internal/app/engine/alertdelivery.go:alertDeliverer.release
[logic-map] missing evidence for modified function internal/app/engine/auxiliary.go:Context.AlertDeliverer
[logic-map] missing evidence for modified function internal/app/engine/engine.go:NewOrderPath
[logic-map] missing evidence for modified function internal/app/engine/engine.go:NewContext
[logic-map] missing evidence for modified function internal/app/engine/engine_test.go:seedExpiredUnconsumedReservation
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:NewExitObserver
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.Run
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.ObserveOnce
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.workingSet
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.openState
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.observe
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.checkOutage
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.judge
[logic-map] internal-app-engine--exitobserver.record: AST hash does not match modified function revision
[logic-map] internal-app-engine--exitobserver.submit: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.release
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.clearTheSymbol
[logic-map] missing evidence for modified function internal/app/engine/exitloop.go:ExitObserver.applyFloor
[logic-map] missing evidence for modified function internal/app/engine/exitloop_test.go:TestABreachDisplacesAnOutstandingTakeProfit
[logic-map] missing evidence for modified function internal/app/engine/exitwiring.go:Context.ExitObserver
[logic-map] missing evidence for modified function internal/app/engine/gateway.go:bindApplyHooks
[logic-map] internal-app-engine--buildgateway: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/app/engine/guardian_production_test.go:TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:riskPolicyFromAutomationGate
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:newProductionRiskGuardian
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring.go:defaultProductionGuardianFactory
[logic-map] missing evidence for modified function internal/app/engine/guardian_wiring_internal_test.go:TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:newAutomationStatus
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:acceptanceDetail
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:verifyGate
[logic-map] missing evidence for modified function internal/app/engine/interlock.go:recordGateSettings
[logic-map] missing evidence for modified function internal/app/engine/interlock_test.go:stubGuardian.ExposureLimits
[logic-map] missing evidence for modified function internal/app/engine/position_policy_command.go:NewPositionPolicyCommandService
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:StartPositionPolicyRuntimeServer
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:PositionPolicyRuntimeServer.Close
[logic-map] missing evidence for modified function internal/app/engine/position_policy_runtime_transport_unix.go:writePositionPolicyRuntimeDescriptor
[logic-map] missing evidence for modified function internal/app/engine/position_policy_transport.go:StartPositionPolicyCommandServer
[logic-map] missing evidence for modified function internal/app/engine/position_policy_transport.go:writePositionPolicyDescriptor
[logic-map] missing evidence for modified function internal/app/engine/reconcileloop.go:NewReconcileDriver
[logic-map] missing evidence for modified function internal/app/engine/reconcileloop.go:Context.ReconcileDriver
[logic-map] missing evidence for modified function internal/app/engine/runtime.go:Runtime.escalate
[logic-map] missing evidence for modified function internal/app/engine/runtime_wiring.go:Context.Recovery
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:strategyAccountAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:newProductionStrategyFirstLegAuthorityLoader
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority.go:strategyFirstLegBindingDigest
[logic-map] missing evidence for modified function internal/app/engine/strategy_account_first_leg_authority_test.go:TestProductionFirstLegAuthorityLoaderPairedKRUS
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle.go:newStrategyDispatchCycle
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle.go:strategyDispatchCycle.dispatch
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:strategyDispatchGatewaySpy.ObserveStrategyEntryGate
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS
[logic-map] missing evidence for modified function internal/app/engine/strategy_dispatch_cycle_test.go:pairedStrategyDispatchCycleFixture
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.NewPairedStrategyEntryProductionAssembly
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.NewRefreshingPairedStrategyEntrySupervisor
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.productionStrategyWorker
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:buildProductionStrategyMarketWorker
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.runProductionStrategyMarketCycle
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:Context.refreshPairedStrategyEntryProductionAssembly
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:NewStrategyEntrySupervisor
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:StrategyEntrySupervisor.Snapshot
[logic-map] missing evidence for modified function internal/app/engine/strategy_entry_supervisor.go:StrategyEntrySupervisor.runMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission.go:strategyFirstLegAdmissionBridge.admit
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:pairedStrategyFirstLegResults
[logic-map] missing evidence for modified function internal/app/engine/strategy_first_leg_admission_test.go:allSixStrategyFirstLegResults
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityPair.ResultAuthority
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityLoader.collect
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority.go:strategyProposalAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:TestStrategyProposalAuthorityLoadsKRUSConcurrently
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:TestStrategyProposalAuthorityKeepsMarketFailureLocal
[logic-map] missing evidence for modified function internal/app/engine/strategy_proposal_authority_test.go:proposalRoutePair
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority.go:strategyRiskAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority_test.go:newStrategyRiskLoaderFixture
[logic-map] missing evidence for modified function internal/app/engine/strategy_risk_authority_test.go:createStrategyRiskLoaderJournal
[logic-map] missing evidence for modified function internal/app/engine/strategy_route_authority.go:strategyRouteAuthorityLoader.collectMarket
[logic-map] missing evidence for modified function internal/app/engine/strategy_runtime_projection.go:Context.Read
[logic-map] missing evidence for modified function internal/app/engine/strategy_runtime_projection.go:strategyProjectionFromAssembly
[logic-map] missing evidence for modified function internal/app/engine/tracer.go:NewTracer
[logic-map] missing evidence for modified function internal/app/engine/tracer_test.go:newTracer
[logic-map] missing evidence for modified function internal/audit/audit.go:Log.Record
[logic-map] missing evidence for modified function internal/client/account.go:Client.ListPendingOrders
[logic-map] missing evidence for modified function internal/client/completed_orders.go:Client.ListCompletedOrdersRange
[logic-map] missing evidence for modified function internal/client/completed_orders_test.go:TestListCompletedOrdersParsesStructuredFields
[logic-map] missing evidence for modified function internal/client/orders_parse_test.go:TestListPendingOrdersParsesStructuredFields
[logic-map] missing evidence for modified function internal/client/watchlist.go:Client.ListWatchlist
[logic-map] missing evidence for modified function internal/client/watchlist_manage.go:Client.ListWatchlistGroups
[logic-map] missing evidence for modified function internal/client/watchlist_test.go:TestListWatchlistFromFixtures
[logic-map] missing evidence for modified function internal/client/watchlist_test.go:TestListWatchlistSendsTypesFilter
[logic-map] missing evidence for modified function internal/config/engine.go:mergeEngine
[logic-map] missing evidence for modified function internal/config/limits.go:GuardianLimits.Validate
[logic-map] missing evidence for modified function internal/config/limits_io.go:Service.LoadRawEngineGate
[logic-map] missing evidence for modified function internal/config/limits_io.go:Service.SaveEngineGateLimits
[logic-map] missing evidence for modified function internal/config/limits_io.go:gateMembersOf
[logic-map] missing evidence for modified function internal/config/limits_io.go:spliceGateLimits
[logic-map] missing evidence for modified function internal/console/a079_quarantine_release_test.go:TestReleaseErrorsAreMappedToOperatorLanguage
[logic-map] missing evidence for modified function internal/console/chrome.go:Console.chromeFor
[logic-map] missing evidence for modified function internal/console/chrome.go:holdingsSnapshot.freshness
[logic-map] missing evidence for modified function internal/console/chrome.go:ordersSnapshot.freshness
[logic-map] missing evidence for modified function internal/console/chrome.go:signalsView.freshness
[logic-map] missing evidence for modified function internal/console/console.go:Console.Serve
[logic-map] missing evidence for modified function internal/console/console.go:Console.settle
[logic-map] missing evidence for modified function internal/console/console.go:Console.routes
[logic-map] missing evidence for modified function internal/console/console_test.go:TestTheDashboardReportsAnUnstartedMachineWithoutFailing
[logic-map] missing evidence for modified function internal/console/data.go:Console.readAttestation
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:TestTheHeadingStepsAreDistinguishable
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:declaredFontSize
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:declaredBodyFontSize
[logic-map] missing evidence for modified function internal/console/display_primitives_test.go:TestTheSummaryCostsNoBrokerCall
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.handleQuarantineReleasePreview
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.handleQuarantineReleaseApply
[logic-map] missing evidence for modified function internal/console/exit_quarantine.go:Console.writeQuarantineError
[logic-map] missing evidence for modified function internal/console/optimization.go:Console.handleOptimization
[logic-map] missing evidence for modified function internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsContainExactMarketMetricMatrix
[logic-map] missing evidence for modified function internal/console/optimization_candidate_filters_test.go:TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls
[logic-map] missing evidence for modified function internal/console/optimization_ui_contract_test.go:TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed
[logic-map] missing evidence for modified function internal/console/optimization_view.go:optimizationCategoryViews
[logic-map] missing evidence for modified function internal/console/pages.go:Console.handleDashboard
[logic-map] missing evidence for modified function internal/console/pages.go:Console.handleReport
[logic-map] missing evidence for modified function internal/console/performance_history.go:Console.handlePerformanceHistory
[logic-map] missing evidence for modified function internal/console/performance_history_test.go:fakePerformanceReader.Dashboard
[logic-map] missing evidence for modified function internal/console/run.go:Console.startRun
[logic-map] missing evidence for modified function internal/console/settings_cadence_test.go:TestTheSubmitContractIsUnchanged
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestEveryCardShowsWhatChangesAndWhen
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestTheCurrencyConsequenceIsRenderedAndNotFolded
[logic-map] missing evidence for modified function internal/console/settings_card_test.go:TestTheSettingsScreensAskForNoTyping
[logic-map] missing evidence for modified function internal/console/settings_limits.go:settingsPage.LimitRows
[logic-map] missing evidence for modified function internal/console/settings_limits.go:Console.handleSettingsLimitPreset
[logic-map] missing evidence for modified function internal/console/settings_limits.go:Console.handleSettingsLimits
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:fakeLimits.Save
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:TestThePresetControlsAskForNoTyping
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:TestTheAdvancedFormSavesIndividualValues
[logic-map] missing evidence for modified function internal/console/settings_limits_test.go:limitCurrent
[logic-map] missing evidence for modified function internal/console/settings_tabs.go:Console.strategyRuntimeSummary
[logic-map] missing evidence for modified function internal/console/static_test.go:registeredRoutes
[logic-map] missing evidence for modified function internal/console/static_test.go:TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate
[logic-map] missing evidence for modified function internal/console/strategy_runtime_multimarket.go:Console.buildMultiMarketStrategyRuntimePage
[logic-map] missing evidence for modified function internal/console/strategy_runtime_multimarket.go:multiMarketStrategyRuntimePage.project
[logic-map] missing evidence for modified function internal/console/trading_views_test.go:TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA
[logic-map] missing evidence for modified function internal/execgw/a098_sender_down_reason_test.go:TestTheSenderDownReasonIsRegisteredInTheEnumeration
[logic-map] missing evidence for modified function internal/execgw/classify.go:classifyMutation
[logic-map] missing evidence for modified function internal/execgw/failclosed.go:AllReasonCodes
[logic-map] missing evidence for modified function internal/execgw/gateway.go:New
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.submit
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkSymbolFree
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkEntry
[logic-map] missing evidence for modified function internal/execgw/gateway.go:Gateway.checkReservation
[logic-map] missing evidence for modified function internal/execgw/gateway_test.go:holdExposure
[logic-map] missing evidence for modified function internal/execgw/modegate.go:EntryGate.ProjectOperatingMode
[logic-map] missing evidence for modified function internal/execgw/observation_test.go:newGuardianOn
[logic-map] missing evidence for modified function internal/execgw/replay.go:Gateway.ReplayInDoubt
[logic-map] missing evidence for modified function internal/execgw/retry.go:Retrier.Query
[logic-map] missing evidence for modified function internal/execgw/retry.go:Retrier.escalateCredentialFailure
[logic-map] missing evidence for modified function internal/execgw/retry.go:EntryGate.Clear
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:NewRiskGuardian
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.LimitsDigest
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.IssueEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian.go:RiskGuardian.escalateFor
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalAccountBaseGuardianPairedKRUS
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS
[logic-map] missing evidence for modified function internal/execgw/riskguardian_account_base_testseam_test.go:qFinalCampaignFirstLegIssuanceFixture
[logic-map] missing evidence for modified function internal/execgw/riskguardian_first_leg.go:RiskGuardian.PrecheckQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/execgw/riskguardian_first_leg.go:RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal.go:RiskGuardian.PrecheckQFinalEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal.go:RiskGuardian.IssuePrecheckedQFinalEntry
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal_test.go:TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker
[logic-map] missing evidence for modified function internal/execgw/riskguardian_qfinal_test.go:qFinalKRRequest
[logic-map] missing evidence for modified function internal/execgw/riskguardian_test.go:newGuardian
[logic-map] missing evidence for modified function internal/execgw/riskguardian_test.go:mustGuardian
[logic-map] missing evidence for modified function internal/execgw/roundtrip.go:Gateway.confirmCreatedOrder
[logic-map] missing evidence for modified function internal/execgw/strategy_entry_gate_authority.go:Gateway.ObserveStrategyEntryGate
[logic-map] missing evidence for modified function internal/execgw/strategy_entry_gate_authority.go:Gateway.withStrategyEntryGateAuthority
[logic-map] missing evidence for modified function internal/httpapi/router.go:router.read
[logic-map] missing evidence for modified function internal/httpapi/strategy_runtime.go:StrategyRuntimeSnapshotFunc
[logic-map] missing evidence for modified function internal/journal/alert_claim.go:Journal.settleUnderClaim
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:Journal.runApplyHooks
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:Journal.ResolveExitProposal
[logic-map] missing evidence for modified function internal/journal/apply_hook.go:ApplyExitFill
[logic-map] missing evidence for modified function internal/journal/exit_snapshot.go:knownArmSuppression
[logic-map] missing evidence for modified function internal/journal/exit_state.go:Journal.recordExitJudgementTx
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestARefusalReArmsTheLevel
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestResolvingNothingIsNotAnError
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestACancelledRungIsProposableAgain
[logic-map] missing evidence for modified function internal/journal/exit_state_test.go:TestARefusedProposalIsReArmedAfterARestart
[logic-map] missing evidence for modified function internal/journal/fills.go:Journal.RecordFill
[logic-map] missing evidence for modified function internal/journal/fills.go:Journal.LiveOrdersForSymbol
[logic-map] missing evidence for modified function internal/journal/issuance_test.go:issueRequest
[logic-map] missing evidence for modified function internal/journal/observation_crash_linux_test.go:crashChildObservationGap
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.TransitionOperatingMode
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.RestoreOperatingModeProjection
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:Journal.OperatingModeHistory
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:currentModeFromRow
[logic-map] missing evidence for modified function internal/journal/operating_mode.go:scanOperatingMode
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:Journal.LinkCampaignOrder
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:ApplyPositionCampaignFill
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:updateCampaignSuccessorRemaining
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignExposureBlockedInTx
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignQuantity
[logic-map] missing evidence for modified function internal/journal/position_campaign.go:campaignRemaining
[logic-map] missing evidence for modified function internal/journal/position_campaign_test.go:TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce
[logic-map] missing evidence for modified function internal/journal/readonly_test.go:TestTheReadOnlyHandleHasNoWriteMethods
[logic-map] missing evidence for modified function internal/journal/reservations.go:reserveRows
[logic-map] missing evidence for modified function internal/journal/reservations_test.go:exposureReserve
[logic-map] missing evidence for modified function internal/journal/reservations_test.go:TestFractionalShareAmountsSurviveTheLedger
[logic-map] missing evidence for modified function internal/journal/risk_bucket.go:Journal.CommitRiskBucketAdmission
[logic-map] missing evidence for modified function internal/journal/risk_bucket.go:ensureRiskBucketEntryScopeClean
[logic-map] missing evidence for modified function internal/journal/risk_bucket_admission_test.go:TestRiskBucketAdmissionRejectsImmutablePolicyAndSnapshotCollision
[logic-map] missing evidence for modified function internal/journal/risk_bucket_admission_test.go:seedExistingRiskReservation
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:loadRiskBucketOrderAuthority
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:riskBucketOrderRemaining
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:loadRiskBucketFillTransition
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill.go:verifyRiskBucketStateDigest
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketAmbiguousSidecarLatchesAllApplicableReservationsWithoutDroppingFill
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported
[logic-map] missing evidence for modified function internal/journal/risk_bucket_fill_test.go:TestRiskBucketActualAndReleaseRequireExactOwnerDecisionScopeForCollidingOrderID
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:commitFreshRiskBucketAdmissionTx
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:insertFreshRiskBucketReservations
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance.go:Journal.RevalidateQFinalAdmission
[logic-map] missing evidence for modified function internal/journal/risk_bucket_issuance_test.go:qFinalIssueFixture
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner.go:Journal.applyRiskBucketOwnerBindingInTx
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner.go:Journal.releaseRiskBucketOwner
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner_test.go:seedRiskBucketOwnerLifecycle
[logic-map] missing evidence for modified function internal/journal/risk_bucket_owner_test.go:closeRiskBucketOwnerLifecycle
[logic-map] missing evidence for modified function internal/journal/schema_test.go:TestSchemaTablesAndColumns
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime.go:linkConfirmedStrategyCampaignTx
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime.go:Journal.RequireCurrentStrategyDispatchTransportAuthority
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime_test.go:prepareStrategyDispatchLease
[logic-map] missing evidence for modified function internal/journal/strategy_dispatch_runtime_test.go:openStrategyDispatchV25Journal
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic.go:Journal.RecordQFinalCampaignFirstLeg
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic_test.go:TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser
[logic-map] missing evidence for modified function internal/journal/strategy_first_leg_atomic_test.go:firstLegAtomicFixture
[logic-map] missing evidence for modified function internal/journal/strategy_lineage.go:insertExactStrategyDecision
[logic-map] missing evidence for modified function internal/journal/strategy_lineage.go:verifyStrategyRiskBinding
[logic-map] missing evidence for modified function internal/journal/strategy_weekly_first_leg_test.go:weeklyFirstLegFixture
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:ProjectAcceptedStrategyflowLineage
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:verifyProjectionRiskIntent
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:strategyflowDecisionIdentity
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:strategyflowLineageFromProjection
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection.go:decodeStrategyflowRiskBinding
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestProjectAcceptedStrategyflowLineagePairsAllSixKRUSLanes
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestVerifyStrategyflowRiskBindingV2Compatibility
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:strategyflowFirstLegFixture
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestRecordQFinalCampaignFirstLegRejectsStrategyflowTamperWithoutRows
[logic-map] missing evidence for modified function internal/journal/strategyflow_projection_test.go:TestVerifyAcceptedStrategyflowProjectionRejectsOuterCanonicalTamper
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:freezeTradeOutcomeTx
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:computeTradeOutcome
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:adoptedRoundTripLegs
[logic-map] missing evidence for modified function internal/journal/trade_outcomes.go:Journal.BackfillTradeOutcome
[logic-map] missing evidence for modified function internal/mcp/server.go:toolResult
[logic-map] missing evidence for modified function internal/mcp/server.go:trim
[logic-map] missing evidence for modified function internal/mcp/server_test.go:TestToolResultKeepsSmallPayload
[logic-map] missing evidence for modified function internal/measure/counterfactual.go:gridPolicy
[logic-map] missing evidence for modified function internal/measure/reconstruct_test.go:issueEntry
[logic-map] missing evidence for modified function internal/obs/a096_one_send_per_condition_test.go:TestAcknowledgeCannotClearTheGateMidSend
[logic-map] missing evidence for modified function internal/obs/a099_round4_test.go:TestAHeldRowIsNotWhispered
[logic-map] missing evidence for modified function internal/obs/log.go:NewLogger
[logic-map] missing evidence for modified function internal/obs/mode.go:Notifier.AnnounceOperatingMode
[logic-map] missing evidence for modified function internal/obs/mode_test.go:TestTheTransitionLogLineIsCountable
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.publishBestEffort
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.notifyCritical
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.claimAndDeliver
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.escalate
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.deliver
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.logLeaseLost
[logic-map] missing evidence for modified function internal/obs/notifier.go:Notifier.logClaimHeld
[logic-map] missing evidence for modified function internal/obs/notifier.go:renderFields
[logic-map] missing evidence for modified function internal/obs/ntfy.go:Ntfy.Publish
[logic-map] missing evidence for modified function internal/official/candle_reads.go:adaptCandles
[logic-map] missing evidence for modified function internal/official/candle_reads_test.go:TestAdaptCandlesEmpty
[logic-map] missing evidence for modified function internal/official/client.go:Client.hcTransport
[logic-map] internal-official--client.dorequest: AST hash does not match modified function revision
[logic-map] missing evidence for modified function internal/official/market_reads.go:adaptPrices
[logic-map] missing evidence for modified function internal/official/reads.go:Client.accountsLocked
[logic-map] missing evidence for modified function internal/official/stock_reads.go:adaptStocks
[logic-map] missing evidence for modified function internal/official/token.go:tokenManager.exchange
[logic-map] missing evidence for modified function internal/ops/wts_operations.go:wtsOperations
[logic-map] missing evidence for modified function internal/performance/attribution_persistence.go:validateAttributionGeneration
[logic-map] missing evidence for modified function internal/performance/attribution_persistence_test.go:TestAttributionEvidenceRejectsHeadEnvelopeMismatch
[logic-map] missing evidence for modified function internal/positioncampaign/campaign.go:TransitionLeg
[logic-map] missing evidence for modified function internal/positioncampaign/command.go:ValidateCommand
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:Replay
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:replayCampaignTransition
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:replayLegFillTransition
[logic-map] missing evidence for modified function internal/positioncampaign/replay.go:remainingFromCap
[logic-map] missing evidence for modified function internal/positioncampaign/review_regression_test.go:TestProductionCoreContainsNoLaneRatiosOrBrokerDependency
[logic-map] missing evidence for modified function internal/positioncampaign/watermark.go:LegLedger.Observe
[logic-map] missing evidence for modified function internal/reconcile/recovery.go:Recovery.Run
[logic-map] missing evidence for modified function internal/risk/chain_test.go:usdPolicy
[logic-map] missing evidence for modified function internal/risk/chain_test.go:TestOrderAboveThePerOrderCapIsRefused
[logic-map] missing evidence for modified function internal/risk/contract_test.go:TestQuantityAboveTheRiskBudgetIsRefused
[logic-map] missing evidence for modified function internal/risk/input.go:DefaultPolicy
[logic-map] missing evidence for modified function internal/risk/input.go:Policy.Validate
[logic-map] missing evidence for modified function internal/riskbucket/fill.go:recomputeOverageLatches
[logic-map] missing evidence for modified function internal/riskbucket/fill.go:cloneFillState
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:LoadProductionRiskSnapshotAuthority
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:validProductionRiskPolicyContents
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:bindProductionRiskInputs
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:loadProductionRiskEntries
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:readProductionRiskUsage
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority.go:aggregateProductionRiskUsage
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority_test.go:newProductionRiskFixture
[logic-map] missing evidence for modified function internal/riskbucket/production_snapshot_authority_test.go:createProductionRiskDB
[logic-map] missing evidence for modified function internal/scheduler/budget.go:BudgetCoordinator.TryAcquire
[logic-map] missing evidence for modified function internal/scheduler/budget.go:BudgetCoordinator.Complete
[logic-map] missing evidence for modified function internal/soak/attest.go:Summary.Evaluate
[logic-map] missing evidence for modified function internal/soak/attest.go:BuildAttestation
[logic-map] missing evidence for modified function internal/soak/attest_test.go:TestBuildAttestationRefusesAnIncompleteSoak
[logic-map] missing evidence for modified function internal/soak/record.go:OpenRecorder
[logic-map] missing evidence for modified function internal/soak/record.go:Recorder.Append
[logic-map] missing evidence for modified function internal/soak/record.go:LoadCycles
[logic-map] missing evidence for modified function internal/strategycandle/adapter_test.go:officialPage
[logic-map] missing evidence for modified function internal/strategyengine/lane_test.go:verifiedBarFixtureAt
[logic-map] missing evidence for modified function internal/strategyengine/wiring_test.go:TestA047ShipsNoRuntimeOrderOrExitWiring
[logic-map] missing evidence for modified function internal/strategyengine/wiring_test.go:assertConsoleOnlyCallsDormantDescriptor
[logic-map] missing evidence for modified function internal/strategyevidence/model.go:kindSupportsMarket
[logic-map] missing evidence for modified function internal/strategyevidence/model.go:validateTypedPayload
[logic-map] missing evidence for modified function internal/strategyevidence/official_source.go:NewOfficialAdapter
[logic-map] missing evidence for modified function internal/strategyevidence/official_source.go:OfficialAdapter.collectSEC
[logic-map] missing evidence for modified function internal/strategyevidence/projection.go:Project
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:SourcePolicy.validate
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:NewAdapter
[logic-map] missing evidence for modified function internal/strategyevidence/source.go:Adapter.Fetch
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestIncompleteOrUnverifiedPolicyMakesZeroCalls
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterKeepsCredentialOutOfRequestMetadata
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterRejectsRuntimeBoundExpansionBeforeTransport
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterEnforcesResponseBytesWindowAndRetryAfter
[logic-map] missing evidence for modified function internal/strategyevidence/source_test.go:TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest
[logic-map] missing evidence for modified function internal/strategyflow/adapters.go:defaultRegistry
[logic-map] missing evidence for modified function internal/strategyflow/adapters.go:proposalRegistry
[logic-map] missing evidence for modified function internal/strategyflow/canonical_projection_test.go:acceptedProjectionFixture
[logic-map] missing evidence for modified function internal/strategyflow/canonical_projection_test.go:TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether
[logic-map] missing evidence for modified function internal/strategyflow/execution_terms_test.go:TestEvaluateSealsExactExecutionTerms
[logic-map] missing evidence for modified function internal/strategyflow/execution_terms_test.go:TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:Evaluate
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:Propose
[logic-map] missing evidence for modified function internal/strategyflow/flow.go:evaluateWith
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestPairedRegistryCoversKRUSContinuationReversalAndWeekly
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestCompleteLineageSealDetectsMutation
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved
[logic-map] missing evidence for modified function internal/strategyflow/flow_test.go:inputFor
[logic-map] missing evidence for modified function internal/strategyflow/production_integration_test.go:TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators
[logic-map] missing evidence for modified function internal/strategyflow/registry.go:LaneInput.matches
[logic-map] missing evidence for modified function internal/strategymarket/bars.go:aggregateClosedKRXFiveMinute
[logic-map] missing evidence for modified function internal/strategymarket/bars_test.go:minutes
[logic-map] missing evidence for modified function internal/strategymarket/bars_test.go:TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:DormantSnapshot
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:UnavailableSnapshot
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:Clone
[logic-map] missing evidence for modified function internal/strategyprojection/model.go:Validate
[logic-map] missing evidence for modified function internal/strategyprojection/projection_test.go:currentPair
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/a108_publication_is_total_test.go:TestProjectionLivenessClausesEachDecideOnTheirOwn
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport.go:Client.Read
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:reclaimStaleControlDirectory
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:verifyStaleSocketShape
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:projectionSocketAccepts
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix.go:Dial
[logic-map] missing evidence for modified function internal/strategyprojectionrpc/transport_unix_test.go:TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.Len
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.ManifestDigest
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:ProductionBatchAuthority.For
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:LoadProductionAuthorityBatch
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:buildLaneInput
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:validScopes
[logic-map] missing evidence for modified function internal/strategyproposal/production.go:laneMatchesMarket
[logic-map] missing evidence for modified function internal/strategyproposal/production_test.go:productionFixture
[logic-map] missing evidence for modified function internal/strategyproposal/production_testseam.go:ProductionBatchAuthorityForTest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:ProductionRouteAuthority.OwnerDigest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:LoadProductionRouteAuthorityBatch
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:verifyProductionRouteManifest
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:validProductionRouteCandidates
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:productionRouteDescriptors
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:openProductionRouteSnapshot
[logic-map] missing evidence for modified function internal/strategyrouter/production.go:loadProductionRouteOwnersFrom
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthorityFailureIsMarketLocal
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthoritySelectsEverySignedSymbolScope
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:newProductionRouteFixture
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:productionRouteFixture.body
[logic-map] missing evidence for modified function internal/strategyrouter/production_test.go:productionRouteFixture.write
[logic-map] missing evidence for modified function internal/verifylive/record.go:OpenRecorder
[logic-map] missing evidence for modified function internal/verifylive/record.go:Recorder.Append
[logic-map] missing evidence for modified function internal/verifylive/record.go:LoadEntries
[logic-map] missing evidence for modified function internal/verifylive/record.go:Artifact.terminal
[logic-map] missing evidence for modified function internal/verifylive/record_test.go:TestRecordCarriesNoUnmaskedAccountNumber
[logic-map] missing evidence for modified function internal/verifylive/report.go:Report.WriteText
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:StrategyflowKRFixture
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:StrategyflowUSFixture
[logic-map] missing evidence for modified function internal/weeklyvaluelane/strategyflow_testseam.go:strategyflowEvaluation
rc=1
```
