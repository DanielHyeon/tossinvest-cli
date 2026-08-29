# Function Logic Map: `parseRatio`

- Source: `internal/exitpolicy/decimal.go`
- Qualified function: `parseRatio`
- AST evidence: `ast.json` (`eded0b568d7c050b7249fb7ca92cebca1cb1cde5949540f2b7b490b740d77ca4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/exitpolicy/decimal.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/exitpolicy/decimal.go:82` — if !strings.Contains(raw, "/") { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B2 | `if` at `internal/exitpolicy/decimal.go:86` — if !ok \|\| numerator == "" \|\| denominator == "" \|\| | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B3 | `if` at `internal/exitpolicy/decimal.go:91` — if !ok \|\| r.Denom().Sign() <= 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`, `strings.Contains`, `parseRat`, `strings.Cut`, `onlyDigits`, `fmt.Errorf`, `SetString`, `new`, `Sign`, `r.Denom` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 3 assignment(s) and 4 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
