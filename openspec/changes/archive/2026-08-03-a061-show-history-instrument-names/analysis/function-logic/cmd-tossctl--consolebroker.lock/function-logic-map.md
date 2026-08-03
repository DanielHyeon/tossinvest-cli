# Function Logic Map: `consoleBroker.lock`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | non-nil request or background context | caller | cancellation is returned before shared broker state is accessed |
| `gate` | one-token channel initialized exactly once | `gateOnce` | concurrent readers serialize on one client/token manager |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | context finishes before token acquisition | none | `ctx.Err()` | contended metadata cancellation test |
| tail | token acquired | removes one gate token | nil | shared-client concurrency test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sync.Once.Do` | make zero-value broker gate usable | exactly once | AST + race test |
| `ctx.Done` | bound request-side waiting | caller deadline | cancellation test |

## State mutations and fallbacks

- Initializes the one-token gate once; acquiring only transfers ownership and does not build or mutate a broker.

## Safety conclusion

- Safe edit boundary: synchronization only; no API call, credential, journal, or trading capability.
- High-risk impact: yes, because it protects shared authentication lifecycle and request cancellation.
