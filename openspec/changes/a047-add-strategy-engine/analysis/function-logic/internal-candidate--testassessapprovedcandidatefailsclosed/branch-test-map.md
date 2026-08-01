# Branch Test Map: `TestAssessApprovedCandidateFailsClosed`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | exact AST `range` at source line 358: `for _, tc := range []struct {` | `TestAssessApprovedCandidateFailsClosed` | not independently measured | source test passes; branch-side coverage not claimed |
| B2 | exact AST `if` at source line 381: `if got != (ApprovedCandidate{}) {` | `TestAssessApprovedCandidateFailsClosed` | not independently measured | source test passes; branch-side coverage not claimed |
| B3 | exact AST `if` at source line 385: `if !errors.As(err, &approvalErr) {` | `TestAssessApprovedCandidateFailsClosed` | not independently measured | source test passes; branch-side coverage not claimed |
| B4 | exact AST `if` at source line 388: `if approvalErr.Kind() != tc.wantKind {` | `TestAssessApprovedCandidateFailsClosed` | not independently measured | source test passes; branch-side coverage not claimed |
| B5 | exact AST `if` at source line 391: `if !reflect.DeepEqual(approvalErr.Vetoes(), tc.wantVetoes) {` | `TestAssessApprovedCandidateFailsClosed` | not independently measured | source test passes; branch-side coverage not claimed |
