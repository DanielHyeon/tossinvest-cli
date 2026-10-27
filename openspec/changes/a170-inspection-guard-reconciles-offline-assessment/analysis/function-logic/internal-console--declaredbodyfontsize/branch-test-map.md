# Branch Test Map: declaredBodyFontSize

- Source: `internal/console/display_primitives_test.go` (250-252). AST branches 0.
- 분기 없음. 호출 시험 `TestTheHeadingStepsAreDistinguishable` 이 통과(B3 갈래, 본문 최소 ≥15px)함을 확인. GREEN: 2026-10-28 `go test ./internal/console -run` 로 통과 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 위임 한 줄(통과 경로) | `TestTheHeadingStepsAreDistinguishable` | no(사후 작성) | yes |
