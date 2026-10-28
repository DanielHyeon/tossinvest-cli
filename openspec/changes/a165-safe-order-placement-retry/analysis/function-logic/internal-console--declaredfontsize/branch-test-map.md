# Branch Test Map: declaredFontSize

- Source: `internal/console/display_primitives_test.go` (230-248). AST branches 4.
- 호출자는 `TestTheHeadingStepsAreDistinguishable`(직접 · `declaredBodyFontSize` 경유). 현재 CSS 에서 h1·h2·h3·body 는 전부 `var(--text-N)` 이므로 통과 시 지나는 갈래는 B3 이다(B4 는 rem 리터럴 규칙용, 현재 CSS 에서 미실행). GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 규칙 시작 없음 | `TestTheHeadingStepsAreDistinguishable`(규칙이 사라지면 B2/B3 실패 갈래로 보고) | no(사후 작성) | 미실행(현재 CSS 에 네 규칙 모두 있음) |
| B2 | 닫는 괄호 없음 | `TestTheHeadingStepsAreDistinguishable` | no | 미실행 |
| B3 | 토큰 크기 | `TestTheHeadingStepsAreDistinguishable` | no | yes |
| B4 | rem 리터럴 크기 | `TestTheHeadingStepsAreDistinguishable` | no | 미실행(현재 CSS 에 해당 규칙 없음) |
