# Branch Test Map: `reconcileBlockFrom`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 221 | `if value == nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 227 | `if !value.StartedAt.IsZero() {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
