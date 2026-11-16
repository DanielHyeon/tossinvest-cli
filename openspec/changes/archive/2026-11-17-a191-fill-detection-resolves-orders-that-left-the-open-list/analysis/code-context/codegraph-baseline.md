# CodeGraph baseline (a191, 1단계)

- 도구: codegraph **1.6.0**(`~/.local/bin/codegraph`). 이 worktree 전용 색인을 새로 만듦: `codegraph init -y .` → 3,203 파일, 55,410 노드,
  204,056 엣지(3.7s). 공유 체크아웃의 `.codegraph/` 는 쓰지 않음.
- 기준 commit: `a392308e8`(worktree HEAD = base, 작업 트리에 Go 변경 0).
- 원문: `codegraph-callers.txt`, `codegraph-affected.txt`.

## 질의와 결과

| 질의 | 결과 | 쓰임 |
|---|---|---|
| `callers Detector.collect` | `pollLocked`(detect.go:283) 1 | 편집 함수의 호출 자리 1 — poll 뮤텍스 아래에서만 돎 |
| `callers Detector.pollLocked` | `PollOnce`(detect.go:277) 1 | |
| `callers Detector.PollOnce` (-l 50) | 32: `Run`(detect.go:515), `Hints`(hints.go), 시험 30(detect_test 18·ledger_test 6·payload_test 3·a087 pin 2·unknown_status 1) | 운영 진입점은 `Run`(SSE 힌트는 운영에서 nil — engine.go:756~758 주석, 손으로 읽음) |
| `callers engineFillDetector` | `engineRuntime`(engine.go:655) + 시험 2(`engine_runtime_branch_test.go:134`, `engine_test.go:420`) | 감지기 생성 자리 1 — D8 |
| `callers Journal.TrackedFillOrders` (-l 50) | `JournalTracked.TrackedOrders`(ledger.go:125), `LocalStateFromJournal`, 시험 다수 | Q5(a)(b) |
| `callers LocalStateFromJournal` (-l 50) | `buildReconcileResolveRuntime`(engine_reconcile.go:237), `ReconcileDriver.RunOnce`(reconcileloop.go:414), 시험 | Q5(b)(d) |
| `callers Derive` | `DeriveOfficialOrder`, `rowFromOrder`(console), `collect` + 시험 | `Derive` 는 바꾸지 않음 |
| `callers ScanOrders` | `findSuccessors`, `scanBoth`, `collect`, `flatten.enumerate`, `reconcile.Collector.Collect` + 시험 | 재사용만, 편집 없음 |
| `callers RecordFill` | `JournalLedger.Apply` + 엔진·journal 시험 | 바꾸지 않음 |

## 영향 시험 (`codegraph affected`)

⚠ 기본 실행은 Go 시험을 못 찾는다(하네스 주석). 원문 그대로:

| 명령 | 결과 |
|---|---|
| `codegraph affected internal/filldetect/detect.go -p .`(기본) | **1** (`auth-helper/tests/test_cli.py` — Go 아님) |
| `… --filter '*_test.go'`(깊이 기본 5) | **1336** |
| `… --filter '*_test.go' -d 2` | **563** |
| `… --filter '*_test.go' -d 1` | **72** (engine 8 · candidate 8 · console 7 · cmd/tossctl 7 · strategyevidence 6 · soak 5 · filldetect 5 · 그 밖) |

파일 단위 import 엣지라 깊이 1 도 넓다. 선택 시험(2단계)은 이것을 상한으로 보고 아래로 고른다: `./internal/filldetect/...` 전부, `cmd/tossctl`
의 `engineFillDetector` 호출 시험 2(M1-A 면), `internal/journal` 의 `TestTrackedFillOrders*`·`RecordFill` 종결 시험, `internal/reconcile` 의
`LocalStateFromJournal` 시험. 전체 시험·전체 게이트는 사용자 요청이 없어 돌리지 않는다.
