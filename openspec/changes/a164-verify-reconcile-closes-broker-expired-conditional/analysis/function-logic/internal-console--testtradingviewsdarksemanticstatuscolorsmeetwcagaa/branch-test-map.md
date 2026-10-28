# Branch Test Map: 삭제된 다크 의미색 대비 시험(base 판)

- Source: `internal/console/trading_views_test.go` (170-195) — base `fd6afe23b` 판. AST branches 5.
- 함수는 a172 에서 삭제됐으므로 각 갈래의 계약은 대체 시험 `TestTradingViewsThemeColorsMeetWCAGInBothThemes`(현재 `trading_views_test.go:177`)가 넘겨받음. GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 첫 다크 블록 없음 | `TestTradingViewsThemeColorsMeetWCAGInBothThemes` | no(사후 작성) | yes(대체 시험 통과) |
| B2 | 둘째 다크 블록 없음 | `TestTradingViewsThemeColorsMeetWCAGInBothThemes` | no | yes |
| B3 | 계약 순회 | `TestTradingViewsThemeColorsMeetWCAGInBothThemes` | no | yes |
| B4 | 규칙 리터럴 없음 | `TestTradingViewsThemeColorsMeetWCAGInBothThemes`(리터럴 대신 `:root` 변수 값) | no | yes |
| B5 | 대비 미달 | `TestTradingViewsThemeColorsMeetWCAGInBothThemes` | no | yes |
