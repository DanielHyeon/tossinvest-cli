# Branch Test Map: `ApprovedCandidate.Valid`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): approved and zero-value accessor behavior | `TestAssessApprovedCandidateReturnsPassWithImmutableProvenance` plus zero-value refusal tables | existing | yes |
