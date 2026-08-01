# Function Logic Map: `FlattenAuthorization.Consume`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque authorization, target, quantity | same permit pointer, exact scope/broker/quantity, permit-owned current time in issue/deadline, unused | prior `decideFlatten` result | `ErrFlattenAuthorization` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil clock/permit or target/quantity mismatch | none | authorization error | zero/wrong scope/broker/quantity table |
| B2 | permit clock before issue or after deadline | clock read only | authorization error | +1h replay test |
| B3 | atomic one-shot CAS loses | atomic read-modify-write | authorization error | copied permit replay and race |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| permit-owned clock and `atomic.Bool.CompareAndSwap` | prevent caller-time replay and copy replay | no reset/fallback | CodeGraph + AST |

## State mutations and fallbacks

- Exactly one successful call flips shared consumed state; it performs no broker mutation.

## Safety conclusion

- Safe edit boundary: narrow consumable proof bound to the previously verified cancel/sellable snapshot.
- High-risk impact: yes; destructive-action precondition, still dormant.
