# Function Logic Map: `openHTTPAPIResources`

- Source: `cmd/tossctl/httpapi.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal/performance/optimization resources | read-only journal plus owned local read-model stores | resolved console journal directory | required registry/store errors close acquired resources and return |
| desired adoption settings | optional config-file loader | `newAdoptionSettingsSeam` | unavailable desired state stays unavailable and never becomes effective state |
| effective runtime descriptor path | fixed private path; descriptor/socket/token resolved on each later API read | `positionpolicyrpc.RuntimeDescriptorPath(engineDir)` | engine absence/restart/replacement yields unknown or fresh truth without HTTP daemon restart |
| safety boundary | HTTP reader exposes projections only | a052 design and TossOS invariants | no cached startup runtime, reconciliation resolver, live order, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 516 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B2 | `if` at line 521 | `if journalErr == nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | go test ./cmd/tossctl |
| B3 | `if` at line 526 | `if performanceErr == nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B4 | `if` at line 530 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B5 | `if` at line 538 | `if err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B6 | `if` at line 556 | `if adoptionSettings != nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | TestHTTPAPIRuntimeFailureRemainsUnknownData; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate |
| B7 | `if` at line 559 | `if dir, err := engineJournalDir(root); err == nil {` | local resource/wiring assignment; no live command | continues with fail-closed optional seam | TestHTTPAPIRuntimeFailureRemainsUnknownData; TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate |
| B8 | `if` at line 565 | `if journalReader != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |
| B9 | `if` at line 568 | `if err := reader.validate(); err != nil {` | local resource/wiring assignment; no live command | early return/error nearby | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleJournalPath` | explicit dependency at line 515 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `journal.OpenReadOnly` | explicit dependency at line 520 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `openConsolePerformanceCapabilities` | explicit dependency at line 525 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `filepath.Dir` | explicit dependency at line 525 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `optimization.CoreRegistry` | explicit dependency at line 529 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `resources.Close` | explicit dependency at line 531 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `optimization.Open` | explicit dependency at line 534 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `filepath.Join` | explicit dependency at line 535 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newConsoleBroker` | explicit dependency at line 544 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `newAdoptionSettingsSeam` | read desired configuration separately from effective runtime state at line 545 | nil/unavailable desired seam does not fabricate effective settings | current AST + focused tests |
| `newConsoleHoldings` | explicit dependency at line 547 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleOrdersSeam` | explicit dependency at line 547 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `consoleSignalsSeam` | explicit dependency at line 548 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `shared.resolve` | explicit dependency at line 552 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `engineJournalDir` | explicit dependency at line 559 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `enginelock.MarkerPath` | explicit dependency at line 560 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionpolicyrpc.RuntimeDescriptorPath` | store the fixed private runtime descriptor path for a fresh dial on each later read at line 562 | path construction only; descriptor absence or replacement is evaluated by the per-read adapter | current AST + focused tests |
| `performancejournal.New` | explicit dependency at line 566 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `reader.validate` | explicit dependency at line 568 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 20 assignment(s), 6 return statement(s), and 0 goroutine launch(es).
- Desired adoption settings use a separate file seam. Effective runtime wiring stores only the private runtime descriptor path and reconnects on each API read.
- Engine absence/restart/descriptor replacement therefore becomes explicit unknown or fresh runtime truth without daemon restart; no cached startup snapshot is promoted.
- Resource cleanup covers only read-only journal/performance/optimization handles; no reconciliation resolver or live-order authority is added.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
