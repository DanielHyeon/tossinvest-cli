# Branch Test Map: `httpAPIReader.readPositions`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 92 | `if r == nil \|\| r.holdings == nil \|\| r.accountRef == nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 96 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B3 | `if` line 100 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B4 | `if` line 107 | `if journalReadable {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B5 | `if` line 109 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B6 | `if` line 113 | `if policyErr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B7 | `range` line 116 | `for _, state := range policyStates {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B8 | `range` line 121 | `for _, row := range journalRows {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B9 | `range` line 129 | `for _, broker := range brokerRows {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B10 | `if` line 133 | `if stored, ok := byKey[key]; ok {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B11 | `else` line 141 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B12 | `if` line 137 | `if released {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B13 | `if` line 143 | `if !journalReadable {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B14 | `range` line 151 | `for _, stored := range journalRows {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B15 | `if` line 152 | `if _, ok := seen[positionKey(stored.Position.Market, stored.Position.Symbol)]; ok {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
| B16 | `if` line 162 | `if released {` entered and complementary path | go test ./cmd/tossctl | a052 RED contract or pre-existing regression | verified by focused package suite |
