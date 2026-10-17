# a121 CodeGraph hard evidence — task 1.2

- 기준: 구현 base `de147cc285c5274cad6d6ab7b208513027a70b40` (feat/a112-four-family-runtime tip, 2026-10-05)
- 도구: CodeGraph **1.6.0** (`codegraph --version`), 공유 트리 색인(파일 2,580 · 노드 45,527 · 엣지 169,826) — 읽기 전용 질의
- 대상 Go 소스는 옛 base `9408fc95` 이후 **바뀌지 않았다**(`git diff --name-status 9408fc95 de147cc2 -- internal/verifylive
  internal/official internal/attest cmd/tossctl/verify.go` = 0 줄). 그래서 색인 신선도가 이 질의 결과를 흔들 수 없다.
- 모든 결과는 HEAD 에서 `grep` 호출 자리 열거로 다시 셌다(아래 「HEAD 대조」).

## 편집 예정 함수

| 함수 | callers (CodeGraph) | HEAD 호출 자리 | 비고 |
|---|---|---|---|
| `verifylive.Artifact.terminal` | 3: `outstandingLines`(`record.go:542`), `Runner.joinChain`(`runner.go:1050`), `TestTheRealRecordsAreJudgedExactlyAsBefore`(`record_replay_test.go:73`) | 5: `record.go:552`(×2 — `prev`·`a`), `record.go:562`, `runner.go:1052`, `record_replay_test.go:109`·`:111` | 정의는 저장소 유일(`record.go:575`) |
| `cmd/tossctl.newVerifyCmd` | 1: `newRootCmd`(`root.go:52`) | 1: `root.go:170` | callees 6: `newVerifyRunCmd`·`newVerifyStatusCmd`·`newVerifyReportCmd`·`newVerifyAbortCmd` + 타입 2 |
| `verifylive.BuildReport` | 9: `runVerifyReport`(`verify.go:718`), `Console.report`(`internal/console/data.go:423`), 시험 7 | 9: `verify.go:727`, `data.go:428`, `http_test.go:445`, `report_test.go:18·34·77·91·109·143` | callees 중 `Report`·`Outcome`·`Entry` 를 다른 패키지로 잘못 해소(아래) |
| `verifylive.BuildProgress` | 6: `runVerifyStatus`(`verify.go:699`), `readVerify`(`internal/console/data.go:363`), 시험 4 | 6: `verify.go:708`, `data.go:375`, `m0_status_test.go:12`, `plan_test.go:507`, `report_test.go:167·187` | callees: `Steps`·`LastEntry`·`Outstanding` |

## 호출만 하는 함수 — 호출자 증거 (편집 아님)

### `verifylive.PendingCleanup` (`cleanup.go:119-121`)

- CodeGraph callers 19 = 생산 1(`readVerify`, `internal/console/data.go:363`) + 시험 함수 18.
- HEAD 호출 자리 20 = 생산 1(`internal/console/data.go:384`) + 시험 19(`hold_test.go` 의 `:144`·`:151` 이 한 함수
  `TestAReDeclaredHoldOutlivesAnOlderVerdict` 안의 두 자리 — 함수 18 개, 자리 19 개).
- 시험 자리: `record_filled_test.go:48`·`:136`, `abort_test.go:115`, `cleanup_test.go:298`·`:312`·`:322`·`:336`,
  `hold_test.go:52`·`:67`·`:84`·`:101`·`:121`·`:144`·`:151`·`:227`, `m0_causal_test.go:173`, `m0_recovery_test.go:115`,
  `record_replay_test.go:134`, `steps_trigger_test.go:333`.
- **`cmd/tossctl` 은 `PendingCleanup` 을 부르지 않는다.** 러너의 재개 정리는 같은 규칙의 다른 입구
  `Runner.cleanupTargets`(`cleanup.go:103-108`)를 쓴다 — 둘 다 `withoutM0ManualReconcile(… cleanupFrom(…))` 이고 차이는
  `settled` 술어뿐(`Settled(entries, id)` 대 `r.settled(id)`). 둘 다 `outstandingLines` → `Artifact.terminal` 을 거치므로
  셋째 종결은 두 입구에 같이 적용된다.
- callees: `withoutM0ManualReconcile`(`m0_manual.go:45`), `cleanupFrom`(`cleanup.go:123`), `Settled`(`record.go:487`).

### `verifylive.M0Unsettled` (`m0_recovery.go:23`)

- CodeGraph callers 3 = HEAD 호출 자리 3, 전부 생산: `validateM0TriggerMode`(`cmd/tossctl/verify.go:414`),
  `Runner.m0PendingCheckpoint`(`m0_recovery.go:110`), `New`(`runner.go:349`). 시험에서 직접 부르는 자리 0.
- 세 자리 모두 `(checkpoint, ok, err)` 를 받아 `err` → 오류, `ok && Kind != "pending-create"` → HOLD 거절로 쓴다
  (AST `tossctl--validatem0triggermode` B3·B5, `verifylive--new` B10·B12).
- callees: `m0CheckpointScopeEqual`(`m0_recovery.go:120`).

## `Outstanding` 를 거쳐 셋째 종결을 따르는 생산 소비자 (HEAD grep)

`abort.go:66`·`:143`, `cleanup.go:282`, `mutate.go:679`(`liveCount`), `redo.go:122`, `report.go:212`·`:341`,
`runner.go:346`(`New` B9)·`:877`, `steps.go:1094`·`:1138`, `cmd/tossctl/verify.go:427`(`validateM0TriggerMode` B9).
`outstandingLines` 자체의 호출자는 `cleanupFrom`(`cleanup.go:125`)과 `Outstanding`(`record.go:508`) 둘(CodeGraph 2 = HEAD 2).

## 영향 시험 (`codegraph affected`, 1.6.0)

| 편집 파일 | 기본 판별식 | `--filter '*_test.go'` | 그중 `internal/verifylive/` · `cmd/tossctl/` |
|---|---:|---:|---|
| `internal/verifylive/record.go` | 1 (`auth-helper/tests/test_cli.py` — 오탐) | 1,051 | 30 · 85 |
| `internal/verifylive/report.go` | 1 (같은 오탐) | 944 | 27 · 81 |
| `cmd/tossctl/verify.go` | 1 (같은 오탐) | 975 | 30 · 84 |

기본 판별식이 Go 시험을 못 찾는 것은 `.claude/CLAUDE.md` 하네스 주석의 알려진 결함 그대로다. 필터 결과는 패키지 import
그래프의 과대 근사이므로 시험 선택은 위 호출자 열거와 FLM 의 Branch Test Map 이 정본이다.

## CodeGraph 오해소 (advisory 한계 기록)

`callees BuildReport`·`callees BuildProgress` 가 같은 패키지 타입 `verifylive.Report`·`verifylive.Outcome`·
`verifylive.Entry` 를 각각 `internal/doctor/service.go:43`·`internal/continuationlane/evaluator.go:170`·
`internal/audit/audit.go:115` 로 해소했다. 함수 호출 엣지(위 표)는 HEAD grep 과 전부 일치했고 어긋난 것은 구조체
이름 해소뿐이다.
