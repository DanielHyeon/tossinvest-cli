# Branch Test Map: `sweepTradingDayBoundary`

- Source: `internal/journal/reservation_release.go` (390-453) — base 7b073f55, 편집 여부는 D1·D2 결정 뒤

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | 395 계좌 범위 | 계좌 한정 sweep | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary` | 근거 |
| B2 | 400 질의 실패 | DB 오류 | 없음(결함 주입 불가 — 기존 부채) | 영향 없음 |
| B3 | 405 후보 순회 | — | B6 과 같음 | 영향 없음 |
| B4 | 410 Scan 실패 | DB 오류 | 없음(기존 부채) | 영향 없음 |
| B5 | 417 rows.Err | DB 오류 | 없음(기존 부채) | 영향 없음 |
| B6 | 424 후보별 판정 | 거래일 경계 | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary` | 근거 |
| B7 | 426 시장 해석 불가 | 보존 + UNKNOWN_MARKET | 구현 단계에서 확인 | 편집 시 시험 필수 |
| B8 | 428 경보 행 읽기 실패 | DB 오류 | 없음(기존 부채) | 영향 없음 |
| B9 | 441 같은 거래일 | 보유 유지 | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary` · `TestOpenExposureDoesNotLapseWithTheDay` | 근거(시장별 경계) |
| B10 | 447 해제 실패 | DB 오류 | 없음(기존 부채) | 영향 없음 |
