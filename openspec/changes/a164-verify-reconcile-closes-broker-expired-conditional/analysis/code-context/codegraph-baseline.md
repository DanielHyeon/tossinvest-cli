# a164 CodeGraph baseline

- 도구: codegraph **1.6.0**(`codegraph-status.txt` — files 2,932 · nodes 51,274 · edges 189,790). HEAD `4cd7af19db013435576e44302c964ebc5c2f0557`,
  2026-10-22. 이 문서 로트는 `make sdd-sync` 를 돌리지 않았다(기존 색인 사용) → tasks 0.9 에서 재질의.
- 영향 시험: `../codegraph-affected.txt`(대상 7 파일, `--filter '*_test.go'` → 1,212) · `../codegraph-affected-narrow.txt`(record.go·official 제외 5 파일
  → 1,186). 둘 다 파일 단위 전이 의존이라 넓다(cmd/tossctl 이 거의 전 패키지를 가져옴). 선택 시험은 tasks 5.2(verifylive · official ·
  cmd/tossctl 의 Reconcile·Outstanding·Cleanup·Hold·Report·M0) — 나머지는 미실행 범위로 기록한다.
- callers(`codegraph-callers.txt`):
  - `Outstanding` 20+(생산: `AbortTargets` · `Abort` · `sweepStep` · cmd `validateM0TriggerMode` · `New` 등), `PendingCleanup` 20+(생산: console `readVerify` `internal/console/data.go:363`),
    `ReconciledArtifacts` 3(`BuildReport` · `BuildProgress` · 시험 1), `checkReconcileAbsence`·`checkReconcileClosedStatus` 각 1(`Reconcile`).
  - `Reconcile` 은 이름 충돌로 `journal.ProjectPosition`·`position.Closed`·`protection` 시험 같은 거짓 엣지를 섞는다 — 실제 생산 호출자는
    `cmd/tossctl.runVerifyReconcile` 하나(HEAD `git grep 'verifylive.Reconcile('` 로 확인).
  - `terminal` 은 `joinChain`(`runner.go:1050`) 하나만 낸다 — **누락**. HEAD `git grep '\.terminal()'` 은 생산 4곳:
    `projection_reconciled.go:18` · `record.go:712` · `record.go:722` · `runner.go:1052`. 판정은 HEAD 를 따른다.
