# Function Logic Map: `ExitPolicy.validate`

- Source: `internal/config/engine.go`
- Qualified function: `ExitPolicy.validate`
- AST evidence: `ast.json` (`01f2158931852abd45c063f40ba7d9c6d9a346e28a1d8128daf4a6b3b8126a13`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/config/engine.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/config/engine.go:61` — if id == "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests |
| B2 | `if` at `internal/config/engine.go:64` — if _, ok := exitpolicy.CommonPolicyByID(id); !ok { | only mutations visible in this branch and its callees | existing return/error contract | `TestMissingCommonExitPolicyPreservesLegacyRatchetSelection`; `TestSaveCommonExitPolicyChangesOnlyItsValueBlock`; config engine tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`, `exitpolicy.CommonPolicyByID` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 2 assignment(s) and 3 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: no; current AST hash and affected-package tests are required.
