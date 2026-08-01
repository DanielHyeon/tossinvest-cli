# Branch Test Map: `TestApprovedCandidateTaintRejectsAuthorityInterface`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 167: `if err != nil {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 171: `if !findingContains(findings, "Gateway.Place") \|\| findingContains(findings, "Reader.ReadSymbolState") {` | `TestApprovedCandidateTaintRejectsAuthorityInterface` | not independently measured | source test passes; branch-side coverage not claimed |
