# CodeGraph hard evidence (a194, tasks 1.2)

- 도구: **codegraph 1.6.0**(`codegraph --version`). 색인: 격리 worktree `~/tossos-work/a194-impl` 에서 `codegraph init .`(2026-11-23, 3,237 파일 · 55,950 노드 · 206,351 엣지). `make sdd-sync` 는 돌리지 않음(지시).
- 기준 커밋: `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`(= `base-commit.txt`), 워킹트리 Go 편집 0.
- 원문: `codegraph-callers.txt`(callers 10개 심볼, 출력 앞 40줄씩), `codegraph-impact-callees.txt`(`impact judgeHoldings` · `callees judgeHoldings`), `codegraph-affected.txt`(`affected adoption.go reconcileloop.go --filter '*_test.go'` → 시험 파일 1,037 — 패키지 전체가 걸려 선택 근거로는 못 씀).
- CodeGraph 는 **함수 단위**로 답한다. 호출 **자리** 수는 `grep -rn`(비시험 .go)로 따로 셌다(아래 「자리」 열).

| 심볼 | 정의 | CodeGraph 비시험 호출자(함수) | 비시험 호출 자리(grep) |
|---|---|---|---|
| `ReconcileDriver.judgeHoldings` | internal/app/engine/adoption.go:107-193 | `RunOnce`(reconcileloop.go:418) 1 | 1 — reconcileloop.go:470 |
| `ReconcileDriver.blocked` | adoption.go:196-201 | `judgeHoldings` 1 | 1 — adoption.go:152 |
| `ReconcileDriver.alertUnmanaged` | adoption.go:449-479 | `judgeHoldings` 1(+ 동명 `ExitObserver.alertUnmanaged` exitloop.go:1848 의 호출자 `workingSet` 를 섞어 보고 — 아래 불일치 1) | 1 — adoption.go:191 |
| `ReconcileDriver.alertAdoptionFailed` | adoption.go:486-508 | `alertUnmanaged` 1 | 1 — adoption.go:452 |
| `reconcileUnmanagedKey` | adoption.go:542-544 | `alertUnmanaged`·`alertAdoptionFailed` 2 | 2 — adoption.go:465 · :493 |
| `ReconcileDriver.unmanagedFact` | adoption.go:514-537 | `alertUnmanaged` 1 | 1 — adoption.go:450 |
| `ReconcileDriver.RunOnce` | reconcileloop.go:418-472 | `Run` 1 | 1 — reconcileloop.go:405 |
| `Tracker.Blocks` | internal/reconcile/mismatch.go:1296-1302 | 7 함수(`position_policy_command.go` Runtime·Blocks, `RunOnce`, `cmd/tossctl/engine_reconcile.go` buildReconcileResolveRuntime, `Tracker.EntryAllowed`, `Tracker.Resolve`, `alertops.go` entryBlockReasons) | 6 자리 — position_policy_command.go:133 · reconcileloop.go:460 · alertops.go:191(이것은 `gate.Blocks()` — EntryGate 의 동명 메서드, 불일치 2) · mismatch.go:1327 · :1361 · cmd/tossctl/engine_reconcile.go:303 |
| `Tracker.EntryAllowed` | mismatch.go:1326-1333 | `blocked` 1 | **2** — adoption.go:200 · exitwiring.go:197(`reconcileFloor.ConfirmedFloor`, 불일치 3) |
| `Tracker.Permanent` | mismatch.go:1314-1318 | `blocked` 1 | **2** — adoption.go:200 · exitwiring.go:197 |
| `Context.ReconcileDriver` | reconcileloop.go:357-388 | (callers 미질의 — 1.3 에서 D5 배선 후보로 드러남) | cmd/tossctl 조립(`engineRuntime` B2 if@667 앞 call@662, a193 FLM 인용) |

## impact · callees (judgeHoldings)

- `impact judgeHoldings` → 3 심볼: `judgeHoldings` · `RunOnce` · `Run`. 시험 파일은 세지 않음(Go `_test.go` 판별식 결함 — `.claude/CLAUDE.md` 하네스 주석).
- `callees judgeHoldings` → 18. 그중 둘은 잘못 해소: `Context` → `internal/accountbudget/prewire_fence_linux.go:37`, `Snapshot` → `internal/filldetect/detect.go:202`(실제는 `context.Context`·`reconcile.Snapshot`) — 불일치 4.

## 구조 고정·census (tasks 1.4 와 같이 봄)

- a166 census `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`(키 92) · `sites.json`(116행): `adoption.go`·`reconcileloop.go`·`judgeHoldings`·`ReconcileDriver`·`RunOnce`·`alertUnmanaged` 문자열 0건. engine 패키지 키는 `engine.go||NewContext` · `engine.go||NewOrderPath` · `gateway.go||buildGateway` · `gateway.go||newGatewayReplay` 넷뿐.
- a163 판정 핀(`a163_judgement_pin_test.go:41-52`): ExitObserver 열 함수 — ReconcileDriver 없음.
- a092 critical 기록자 열거(`a092_structure_pins_test.go:116`): `RecordAlert`·`ClaimAlertForDelivery`·`EnqueueAlert` 의 비시험 호출자 — `Notifier.RecordCritical` 경유는 이 목록을 바꾸지 않음.
- 소스 문자열 스캔 둘이 새 파일에도 걸린다: `TestExitEligibilityHasOneSpelling`(internal/position/eligibility_drift_test.go — internal 아래 모든 비시험 .go 에서 `EntryDecisionID`·`entry_decision_id`·`AdoptionID`·`adoption_id` 금지, 허용 목록에 `internal/app/engine/adoption.go` 있음, 새 파일은 없음), `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`(a071_readiness_assembly_test.go:27 — `reconcileloop.go` 에서 `ProtectionReadiness`·`ReadinessAdapter`·`checkProtection` 금지).
