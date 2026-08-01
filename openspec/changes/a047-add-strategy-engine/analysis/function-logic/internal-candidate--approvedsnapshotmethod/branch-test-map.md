# Branch Test Map: `approvedSnapshotMethod`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/missing/multiple receiver declaration is rejected | `TestApprovedCandidateConsumersStayInsidePureBoundaries` | existing guard baseline | pass |
| B2 | receiver must be exactly `ApprovedSnapshot` | same static boundary test | existing guard baseline | pass |
| B3 | method name, zero-parameter and one-result contract must match the closed accessor allowlist | same static boundary test | current-life accessor additions initially absent | pass |
