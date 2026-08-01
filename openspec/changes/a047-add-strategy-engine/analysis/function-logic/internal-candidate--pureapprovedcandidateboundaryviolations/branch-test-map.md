# Branch Test Map: `pureApprovedCandidateBoundaryViolations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): aggregates dependency, type and AST violation slices | `TestApprovedCandidateConsumersStayInsidePureBoundaries`, synthetic taint rejection test | existing guard baseline | pass |
