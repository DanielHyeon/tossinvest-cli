# Branch Test Map: `ReserveRequest.build`

- Source: `internal/journal/reservations.go` (565-672) — base 7b073f55, a136 추천안(D2-b)에서는 무편집

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | 576 | 머리 검사 | B2~B6 | 영향 없음 |
| B2 | 577 | 결정 없음 | `TestReservationRequiresARecordedUnexpiredDecision` | 영향 없음 |
| B3 | 579 | 계좌 없음 | 기존 | 영향 없음 |
| B4 | 581 | as-of 없음 | `TestStaleSnapshotIsRefused` | 영향 없음 |
| B5 | 583 | 버전 0 | `TestReservationVersionStartsAboveZero` | 영향 없음 |
| B6 | 586 | 행 0 | 기존 | 영향 없음 |
| B7 | 590 | 행 순회 | B8~B16 | 영향 없음 |
| B8 | 595 | 행 검사 | B9~B13 | 영향 없음 |
| B9 | 596 | id 없음 | 기존 | 영향 없음 |
| B10 | 598 | 모르는 종류 | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` | 근거(닫힌 집합) |
| B11 | 600 | 통화 없음 | `TestMixedCurrenciesAreRefused` | 영향 없음 |
| B12 | 602 | DAILY_LOSS 거래일 없음 | `TestDailyLossReservationNeedsItsTradingDay` | 근거 |
| B13 | 605 | 비 DAILY_LOSS 거래일 있음 | `TestDailyLossReservationNeedsItsTradingDay` | 근거 |
| B14 | 610 | 금액 오류 | `TestReservationArithmeticIsExact` | 영향 없음 |
| B15 | 618 | 누계 시작 | `TestReservationArithmeticIsExact` | 영향 없음 |
| B16 | 622 | 누계 오류 | 없음(기존 부채) | 영향 없음 |
| B17 | 629 | usage 순회 | B18 | 영향 없음 |
| B18 | 631 | usage 오류 | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` | 영향 없음 |
| B19 | 636 | limit 순회 | B20 | 영향 없음 |
| B20 | 638 | limit 오류 | 기존 | 영향 없음 |
| B21 | 647 | 대조 순회 | B22~B26 | 영향 없음 |
| B22 | 649 | usage 부재 | `TestUnknownAggregateIsRefusedRatherThanTreatedAsZero` | 영향 없음 |
| B23 | 654 | limit 부재 | 기존 | 영향 없음 |
| B24 | 657 | 통화 불일치 | `TestMixedCurrenciesAreRefused` | 영향 없음 |
| B25 | 663 | 비교 오류 | 없음(기존 부채) | 영향 없음 |
| B26 | 666 | limit ≤ 0 | 기존 | 영향 없음 |
