# 인용 함수 AST (a193 1단계)

이 change 의 proposal·design 이 분기·early return·side effect 를 근거로 쓴 함수들의 `tools/logic-map` 산출물이다. 1단계는 production Go 를 고치지 않으므로
전부 **인용만** 하는 base 판본이다. 2단계에서 편집하는 기존 함수(`engineRuntime`, 시험 `TestProductionRuntimeStartsExactlyOneAlertDeliverer`)는
`../function-logic/<pkg>--<fn>/` 번들(ast·FLM·BTM·risk)을 따로 가진다.

- 기준: worktree `~/tossos-work/a193-in-doubt-runtime-resolve`, HEAD `791fb6ef1`(= `base-commit.txt`), 2026-11-18. 작업 트리의 Go 변경 0.
- 명령: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수>` → `<패키지-경로>--<함수 소문자>.ast.json`. **40개 전부 rc 0**
  (생성 루프는 실패 시 `FAIL` 을 찍고 파일을 옮기지 않게 짰고, 찍힌 줄 없음).
- 요약: `summary.txt`(생성기 `astsum.py.txt` — a191·a192 의 것을 그대로 복사. 분기 `B<n>:kind@line`, return 줄, call `text@line`).
- 좌표 표기: design·proposal 의 `B<n>(kind@line)` 은 `branches[].id`·`at.line`, `call@line` 은 `calls[].at.line`, `return@line` 은 `returns[].at.line`.
- 주의: 이 판의 `assignments` 는 줄과 종류만 싣고 대상 이름을 싣지 않는다 — 「수신자 필드에 쓰지 않는다」 같은 주장은 줄 목록 + 손으로 읽음.

## AST 가 말하지 못해서 손으로 읽은 것 (design 에 「손으로 읽음」 표시)

| 무엇 | 자리 | 왜 AST 밖인가 |
|---|---|---|
| `DefaultResolveConfig` 값(안정 3회·45초·5초·5분·50쪽·5분) | indoubt.go:157-164 | 복합 리터럴 |
| production `Resolver` 의 필드(Journal·Orders·Order·Account·Clock·Gate=entry) | gateway.go:298-305 | 복합 리터럴 |
| `OfficialOrders`·`OfficialAccount` 가 `in.official` 을 씀 | gateway.go:259-260 | 복합 리터럴 |
| `Context.Resolver` 필드 | engine.go:186-188 | 구조체 선언 |
| exit 제출 요청에 `Baseline` 없음 | exitloop.go:1556-1561 | 복합 리터럴 |
| `PendingAttempts` 의 상태 4개·정렬 | recovery.go:31-33 | SQL 문자열 |
| `TrackedFillOrders` 가 CONFIRMED PLACE/AMEND + broker_order_id 를 추적 | fills.go:1554-1559 | SQL 문자열 |
| `*sql.DB` 의 ctx 검사(`DB.conn`·`Tx.Commit`) | go1.26.5 `database/sql/sql.go:1316-1328`·`:2287-2298` | 표준 라이브러리 |
| `j.db` 가 `*sql.DB` | journal.go:86 | 구조체 선언 |
| 시계 `Sleep` 이 ctx 취소면 `ctx.Err()` | clock/clock.go:87-95, clock/fake.go:44-56 | 인용 AST 안 만듦(짧은 함수, 문장 그대로 인용) |
| `accountbudget.PriorityFromContext` 기본값 Normal, 대기열이 Critical 먼저 | transport_linux.go:24-29, budget_linux.go:551 | a166 진행 중 코드 — AST 안 만듦(이 change 는 그 함수를 안 바꾸고 영향만 적음) |
| 진입 게이트 검사가 노출 증가 mutation 에만 | gateway.go:860 | 조건식(이 change 는 그 함수를 안 바꿈) |
| `ReasonUnresolvedInDoubt` 래치를 푸는 production 호출자 0 | `rg '\.Clear\('` 시험 제외(notifier.go:952·981, recovery.go:338, mismatch.go:1532 — 이유 목록, filldetect detect.go:672) | grep |
| `Resolve` 의 production 호출자 2 | `rg '\.Resolve\(ctx'` 시험 제외 | grep(CodeGraph 이름 단위 결과의 재확인) |
| `order.in_doubt` 가 critical | event.go:55·364 | 상수·맵 리터럴 |
| 공식 HTTP 클라이언트 시한 15초, 원장 busy_timeout | official/client.go:22, journal.go:224 | 상수 |
| a166 census 키 50개와 세는 선택자 | shapes.json, a166_wiring_census_test.go:81-94 | JSON·시험 코드 |
| IN_DOUBT 로 전이하는 원장 자리 | dispatch.go:174·191·206·215, recovery.go:105 | grep — 전이 뒤 같은 attempt 를 다시 쓰는 단계 없음 확인 |
| 기존 시험이 영값 Journal 로 런타임을 돌림 | engine_runtime_branch_test.go:36·118·126 | 시험 코드 |

## 목록 (40)

| 파일 | 함수 | 위치 | 분기 |
|---|---|---|---|
| `cmd-tossctl--enginerecoveryobserver.ast.json` | `engineRecoveryObserver` | cmd/tossctl/engineready.go:101-112 | 1 |
| `cmd-tossctl--engineruntime.ast.json` | `engineRuntime` | cmd/tossctl/engine.go:655-752 | 6 |
| `cmd-tossctl--recoverthenready.ast.json` | `recoverThenReady` | cmd/tossctl/engineready.go:58-88 | 5 |
| `cmd-tossctl--testproductionruntimestartsexactlyonealertdeliverer.ast.json` | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` | cmd/tossctl/a098_the_engine_registers_a_sender_test.go:20-31 | 2 |
| `internal-app-engine--context.manualbuyrefresher.ast.json` | `Context.ManualBuyRefresher` | internal/app/engine/manual_buy_refresh.go:62-81 | 5 |
| `internal-app-engine--context.recovery.ast.json` | `Context.Recovery` | internal/app/engine/runtime_wiring.go:175-205 | 5 |
| `internal-app-engine--exitobserver.noteexitorder.ast.json` | `ExitObserver.noteExitOrder` | internal/app/engine/exit_order_log.go:198-231 | 9 |
| `internal-app-engine--exitobserver.noteheldproposal.ast.json` | `ExitObserver.noteHeldProposal` | internal/app/engine/exit_held_proposal.go:126-173 | 11 |
| `internal-app-engine--exitobserver.recordepisodecritical.ast.json` | `ExitObserver.recordEpisodeCritical` | internal/app/engine/exit_held_proposal.go:318-332 | 3 |
| `internal-app-engine--exitobserver.recordparkcause.ast.json` | `ExitObserver.recordParkCause` | internal/app/engine/exit_held_proposal.go:217-238 | 1 |
| `internal-app-engine--exitobserver.submit.ast.json` | `ExitObserver.submit` | internal/app/engine/exitloop.go:1495-1616 | 17 |
| `internal-app-engine--runauxiliarybody.ast.json` | `runAuxiliaryBody` | internal/app/engine/auxiliary.go:130-137 | 1 |
| `internal-app-engine--runtime.gracefulstop.ast.json` | `Runtime.gracefulStop` | internal/app/engine/runtime.go:387-390 | 0 |
| `internal-app-engine--runtime.run.ast.json` | `Runtime.Run` | internal/app/engine/runtime.go:288-379 | 5 |
| `internal-app-engine--runtime.runauxiliary.ast.json` | `Runtime.runAuxiliary` | internal/app/engine/auxiliary.go:90-121 | 3 |
| `internal-execgw--defaultresolveconfig.ast.json` | `DefaultResolveConfig` | internal/execgw/indoubt.go:156-165 | 0 |
| `internal-execgw--entrygate.block.ast.json` | `EntryGate.Block` | internal/execgw/retry.go:558-565 | 1 |
| `internal-execgw--gateway.checksymbolfree.ast.json` | `Gateway.checkSymbolFree` | internal/execgw/gateway.go:802-838 | 10 |
| `internal-execgw--gateway.placeclaimedstrategy.ast.json` | `Gateway.PlaceClaimedStrategy` | internal/execgw/strategy_gateway.go:57-67 | 0 |
| `internal-execgw--resolver.absencecorroborated.ast.json` | `Resolver.absenceCorroborated` | internal/execgw/indoubt.go:445-500 | 8 |
| `internal-execgw--resolver.park.ast.json` | `Resolver.park` | internal/execgw/indoubt.go:372-384 | 2 |
| `internal-execgw--resolver.resolve.ast.json` | `Resolver.Resolve` | internal/execgw/indoubt.go:229-357 | 26 |
| `internal-execgw--resolver.scanboth.ast.json` | `Resolver.scanBoth` | internal/execgw/indoubt.go:409-440 | 6 |
| `internal-execgw--resolver.waitorexpire.ast.json` | `Resolver.waitOrExpire` | internal/execgw/indoubt.go:360-365 | 1 |
| `internal-execgw--resolver.windowcontaminated.ast.json` | `Resolver.windowContaminated` | internal/execgw/indoubt.go:512-536 | 4 |
| `internal-flatten--saga.resolvestep.ast.json` | `Saga.resolveStep` | internal/flatten/flatten.go:423-446 | 6 |
| `internal-flatten--saga.sell.ast.json` | `Saga.sell` | internal/flatten/liquidate.go:444-540 | 13 |
| `internal-journal--attempt.resolveconfirmed.ast.json` | `Attempt.ResolveConfirmed` | internal/journal/resolution.go:48-59 | 1 |
| `internal-journal--attempt.resolvefailed.ast.json` | `Attempt.ResolveFailed` | internal/journal/resolution.go:63-70 | 0 |
| `internal-journal--attempt.resolveunresolved.ast.json` | `Attempt.ResolveUnresolved` | internal/journal/resolution.go:77-84 | 0 |
| `internal-journal--attempt.transition.ast.json` | `Attempt.transition` | internal/journal/durability.go:548-674 | 23 |
| `internal-journal--exitintentattemptsq.ast.json` | `exitIntentAttemptsQ` | internal/journal/exit_proposal_release.go:182-241 | 19 |
| `internal-journal--journal.pendingattempts.ast.json` | `Journal.PendingAttempts` | internal/journal/recovery.go:30-51 | 4 |
| `internal-journal--journal.recoverpending.ast.json` | `Journal.RecoverPending` | internal/journal/recovery.go:86-125 | 10 |
| `internal-journal--journal.releaseexitproposal.ast.json` | `Journal.releaseExitProposal` | internal/journal/exit_proposal_release.go:96-131 | 8 |
| `internal-journal--journal.unresolvedattempts.ast.json` | `Journal.UnresolvedAttempts` | internal/journal/resolution.go:92-112 | 4 |
| `internal-obs--notifier.recordcritical.ast.json` | `Notifier.RecordCritical` | internal/obs/record_only.go:86-101 | 2 |
| `internal-reconcile--recovery.blockedsymbol.ast.json` | `Recovery.blockedSymbol` | internal/reconcile/recovery.go:423-434 | 1 |
| `internal-reconcile--recovery.replay.ast.json` | `Recovery.replay` | internal/reconcile/recovery.go:363-379 | 3 |
| `internal-reconcile--recovery.run.ast.json` | `Recovery.Run` | internal/reconcile/recovery.go:247-345 | 13 |
