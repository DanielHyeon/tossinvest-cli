# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context and console flags | validated Cobra inputs | `cobra.Command`, `consoleOptions` | invalid remote/record/OpenAPI configuration returns before serving |
| journal path | resolved engine-profile path or unavailable | existing console journal resolver | failure leaves journal-backed screens and optimization lifecycle read-only |
| optimization store | separate `optimization-control.db` beside the journal | `newConsoleOptimizationCommander` | open failure warns and injects no command capability |
| runtime seams | least-capability interfaces | `console.Options` | individual optional seams fail closed without widening another seam |

## Branches and early returns

| Branch group | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | missing context or configuration resolution failure | signal context only | configured error or background fallback | existing console command tests |
| B9-B13 | journal/optimization store resolution | opens separate private control DB only | warning and read-only fallback on failure | `TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore` |
| B14-B26 | engine marker and updater availability | optional updater/lock wiring | warning and disabled optional seam | system-update and engine wiring tests |
| B27-B34 | autostart and position-policy RPC availability | optional runtime calls; no LIVE authorization | warning/read-only fallback | engine autostart and policy wiring tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newConsoleOptimizationCommander` | binds the a050 lifecycle to its own SQLite store | one open attempt; warning and nil seam on failure | CodeGraph + AST + dedicated store test |
| `console.ListenAndServe` | injects bounded read/command capabilities | returns the server error to `finishConsole` | CodeGraph + AST |
| `finishConsole` | preserves graceful container shutdown behavior | server failure retains precedence | existing finish-console tests |

## State mutations and fallbacks

- a050 opens only its separate control-plane database and closes it on server exit.
- It does not write the execution journal or call a broker, lane, gate, kill switch, or LIVE order seam.
- Missing evidence/provider/store wiring remains visible and read-only.

## Safety conclusion

- Safe edit boundary: optional a050 store construction and narrow interface injection.
- High-risk impact: yes; existing startup refusals remain unchanged and the new seam fails closed.
