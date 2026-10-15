# Branch Test Map: `reserveRows`

- Source: `internal/journal/reservations.go` (328-384)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 331 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B2 | if 337 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B3 | if 340 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B4 | if 344 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B5 | if 348 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B6 | range 354 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B7 | if 355 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
| B8 | if 379 | 분기 동작 | `go test ./internal/journal -run 'Reserve|A136|A138'` · `TestA138ReconcileScopesMatchExactlyWhatTheyCover` · `TestA138TheSeventhSymbolIsRefused` · `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` · `TestA138ConcurrentEntriesCannotBothTakeTheLastSlot` |
