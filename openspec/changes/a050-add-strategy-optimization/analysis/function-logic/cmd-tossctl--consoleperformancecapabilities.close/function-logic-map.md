# Function Logic Map: `consolePerformanceCapabilities.Close`

- Source: `cmd/tossctl/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| close closure | nil or performance store close | constructor | nil is no-op; close error returned |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | closure absent | none | nil | lifecycle test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| stored close closure | release derived DB | no retry | lifecycle test |

## State mutations and fallbacks

- Releases only the derived read store.

## Safety conclusion

- Safe edit boundary: resource lifecycle; high-risk impact: no.
