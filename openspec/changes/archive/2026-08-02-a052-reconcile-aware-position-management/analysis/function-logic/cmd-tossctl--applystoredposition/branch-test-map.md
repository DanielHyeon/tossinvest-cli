# Branch Test Map: `applyStoredPosition`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 245 | `if out.Eligible {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `else` line 247 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `if` line 252 | `if stored.HasExit {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 255 | `if view.Snapshot != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `else` line 259 | `} else if hasStoredExitEvidence(stored.Exit) {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `if` line 259 | `} else if hasStoredExitEvidence(stored.Exit) {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
