# Branch Test Map: `TestGateOnRefusals`

- Source: `internal/app/engine/interlock_test.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` @ 537 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B2 | `if` @ 541 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B3 | `if` @ 546 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B4 | `if` @ 552 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B5 | `if` @ 555 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 558 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B7 | `if` @ 561 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
| B8 | `if` @ 564 | `TestGateOnRefusals` 의 기존/신규 커버리지 | yes | yes |
