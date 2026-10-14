# Branch Test Map: `Gateway.ObserveStrategyEntryGate`

- Source: `internal/execgw/strategy_entry_gate_authority.go` (22-35, 편집 뒤). 편집 전 분기수 2 → 편집 뒤 2. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 24 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run StrategyEntryGate` | 무편집 분기 |
| B2 | if 27 | a137 편집 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(양방향 6 경우) | 편집 분기 |
