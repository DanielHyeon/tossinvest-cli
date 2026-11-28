# a205 evidence reconciliation

| 질의 | CodeGraph 1.6.0 | AST · grep · 실행 | 결론(base `e2921fc00`) |
|---|---|---|---|
| 기본 주기의 소비자 | `Interval` 1 | grep: exitloop.go:405 1 | 일치 |
| `Interval()` 의 생산 호출자 | `Run` 1 | AST: Run :434 · :435(2회), Tracer.Run :315 | CodeGraph 누락(필드 · 인터페이스 경유) — AST 채택. Tracer 는 생산 호출자 0(claims C4) |
| 생산 배선이 Interval 을 넘기는가 | — | `cmd/tossctl/engine.go:675-685` ExitObserverOptions 에 Interval 없음 | 생산 주기 = 상수 |
| 하한가 갱신이 판정 goroutine 안인가 | — | AST: Run 의 go_statements 0, cycleComplete B2 select(막히지 않음), 갱신은 보조 실행자 | 밖 — design D2 |
| 상수 2s 에서 엔진 패키지 시험 | — | `go test -count=1 ./internal/app/engine/` rc 0, 406.7s(사전 탐침, 상수만 바꾼 상태, 탐침 뒤 되돌림) | 엔진 시험 함수 편집 불필요 |
| 콘솔이 엔진을 import 할 수 있나 | — | `static_test.go:598` 금지 목록 | 불가 — 사본 유지(design D4) |

미해소 불일치 0. 편집을 막는 것 없음.
