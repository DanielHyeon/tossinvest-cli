# CodeGraph baseline (a192, 1단계)

- 도구: codegraph **1.6.0**(`codegraph --version`). 이 worktree 전용 색인을 새로 만듦: `codegraph init -y .` → 55,568 노드, 204,866 엣지(3.6s).
  공유 체크아웃의 `.codegraph/` 는 쓰지 않음(`.codegraph/` 는 `.gitignore:72` 로 무시됨).
- 기준 commit: `6277ec34e`(worktree HEAD = base, 작업 트리의 Go 변경 0).
- 원문: `codegraph-callers.txt`(26 심볼), `codegraph-affected.txt`.
- 주의: CodeGraph 의 callers 는 **이름** 단위라 같은 이름의 다른 패키지 함수가 섞인다(예: `persist` 에 `accountbudget`·`degrade` 가 섞임). 아래 표의
  production 호출자는 HEAD 의 `rg` 로 다시 좁힌 값이다.

## 질의와 결과(production 호출자만, 시험 제외)

| 질의 | 결과 | 쓰임 |
|---|---|---|
| `Tracker.Observe` | `ReconcileDriver.RunOnce`(reconcileloop.go:453) 1 — `rg '\.Observe\(ctx, diff\)'` 로 재확인 | D4 편집 함수의 호출 자리 1 |
| `blocksFor` | `Observe` 1 | D11 |
| `persist`(reconcile) | `Observe` 1 | D11 |
| `Restore` | `buildGateway`(gateway.go:277), `buildReconcileResolveRuntime`(engine_reconcile.go:278) | Q4·D4(`Terminal` nil 인 tracker) |
| `Refresh` | `RunOnce`(call@447), 감지기 `JournalLedger.Apply`(engine.go:780 배선 → ledger.go call@85) | Q6 |
| `Resolve` | `buildReconcileResolveRuntime` 의 `resolve: tracker.Resolve` | Q2 |
| `AdjustmentApplied` | `Converger.ConvergeQuantities`(converge.go:189) 1 | Q2 |
| `EntryAllowed` | `ReconcileDriver.blocked`, `reconcileFloor.ConfirmedFloor` | Q5 ②③ |
| `Permanent` | `blocked`, `ConfirmedFloor` | Q5 |
| `blocked` | `judgeHoldings` 1 | Q5 ② |
| `applyFloor` | `ExitObserver.submit` 1 | Q5 ③ |
| `EnterReconcile` | `Tracker.persist`, `guardTrackedFillIdentity`(fills.go), `recordScopedLineageConflict`(lineage.go) | D2 W1 — 다른 두 호출자는 `IDENTIFIER_CONFLICT` 계열(결속 없이 호출 → 결속 0, 행동 불변) |
| `ReleaseReconcile` | `Tracker.persist` 1 | D5 |
| `ReleaseReconciles` | `Tracker.Resolve` 1 | Q2 |
| `ValidReconcileReleaseCause` | `ReleaseReconcile`, `ReleaseReconciles` | D5 |
| `LookupFillScoped` | `followExitFills`(exit_order_log.go:273) | D3 C6 — 새 호출자가 하나 는다 |
| `LocalStateFromJournal` | `RunOnce`, `buildReconcileResolveRuntime` | Q3 |
| `enterReconcileScopeInTx` | `ApplyPositionCampaignFill`, `latchAmbiguousCampaignFill`, `resolveFillOrigin`, `enterReconcileInTx` | Q1·R4 — 투영 계열 생산자들(결속 안 씀) |
| `thresholdKey`·`permanentPromotionDetail` | `Observe` 1 | Q8 반증 ① |
| `ExitAllowed` | **0**(`rg 'ExitAllowed\('` 시험 제외) | Q5 |
| `Due` | **0**(`rg '\.Due\('` 시험 제외) | Q4(30s 는 운영 주기가 아님) |

## 영향 시험 (`codegraph affected`)

⚠ 기본 실행은 Go 시험을 못 찾는다(하네스 주석). 원문 그대로(대상 `internal/reconcile/mismatch.go internal/journal/reconcile_states.go`):

| 명령 | 결과 |
|---|---|
| 기본 | **1** (`auth-helper/tests/test_cli.py` — Go 아님) |
| `--filter '*_test.go'`(깊이 기본) | **1317** |
| `--filter '*_test.go' -d 2` | **278** |
| `--filter '*_test.go' -d 1` | **24** (reconcile 8 · journal 7 · engine 4 · execgw 3 · cmd/tossctl 2) |

파일 단위 import 엣지라 깊이 1 도 넓다. 2단계 선택 시험은 design §5 의 목록을 이 24 안에서 고른다. 전체 시험·전체 게이트는 사용자 요청이 없어 돌리지 않는다.
