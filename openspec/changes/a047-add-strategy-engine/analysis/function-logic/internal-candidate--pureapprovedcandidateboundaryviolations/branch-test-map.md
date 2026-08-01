# Branch Test Map: `pureApprovedCandidateBoundaryViolations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch B1: aggregates dependency, type, and AST violations for approved-candidate consumers | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | existing guard baseline / a047 RED where current-life accessors were absent | yes |
