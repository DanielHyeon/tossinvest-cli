# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| lifecycle descriptor/client | optional private engine command descriptor resolved once at console startup | `positionpolicyrpc.DescriptorPath` + engine-owned Unix server | missing/dial failure leaves policy commands read-only |
| runtime descriptor path | fixed private path only; the descriptor/socket/token are resolved afresh by the adapter on every read | `positionpolicyrpc.RuntimeDescriptorPath(engineDir)` | late engine start, restart, or descriptor replacement is observed without console restart; failure remains runtime-unknown |
| desired adoption settings | config-file seam separate from runtime authority | console settings loader | never substitutes for unavailable effective runtime settings |
| safety boundary | console receives clients/read seams, never an engine journal write handle | a052 design and TossOS invariants | no reconciliation resolution, live order, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 203 | `if ctx == nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B2 | `if` at line 212 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B3 | `if` at line 217 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B4 | `if` at line 221 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B5 | `if` at line 225 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B6 | `if` at line 229 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B7 | `if` at line 233 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B8 | `if` at line 238 | `if err != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B9 | `if` at line 246 | `if journalPath != "" {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B10 | `if` at line 248 | `if err != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B11 | `else` at line 251 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B12 | `if` at line 255 | `if err != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B13 | `else` at line 258 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B14 | `if` at line 269 | `if dir, derr := engineJournalDir(root); derr == nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B15 | `else` at line 272 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B16 | `if` at line 280 | `if os.Getenv("TOSSOS_CONTAINER") == "1" {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B17 | `else` at line 283 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B18 | `if` at line 283 | `} else if self, serr := binstamp.SelfPath(); serr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B19 | `else` at line 285 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B20 | `if` at line 287 | `if cerr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B21 | `else` at line 295 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B22 | `if` at line 289 | `if updater, uerr := localupdate.New(self); uerr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B23 | `else` at line 291 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B24 | `if` at line 298 | `if updater != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B25 | `if` at line 302 | `if uerr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B26 | `else` at line 304 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B27 | `if` at line 311 | `if engineDir != "" {` | local resource/wiring assignment; no live command | early return/error nearby | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate |
| B28 | `if` at line 314 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B29 | `if` at line 330 | `if engineBoot != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B30 | `if` at line 337 | `if engineBootNote != "" {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B31 | `if` at line 345 | `if engineDir != "" {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate |
| B32 | `if` at line 347 | `if _, statErr := os.Stat(descriptorPath); statErr == nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes; TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate |
| B33 | `else` at line 359 | `} else if !errors.Is(statErr, os.ErrNotExist) {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B34 | `if` at line 349 | `if dialErr != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B35 | `else` at line 351 | `} else {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B36 | `if` at line 359 | `} else if !errors.Is(statErr, os.ErrNotExist) {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.Context` | explicit dependency at line 202 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `context.Background` | explicit dependency at line 204 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `signal.NotifyContext` | explicit dependency at line 208 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleTerminationSignals` | explicit dependency at line 208 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `stop` | explicit dependency at line 209 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `remoteAccessOptions` | explicit dependency at line 211 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resolveVerifyRecord` | explicit dependency at line 216 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resolveVerifyRecordFor` | explicit dependency at line 220 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resolveSoakRecord` | explicit dependency at line 224 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resolveSoakAttestationPath` | explicit dependency at line 228 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newConsoleOpenAPISeam` | explicit dependency at line 232 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `defaultConsoleOpenAPIDeps` | explicit dependency at line 232 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleJournalPath` | explicit dependency at line 237 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `fmt.Fprintf` | explicit dependency at line 241 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `cmd.ErrOrStderr` | explicit dependency at line 241 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `openConsolePerformanceCapabilities` | explicit dependency at line 247 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `filepath.Dir` | explicit dependency at line 247 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `performanceCapabilities.Close` | explicit dependency at line 252 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newConsoleOptimizationCommander` | explicit dependency at line 254 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `optimizationCommander.Close` | explicit dependency at line 259 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `engineJournalDir` | explicit dependency at line 269 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `enginelock.MarkerPath` | explicit dependency at line 271 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `cmd.OutOrStdout` | explicit dependency at line 276 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `os.Getenv` | explicit dependency at line 280 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `fmt.Fprintln` | explicit dependency at line 281 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `binstamp.SelfPath` | explicit dependency at line 283 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resolveUpdateCachePath` | explicit dependency at line 286 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `localupdate.New` | explicit dependency at line 289 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `assembleConsoleSystemUpdate` | explicit dependency at line 296 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `filepath.Join` | explicit dependency at line 297 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `enginelock.Acquire` | explicit dependency at line 313 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newConsoleBroker` | explicit dependency at line 327 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleEngineBootSeam` | explicit dependency at line 328 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `runConfiguredEngineAutostart` | explicit dependency at line 333 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `startEngine` | explicit dependency at line 335 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicyrpc.DescriptorPath` | explicit dependency at line 346 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `os.Stat` | explicit dependency at line 347 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicyrpc.Dial` | open the lifecycle command client owned by the running engine at line 348 | startup failure leaves the console read-only; it never opens the journal writable | current AST + focused tests |
| `positionpolicyrpc.RuntimeDescriptorPath` | store the fixed private runtime descriptor path for a fresh dial on each later read at line 355 | path construction only; descriptor absence or replacement is evaluated by the per-read adapter | current AST + focused tests |
| `errors.Is` | explicit dependency at line 359 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `console.ListenAndServe` | explicit dependency at line 364 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleVerifyStarter` | explicit dependency at line 367 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `soak.DefaultCriteria` | explicit dependency at line 372 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `engine.RequiredEndpoints` | explicit dependency at line 373 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strictVerifyActivity` | explicit dependency at line 380 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `verifyRunLockPath` | explicit dependency at line 380 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `UTC` | explicit dependency at line 380 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `time.Now` | explicit dependency at line 380 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newConsoleHoldings` | explicit dependency at line 386 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleOrdersSeam` | explicit dependency at line 389 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleSettingsSeam` | explicit dependency at line 392 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleExitPolicySettingsSeam` | explicit dependency at line 393 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleMarketScheduleSeam` | explicit dependency at line 400 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleGateLimitsSeam` | explicit dependency at line 405 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleLimitSettingsSeam` | explicit dependency at line 411 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleTradingPolicySeam` | explicit dependency at line 417 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleGateSwitchSeam` | explicit dependency at line 418 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleSignalsSeam` | explicit dependency at line 425 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleRelaunch` | explicit dependency at line 430 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `handoff.New` | explicit dependency at line 431 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleHandoffPath` | explicit dependency at line 431 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `restartSoak` | explicit dependency at line 433 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `toConsoleOpenAPICredentialCheck` | explicit dependency at line 436 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `openAPISeam.Check` | explicit dependency at line 436 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `openAPISeam.Save` | explicit dependency at line 439 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `stopEngine` | explicit dependency at line 452 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `finishConsole` | explicit dependency at line 455 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 41 assignment(s), 17 return statement(s), and 0 goroutine launch(es).
- Lifecycle command connectivity is resolved at startup, while runtime truth stores only `RuntimeDescriptorPath`; the adapter re-dials that descriptor on every read so an engine started or replaced later becomes visible without console restart.
- Missing lifecycle descriptor keeps the page read-only; missing/replaced runtime descriptor yields explicit runtime-unavailable/unknown, never desired-as-effective.
- The console receives no journal write handle, reconciliation resolver, or live-order authority.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
