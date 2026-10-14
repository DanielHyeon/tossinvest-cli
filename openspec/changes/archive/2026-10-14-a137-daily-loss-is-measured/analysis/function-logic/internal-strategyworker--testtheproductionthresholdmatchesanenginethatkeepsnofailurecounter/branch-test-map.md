# Branch Test Map: `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter`

- Source: `internal/strategyworker/policy_receipt_test.go` (362-422, 편집 뒤). 편집 전 분기수 8 → 편집 뒤 8. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | range 364 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B2 | if 367 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B3 | if 371 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B4 | range 375 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B5 | range 376 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B6 | if 383 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B7 | if 412 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
| B8 | if 418 | 기존 동작 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` | 무편집 분기 |
