# Branch Test Map: `authorityInterfaceDeclarations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `range` at source line 140: `for _, file := range files {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 143: `if !ok {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B3 | exact AST `if` at source line 147: `if !ok {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B4 | exact AST `range` at source line 150: `for _, field := range contract.Methods.List {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B5 | exact AST `range` at source line 151: `for _, name := range field.Names {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B6 | exact AST `if` at source line 152: `if authorityMethodNames[name.Name] {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
