# Branch Test Map: `seedConditionalDeadlock`

- Source: `internal/console/remeasure_test.go`
- Function: `internal/console/remeasure_test.go:seedConditionalDeadlock`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED는 `internal-verifylive--cleanupfrom`과 `internal-verifylive--redoset`의
branch-test-map에 기록돼 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 373: `if err != nil {` | `seedConditionalDeadlock` | no | yes |
| B2 | `if` line 380: `if err := rec.Append(verifylive.Entry{` | `seedConditionalDeadlock` | no | yes |
| B3 | `range` line 389: `for _, step := range verifylive.Steps() {` | `seedConditionalDeadlock` | no | yes |
| B4 | `range` line 398: `for _, id := range []verifylive.StepID{` | `seedConditionalDeadlock` | no | yes |
