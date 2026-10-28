# Branch Test Map: limitCurrent

- Source: `internal/console/settings_limits_test.go` (597-608) — base 판. AST branches 2. 본문 무변경.
- 호출 시험(`settings_limits_test.go:210` · `a136_daily_buy_settings_test.go:60` · `a138_open_positions_settings_test.go:79`): `TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault` · `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestA138TheScreenSaysWhatTheOpenPositionMaximumIs`. GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인 — 통과 경로(표가 있고 닫힘)는 세 시험 모두 지남.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표 시작 없음 | `TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault`(실패 갈래 — 통과 시 미실행) | no(사후 작성) | 미실행(설계상) |
| B2 | 표 닫힘 없음 | `TestUnsetLimitsRenderAsUnsetAndNotAsTheDefault`(실패 갈래) | no | 미실행(설계상) |
