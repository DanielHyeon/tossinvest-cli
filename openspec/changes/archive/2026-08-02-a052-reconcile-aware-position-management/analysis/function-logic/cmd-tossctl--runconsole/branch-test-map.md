# Branch Test Map: `runConsole`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 203 | `if ctx == nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B2 | `if` line 212 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B3 | `if` line 217 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B4 | `if` line 221 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B5 | `if` line 225 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B6 | `if` line 229 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B7 | `if` line 233 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B8 | `if` line 238 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B9 | `if` line 246 | `if journalPath != "" {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B10 | `if` line 248 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B11 | `else` line 251 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B12 | `if` line 255 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B13 | `else` line 258 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B14 | `if` line 269 | `if dir, derr := engineJournalDir(root); derr == nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B15 | `else` line 272 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B16 | `if` line 280 | `if os.Getenv("TOSSOS_CONTAINER") == "1" {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B17 | `else` line 283 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B18 | `if` line 283 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B19 | `else` line 285 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B20 | `if` line 287 | `if cerr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B21 | `else` line 295 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B22 | `if` line 289 | `if updater, uerr := localupdate.New(self); uerr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B23 | `else` line 291 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B24 | `if` line 298 | `if updater != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B25 | `if` line 302 | `if uerr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B26 | `else` line 304 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B27 | `if` line 311 | `if engineDir != "" {` entered and complementary path | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B28 | `if` line 314 | `if err != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B29 | `if` line 330 | `if engineBoot != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B30 | `if` line 337 | `if engineBootNote != "" {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B31 | `if` line 345 | `if engineDir != "" {` entered and complementary path | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B32 | `if` line 347 | `if _, statErr := os.Stat(descriptorPath); statErr == nil {` entered and complementary path | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B33 | `else` line 359 | `} else if !errors.Is(statErr, os.ErrNotExist) {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B34 | `if` line 349 | `if dialErr != nil {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B35 | `else` line 351 | `} else {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
| B36 | `if` line 359 | `} else if !errors.Is(statErr, os.ErrNotExist) {` entered and complementary path | go test ./cmd/tossctl | a052 runtime-wiring regression | verified by focused cmd/tossctl suite |
