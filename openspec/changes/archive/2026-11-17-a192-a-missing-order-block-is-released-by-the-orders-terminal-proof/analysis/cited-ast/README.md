# 인용 함수 AST (a192 1단계)

이 change 의 proposal·design 이 분기·early return·side effect 를 근거로 쓴 함수들의 `tools/logic-map` 산출물이다. 1단계는 production
Go 를 고치지 않으므로 전부 **인용만** 하는 base 판본이다. 2단계에서 편집 대상이 되는 함수(design D11 목록)는 그때
`analysis/function-logic/<pkg>--<fn>/` 번들(ast·FLM·BTM·risk)을 따로 만든다 — 여기 파일은 그 번들의 `pre-edit` 원본으로 재사용할 수 있다.

- 기준: worktree `~/tossos-work/a192-missing-order-block-release`, HEAD `6277ec34e`(= `base-commit.txt`), 2026-11-17. 작업 트리의 Go 변경 0.
- 명령: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수>` → `<패키지-경로>--<함수 소문자>.ast.json`. **53개 전부 rc 0**
  (실패 줄 0 — 생성 스크립트는 실패 시 `FAIL` 을 찍고 파일을 지우게 짰고, 찍힌 줄 없음).
- 요약: `summary.txt`(생성기 `astsum.py.txt` — a191 의 것을 그대로 복사. 분기 `B<n>:kind@line`, return 줄, call `text@line`).
- 좌표 표기: design·proposal 의 `B<n>(kind@line)` 은 `branches[].id`·`at.line`, `call@line` 은 `calls[].at.line`, `return@line` 은 `returns[].at.line`.

## AST 가 말하지 못해서 손으로 읽은 것 (design 에 「손으로 읽음」 표시)

| 무엇 | 자리 | 왜 AST 밖인가 |
|---|---|---|
| missing-order 블록의 `Cause`·`Release`·`Market`·`Detail` 값 | mismatch.go:1419~1430 | 복합 리터럴의 필드 값 |
| 영구 블록의 `Cause`·`Release`·`Detail` 값 | mismatch.go:728~737 | 복합 리터럴 |
| `persist` 가 `EnterReconcile` 에 넘기는 필드(ScopeMarket·ID 없음) | mismatch.go:880~885 | 복합 리터럴 |
| `persist` 의 해제 cause 상수 `ADJUSTMENT_APPLIED` | mismatch.go:903 | 복합 리터럴 |
| `promotionDisputeKeyLess` 의 비교 순서(kind 가 첫 원소) | mismatch.go:1087~1090 | 배열 리터럴 |
| `Restore` 가 지우는 것(streaks·pendingPermanent·adjusted·failures) | mismatch.go:959~969 | 대입문(AST 는 분기만 열거) |
| `Diff.BlocksEntry` 의 식 | compare.go:313 | 분기 0, 식 본문 |
| `DefaultReconcilePeriod = 60s`, `DefaultReconcileInterval = 30s`, `DefaultMaxFailures = 3`, `DefaultStabilisationInterval = 2s`, 감지기 `PollInterval 3s`, `closedLookupInterval 30s` | reconcileloop.go:95, mismatch.go:281·285, snapshot.go:376, detect.go:130, closed_lookup.go:82 | 상수 선언 |
| `Tracker.Due` 의 production 호출자 0 | `rg '\.Due\(' internal cmd`(시험 제외) | grep — 30s 최소 간격은 운영 주기를 정하지 않음 |
| `Tracker.ExitAllowed` 의 production 호출자 0 | `rg 'ExitAllowed\(' internal cmd`(시험 제외) | grep |
| 추적 SQL 의 `f.terminal = 0` | fills.go:1604·1627 | SQL 문자열(a191 이 이미 손으로 읽음) |
| 실패 갱신이 `terminal` 을 그대로 두고 `fail_closed = 1` 만 세움 | fills.go:834~857 `updateFillSnapshotRefusal` | SQL 문자열 |
| `LiveOrdersForSymbol` 의 종결 증거(`confirmedOrderTerminalEvidence` — `terminal = 1`, fail_closed 를 안 봄) | fills.go:1854~1885, exit_proposal_release.go:281~294 | SQL 문자열 |
| `reconcile_states` DDL·부분 UNIQUE·겹침 트리거·추가 열 | execution_contract.go:202~223, risk_bucket_owner.go:55~78 | SQL 문자열 |
| 열린 포지션 상한이 활성 종목 RECONCILE 행을 슬롯으로 셈 | open_positions.go:152~180 | SQL 문자열 |
| 위험 버킷 진입이 활성 RECONCILE 행 존재로 막힘 | risk_bucket.go:295~312 | SQL 문자열 |
| 위험 버킷 broker-zero 가 최신 해제 행의 release cause 를 `RECHECK_MATCHED`/`ADJUSTMENT_APPLIED` 로만 받음 | risk_bucket_owner.go:265~281·676~677 | SQL + 조건식(AST 안 만듦 — 이 change 는 그 함수를 바꾸지 않고 영향만 적음) |
| 투영 거절(과매도)이 쓰는 RECONCILE cause 가 `QUANTITY_MISMATCH` | position_projection.go:428~433(`reconcileCauseFor` AST 는 있음, 상수 값은 손으로) | 상수 |
| 포지션 정책 명령의 `Runtime` 이 `Blocks()` 를 읽어 화면에 넘김 | position_policy_command.go:133~137 | 읽기 전용 소비자, AST 안 만듦 |

## 목록

| 파일 | 함수 | 근거로 쓴 곳(design) |
|---|---|---|
| `internal-reconcile--tracker.observe.ast.json` | `Tracker.Observe` | Q1·Q3·Q4·D4 |
| `internal-reconcile--blocksfor.ast.json` | `blocksFor` | Q1 |
| `internal-reconcile--tracker.persist.ast.json` | `Tracker.persist` | Q1·Q2·D5 |
| `internal-reconcile--tracker.advancestreaks.ast.json` | `Tracker.advanceStreaks` | Q4 |
| `internal-reconcile--promotionkeys.ast.json` | `promotionKeys` | Q4 |
| `internal-reconcile--missingorderpromotionkey.ast.json` | `missingOrderPromotionKey` | Q4 |
| `internal-reconcile--quantitypromotionkey.ast.json` | `quantityPromotionKey` | Q4 |
| `internal-reconcile--tracker.thresholdkey.ast.json` | `Tracker.thresholdKey` | Q8 반증 ① |
| `internal-reconcile--promotiondisputekeyless.ast.json` | `promotionDisputeKeyLess` | Q8 반증 ① |
| `internal-reconcile--permanentpromotiondetail.ast.json` | `permanentPromotionDetail` | Q4·Q8 |
| `internal-reconcile--tracker.restore.ast.json` | `Tracker.Restore` | Q4·Q10 |
| `internal-reconcile--tracker.refresh.ast.json` | `Tracker.Refresh` | Q2·Q6 |
| `internal-reconcile--tracker.resolve.ast.json` | `Tracker.Resolve` | Q2 |
| `internal-reconcile--tracker.adjustmentapplied.ast.json` | `Tracker.AdjustmentApplied` | Q2 |
| `internal-reconcile--creditusableby.ast.json` | `creditUsableBy` | Q3·D3 |
| `internal-reconcile--hasblockfor.ast.json` | `hasBlockFor` | Q2 |
| `internal-reconcile--symbolsindispute.ast.json` | `symbolsInDispute` | Q3·D3 |
| `internal-reconcile--blockfromreconcilestate.ast.json` | `blockFromReconcileState` | Q4 |
| `internal-reconcile--tracker.syncgate.ast.json` | `Tracker.syncGate` | Q5 |
| `internal-reconcile--block.key.ast.json` | `Block.Key` | Q1(같은 종목 블록 충돌) |
| `internal-reconcile--block.covers.ast.json` | `Block.Covers` | Q5 |
| `internal-reconcile--tracker.entryallowed.ast.json` | `Tracker.EntryAllowed` | Q5 |
| `internal-reconcile--tracker.exitallowed.ast.json` | `Tracker.ExitAllowed` | Q5 |
| `internal-reconcile--tracker.permanent.ast.json` | `Tracker.Permanent` | Q5 |
| `internal-reconcile--haspermanentquantityaccountblock.ast.json` | `hasPermanentQuantityAccountBlock` | Q4 |
| `internal-reconcile--comparer.compare.ast.json` | `Comparer.Compare` | Q3·Q10 |
| `internal-reconcile--localordersforcomparison.ast.json` | `localOrdersForComparison` | Q3 |
| `internal-reconcile--localstatefromjournal.ast.json` | `LocalStateFromJournal` | Q3·Q6 |
| `internal-reconcile--diff.blocksentry.ast.json` | `Diff.BlocksEntry` | Q3 |
| `internal-reconcile--converger.convergequantities.ast.json` | `Converger.ConvergeQuantities` | Q2 |
| `internal-app-engine--reconciledriver.runonce.ast.json` | `ReconcileDriver.RunOnce` | Q6·D7 |
| `internal-app-engine--reconciledriver.run.ast.json` | `ReconcileDriver.Run` | Q6 |
| `internal-app-engine--reconciledriver.logcycle.ast.json` | `ReconcileDriver.logCycle` | Q11 |
| `internal-app-engine--reconciledriver.judgeholdings.ast.json` | `ReconcileDriver.judgeHoldings` | Q5·Q11 |
| `internal-app-engine--reconciledriver.blocked.ast.json` | `ReconcileDriver.blocked` | Q5 |
| `internal-app-engine--reconcilefloor.confirmedfloor.ast.json` | `reconcileFloor.ConfirmedFloor` | Q5(블록이 청산을 바꾸는 자리) |
| `internal-app-engine--exitobserver.applyfloor.ast.json` | `ExitObserver.applyFloor` | Q5 |
| `internal-app-engine--runtime.run.ast.json` | `Runtime.Run` | Q6(루프가 따로 도는 goroutine) |
| `cmd-tossctl--engineruntime.ast.json` | `engineRuntime` | Q6 |
| `cmd-tossctl--enginefilldetector.ast.json` | `engineFillDetector` | Q6(감지기 → `Refresh`) |
| `cmd-tossctl--runenginereconcileresolve.ast.json` | `runEngineReconcileResolve` | Q2·Q8 |
| `cmd-tossctl--buildreconcileresolveruntime.ast.json` | `buildReconcileResolveRuntime`(validate 클로저 포함) | Q2 |
| `internal-journal--journal.enterreconcile.ast.json` | `Journal.EnterReconcile` | Q1·Q7 |
| `internal-journal--journal.releasereconcile.ast.json` | `Journal.ReleaseReconcile` | Q2·D5 |
| `internal-journal--journal.releasereconciles.ast.json` | `Journal.ReleaseReconciles` | Q2 |
| `internal-journal--validreconcilereleasecause.ast.json` | `ValidReconcileReleaseCause` | D5 |
| `internal-journal--validreconcilecause.ast.json` | `ValidReconcileCause` | Q7 대안 |
| `internal-journal--lookupfillsnapshotscoped.ast.json` | `lookupFillSnapshotScoped` | D3·Q10 |
| `internal-journal--enterreconcilescopeintx.ast.json` | `enterReconcileScopeInTx` | Q7(b)·Q10 R4 |
| `internal-journal--reconcilecausefor.ast.json` | `reconcileCauseFor` | Q7(b) |
| `internal-filldetect--journalledger.apply.ast.json` | `JournalLedger.Apply` | Q6 |
| `internal-execgw--reconcilereasonfor.ast.json` | `ReconcileReasonFor` | Q7 대안 |
| `internal-execgw--entrygate.rebuildreconcileprojection.ast.json` | `EntryGate.RebuildReconcileProjection` | Q4·Q5 |
