# Function Logic Map: `consoleMarketScheduleSeam`

- Source: `cmd/tossctl/marketschedule.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root | selected config directory/profile | `configFilePath` | path error is retained for read-time fail-closed UI |
| optional calendar reader | zero or one narrow typed official read seam | `runConsole` shared broker | absent reader is safe for market-none; selected market read fails |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | config path resolution fails | stores error only | returns inert reader | console path/error contract |
| B2 | calendar reader supplied | stores first narrow reader | continue | production provenance test |
| success | path valid | constructs desired store and UTC clock | reader only; no I/O yet | closed-default test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `configFilePath` | profile-local desired state path | errors deferred to page read | CodeGraph + AST |
| `scheduler.NewDesiredStore` | read desired state beside config | constructor has no side effect | CodeGraph + AST |

## State mutations and fallbacks

- Constructor performs no network, activation, save or toggle. Calendar fetch is request-time and only for a selected market.

## Safety conclusion

- Safe edit boundary: read-only console projection assembly.
- High-risk impact: no mutation authority; provenance failure only degrades the page.
