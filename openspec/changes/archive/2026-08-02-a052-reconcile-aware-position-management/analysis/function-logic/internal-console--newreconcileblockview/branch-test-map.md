# Branch Test Map: `newReconcileBlockView`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `switch` line 138 | `switch block.Scope {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `case` line 139 | `case positionpolicy.ScopeMarket:` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `case` line 141 | `case positionpolicy.ScopeSymbol:` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 146 | `if !block.StartedAt.IsZero() {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 149 | `if d < 0 {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `switch` line 152 | `switch {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B7 | `case` line 153 | `case d < time.Minute:` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B8 | `case` line 155 | `case d < 24*time.Hour:` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B9 | `case` line 157 | `default:` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
