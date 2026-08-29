# Branch Test Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` @ 454 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B2 | `if` @ 467 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B3 | `if` @ 473 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B4 | `else` @ 475 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B5 | `if` @ 475 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 483 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B7 | `else` @ 485 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B8 | `if` @ 485 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B9 | `if` @ 493 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B10 | `if` @ 509 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B11 | `if` @ 514 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B12 | `if` @ 520 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B13 | `if` @ 521 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B14 | `if` @ 530 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B15 | `if` @ 533 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B16 | `if` @ 540 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B17 | `if` @ 558 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B18 | `if` @ 561 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B19 | `if` @ 564 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B20 | `if` @ 576 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B21 | `if` @ 580 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B22 | `if` @ 594 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
| B23 | `if` @ 597 | `Gateway.submit` 의 기존/신규 커버리지 | yes | yes |
