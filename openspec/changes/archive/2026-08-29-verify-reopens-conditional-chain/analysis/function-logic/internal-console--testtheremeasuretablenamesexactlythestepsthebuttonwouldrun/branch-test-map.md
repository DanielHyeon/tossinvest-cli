# Branch Test Map: `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun`

- Source: `internal/console/remeasure_test.go`
- Function: `internal/console/remeasure_test.go:TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED는 `internal-verifylive--cleanupfrom`과 `internal-verifylive--redoset`의
branch-test-map에 기록돼 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 434: `if err != nil {` | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | no | yes |
| B2 | `if` line 438: `if len(want) == 0 {` | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | no | yes |
| B3 | `range` line 443: `for _, id := range want {` | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | no | yes |
| B4 | `if` line 444: `if !strings.Contains(section, "<code>"+string(id)+"</code>") {` | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | no | yes |
| B5 | `if` line 449: `if got := strings.Count(section, "<tr><td><code>"); got != len(want) {` | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | no | yes |
