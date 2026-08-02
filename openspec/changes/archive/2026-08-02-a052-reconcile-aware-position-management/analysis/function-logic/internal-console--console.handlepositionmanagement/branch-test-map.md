# Branch Test Map: `Console.handlePositionManagement`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 200 | `if c.opts.Settings == nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `else` line 202 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `if` line 202 | `} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `else` line 204 | `} else {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 208 | `if c.opts.PositionPolicies == nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `if` line 213 | `if runtimeErr != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B7 | `else` line 215 | `} else {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B8 | `range` line 218 | `for _, block := range runtime.Blocks {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B9 | `if` line 223 | `if err != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B10 | `range` line 228 | `for _, state := range states {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B11 | `if` line 236 | `if management.Block != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B12 | `if` line 239 | `if state.Status == positionpolicy.StatusManaged {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B13 | `else` line 248 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecycleEligible() {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B14 | `range` line 241 | `for _, policy := range exitpolicy.RegisteredCommonPolicies() {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B15 | `if` line 245 | `if state.ExternalLifecycleEligible() {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B16 | `if` line 248 | `} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecycleEligible() {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
