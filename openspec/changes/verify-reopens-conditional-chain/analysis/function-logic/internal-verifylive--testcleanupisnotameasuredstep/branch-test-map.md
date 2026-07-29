# Branch Test Map: `TestCleanupIsNotAMeasuredStep`

- Source: `internal/verifylive/cleanup_test.go`
- Function: `internal/verifylive/cleanup_test.go:TestCleanupIsNotAMeasuredStep` (base revision `de14674974ab`)

이 change는 이 함수를 수정하지 않았다. RED 열이 전부 `no`인 것은 그 사실의 기록이다 —
실패 상태를 만들지 않았고, 이 함수는 구현 전후 모두 통과했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` base line 228: `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B2 | `range` base line 234: `for i := range entries {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B3 | `if` base line 235: `if entries[i].StepID == StepCleanup {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B4 | `if` base line 239: `if cleanup == nil {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B5 | `if` base line 242: `if cleanup.Kind != KindCleanup {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B6 | `range` base line 248: `for _, e := range entries {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B7 | `if` base line 249: `if e.Kind == KindStep && e.StepID != StepCleanup {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B8 | `if` base line 253: `if stepsAfter != measured {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B9 | `range` base line 257: `for _, id := range RedoSet(entries) {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
| B10 | `if` base line 258: `if id == StepCleanup {` | `TestCleanupIsNotAMeasuredStep` | no | yes |
