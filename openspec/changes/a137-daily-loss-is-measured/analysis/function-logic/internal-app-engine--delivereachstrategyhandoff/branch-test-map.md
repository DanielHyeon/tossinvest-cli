# Branch Test Map: `deliverEachStrategyHandoff`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (62-87, 편집 뒤). 편집 전 분기수 4 → 편집 뒤 5. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | range 64 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` | 무편집 분기 |
| B2 | if 66 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` | 무편집 분기 |
| B3 | if 70 | 기존 동작 | 기존 `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` | 무편집 분기 |
| B4 | if 77 | a137 편집 | `TestA137ADailyLossModeBlockEndsTheCycleQuietly` · `TestA137AMixedCycleStillReportsItsSkippedScopes` · `TestA137OtherModeBlocksBehaveAsBefore` | 편집 분기 |
| B5 | if 81 | 기존 동작 | `go test -tags tossos_testseams ./internal/app/engine -run ScopeRefusal` | 무편집 분기 |
