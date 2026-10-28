# Branch Test Map: TestTheHeadingStepsAreDistinguishable

- Source: `internal/console/display_primitives_test.go` (184-209). AST branches 6.
- 시험 함수 자체의 갈래다. 통과 갈래(B1·B4)는 시험이 돌 때마다 지나고, 실패 갈래(B2·B3·B5·B6)는 CSS 계약이 깨질 때만 지난다. GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | h1·h2·h3 순회 | `TestTheHeadingStepsAreDistinguishable` | no(사후 작성) | yes |
| B2 | 단계 크기 선언 없음 | `TestTheHeadingStepsAreDistinguishable`(실패 갈래 — 통과 시 미실행) | no | 미실행(설계상) |
| B3 | 본문 크기 선언 없음 | `TestTheHeadingStepsAreDistinguishable`(실패 갈래) | no | 미실행(설계상) |
| B4 | 양끝 뷰포트 순회 | `TestTheHeadingStepsAreDistinguishable` | no | yes |
| B5 | 단계 순서 깨짐 | `TestTheHeadingStepsAreDistinguishable`(실패 갈래) | no | 미실행(설계상) |
| B6 | 본문 최소 < 15px | `TestTheHeadingStepsAreDistinguishable`(실패 갈래) | no | 미실행(설계상) |
