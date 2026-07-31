# Function Logic Map: `managedPolicyIdentity`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| runtime/read identity | complete valid identity or pre-a042 zero | journal read model | invalid/unknown fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | any identity field supplied | validate and preserve tuple | validation error | seed/read tests |
| B2 | zero ladder identity | resolve pinned ID/adoption variant | unknown conflict | ladder reinterpretation test |
| B3 | zero ratchet identity | resolve pinned default | conflict on active mismatch later | ratchet tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PolicyIdentity.Validate` | reject partial tuple | fail closed | AST |
| `Legacy*PolicyIdentity` | exact compatibility only | unknown is conflict | AST |

## State mutations and fallbacks

- Does not consult today's mutable registry for a legacy row.

## Safety conclusion

- Safe edit boundary: pure identity resolution.
- High-risk impact: yes — controls whether evaluation may run.
