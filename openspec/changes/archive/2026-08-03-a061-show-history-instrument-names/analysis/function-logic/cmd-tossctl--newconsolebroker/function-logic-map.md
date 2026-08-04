# Function Logic Map: `newConsoleBroker`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | current console profile options | `runConsole` | retained for lazy credential/client construction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| tail | unconditional constructor | allocate resolver with production account-resolved builder | non-nil resolver | console shared-resolver tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `buildConsoleAccountBroker` method value | defer credential load, client construction, and account resolution until a read needs it | receives request context on metadata-first path | A061 cold resolver test |

## State mutations and fallbacks

- Allocates process-local resolver state only; no file or network read occurs.

## Safety conclusion

- Safe edit boundary: add a context-aware account builder field while preserving lazy construction.
- High-risk impact: no direct side effect, but the returned resolver feeds account-scoped reads and is covered by identity/account-resolution tests.
