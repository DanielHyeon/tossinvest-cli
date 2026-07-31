# Function Logic Map: `validateRecoveryProjection`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| snapshot output + remaining quantity | known suppression; orderable ratio `(0,1]`; exact integer projection; coherent flags | immutable evaluator output | `ErrRecoveryIdentity`, no normalization |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unknown suppression | none | refuse | semantic output table |
| B2 | orderable ratio/level/suppression invalid | none | refuse | zero/over-one/invalid ratio tests |
| B3 | recomputed whole-share projection differs | none | refuse | fractional/wrong projection tests |
| B4 | orderable/state-only flags disagree | none | refuse | flag mismatch tests |
| B5 | nonorderable carries proposal fields | none | refuse | nonorderable table |
| B6 | ladder hold vs ordinary state-only semantics | none | accept only valid hold/zero projection | ladder snapshot tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ProjectWholeShares` | independently recompute integer quantity from remaining and ratio | error/refusal; no rounding fallback | CodeGraph + AST |
| `parseRatio` | exact rational range validation | error/refusal | CodeGraph + AST |

## State mutations and fallbacks

- Pure semantic validation; stored strings are never repaired or coerced.

## Safety conclusion

- Safe edit boundary: snapshot proposal/projection evidence.
- High-risk impact: yes; malformed order evidence must never become executable.
