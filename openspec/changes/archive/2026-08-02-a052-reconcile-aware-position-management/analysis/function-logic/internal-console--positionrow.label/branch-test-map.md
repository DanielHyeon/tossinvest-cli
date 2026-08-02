# Branch Test Map: `positionRow.Label`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 350 | `if r.HasManagementProjection() {` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `switch` line 353 | `switch {` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `case` line 354 | `case r.Unknown():` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `case` line 356 | `case !r.Managed() && r.Excluded:` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `case` line 358 | `case !r.Managed() && r.Designated:` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `case` line 360 | `case !r.Managed():` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B7 | `case` line 362 | `case r.HasExit && r.Exit.Completed:` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B8 | `case` line 364 | `case r.HasExit:` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B9 | `case` line 366 | `default:` true/entered and complementary path | go test ./internal/console | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
