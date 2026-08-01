# Branch Test Map: `SealApproved`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | branchless happy-path sentinel (AST has no conditional): copies only opaque candidate accessors into a value-only snapshot, preserving zero/refused values as invalid | `TestApprovedCandidateConsumersStayInsidePureBoundaries`, Parker current-life tests | yes | yes |
