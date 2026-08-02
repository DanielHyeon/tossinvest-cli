# Branch Test Map: `httpAPIReader.validate`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 479 | `if r == nil \|\| r.holdings == nil \|\| r.orders == nil \|\| r.signals == nil \|\| r.accountRef == nil \|\|` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
