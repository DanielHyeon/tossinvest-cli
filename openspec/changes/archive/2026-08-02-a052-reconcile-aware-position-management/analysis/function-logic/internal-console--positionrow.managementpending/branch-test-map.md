# Branch Test Map: `positionRow.ManagementPending`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path | valid invocation preserves the function contract | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
