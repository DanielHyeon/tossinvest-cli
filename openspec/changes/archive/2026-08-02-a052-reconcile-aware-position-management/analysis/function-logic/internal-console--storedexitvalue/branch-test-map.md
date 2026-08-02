# Branch Test Map: `storedExitValue`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 151 | `if value = strings.TrimSpace(value); value != "" {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
