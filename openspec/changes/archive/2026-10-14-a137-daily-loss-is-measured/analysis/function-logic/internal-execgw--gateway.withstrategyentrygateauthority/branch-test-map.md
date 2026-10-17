# Branch Test Map: `Gateway.withStrategyEntryGateAuthority`

- Source: `internal/execgw/strategy_entry_gate_authority.go` (46-81, 편집 뒤). 편집 전 분기수 6 → 편집 뒤 8. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 49 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` | 무편집 분기 |
| B2 | if 56 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` | 무편집 분기 |
| B3 | if 57 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` | 무편집 분기 |
| B4 | if 64 | a137 편집 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` | 편집 분기 |
| B5 | if 66 | a137 편집 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` · `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` | 편집 분기 |
| B6 | if 71 | a137 편집 | `TestA137OtherFencesBehaveAsBefore` · 기존 `TestStrategyGatewayRechecksEntryGateAfterSubmittingFencePairedKRUS` | 편집 분기 |
| B7 | if 72 | a137 편집 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` | 편집 분기 |
| B8 | if 77 | 기존 동작 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` | 무편집 분기 |
