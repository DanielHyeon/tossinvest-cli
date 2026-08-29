# Function Logic Map: `QFinalEntryIssuance.ClearFXAuthorityForTest`

- Source: `internal/execgw/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| issuance request | test-built request | test helper | subsequent precheck fails currency unresolved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | clears private evidence and presence bit | no error | forged q_final authority test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct private-field reset | test binary only | AST |

## State mutations and fallbacks

- Removes the only test-only authority bypass so public DTO substitution can be tested.

## Safety conclusion

- Safe edit boundary: method remains in `_test.go`.
- High-risk impact: test coverage for fail-closed production behavior.
