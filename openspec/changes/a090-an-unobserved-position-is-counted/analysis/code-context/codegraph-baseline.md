# CodeGraph baseline — a090

- 날짜: 2026-09-29 · 기준 base `d3bd1843` (`exitloop.go` sha256 `522d5d81…` = AST 추출·진입 실측 커밋 `b0a202b8` 의 것과 같음)
  · codegraph **1.6.0** · 질의 직전 `codegraph sync .` 완료 · 모든 질의 `--limit 1000 --json`, 원자료 `raw/`.

| 질의 | 결과 | 비고 |
|---|---|---|
| `callers ExitObserver.ObserveOnce` | 8: 생산 `Run` 2(`exitloop.go:354` · `tracer.go:273`) + 시험 6 | HEAD grep `-F '.ObserveOnce('` 비시험 2 자리(`exitloop.go:359` · `tracer.go:297`)와 일치. 시험 대부분은 하네스 `h.observe()` 경유라 직접 호출자로 안 잡힌다 — reconciliation R1 |
| `callees ExitObserver.ObserveOnce` | 11: `Now` · `FillDetectionBehind` · `checkOutage` · `workingSet` · `symbolsOf` · `quoteUsable` · `observe` · `judge` · 타입 3 | AST `calls` 16 과 대조 — 누락 없음(차이는 `strings.*`·`len` 표준 호출) |
| `callers ExitObserver.alert` | 7: `announceQuarantine` · `checkOutage` · `applyFloor` · `alertUnmanaged` · `alertRefused` · `alertProposalRefused` · `noteDelay` | a090 이 여덟째 호출자를 더한다 |
| `callers ExitObserver.checkOutage` | 1: `ObserveOnce` | 계정 사다리의 유일 진입 |
| `callers ExitObserver.quoteUsable` | 6: `ObserveOnce` · `judge` · `judgeRatchet` · `judgeLadder` · `refreshObservation` · `record` | grep 비시험 6 자리(`:459` `:859` `:956` `:1027` `:1050` `:1180`)와 일치 — design D1 |
| `affected exitloop.go` 기본 | 시험 1(`auth-helper/tests/test_cli.py`) | 하네스 주석의 `isTestPath` 결함 재현 — 결론으로 쓰지 않는다 |
| `affected exitloop.go --filter '*_test.go'` | 898 줄 | `raw/affected-filter.txt` |
