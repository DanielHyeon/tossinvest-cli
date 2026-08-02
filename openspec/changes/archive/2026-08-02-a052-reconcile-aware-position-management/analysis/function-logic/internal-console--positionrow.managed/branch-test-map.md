# Branch Test Map: `positionRow.Managed`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 323 | `if r.LifecycleKnown {` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
