# CodeGraph baseline — a129

- 날짜: 2026-10-12 · base `381e73da` (`exitloop.go` sha256 `0733bd86…`, `engine.go` sha256 `6844d379…` = 이 change 의 편집 전 AST 와 같은 바이트)
  · codegraph **1.6.0** · 질의 직전 `codegraph sync .` · 모든 질의 `--limit 1000`, 원자료 `raw/`.

| 질의 | 결과 | 비고 |
|---|---|---|
| `callers ExitObserver.submit` | 2: 생산 `record`(`exitloop.go:1232`) + `internal/journal/durability_test.go:539`(이름 충돌, 다른 `submit`) | HEAD grep `o.submit(` 비시험 1 자리(`:1360`)와 일치 |
| `callees ExitObserver.submit` | 20: `applyFloor` · `isProtective` · `isZeroQuantity` · `IssueReduction` · `alertProposalRefused` · `AttachExitIntent` · `sellIntent` · `alertRefused` · `Place` · `log` · `noteDelay` · `release` … | AST `calls` 27 과 대조 — 차이는 `fmt`·`strings`·`costs.Market`·`err.Error` 표준 호출 |
| `callers ObserveOnce` | 11: 생산 `ExitObserver.Run`(`exitloop.go:382`) · `Tracer.Run`(`tracer.go:273`) + 시험 9 | 하네스 `h.observe()` 경유 시험은 직접 호출자로 안 잡힘 — reconciliation R1 |
| `callers engineRuntime` | 4 시험 + `engine.go` 파일 노드(변수 `engineRuntimeFactory`) | HEAD grep `engineRuntime` 비시험 사용 1(`engineRuntimeFactory = engineRuntime`) |
| `callers ExitObserver.log` | 2: `settleUnobserved` · `logUnobservedFailure` | **불일치** — HEAD grep `o.log(` 비시험 6 자리(아래 R2) |
| `callers ExitObserver.logErr` | 3: `reportCycle` · `checkOutage` · `alert` | HEAD grep `o.logErr(` 4 자리(`:414` `:893` `:898` `:1837`)와 함수 단위로 일치 |
| `affected exitloop.go engine.go` 기본 | 시험 1(`auth-helper/tests/test_cli.py`) | 하네스 주석의 `isTestPath` 결함 재현 — 결론으로 쓰지 않음 |
| `affected … --filter '*_test.go'` | 1091 줄(그중 `internal/app/engine/`·`cmd/tossctl/` 257) | `raw/affected-filter.txt` — 검증은 두 패키지 전체 `go test -race` 로 함 |
