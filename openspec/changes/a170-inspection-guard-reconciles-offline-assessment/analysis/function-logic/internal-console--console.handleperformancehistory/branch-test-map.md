# Branch Test Map: `Console.handlePerformanceHistory`

- Source: `internal/console/performance_history.go` (73-95) — 편집 뒤. AST branches 5.
- 측정: 시험별 `-coverprofile` 로 분기 본문 블록 count>0 확인(2026-10-25, a177 close).

| Branch | Anchor | Scenario | Test | a177 |
|---|---|---|---|---|
| B1 | `74:2` | POST → 405 | `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` | 무변경 |
| B2 | `82:2` | 리더 배선 | `TestPerformanceWindowIsSingleSourceAcrossScreens`(wired · loaderr) · `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` | 미배선 갈래(B2 거짓)는 `TestPerformanceWindowIsSingleSourceAcrossScreens`(unwired) 가 `미측정 · seam_unwired` 로 고정 |
| B3 | `84:3` | Dashboard 오류 | `TestPerformanceWindowIsSingleSourceAcrossScreens`(loaderr — 창 `최근 30일`, `최근 0일` 아님) | a177 이 고정한 회귀 |
| B4 | `86:10` | Dashboard 성공 | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` · `TestPerformanceWindowIsSingleSourceAcrossScreens`(wired) | 무변경 |
| B5 | `89:4` | attribution 오류 | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` · `TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible` | 무변경 |
