# Branch Test Map: `adoptionSettingsView.StopText`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 83 | `if !v.Known {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 86 | `if v.DefaultStopPct == 0 {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
