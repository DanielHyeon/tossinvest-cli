# Function Logic Map: `newRootCmd`

- Source: `cmd/tossctl/root.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| process environment, persistent flags and command arguments | Cobra-parsed values; output is table/json/csv | `rootOptions`, Cobra flag registry | validation errors are returned before a leaf handler |
| session/config/update state used by persistent hooks | profile-resolved read-only state | existing session/config/update services | missing advisory state suppresses a notice; it never enables a command |
| root command registry | fixed, statically constructed command set | `newRootCmd` | an unregistered command is unreachable and Cobra returns an error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | output format parse fails | none | return validation error | existing root output-format tests |
| B2 | update cache path resolves | constructs advisory gates only | continue | existing update-notice tests |
| B3 | table output and onboarding needed | writes one hint to stderr | continue | existing onboarding hint tests |
| B4 | command registration | adds fixed leaf/subcommand constructors | return fully built root command | `TestHTTPAPICommandIsRegisteredReadOnlyAndNonMutating` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHTTPAPICmd` | register the isolated private REST/SSE daemon | construction only; no listener or engine starts during root assembly | CodeGraph + AST + command contract test |
| existing `new*Cmd` constructors | preserve all prior CLI capabilities | unchanged constructors and order-independent registration | CodeGraph + AST + full command convention tests |
| `writeExpiryWarningIfNeeded`, `writeConfigLegacyWarningIfNeeded` | keep read-only startup advisories | failures never broaden authority | CodeGraph + AST + existing root tests |

## State mutations and fallbacks

- Allocates only in-memory Cobra command/flag structures.
- Persistent hooks read profile/session/update state and may write notices; they do not mutate trading state.
- Adding `httpapi` must not annotate it `mutating:true`, start the engine, or construct a broker writer.

## Safety conclusion

- Safe edit boundary: append the fixed `newHTTPAPICmd(opts)` constructor to the root registry; leave persistent hooks and every existing command untouched.
- High-risk impact: low at construction time, security-sensitive at runtime. Static command tests must prove no LIVE/order authority and runtime tests must prove read-only journal seams.
