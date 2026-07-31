# Function Logic Map: `consoleOpenAPISeam.environmentManaged`

- Source: `cmd/tossctl/console_openapi.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| key environment value | raw string | `official.LoadCredentials` | only raw empty means absent |
| secret environment value | raw string | `official.LoadCredentials` | only raw empty means absent |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | both raw values are non-empty | none | true | whitespace environment test |
| B2 | either raw value is empty | none | false | file credential tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `deps.getenv` | read the same two variable names as the official loader | no fallback logic in this helper | CodeGraph + AST |

## State mutations and fallbacks

- This helper must exactly mirror `official.LoadCredentials`; trimming here creates a different generation than the child process.

## Safety conclusion

- Safe edit boundary: remove trimming and compare raw emptiness only.
- High-risk impact: yes — a mismatch can validate one pair and spawn with another.
