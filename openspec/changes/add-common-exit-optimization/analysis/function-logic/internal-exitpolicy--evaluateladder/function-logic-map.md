# Function Logic Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go`
- Qualified function: `EvaluateLadder`
- AST evidence: `ast.json` (`66c4f4356e33a53bca02fa80c4c064058e4d03574d89138fcd85672fd07e8e40`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/exitpolicy/ladder.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/exitpolicy/ladder.go:299` — if err := in.Policy.Validate(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B2 | `if` at `internal/exitpolicy/ladder.go:302` — if in.State.PolicyID != in.Policy.PolicyID { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B3 | `if` at `internal/exitpolicy/ladder.go:309` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B4 | `if` at `internal/exitpolicy/ladder.go:313` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B5 | `if` at `internal/exitpolicy/ladder.go:317` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B6 | `if` at `internal/exitpolicy/ladder.go:321` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B7 | `if` at `internal/exitpolicy/ladder.go:324` — if _, err := fraction("taken ratio total", in.State.TakenRatioTotal); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B8 | `if` at `internal/exitpolicy/ladder.go:327` — if in.State.ActivatedRung < NoRung \|\| in.State.ActivatedRung >= len(in.Policy.Rungs) { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B9 | `if` at `internal/exitpolicy/ladder.go:334` — if observed.Cmp(probe) > 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B10 | `range` at `internal/exitpolicy/ladder.go:353` — for i, rung := range in.Policy.Rungs { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B11 | `if` at `internal/exitpolicy/ladder.go:355` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B12 | `if` at `internal/exitpolicy/ladder.go:358` — if i > newIndex && returnPct.Cmp(target) >= 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B13 | `if` at `internal/exitpolicy/ladder.go:365` — if newIndex > NoRung { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B14 | `if` at `internal/exitpolicy/ladder.go:367` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B15 | `if` at `internal/exitpolicy/ladder.go:371` — if newIndex == len(in.Policy.Rungs)-1 && in.Policy.RunnerTrailPct != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B16 | `if` at `internal/exitpolicy/ladder.go:373` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B17 | `if` at `internal/exitpolicy/ladder.go:383` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B18 | `if` at `internal/exitpolicy/ladder.go:387` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B19 | `if` at `internal/exitpolicy/ladder.go:399` — if newIndex > in.State.ActivatedRung { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B20 | `if` at `internal/exitpolicy/ladder.go:406` — if in.State.Completed { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B21 | `if` at `internal/exitpolicy/ladder.go:415` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B22 | `if` at `internal/exitpolicy/ladder.go:418` — if observed.Cmp(baseline) < 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B23 | `if` at `internal/exitpolicy/ladder.go:420` — if in.State.PendingAction == ActionLadderStop { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B24 | `if` at `internal/exitpolicy/ladder.go:432` — if out.RungPromotedTo == NoRung { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B25 | `switch` at `internal/exitpolicy/ladder.go:436` — switch { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B26 | `case` at `internal/exitpolicy/ladder.go:437` — case newIndex == len(in.Policy.Rungs)-1 && in.Policy.FinalTakeFull: | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B27 | `case` at `internal/exitpolicy/ladder.go:440` — case isPositive(rung.PartialRatio): | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B28 | `case` at `internal/exitpolicy/ladder.go:445` — default: | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |
| B29 | `if` at `internal/exitpolicy/ladder.go:454` — if in.State.PendingAction != ActionNone { | only mutations visible in this branch and its callees | existing return/error contract | `TestCommonPolicyRegistryHasExactlyTheThreeApprovedProfiles`; ladder and decimal boundary tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `in.Policy.Validate`, `refusal`, `fmt.Sprintf`, `positive`, `fraction`, `len`, `observed.Cmp`, `percentOf`, `formatPrice`, `formatRMultiple`, `parseRat`, `err.Error` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 43 assignment(s) and 20 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
