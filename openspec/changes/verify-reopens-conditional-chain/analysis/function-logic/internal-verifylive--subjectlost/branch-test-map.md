# Branch Test Map: `subjectLost`

- Source: `internal/verifylive/redo.go`
- Function: `internal/verifylive/redo.go:subjectLost`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED는 `internal-verifylive--cleanupfrom`과 `internal-verifylive--redoset`의
branch-test-map에 기록돼 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 109: `if newest.Verdict != VerdictPass {` | `subjectLost` | no | yes |
| B2 | `range` line 113: `for _, a := range newest.Artifacts {` | `subjectLost` | no | yes |
| B3 | `if` line 114: `if a.Kind == KindConditional && a.Deliberate {` | `subjectLost` | no | yes |
| B4 | `if` line 119: `if !left {` | `subjectLost` | no | yes |
| B5 | `range` line 122: `for _, a := range Outstanding(entries) {` | `subjectLost` | no | yes |
| B6 | `if` line 123: `if a.Kind == KindConditional {` | `subjectLost` | no | yes |
| B7 | `range` line 127: `for _, dep := range Steps() {` | `subjectLost` | no | yes |
| B8 | `if` line 128: `if dep.Deferred != "" \|\| !dependsOn(dep, step.ID) {` | `subjectLost` | no | yes |
| B9 | `if` line 131: `if !Passed(entries, dep.ID) {` | `subjectLost` | no | yes |
