# Branch Test Map: `reserveRows`

- Source: `internal/journal/reservations.go` (323-375, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 326 | 기존 동작 | `TestReservationRequiresARecordedUnexpiredDecision` | 무편집 분기 |
| B2 | if 332 | 기존 동작 | `go test ./internal/journal -run TestA136` · 예약 시험군 | 무편집 분기 |
| B3 | if 335 | 기존 동작 | `TestReachingTheAggregateLimitIsRefused` | 무편집(바로 뒤에 신규 B4) |
| B4 | if 339 | 하루 매수 상한 도달 · 상한 · 수동 합계 없음 · 계산 불가 | `TestA136ReachingTheDailyBuyCeilingIsRefused` · `TestA136AnExposureRaisingReservationWithoutTheCeilingIsRefused` · `TestA136TheManualBuySnapshotCountsAndMustBeFresh` | **a136 신규** — 하루 매수 판정 거부 |
| B5 | range 345 | 기존 동작 | `TestReservationArithmeticIsExact` | base B4(위치만 이동) |
| B6 | if 346 | 기존 동작 | 기존(DB 오류) | base B5(위치만 이동) |
| B7 | if 370 | 기존 동작 | 기존(DB 오류) | base B6(위치만 이동) |
