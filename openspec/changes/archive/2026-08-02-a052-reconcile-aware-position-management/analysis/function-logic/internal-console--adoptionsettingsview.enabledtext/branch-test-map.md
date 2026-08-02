# Branch Test Map: `adoptionSettingsView.EnabledText`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 73 | `if !v.Known {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 76 | `if v.Enabled {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
