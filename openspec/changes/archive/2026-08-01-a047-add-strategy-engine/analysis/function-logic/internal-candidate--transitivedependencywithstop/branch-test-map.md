# Branch Test Map: `transitiveDependencyWithStop`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `for` at source line 285: `for len(queue) != 0 {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 288: `if seen[current.name] {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B3 | exact AST `if` at source line 292: `if matches(current.name) {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B4 | exact AST `if` at source line 295: `if stop(current.name) {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B5 | exact AST `range` at source line 298: `for _, dependency := range graph[current.name] {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
