# Branch Test Map: `allowedApprovedSnapshotDeclaration`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1: allows only SealApproved and audited ApprovedSnapshot methods | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | existing guard baseline / a047 RED where current-life accessors were absent | yes |
| B2 | AST branch B2: allows only SealApproved and audited ApprovedSnapshot methods | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | existing guard baseline / a047 RED where current-life accessors were absent | yes |
