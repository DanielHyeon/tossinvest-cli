# Branch Test Map: `TestGatewayKeepsExitsOpenUnderAMismatch`

- Source: `internal/reconcile/mismatch_test.go`

RED는 이 change의 새 테스트에서 관측했고(게이트웨이 매수 거부, 인터록 분리, 구조 단언),
GREEN은 `go test ./...` 3,889건 전수 통과로 관측했다. review.md에 기록.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` @ 512 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B2 | `if` @ 537 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B3 | `if` @ 544 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B4 | `if` @ 547 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B5 | `if` @ 550 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B6 | `if` @ 560 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B7 | `if` @ 563 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B8 | `if` @ 570 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B9 | `if` @ 583 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B10 | `if` @ 586 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
| B11 | `if` @ 592 | `TestGatewayKeepsExitsOpenUnderAMismatch` 의 기존/신규 커버리지 | yes | yes |
