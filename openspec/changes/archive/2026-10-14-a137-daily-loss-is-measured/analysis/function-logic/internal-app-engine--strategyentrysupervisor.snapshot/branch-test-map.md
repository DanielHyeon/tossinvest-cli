# Branch Test Map: `StrategyEntrySupervisor.Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (782-804, 편집 뒤). 편집 전 분기수 2 → 편집 뒤 2. RED 는 `analysis/red/`, GREEN 은 `analysis/verify/receipts.md`.

| Branch | Anchor | Scenario | Test | a137 |
|---|---|---|---|---|
| B1 | if 783 | 기존 동작 | 기존 스냅숏 시험 | 무편집 분기 |
| B2 | if 789 | 기존 동작 | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay`(effective · refreshOnly) | 무편집 분기 |
