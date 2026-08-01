# Function Logic Map: `Store.Preview`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request and store actor | finite server-owned preview request; actor fixed at store construction | caller and `Store.actor` | delegated preview validates everything and inserts nothing on error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper happy path | delegates once | preview or typed error | direct store actor-binding and preview lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.preview` | binds empty rollback salt and immutable store actor | one call/no retry | actor-binding tests |

## State mutations and fallbacks

- Wrapper has no mutation of its own and never accepts a request-supplied actor.

## Safety conclusion

- Safe edit boundary: actor-bound public preview capability.
- High-risk impact: yes; capability issuance remains delegated to validated lifecycle logic.
