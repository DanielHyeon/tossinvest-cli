# Branch Test Map: `fakePerformanceReader.Dashboard`

- Source: `internal/console/performance_history_test.go` (20-28) — AST branches 1. 시험 파일은 coverprofile 계측 대상이 아니라, 오류를 주입하는 유일한 픽스처로 도달을 보임.

| Branch | Anchor | Scenario | Test | a177 |
|---|---|---|---|---|
| B1 | `24:2` | `err: errors.New("boom")` 주입 | `TestPerformanceWindowIsSingleSourceAcrossScreens`(loaderr 하위시험, `performance_window_test.go:67`) | 추가 |
