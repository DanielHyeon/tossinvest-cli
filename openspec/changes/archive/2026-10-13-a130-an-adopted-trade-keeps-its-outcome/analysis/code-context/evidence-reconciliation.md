# Evidence reconciliation (a130)

| # | CodeGraph | 현재 HEAD / AST | 결론 |
|---|---|---|---|
| R1 | `ApplyExitFill` 호출자 없음 | `gateway.go:191` 이 함수 값으로 바인딩, `runApplyHooks` 가 `hooks.Exit(...)` 로 부름 | 동적 호출 — 생산 경로는 RecordFill → runApplyHooks → ApplyExitFill. HEAD 채택 |
| R2 | `BackfillTradeOutcome` 생산 호출자 0 | 같음(grep) | 일치 — 주석의 「BackfillTradeOutcome recovers it」 는 지금까지 호출하는 생산 코드가 없었다 |
| R3 | `freezeTradeOutcomeTx` 호출자 1 | AST: ApplyExitFill B14 안 :1023, 반환값 미사용 | 일치 — 거절이 사라지는 자리 하나 |
| R4 | affected 기본 1건 | Go 시험은 `--filter` 로 1109 | 하네스 주석대로 — 필터 결과 채택 |
| R5 | `adoptedRoundTripLegs` 피호출 `Context`(app.go) | 소스에 없음 | CodeGraph 이름 해석 잡음 — 무시 |

불일치는 전부 HEAD 로 해소됨 — 편집 차단 없음.
