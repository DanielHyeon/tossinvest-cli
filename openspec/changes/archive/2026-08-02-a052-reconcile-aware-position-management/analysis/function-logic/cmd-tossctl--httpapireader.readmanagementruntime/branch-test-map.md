# Branch Test Map: `httpAPIReader.readManagementRuntime`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 172 | `if r == nil \|\| r.managementRuntime == nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 176 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
