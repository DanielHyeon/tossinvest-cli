# Branch Test Map: `decidedAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:decidedAfter`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED는 `internal-verifylive--cleanupfrom`과 `internal-verifylive--redoset`의
branch-test-map에 기록돼 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 152: `for i := range entries {` | `decidedAfter` | no | yes |
| B2 | `range` line 153: `for _, x := range entries[i].Artifacts {` | `decidedAfter` | no | yes |
| B3 | `if` line 154: `if x.Kind == a.Kind && x.ID == a.ID {` | `decidedAfter` | no | yes |
| B4 | `if` line 159: `if created >= 0 {` | `decidedAfter` | no | yes |
| B5 | `if` line 163: `if created < 0 {` | `decidedAfter` | no | yes |
| B6 | `range` line 167: `for i := range entries {` | `decidedAfter` | no | yes |
| B7 | `if` line 168: `if entries[i].StepID == id {` | `decidedAfter` | no | yes |
