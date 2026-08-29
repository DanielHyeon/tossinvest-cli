# Branch Test Map: `NewContext`

- Source: `internal/app/engine/engine.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` @ 390 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B2 | `if` @ 397 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B3 | `if` @ 401 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B4 | `if` @ 412 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B5 | `if` @ 426 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 433 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B7 | `if` @ 442 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B8 | `if` @ 459 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B9 | `if` @ 477 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
| B10 | `if` @ 482 | `NewContext` 의 기존/신규 커버리지 | yes | yes |
