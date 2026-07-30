# Branch Test Map: `New`

- Source: `internal/execgw/gateway.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch` @ 146 | `New` 의 기존/신규 커버리지 | yes | yes |
| B2 | `case` @ 147 | `New` 의 기존/신규 커버리지 | yes | yes |
| B3 | `case` @ 149 | `New` 의 기존/신규 커버리지 | yes | yes |
| B4 | `case` @ 151 | `New` 의 기존/신규 커버리지 | yes | yes |
| B5 | `if` @ 170 | `New` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 173 | `New` 의 기존/신규 커버리지 | yes | yes |
| B7 | `if` @ 176 | `New` 의 기존/신규 커버리지 | yes | yes |
