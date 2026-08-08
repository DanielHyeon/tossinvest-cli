# Branch Test Map: `Runtime.alert`

Source: `internal/app/engine/runtime.go` (444-456). AST 기준 분기 2 / 이탈 1 / defers 1.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:445` `Alerts == nil` → 무동작 | 간접 — `Alerts` 없이 도는 runtime 테스트 | no | yes |
| B2 | `:453` `Notify` 오류 → `alert_undelivered` warn | 간접 | no | yes |

## 이 change가 여기 요구하는 것: 없음

`Runtime.alert`는 **변경 대상이 아니다.** 이 표는 근거로 인용한 함수의 분기를
열거해 두기 위한 것이고, 이 change가 추가할 RED는 없다.

인용의 핵심은 분기가 아니라 `:451`의 `context.WithTimeout`과 `:452`의 defer다 —
AST가 defers **1**로 세는 그 한 줄이, `ExitObserver.alert`(defers **0**)와
`ReconcileDriver.alert`(defers **0**)에는 없다.
