# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context and console flags | validated Cobra inputs | `cobra.Command`, `consoleOptions` | invalid remote/record/OpenAPI configuration returns before serving |
| journal path | resolved engine-profile path or unavailable | existing console journal resolver | failure leaves journal-backed screens and optimization lifecycle read-only |
| performance reads | `performance.db` in the selected profile data directory | `performance.Open` wrapped as `console.PerformanceReader` plus `optimization.EvidenceProvider` | open failure warns, leaves performance/evidence unavailable, and console startup continues |
| optimization store | separate `optimization-control.db` beside the journal, supplied only a narrow evidence provider | `newConsoleOptimizationCommander` | open failure warns and injects no command capability |
| runtime seams | least-capability interfaces | `console.Options` | individual optional seams fail closed without widening another seam |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | missing context or configuration resolution failure | signal context only | configured error or background fallback | existing console command tests |
| B9-B13 | journal/performance/optimization store resolution | opens derived performance DB and separate private control DB; closes both at shutdown | warning and read-only/unavailable fallback on either failure | cmd performance/evidence wiring tests |
| B14-B26 | engine marker and updater availability | optional updater/lock wiring | warning and disabled optional seam | system-update and engine wiring tests |
| B27-B34 | autostart and position-policy RPC availability | optional runtime calls; no LIVE authorization | warning/read-only fallback | engine autostart and policy wiring tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newConsoleOptimizationCommander` | binds the a050 lifecycle to its own SQLite store | one open attempt; warning and nil seam on failure | CodeGraph + AST + dedicated store test |
| `openConsolePerformanceCapabilities` | opens the profile's derived performance DB and reduces it to two read capabilities | one open attempt; warning and nil capabilities on failure | cmd production wiring tests |
| `console.ListenAndServe` | injects bounded read/command capabilities | returns the server error to `finishConsole` | CodeGraph + AST |
| `finishConsole` | preserves graceful container shutdown behavior | server failure retains precedence | existing finish-console tests |

## State mutations and fallbacks

- a050 opens the separate derived performance and control-plane databases and closes both on server exit.
- It does not open the execution journal for collection, write it, or call a broker, lane, gate, kill switch, or LIVE order seam.
- Missing evidence/provider/store wiring remains visible and read-only.

## Safety conclusion

- Safe edit boundary: optional a050 store construction and narrow interface injection.
- High-risk impact: yes; existing startup refusals remain unchanged and the new seam fails closed.
