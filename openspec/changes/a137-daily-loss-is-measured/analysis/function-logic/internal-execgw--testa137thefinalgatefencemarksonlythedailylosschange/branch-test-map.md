# Branch Test Map: `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange`

- Source: `internal/execgw/a137_daily_loss_internal_test.go` (283-337, 편집 뒤). 편집 전 분기수 8 → 편집 뒤 10. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 287 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B2 | if 292 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B3 | if 296 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B4 | if 303 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B5 | if 308 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B6 | if 315 | a137 편집 | `go test ./internal/execgw -run TestA137TheFinalGate` | 편집 분기 |
| B7 | if 320 | a137 편집 | `go test ./internal/execgw -run TestA137TheFinalGate` | 편집 분기 |
| B8 | if 326 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B9 | if 331 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
| B10 | if 334 | 기존 동작 | `go test ./internal/execgw -run TestA137TheFinalGate` | 무편집 분기 |
