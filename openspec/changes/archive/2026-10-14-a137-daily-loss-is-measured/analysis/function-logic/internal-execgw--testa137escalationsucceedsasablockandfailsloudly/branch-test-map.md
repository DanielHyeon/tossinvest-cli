# Branch Test Map: `TestA137EscalationSucceedsAsABlockAndFailsLoudly`

- Source: `internal/execgw/a137_daily_loss_internal_test.go` (328-346, base 리비전). 편집 전 분기수 4 → 편집 뒤 4. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 331 | 기존 동작 | `go test ./internal/execgw -run TestA137EscalationSucceeds` | 무편집 분기 |
| B2 | if 334 | 기존 동작 | `go test ./internal/execgw -run TestA137EscalationSucceeds` | 무편집 분기 |
| B3 | if 339 | 기존 동작 | `go test ./internal/execgw -run TestA137EscalationSucceeds` | 무편집 분기 |
| B4 | if 343 | 기존 동작 | `go test ./internal/execgw -run TestA137EscalationSucceeds` | 무편집 분기 |
