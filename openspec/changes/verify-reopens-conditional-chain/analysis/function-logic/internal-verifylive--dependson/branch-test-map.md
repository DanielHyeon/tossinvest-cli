# Branch Test Map: `dependsOn`

- Source: `internal/verifylive/redo.go`
- Function: `internal/verifylive/redo.go:dependsOn`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED는 `internal-verifylive--cleanupfrom`과 `internal-verifylive--redoset`의
branch-test-map에 기록돼 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 139: `for _, d := range s.DependsOn {` | `dependsOn` | no | yes |
| B2 | `if` line 140: `if d == id {` | `dependsOn` | no | yes |
