# Function Logic Map: `originOf`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| composite key, validity, scoped engine set, journal state | exact account/day identity | bounded journal result | unreadable/invalid is unknown or other, never engine |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal unreadable | none | unknown | unreadable ledger test |
| B2 | valid scoped engine hit | none | engine | origin distinction test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure classification | no error/retry | AST + origin tests |

## State mutations and fallbacks

- No bare broker-id lookup remains.

## Safety conclusion

- Safe edit boundary: presentation classification.
- High-risk impact: fail-closed display only.
