# Branch Test Map: `syntheticCandidatePackage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `range` at source line 35: `for name := range approvedCandidateAccessors {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 37: `if name == "Valid" {` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B3 | exact AST `else` at source line 39: `} else if name == "FirstSeenUnixNano" \|\| name == "LastSeenUnixNano" \|\|` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
| B4 | exact AST `if` at source line 39: `} else if name == "FirstSeenUnixNano" \|\| name == "LastSeenUnixNano" \|\|` | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | not independently measured | source test passes; branch-side coverage not claimed |
