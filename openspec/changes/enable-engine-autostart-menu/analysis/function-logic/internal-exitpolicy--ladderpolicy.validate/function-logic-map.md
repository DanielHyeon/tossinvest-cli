# Function Logic Map: `LadderPolicy.Validate`

- Source: `internal/exitpolicy/ladder.go`
- Qualified function: `LadderPolicy.Validate`
- AST evidence: `ast.json` (`66c4f4356e33a53bca02fa80c4c064058e4d03574d89138fcd85672fd07e8e40`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/exitpolicy/ladder.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/exitpolicy/ladder.go:167` — if p.PolicyID == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B2 | `if` at `internal/exitpolicy/ladder.go:170` — if len(p.Rungs) == 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B3 | `range` at `internal/exitpolicy/ladder.go:173` — for i, rung := range p.Rungs { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B4 | `if` at `internal/exitpolicy/ladder.go:174` — if err := rung.Validate(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B5 | `if` at `internal/exitpolicy/ladder.go:177` — if i == 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B6 | `if` at `internal/exitpolicy/ladder.go:183` — if target.Cmp(prevTarget) <= 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B7 | `if` at `internal/exitpolicy/ladder.go:189` — if stop.Cmp(prevStop) < 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B8 | `if` at `internal/exitpolicy/ladder.go:194` — if p.RunnerTrailPct != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B9 | `if` at `internal/exitpolicy/ladder.go:196` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B10 | `if` at `internal/exitpolicy/ladder.go:199` — if trail.Sign() <= 0 \|\| trail.Cmp(big.NewRat(100, 1)) >= 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `refusal`, `len`, `rung.Validate`, `parseRat`, `target.Cmp`, `fmt.Sprintf`, `stop.Cmp`, `err.Error`, `trail.Sign`, `trail.Cmp`, `big.NewRat` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 7 assignment(s) and 8 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
