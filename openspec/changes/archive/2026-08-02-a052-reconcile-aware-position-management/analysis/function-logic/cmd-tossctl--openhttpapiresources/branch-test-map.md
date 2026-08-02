# Branch Test Map: `openHTTPAPIResources`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 516 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B2 | `if` line 521 | `if journalErr == nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B3 | `if` line 526 | `if performanceErr == nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B4 | `if` line 530 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B5 | `if` line 538 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B6 | `if` line 556 | `if adoptionSettings != nil {` entered and complementary path | TestHTTPAPIRuntimeFailureRemainsUnknownData; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B7 | `if` line 559 | `if dir, err := engineJournalDir(root); err == nil {` entered and complementary path | TestHTTPAPIRuntimeFailureRemainsUnknownData; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B8 | `if` line 565 | `if journalReader != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B9 | `if` line 568 | `if err := reader.validate(); err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
