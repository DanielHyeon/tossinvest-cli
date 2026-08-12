# Branch Test Map: `Runtime.alert`

Source: `internal/app/engine/runtime.go` (444-456). AST 기준 분기 2 / 이탈 1 / defers 1.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:445` `Alerts == nil` → 무동작 | **없음** — harness가 항상 채운다 | no | **no** |
| B2 | `:453` `Notify` 오류 → `alert_undelivered` warn | **없음** — 오류를 내는 스텁이 없다 | no | **no** |

> **18라운드 B-P7이 두 칸을 다 내렸다.** `runtime_test.go`의 `Alerts:`는 여덟 자리
> 전부 non-nil이고(`:139`·`:178`·`:219`·`:244`·`:284`·`:346`·`:385`·`:414`),
> 그 값인 `recordingAlerts.Notify`는 **항상 nil을 돌려준다**(`runtime_test.go:37-42`).
> 그러므로 nil 분기도, 오류 분기도 어떤 테스트도 지나지 않는다. `간접`이라는 말이
> 두 칸에 붙어 있었고 두 칸 다 근거가 없었다.

## 이 change가 여기 요구하는 것: 없음

`Runtime.alert`는 **변경 대상이 아니다.** 이 표는 근거로 인용한 함수의 분기를
열거해 두기 위한 것이고, 이 change가 추가할 RED는 없다.

인용의 핵심은 분기가 아니라 `:451`의 `context.WithTimeout`과 `:452`의 defer다 —
AST가 defers **1**로 세는 그 한 줄이, `ExitObserver.alert`(defers **0**)와
`ReconcileDriver.alert`(defers **0**)에는 없다.
