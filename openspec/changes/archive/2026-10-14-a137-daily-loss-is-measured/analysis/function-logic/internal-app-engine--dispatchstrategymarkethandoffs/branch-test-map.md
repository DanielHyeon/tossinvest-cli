# Branch Test Map: `dispatchStrategyMarketHandoffs`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (32-54, 편집 뒤). 편집 전 분기수 3 → 편집 뒤 4. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 38 | 기존 동작 | 기존 dispatch 시험 | 무편집 분기 |
| B2 | if 41 | 기존 동작 | 기존 dispatch 시험 | 무편집 분기 |
| B3 | if 46 | a137 편집 | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay`(effective · refreshOnly) | 편집 분기 |
| B4 | if 49 | 기존 동작 | 기존 dispatch 시험 | 무편집 분기 |
