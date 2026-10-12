# Evidence reconciliation — a129

- 기준: base `381e73da`, `exitloop.go` sha256 `0733bd86…`, `engine.go` sha256 `6844d379…`.
- 확인한 파일·심볼: `internal/app/engine/exitloop.go`(`ExitObserverOptions` · `ObserveOnce` · `record` · `submit` · `log` · `logErr`) ·
  `internal/app/engine/exit_unobserved.go`(`logUnobserved`) · `internal/app/engine/exit_stop_sold_nothing.go`(`ZeroFloorLog`) ·
  `internal/app/engine/exit_held_proposal.go`(`warnEpisode`) · `internal/app/engine/exitwiring.go`(`Context.ExitObserver`) ·
  `cmd/tossctl/engine.go`(`engineRuntime` · `engineFillDetector`) · `internal/obs/log.go` · `internal/obs/event.go`(등급표) ·
  `internal/filldetect/ledger.go` · `internal/journal/fills.go`(`FillSnapshotRecord` · `LookupFill`) · `internal/journal/position_projection.go`(`PositionID`) ·
  `internal/execgw/gateway.go`(`Outcome`) · `cmd/tossctl/a090_engine_runtime_unobserved_log_test.go`.

| # | 불일치 | 현재 HEAD 로 내린 결론 |
|---|---|---|
| R1 | `callers ObserveOnce` 는 시험 9 만 보이고 하네스 경유 시험 수십 개는 안 보인다 | 영향 시험은 `affected --filter '*_test.go'` 와 두 패키지 전체 `go test -race` 로 판단 |
| R2 | `callers ExitObserver.log` = `settleUnobserved` · `logUnobservedFailure` 2 ↔ HEAD grep `o.log(` 비시험 6 자리(`exitloop.go:622` `:735` `:763` `:1455`, `exit_held_proposal.go:320` `:337`) | CodeGraph 가 `o.log` 를 같은 파일의 `logUnobserved` 로 잘못 묶음(이름 접두 해석). 계좌 원문 판정(design D1)은 HEAD grep 의 아홉 자리로 함 |
| R3 | `affected` 기본 실행 시험 1 | 알려진 판별식 결함(CLAUDE.md 하네스 주석). `--filter` 결과를 씀 |
| R4 | 체결의 생산자 | CodeGraph 질의 없이 HEAD 로 확인: 체결은 `filldetect.JournalLedger.Apply` → `Journal.RecordFill`, `filldetect.Detector` 에 로거 필드 없음(`detect.go:151-188`) — design D5 의 근거 |
