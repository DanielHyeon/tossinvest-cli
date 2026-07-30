# Branch Test Map: `TestRedoableVerdictAgreesWithTheSet`

- Source: `internal/verifylive/redo_test.go`
- Function: `internal/verifylive/redo_test.go:TestRedoableVerdictAgreesWithTheSet` (base revision `de14674974ab`)

이 change는 이 함수를 수정하지 않았다. RED 열이 전부 `no`인 것은 그 사실의 기록이다 —
실패 상태를 만들지 않았고, 이 함수는 구현 전후 모두 통과했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` base line 113: `for verdict, want := range cases {` | `TestRedoableVerdictAgreesWithTheSet` | no | yes |
| B2 | `if` base line 114: `if got := RedoableVerdict(verdict); got != want {` | `TestRedoableVerdictAgreesWithTheSet` | no | yes |
