# CodeGraph baseline (a195, 1단계)

- 기준 commit: `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`(= `base-commit.txt`), worktree `~/tossos-work/a195-impl`(ext4).
- 도구: codegraph **1.6.0**. 이 worktree 에서 `codegraph init .` 새 색인(2026-11-23) — 3,237 파일 · 55,950 노드 · 206,351 엣지. 원문: `codegraph-callers.txt`.
- 질의: `codegraph callers <sym> -p .`(기본 limit 20, 정식 이름은 `-l 200`) · `codegraph impact <sym> -p .`.
- CodeGraph 의 Go 호출자는 **이름 매칭**이라 `Resolve`·`Blocks` 는 잡음이 섞인다. 자리 수는 아래처럼 HEAD `grep` 으로 좁혔다.

## 심볼별 결과

| 심볼 | CodeGraph | HEAD grep(production, `_test.go` 제외) | 호출 자리 수 |
|---|---|---|---|
| `buildReconcileResolveRuntime` (`cmd/tossctl/engine_reconcile.go:237`) | callers 2: `productionReconcileResolveDeps`(engine_reconcile.go:67, 함수 값) · `TestA166EngineReconcileLocalTLS`. impact 4(+`newEngineReconcileResolveCmd`) | production 1(`build: buildReconcileResolveRuntime`, :67). 시험 4 자리(a166_reconcile_safety_test.go:87·91·105·109) + AST 이름 대조 1(:33) | production **1** |
| `runEngineReconcileResolve` (:107) | callers 1: `newEngineReconcileResolveCmdWithDeps`(:94 RunE 클로저) | 같음 | production **1** |
| `Tracker.Resolve` (`internal/reconcile/mismatch.go:1352`) | `Tracker.Resolve` 정식 질의 1(`Observe` — 오인), 이름 `Resolve` 17(전부 다른 `Resolve`) | production **1** — `resolve: tracker.Resolve`(engine_reconcile.go:316, 메서드 값). 실행 자리는 `runtime.resolve(ctx, operator, note)`(:204) 하나 | production **1** |
| `Tracker.Blocks` (mismatch.go:1296) | 정식 질의 30(production 3 + 시험) | production **6**: engine_reconcile.go:303(validate 클로저) · position_policy_command.go:133 · reconcileloop.go:460 · alertops.go:191 · mismatch.go:1327(`EntryAllowed`) · mismatch.go:1361(`Resolve`) | 6 |
| `runtime.validate()` | — | production **1** — engine_reconcile.go:201 | 1 |

## 해제 쓰기 자리 (전수, HEAD grep `UPDATE reconcile_states SET released_at` / `.ReleaseReconcile(s)?(`)

| 자리 | 원인 제한 |
|---|---|
| `journal.ReleaseReconcile` (reconcile_states.go:299) ← 호출 production 1: mismatch.go:935(`ADJUSTMENT_APPLIED`, `Observe` 의 `out.Cleared` 만) | `Observe` B7 if@667 이 `QUANTITY_MISMATCH` 외를 `continue` |
| `journal.ReleaseReconciles` (reconcile_states.go:396) ← 호출 production 1: mismatch.go:1380(`Tracker.Resolve`, `OPERATOR`) | CLI validate 클로저(engine_reconcile.go B14 if@308)만이 원인을 거른다 |
| `ReleaseReconcileOnOrderTerminalProof` (reconcile_order_terminal_release.go:126) | 판정 `orderTerminalReleaseVerdict` B3 if@182(조건 줄 184)가 `QUANTITY_MISMATCH`·scope_market 빈 행만 |

## impact 가 못 보는 결속 (HEAD grep 으로 찾음)

- `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 키 `cmd/tossctl/engine_reconcile.go||buildReconcileResolveRuntime` — 함수 **본문 전체**를 `go/format` 문자열로 고정(`TestA166WiringCensusBaseline`, a166_wiring_census_test.go:209-217 `reviewed function shape changed`). 데이터 파일이라 CodeGraph 호출 그래프에 없다.
- `TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor`(engine_reconcile_test.go:97) — `engine_reconcile.go` 파일 **텍스트**의 포함/금지 문자열 검사.
- `TestA166EngineReconcileOptionsGuard`(a166_reconcile_safety_test.go:25) — `buildReconcileResolveRuntime` AST 안 `newRootOfficialClient(… root.officialOptions...)` 호출 모양.
