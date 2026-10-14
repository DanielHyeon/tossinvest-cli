# Branch Test Map: `reserveRows`

- Source: `internal/journal/reservations.go` (323-375, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 326 | 기존 동작 | `TestReservationRequiresARecordedUnexpiredDecision` | 무편집 분기 |
| B2 | if 332 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
| B3 | if 335 | 기존 동작 | `TestReachingTheAggregateLimitIsRefused` · `TestA136ReachingTheDailyBuyCeilingIsRefused` | 뒤에 새 판정 호출(편집 지점) |
| B4 | if 339 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
| B5 | range 345 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
| B6 | if 346 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
| B7 | if 370 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
