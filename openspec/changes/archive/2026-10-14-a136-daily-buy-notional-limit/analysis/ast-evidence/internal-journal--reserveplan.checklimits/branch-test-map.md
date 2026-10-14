# Branch Test Map: `reservePlan.checkLimits`

- Source: `internal/journal/reservations.go` (675-711) — base 7b073f55, a136 무편집 계획

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | 676 | 종류 순회 | B2~B7 | 영향 없음 |
| B2 | 681 | 기존 HELD 합산 | `TestConcurrentDecisionsCannotBothTakeTheLastSlot` | 영향 없음 |
| B3 | 682 | held 통화 불일치 | 없음(기존 부채) | 영향 없음 |
| B4 | 689 | 덧셈 오류 | 없음(기존 부채) | 영향 없음 |
| B5 | 695 | 덧셈 오류 | 없음(기존 부채) | 영향 없음 |
| B6 | 701 | 비교 오류 | 없음(기존 부채) | 영향 없음 |
| B7 | 704 | 도달 = 거부 | `TestReachingTheAggregateLimitIsRefused` | 근거(≥ 경계) |
