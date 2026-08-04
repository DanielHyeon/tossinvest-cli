# Function Logic Map: `consoleBroker.unlock`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| gate ownership | caller previously acquired the sole token | `consoleBroker.lock` | misuse would block; every production call defers unlock after successful lock |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| tail | always | returns the sole token | none | race/shared-client tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| channel send | release broker ownership | no retry; paired with successful lock | AST + race test |

## State mutations and fallbacks

- Restores gate capacity to one; it never touches the broker payload.

## Safety conclusion

- Safe edit boundary: synchronization release only.
- High-risk impact: yes, because an unmatched release would deadlock shared reads.
