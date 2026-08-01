# Branch Test Map: `allowedApprovedSnapshotDeclaration`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil function declaration is rejected | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | existing guard baseline | pass |
| B2 | receiver declaration is accepted only through the audited `approvedSnapshotMethod` predicate | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | current-life accessor additions initially absent | pass |
