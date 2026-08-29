# Function Logic Map: `Rung.Validate`

- Source: `internal/exitpolicy/ladder.go`
- Qualified function: `Rung.Validate`
- AST evidence: `ast.json` (`66c4f4356e33a53bca02fa80c4c064058e4d03574d89138fcd85672fd07e8e40`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/exitpolicy/ladder.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/exitpolicy/ladder.go:112` — if _, err := parseRat("rung target percent", r.TargetPct); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B2 | `if` at `internal/exitpolicy/ladder.go:115` — if _, err := parseRat("rung stop percent", r.StopPct); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B3 | `if` at `internal/exitpolicy/ladder.go:119` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B4 | `if` at `internal/exitpolicy/ladder.go:122` — if ratio.Sign() < 0 \|\| ratio.Cmp(one) > 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseRat`, `refusal`, `err.Error`, `parseRatio`, `ratio.Sign`, `ratio.Cmp`, `fmt.Sprintf` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 3 assignment(s) and 5 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
