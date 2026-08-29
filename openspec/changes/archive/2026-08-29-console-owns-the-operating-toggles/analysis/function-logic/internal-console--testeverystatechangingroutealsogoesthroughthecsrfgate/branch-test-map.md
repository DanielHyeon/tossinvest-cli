# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` @ 382 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
| B2 | `switch` @ 384 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
| B3 | `case` @ 385 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
| B4 | `case` @ 387 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
| B5 | `range` @ 391 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 392 | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` 의 기존/신규 커버리지 | yes | yes |
