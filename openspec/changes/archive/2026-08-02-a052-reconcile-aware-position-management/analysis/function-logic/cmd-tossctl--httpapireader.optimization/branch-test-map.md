# Branch Test Map: `httpAPIReader.Optimization`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 391 | `if r == nil \|\| r.optimization == nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 395 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `if` line 398 | `if r.adoptionDesired == nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 402 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 410 | `if runtime.EffectiveKnown {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
