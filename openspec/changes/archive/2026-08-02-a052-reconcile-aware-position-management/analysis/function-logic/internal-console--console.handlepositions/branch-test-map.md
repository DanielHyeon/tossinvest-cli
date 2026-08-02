# Branch Test Map: `Console.handlePositions`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 49 | `if c.opts.PositionPolicies != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 55 | `if states, err := c.opts.PositionPolicies.List(r.Context()); err == nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `range` line 57 | `for _, state := range states {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 62 | `if c.opts.Settings != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 63 | `if block, _, err := c.opts.Settings.Load(); err == nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `range` line 66 | `for i := range page.Snap.Rows {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B7 | `if` line 72 | `if runtimeAttempted {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B8 | `range` line 73 | `for i := range page.Snap.Rows {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B9 | `if` line 77 | `if row.InJournal {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B10 | `if` line 79 | `if !ok {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B11 | `else` line 81 | `} else {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B12 | `if` line 90 | `if row.Management.Block != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
