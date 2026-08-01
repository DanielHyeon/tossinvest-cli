# Branch Test Map: `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 402: `if err != nil {` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 405: `if !got.Valid() \|\| !got.Chase().Passed() {` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B3 | exact AST `if` at source line 408: `if got.Key() != input.Candidate.Key \|\| !got.FirstSeenAt().Equal(input.Candidate.FirstSeenAt) {` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B4 | exact AST `if` at source line 412: `if got.State() != StateActive \|\| !got.LastSeenAt().Equal(input.Candidate.LastSeenAt) \|\|` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B5 | exact AST `if` at source line 416: `if got.ThresholdVersion() != set.Version() \|\| got.SetDigest() != set.SetDigest() \|\|` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B6 | exact AST `range` at source line 423: `for index := range typ.NumField() {` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
| B7 | exact AST `if` at source line 425: `if field.IsExported() {` | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` | not independently measured | source test passes; branch-side coverage not claimed |
