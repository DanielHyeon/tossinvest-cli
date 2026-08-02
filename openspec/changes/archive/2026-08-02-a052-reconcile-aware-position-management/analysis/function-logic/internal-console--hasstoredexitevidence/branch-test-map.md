# Branch Test Map: `hasStoredExitEvidence`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path | valid invocation | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
