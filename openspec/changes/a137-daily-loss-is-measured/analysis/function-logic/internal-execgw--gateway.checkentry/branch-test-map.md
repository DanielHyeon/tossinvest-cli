# Branch Test Map: `Gateway.checkEntry`

- Source: `internal/execgw/gateway.go` (850-860, 편집 뒤). 편집 전 분기수 1 → 편집 뒤 2. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 851 | 기존 동작 | `go test ./internal/execgw -run Gateway` | 무편집 분기 |
| B2 | if 855 | a137 편집 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(submit :588 · :665 경유) · `TestA137OtherFencesBehaveAsBefore` | 편집 분기 |
