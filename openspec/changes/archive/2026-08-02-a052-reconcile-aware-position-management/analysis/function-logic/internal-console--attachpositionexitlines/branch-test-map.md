# Branch Test Map: `attachPositionExitLines`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `range` line 103 | `for i := range rows {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 105 | `if !row.HasExit {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `if` line 108 | `if row.LifecycleKnown && row.LifecycleStatus == positionpolicy.StatusReleased {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 109 | `if hasStoredExitEvidence(row.Exit) {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 129 | `if snapshot.Snapshot != nil {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `else` line 133 | `} else if hasStoredExitEvidence(row.Exit) {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
| B7 | `if` line 133 | `} else if hasStoredExitEvidence(row.Exit) {` entered and complementary path | go test ./internal/console | a052 RED contract or pre-existing regression | verified by focused package suite |
