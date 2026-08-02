# Branch Test Map: `adoptionStopText`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 398 | `if value <= 0 {` true/entered and complementary path | go test ./internal/httpapi | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
