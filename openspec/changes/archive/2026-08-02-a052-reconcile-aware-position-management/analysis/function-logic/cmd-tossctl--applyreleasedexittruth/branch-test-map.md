# Branch Test Map: `applyReleasedExitTruth`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 211 | `if hasStoredExitEvidence(stored.Exit) {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
